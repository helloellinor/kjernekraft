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
	handlers.DB = db
	handlers.AdminDB = db

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(handlers.WithUser)
	r.Use(handlers.CSRF)

	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// ---- opne rutor ----
	// Alt som lyt naaast fyre innloggingi, og ikkje eit einaste kall meir.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/elev/hjem", http.StatusTemporaryRedirect)
	})
	r.Get("/signup", handlers.SignUpPageHandler)
	r.Post("/signup", handlers.SignUpHandler)
	r.Get("/terms", handlers.TermsHandler)
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
		r.Get("/elev/medlemskap", handlers.MembershipSelectorHandler)
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
		r.Post("/api/admin/membership", handlers.CreateMembershipHandler)
		r.Delete("/api/admin/membership", handlers.DeleteMembershipHandler)
		r.Post("/api/admin/class", handlers.CreateClassHandler)
		r.Put("/api/admin/class/*", handlers.UpdateClassHandler)
		r.Delete("/api/admin/class/*", handlers.DeleteClassHandler)
		r.Post("/api/admin/events/update-time", handlers.UpdateEventTimeHandler)
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
