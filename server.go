package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"kjernekraft/database"
	"kjernekraft/handlers"
	"kjernekraft/handlers/config"
)

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
	statiske := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.Handle("/static/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if handlers.IsDevelopment() {
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

	// ---- innlogga rutor ----
	r.Group(func(r chi.Router) {
		r.Use(handlers.RequireAuth)

		r.Get("/users/roles", handlers.GetUserRolesHandler)
		r.Get("/users/payment-methods", handlers.GetUserPaymentMethodsHandler)

		// Event routes
		r.Get("/api/events", handlers.GetAllEventsHandler)

		// Dashboard component routes (HTMX endpoints)
		r.Get("/api/user/klippekort", handlers.UserKlippekortHandler)
		r.Get("/api/user/membership", handlers.UserMembershipHandler)
		r.Get("/api/user/signups", handlers.UserSignupsHandler)

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
		r.Get("/elev/medlemskap", handlers.MedlemskapetHandler)
		r.Post("/elev/medlemskap/recommendations", handlers.MembershipRecommendationsHandler)
		r.Get("/elev/betaling", handlers.BetalingHandler)
		r.Get("/elev/min-profil", handlers.MinProfilHandler)
		r.Post("/elev/min-profil", handlers.MinProfilHandler)
	})

	// ---- administrasjon ----
	// Rolla vert lesi or basen for kvar soknad, ikkje or kaka.
	r.Group(func(r chi.Router) {
		r.Use(handlers.RequireAdmin)

		r.Get("/admin", handlers.AdminPageHandler)
		r.Post("/users/assign-role", handlers.AssignRoleToUserHandler)

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
		r.Post("/api/admin/rule/teacher", handlers.UpdateRuleTeacherHandler)
		r.Post("/api/admin/rule/klokke", handlers.UpdateRuleClockHandler)
		r.Post("/api/admin/rule/beskriving", handlers.UpdateRuleDescriptionHandler)
		r.Post("/api/admin/rule/lengd", handlers.UpdateRuleLengthHandler)
		r.Post("/api/admin/class/vikar", handlers.UpdateClassVikarHandler)
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

	log.Println("Serving on http://localhost:8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}

func redirectTo(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	}
}
