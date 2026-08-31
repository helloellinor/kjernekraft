package main

import (
	"cmp"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"kjernekraft/database"
	"kjernekraft/handsamarar"
	"kjernekraft/handsamarar/config"
)

// standardPort is 18108 and not 8080.
//
// 8080 is the port most things take first, so "address already in use" is
// not an accident, it is the normal state. Worse is when something *else*
// is already listening: the page answers, but it is not this page, and you
// sit wondering why your changes do not show.
//
// 18108 is not registered with IANA, it is clear of what people take by
// default (3000, 5000, 8000, 8080, 8888, 9000), and it is below 32768 —
// where the kernel starts handing out ports for outgoing connections.
//
// The number is 108, the beads on a mala and the sun salutations in a full
// round. A port you remember is a port you do not guess at.
const standardPort = "18108"

func main() {
	// Initialize global settings (this will set up Oslo timezone by default)
	settings := config.GetInstance()
	log.Printf("Application started with timezone: %s", settings.GetTimezone())

	// Keep backward compatibility with OsloLoc
	handsamarar.OsloLoc = settings.GetLocation()

	// Økti fyrst: manglar nykelen, skal me ikkje starta i det heile.
	if err := handsamarar.InitializeSessionStore(); err != nil {
		log.Fatalf("økt: %v", err)
	}
	if handsamarar.IsDevelopment() {
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
	app := handsamarar.NyApp(db)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Alt som gjeng ut, pakka saman.
	//
	// Stilarket er den store: 301 kB gjekk yver netet ukomprimert til kvar
	// einaste sida. Det er ikkje eit lite tal for eit studio der folk opnar
	// timeplanen på telefonen i garderoben. Malane er tekst dei og, og tekst
	// er det gzip er best på.
	//
	// Han lyt liggja utanfor Recoverer, so feilsida hans vert pakka som alt
	// anna. Han rører ikkje det som alt er pakka (skriftene og bileti), og
	// nettlesarar som ikkje bed um det fær det ikkje.
	r.Use(middleware.Compress(5))

	// A panic in a handler should give 500, not a connection that simply
	// drops. Without this, net/http writes the stack to stderr and closes the
	// connection — the browser sees "connection was reset", which looks like a
	// network problem rather than a fault in the house.
	//
	// Ours rather than chi's: chi wrote 500 without a body, and only when the
	// panic came before the first byte — handlers stream templates straight
	// out, so it rarely did. See handlers/berging.go.
	r.Use(handsamarar.Recoverer)

	// In development nothing should be cached — not the pages, not the
	// fragments htmx fetches. Static files got no-store, but the HTML did not,
	// so you could sit looking at a page drawn while a template was broken
	// long after it was fixed. The page *looks* right; it is just old.
	if handsamarar.IsDevelopment() {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Cache-Control", "no-store, must-revalidate")
				next.ServeHTTP(w, req)
			})
		})
	}
	r.Use(app.WithUser)
	r.Use(handsamarar.CSRF)

	// Serve static files.
	//
	// I utvikling med `no-store`: lesaren bufrar elles CSS og JS, og då
	// ser ein sine eigne endringar fyrst etter ei hard oppdatering. Det
	// er ei felle ein gjeng i om att og om att, av di sida *ser* rett ut
	// — ho er berre gamal.
	//
	// Skriftene er unnatekne. Dei er det einaste under /static/ ein
	// aldri *endrar* medan ein arbeider — dei er binære filer som fylgjer
	// med repoet — og `no-store` på deim tyder at Union Gothic (95 kB)
	// vert lasta ned på nytt for kvar einaste sida ein opnar. Med
	// `font-display: swap` syner nettlesaren då reserveskrifti fyrst og
	// byter når fila er komi: eit blaff i typografien ved kvart klikk.
	//
	// Difor: alt anna `no-store` i utvikling, skriftene bufra hardt i
	// båe. Skiftar ei skriftfil namn, skiftar ho ogso adressa.
	// Stilarket kjem fyre filtenaren: han er sett saman av mange filer
	// under static/css/deler/, men hev framleis éi adresse. Sjå
	// handlers/stilark.go.
	r.Get("/static/css/kjernekraft.css", app.Stilark)

	statiske := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.Handle("/static/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/static/fonts/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if !handsamarar.IsDevelopment() && req.URL.Query().Get("v") != "" &&
			(strings.HasPrefix(req.URL.Path, "/static/css/") || strings.HasPrefix(req.URL.Path, "/static/js/")) {
			// Malane legg eit innhaldsavtrykk på CSS- og JS-adressene. Ein
			// ny utgåve får ny adresse, så denne kan trygt bufrast hardt.
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if strings.HasPrefix(req.URL.Path, "/static/img/") {
			// Bilete er ikkje skrifter — ein *kann* retta ein SVG medan
			// ein arbeider — men dei skal ikkje hentast på nytt for
			// kvar sida heller. Eit minutt: endringar syner seg medan du
			// framleis hugsar kva du gjorde, og ingen navigasjon kostar
			// ei ny henting.
			if handsamarar.IsDevelopment() {
				w.Header().Set("Cache-Control", "public, max-age=60")
			} else {
				w.Header().Set("Cache-Control", "public, max-age=604800")
			}
		} else if handsamarar.IsDevelopment() {
			w.Header().Set("Cache-Control", "no-store, must-revalidate")
		}
		statiske.ServeHTTP(w, req)
	}))

	// ---- opne rutor ----
	// Alt som lyt naaast fyre innloggingi, og ikkje eit einaste kall meir.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/elev/hjem", http.StatusTemporaryRedirect)
	})
	r.Get("/signup", app.SignUpPage)
	r.Post("/signup", app.SignUp)
	r.Get("/terms", app.Terms)
	r.Get("/glemt-passord", app.GløymtPassord)
	r.Get("/innlogging", app.Innlogging)
	r.Post("/innlogging", app.Innlogging)
	r.Post("/logout", app.Logout)

	// Innsjekkskjermen i vestibylen. Han stend utanfor innloggingi med
	// vilje — skulde folk logga seg på fyre dei kryssa av, hadde ingen
	// kryssa av. Personvernet ligg i tidsvindauget i staden: sida syner
	// berre timar som byrjar snart eller gjeng no. Sjå handlers/innsjekk.go.
	r.Get("/innsjekk", app.Innsjekk)
	r.Get("/innsjekk/laas", app.InnsjekkLås)
	r.Post("/innsjekk/laas", app.InnsjekkLås)
	r.Post("/api/innsjekk", app.InnsjekkKryss)
	// Drop-in over disken: søk upp namnet og gakk på timen, um det er
	// plass. Same vindauge, same vakthold.
	r.Get("/api/innsjekk/sok", app.InnsjekkSok)
	r.Post("/api/innsjekk/dropin", app.InnsjekkDropin)
	r.Post("/api/innsjekk/angre", app.InnsjekkAngre)

	// ---- innlogga rutor ----
	r.Group(func(r chi.Router) {
		r.Use(handsamarar.RequireAuth)

		r.Get("/users/payment-methods", app.GetUserPaymentMethods)

		// Event routes
		r.Get("/api/events", app.GetAllEvents)

		// Dashboard component routes (HTMX endpoints)
		r.Get("/api/user/klippekort", app.UserKlippekort)
		r.Get("/api/user/signups", app.UserSignups)
		r.Get("/api/user/helsing", app.Heimehovud)
		r.Get("/api/user/ledig-plass", app.LedigPlass)

		// Payment API routes
		r.Get("/api/payment-methods", app.PaymentMethods)
		r.Get("/api/charges", app.Charges)
		r.Post("/api/payment-methods/set-default", app.SetDefaultPaymentMethod)
		r.Post("/api/payment-methods/remove", app.RemovePaymentMethod)

		// Membership management API routes
		r.Post("/api/membership/freeze", app.FreezeMembership)
		r.Post("/api/membership/cancel-freeze", app.CancelFreezeRequest)
		r.Post("/api/membership/unfreeze", app.UnfreezeMembership)
		r.Post("/api/membership/add", app.AddMembership)
		r.Post("/api/membership/change", app.ChangeMembership)
		r.Get("/api/membership/can-change", app.CanChangeMembership)
		r.Post("/api/membership/remove", app.RemoveMembership)

		// Klippekort management API routes
		r.Post("/api/klippekort/purchase", app.PurchaseKlippekort)

		// Event signup API routes
		r.Post("/api/events/signup", app.EventSignup)
		r.Post("/api/events/cancel-signup", app.EventCancelSignup)

		// Gamle adressor, haldne ved lag
		r.Get("/klippekort", redirectTo("/elev/klippekort"))
		r.Get("/medlemskap", redirectTo("/elev/medlemskap"))

		// Elev dashboard routes
		r.Get("/elev", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/elev/hjem", http.StatusTemporaryRedirect)
		})
		r.Get("/elev/hjem", app.ElevDashboard)
		r.Get("/elev/timeplan", app.ElevTimeplan)
		r.Get("/elev/klippekort", app.KlippekortPage)
		r.Get("/elev/medlemskap", app.MembershipPage)
		r.Get("/elev/betaling", app.Betaling)
		r.Get("/elev/min-profil", app.MinProfil)
		r.Post("/elev/min-profil", app.MinProfil)
	})

	// ---- administrasjon ----
	// Løyvet vert lesi or basen for kvar soknad, ikkje or kaka.
	r.Group(func(r chi.Router) {
		r.Use(handsamarar.RequireAdmin)

		r.Get("/admin", app.AdminPage)
		r.Post("/api/admin/loyve", app.SettLøyve)

		r.Post("/api/events", app.CreateEvent)

		r.Get("/api/admin/membership-rules", app.GetMembershipRules)
		r.Post("/api/admin/membership-rules", app.SaveMembershipRules)
		r.Get("/api/admin/medlemskapsprisar", app.AdminPrisar)
		r.Post("/api/admin/prisar", app.AdminPrisar)
		r.Get("/api/admin/reglar", app.AdminReglar)
		r.Post("/api/admin/reglar", app.AdminReglar)
		r.Get("/api/admin/klippeprisar", app.AdminKlippeprisar)
		r.Post("/api/admin/klippeprisar", app.AdminKlippeprisar)
		r.Get("/api/admin/class/conflict", app.RoomConflict)
		r.Post("/api/admin/class", app.CreateClass)
		r.Post("/api/admin/serie/lagra", app.LagraSerie)
		r.Post("/api/admin/rabattkrav", app.Rabattkrav)
		r.Post("/api/admin/melding", app.HandsamaMelding)
		r.Post("/api/admin/gruppe", app.LagGruppe)
		r.Post("/api/admin/gruppe/slett", app.SlettGruppe)
		r.Post("/api/admin/gruppemedlem", app.Gruppemedlem)
		r.Put("/api/admin/class/*", app.UpdateClass)
		r.Delete("/api/admin/class/*", app.DeleteClass)
		r.Post("/api/admin/freeze-requests/approve", app.ApproveFreezeRequest)
		r.Post("/api/admin/freeze-requests/reject", app.RejectFreezeRequest)
		r.Route("/api/admin/settings", func(r chi.Router) {
			r.Get("/", app.AdminSettings)
			r.Post("/", app.AdminSettings)
		})
	})

	// ---- test data ----
	// These overwrite the database. They answer 404 without
	// KJERNEKRAFT_ENV=development, so they cannot be reached in production
	// even if someone calls them.
	r.Group(func(r chi.Router) {
		r.Use(handsamarar.RequireDevelopment)

		r.Post("/api/shuffle-test-data", app.ShuffleTestData)
		r.Post("/api/shuffle-memberships", app.ShuffleMemberships)
		r.Post("/api/shuffle-user-klippekort", app.ShuffleUserKlippekort)
		r.Post("/api/shuffle-all-test-data", app.ShuffleAllTestData)
		r.Post("/api/setup-test-data", app.SetupTestData)
		r.Get("/elev/testdata", app.TestDataPage)

		// Verkstaden. Han skriv ingen ting, men han syner heile
		// stilarket på ei sida, og det er ikkje noko ein brukar
		// treng koma til.
		r.Get("/arket", app.Arket)
	})

	// The port comes from the environment. It was ":8080" written in here
	// while ./køyr set PORT as though that did something — the script used the
	// number to look up who held the port while the server bound 8080 whatever
	// you said. The two agreed until the day somebody tried something else.
	port := cmp.Or(os.Getenv("PORT"), standardPort)
	// Who can reach the server. Empty means 127.0.0.1 — whoever wants it
	// reachable from outside has to say so, with KJERNEKRAFT_BIND=0.0.0.0.
	//
	// It was the other way round: empty meant all interfaces, and nothing in
	// the repo set the variable, so the protection existed only for those who
	// knew about it. A protection that has to be switched on is one that is
	// off.
	bind := cmp.Or(os.Getenv("KJERNEKRAFT_BIND"), "127.0.0.1")
	// One address, built once, for both the log and the listener. JoinHostPort
	// rather than concatenation, so "::1" becomes "[::1]" and not "too many
	// colons".
	addr := net.JoinHostPort(bind, port)
	log.Printf("Lyder paa http://%s", addr)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatal(err)
	}
}

func redirectTo(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	}
}
