package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"kjernekraft/database"
	"kjernekraft/handlers"
	"kjernekraft/handlers/config"
)

// standardPort er 18108 og ikkje 8080.
//
// 8080 er den porten flest ting tek fyrst. Draheim stend paa honom paa
// heim, og kvar andre tenar ein startar paa ei maskin gjer det same — so
// «address already in use» er ikkje eit uhell, det er normaltilstanden.
// Verre er det naar noko *anna* alt lyder der: daa svarar sida, men det
// er ikkje denne sida, og ein sit og undrar seg yver kvifor endringane
// ikkje syner seg.
//
// 18108 er ikkje skriven upp hjaa IANA, han ligg klaar av alt folk tek
// til vanleg (3000, 5000, 8000, 8080, 8888, 9000), og han ligg under
// 32768 — der byrjar dei portane kjernen deler ut til utgaaande
// sambandi sjølv, og bind ein seg der, tek ein av og til plassen fraa
// noko som alt hev fenge honom.
//
// Talet er 108, som er talet paa perlor i ei mala og paa solhelsingar i
// ei heil rekkja. Ein port ein hugsar er ein port ein ikkje gissar paa.
const standardPort = "18108"

// cmpOr gjev det fyrste som ikkje er tomt.
func cmpOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func main() {
	// Initialize global settings (this will set up Oslo timezone by default)
	settings := config.GetInstance()
	log.Printf("Application started with timezone: %s", settings.GetTimezone())

	// Keep backward compatibility with OsloLoc
	handlers.OsloLoc = settings.GetLocation()

	// Økti fyrst: manglar nykelen, skal me ikkje starta i det heile.
	if err := handlers.InitializeSessionStore(); err != nil {
		log.Fatalf("økt: %v", err)
	}
	if handlers.IsDevelopment() {
		log.Println("KJERNEKRAFT_ENV=development — testdata-rutor er opne og kakone gjeng utan Secure")
	}

	dbConn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// Kjør migrering
	if err := database.Migrate(dbConn); err != nil {
		log.Fatal(err)
	}

	db := &database.Database{Conn: dbConn}

	// Namn på produkt bur i si eiga tabell. Fyrste gongen flyttar ho
	// namna som alt står i basen inn som bokmålsnamn — dei er skrivne på
	// bokmål, so det er det dei er.
	if err := db.MigrerProduktnamn(); err != nil {
		log.Fatal(err)
	}
	handlers.DB = db
	handlers.AdminDB = db

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	// Ein panikk i ein handsamar skal gjeva 500 og ikkje eit samband som
	// berre dett. Utan denne skriv net/http stakken til stderr og stengjer
	// sambandet — nettlesaren ser «tilkoplinga vart broti», som ser ut som
	// eit nettverksproblem og ikkje som ein feil i huset. Tenaren stend
	// like fullt; det er kva *brukaren* ser som er skilnaden.
	r.Use(middleware.Recoverer)

	// I utvikling skal ingen ting bufrast — korkje sidone eller bitane
	// htmx hentar. Statiske filer fekk `no-store` fyrr, men sjølve
	// HTML-en gjorde det ikkje, so ein kunde sitja og sjaa paa ei sida
	// som var teikna medan ein mal var i stykke, lenge etter at han var
	// retta. Sida *ser* rett ut; ho er berre gamal.
	if handlers.IsDevelopment() {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Cache-Control", "no-store, must-revalidate")
				next.ServeHTTP(w, req)
			})
		})
	}
	r.Use(handlers.WithUser)
	r.Use(handlers.CSRF)

	// Serve static files.
	//
	// I utvikling med `no-store`: lesaren bufrar elles CSS og JS, og daa
	// ser ein sine eigne endringar fyrst etter ei hard oppdatering. Det
	// er ei felle ein gjeng i om att og om att, av di sida *ser* rett ut
	// — ho er berre gamal.
	//
	// Skriftene er unnatekne. Dei er det einaste under /static/ ein
	// aldri *endrar* medan ein arbeider — dei er binære filer som fylgjer
	// med repoet — og `no-store` paa deim tyder at Union Gothic (95 kB)
	// vert lasta ned paa nytt for kvar einaste sida ein opnar. Med
	// `font-display: swap` syner nettlesaren daa reserveskrifti fyrst og
	// byter naar fila er komi: eit blaff i typografien ved kvart klikk.
	//
	// Difor: alt anna `no-store` i utvikling, skriftene bufra hardt i
	// baae. Skiftar ei skriftfil namn, skiftar ho ogso adressa.
	// Stilarket kjem fyre filtenaren: han er sett saman av mange filer
	// under static/css/deler/, men hev framleis éi adresse. Sjaa
	// handlers/stilark.go.
	r.Get("/static/css/kjernekraft.css", handlers.StilarkHandler)

	statiske := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.Handle("/static/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/static/fonts/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if !handlers.IsDevelopment() && req.URL.Query().Get("v") != "" &&
			(strings.HasPrefix(req.URL.Path, "/static/css/") || strings.HasPrefix(req.URL.Path, "/static/js/")) {
			// Malane legg eit innhaldsavtrykk på CSS- og JS-adressene. Ein
			// ny utgåve får ny adresse, så denne kan trygt bufrast hardt.
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if strings.HasPrefix(req.URL.Path, "/static/img/") {
			// Bilete er ikkje skrifter — ein *kann* retta ein SVG medan
			// ein arbeider — men dei skal ikkje hentast paa nytt for
			// kvar sida heller. Eit minutt: endringar syner seg medan du
			// framleis hugsar kva du gjorde, og ingen navigasjon kostar
			// ei ny henting.
			if handlers.IsDevelopment() {
				w.Header().Set("Cache-Control", "public, max-age=60")
			} else {
				w.Header().Set("Cache-Control", "public, max-age=604800")
			}
		} else if handlers.IsDevelopment() {
			w.Header().Set("Cache-Control", "no-store, must-revalidate")
		}
		statiske.ServeHTTP(w, req)
	}))

	// ---- opne rutor ----
	// Alt som lyt naaast fyre innloggingi, og ikkje eit einaste kall meir.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/elev/hjem", http.StatusTemporaryRedirect)
	})
	r.Get("/signup", handlers.SignUpPageHandler)
	r.Post("/signup", handlers.SignUpHandler)
	r.Get("/terms", handlers.TermsHandler)
	r.Get("/glemt-passord", handlers.GloymtPassordHandler)
	r.Get("/innlogging", handlers.InnloggingHandler)
	r.Post("/innlogging", handlers.InnloggingHandler)
	r.Post("/logout", handlers.LogoutHandler)

	// Innsjekkskjermen i vestibylen. Han stend utanfor innloggingi med
	// vilje — skulde folk logga seg paa fyre dei kryssa av, hadde ingen
	// kryssa av. Personvernet ligg i tidsvindauget i staden: sida syner
	// berre timar som byrjar snart eller gjeng no. Sjaa handlers/innsjekk.go.
	r.Get("/innsjekk", handlers.InnsjekkHandler)
	r.Get("/innsjekk/laas", handlers.InnsjekkLaasHandler)
	r.Post("/innsjekk/laas", handlers.InnsjekkLaasHandler)
	r.Post("/api/innsjekk", handlers.InnsjekkMerkHandler)
	// Drop-in over disken: søk upp namnet og gakk paa timen, um det er
	// plass. Same vindauge, same vakthold.
	r.Get("/api/innsjekk/sok", handlers.InnsjekkSokHandler)
	r.Post("/api/innsjekk/dropin", handlers.InnsjekkDropinHandler)
	r.Post("/api/innsjekk/angre", handlers.InnsjekkAngreHandler)

	// ---- innlogga rutor ----
	r.Group(func(r chi.Router) {
		r.Use(handlers.RequireAuth)

		r.Get("/users/payment-methods", handlers.GetUserPaymentMethodsHandler)

		// Event routes
		r.Get("/api/events", handlers.GetAllEventsHandler)

		// Dashboard component routes (HTMX endpoints)
		r.Get("/api/user/klippekort", handlers.UserKlippekortHandler)
		r.Get("/api/user/signups", handlers.UserSignupsHandler)
		r.Get("/api/user/helsing", handlers.HeimehovudHandler)
		r.Get("/api/user/ledig-plass", handlers.LedigPlassHandler)

		// Payment API routes
		r.Get("/api/payment-methods", handlers.PaymentMethodsHandler)
		r.Get("/api/charges", handlers.ChargesHandler)
		r.Post("/api/payment-methods/set-default", handlers.SetDefaultPaymentMethodHandler)
		r.Post("/api/payment-methods/remove", handlers.RemovePaymentMethodHandler)

		// Membership management API routes
		r.Post("/api/membership/freeze", handlers.FreezeMembershipHandler)
		r.Post("/api/membership/cancel-freeze", handlers.CancelFreezeRequestHandler)
		r.Post("/api/membership/unfreeze", handlers.UnfreezeMembershipHandler)
		r.Post("/api/membership/add", handlers.AddMembershipHandler)
		r.Post("/api/membership/change", handlers.ChangeMembershipHandler)
		r.Get("/api/membership/can-change", handlers.CanChangeMembershipHandler)
		r.Post("/api/membership/remove", handlers.RemoveMembershipHandler)
		r.Post("/api/membership-recommendations", handlers.MembershipRecommendationsHandler)

		// Klippekort management API routes
		r.Post("/api/klippekort/purchase", handlers.PurchaseKlippekortHandler)

		// Event signup API routes
		r.Post("/api/events/signup", handlers.EventSignupHandler)
		r.Post("/api/events/cancel-signup", handlers.EventCancelSignupHandler)

		// Gamle adressor, haldne ved lag
		r.Get("/klippekort", redirectTo("/elev/klippekort"))
		r.Get("/medlemskap", redirectTo("/elev/medlemskap"))
		r.Post("/medlemskap/recommendations", handlers.MembershipRecommendationsHandler)

		// Elev dashboard routes
		r.Get("/elev", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/elev/hjem", http.StatusTemporaryRedirect)
		})
		r.Get("/elev/hjem", handlers.ElevDashboardHandler)
		r.Get("/elev/timeplan", handlers.ElevTimeplanHandler)
		r.Get("/elev/klippekort", handlers.KlippekortPageHandler)
		r.Get("/elev/medlemskap", handlers.MembershipPageHandler)
		r.Post("/elev/medlemskap/recommendations", handlers.MembershipRecommendationsHandler)
		r.Get("/elev/betaling", handlers.BetalingHandler)
		r.Get("/elev/min-profil", handlers.MinProfilHandler)
		r.Post("/elev/min-profil", handlers.MinProfilHandler)
	})

	// ---- administrasjon ----
	// Løyvet vert lesi or basen for kvar soknad, ikkje or kaka.
	r.Group(func(r chi.Router) {
		r.Use(handlers.RequireAdmin)

		r.Get("/admin", handlers.AdminPageHandler)
		r.Post("/api/admin/loyve", handlers.SettLoyveHandler)

		r.Post("/api/events", handlers.CreateEventHandler)

		r.Get("/api/admin/users", handlers.GetUsersAPIHandler)
		r.Get("/api/admin/membership-rules", handlers.GetMembershipRulesHandler)
		r.Post("/api/admin/membership-rules", handlers.SaveMembershipRulesHandler)
		r.Post("/api/admin/membership-price", handlers.UpdateMembershipPriceHandler)
		r.Get("/api/admin/medlemskapsprisar", handlers.AdminPriserHandler)
		r.Post("/api/admin/priser", handlers.AdminPriserHandler)
		r.Get("/api/admin/reglar", handlers.AdminReglarHandler)
		r.Post("/api/admin/reglar", handlers.AdminReglarHandler)
		r.Get("/api/admin/klippeprisar", handlers.AdminKlippeprisarHandler)
		r.Post("/api/admin/klippeprisar", handlers.AdminKlippeprisarHandler)
		r.Post("/api/admin/membership", handlers.CreateMembershipHandler)
		r.Delete("/api/admin/membership", handlers.DeleteMembershipHandler)
		r.Get("/api/admin/class/conflict", handlers.RoomConflictHandler)
		r.Post("/api/admin/class", handlers.CreateClassHandler)
		r.Post("/api/admin/serie/lagra", handlers.LagraSerieHandler)
		r.Post("/api/admin/rabattkrav", handlers.RabattkravHandler)
		r.Post("/api/admin/melding", handlers.MeldingHandsamaHandler)
		r.Post("/api/admin/gruppe", handlers.LagGruppeHandler)
		r.Post("/api/admin/gruppe/slett", handlers.SlettGruppeHandler)
		r.Post("/api/admin/gruppemedlem", handlers.GruppemedlemHandler)
		r.Put("/api/admin/class/*", handlers.UpdateClassHandler)
		r.Delete("/api/admin/class/*", handlers.DeleteClassHandler)
		r.Post("/api/admin/freeze-requests/approve", handlers.ApproveFreezeRequestHandler)
		r.Post("/api/admin/freeze-requests/reject", handlers.RejectFreezeRequestHandler)
		r.Route("/api/admin/settings", func(r chi.Router) {
			r.Get("/", handlers.AdminSettingsHandler)
			r.Post("/", handlers.AdminSettingsHandler)
		})
	})

	// ---- testdata ----
	// Desse skriva yver basen. Dei svara 404 utan KJERNEKRAFT_ENV=development,
	// so dei kunna ikkje naaast i drift jamvel um nokon kallar paa deim.
	r.Group(func(r chi.Router) {
		r.Use(handlers.RequireDevelopment)

		r.Post("/api/shuffle-test-data", handlers.ShuffleTestDataHandler)
		r.Post("/api/shuffle-memberships", handlers.ShuffleMembershipsHandler)
		r.Post("/api/shuffle-user-klippekort", handlers.ShuffleUserKlippekortHandler)
		r.Post("/api/shuffle-all-test-data", handlers.ShuffleAllTestDataHandler)
		r.Post("/api/setup-test-data", handlers.SetupTestDataHandler)
		r.Get("/elev/testdata", handlers.TestDataPageHandler)

		// Verkstaden. Han skriv ingen ting, men han syner heile
		// stilarket paa ei sida, og det er ikkje noko ein brukar
		// treng koma til.
		r.Get("/arket", handlers.ArketHandler)
	})

	// Porten kjem or umgjevnaden. Han stod som «:8080» skriven inn her,
	// og `./køyr` sette `PORT` som um det gjorde noko — skriptet nytta
	// talet til aa leita upp kven som heldt porten, medan tenaren batt
	// 8080 kva enn ein sa. Dei tvo var samde so lenge ingen prøvde noko
	// anna, og dei var det ikkje den dagen nokon gjorde det: paa heim
	// stend det alt noko paa 8080, og tenaren fall med «address already
	// in use» etter aa ha meldt at han lydde.
	port := os.Getenv("PORT")
	if port == "" {
		port = standardPort
	}
	// Kven som fær naa tenaren. Tom tyder alle grensesnitt, som fyrr.
	//
	// Stend han bak ein tunnel, skal han lyda paa 127.0.0.1 og ingen
	// annan stad: elles svarar han ogso beinveges paa porten sin, utan
	// TLS, og heile arbeidet med aa fœra folk gjenom https er umsonst
	// for den som skriv adressa sjølv. Paa heim var han naaeleg utanfraa
	// paa 18108 medan tunnelen stod ved sida av.
	bind := os.Getenv("KJERNEKRAFT_BIND")
	log.Printf("Serving on http://%s:%s", cmpOr(bind, "localhost"), port)
	err = http.ListenAndServe(bind+":"+port, r)
	if err != nil {
		log.Fatal(err)
	}
}

func redirectTo(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	}
}
