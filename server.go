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

	// Initialize session store
	handlers.InitializeSessionStore()

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

	// Serve static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Redirect to Elev dashboard for now (in the future this could be role-based)
		http.Redirect(w, r, "/elev/hjem", http.StatusTemporaryRedirect)
	})

	r.Get("/signup", handlers.SignUpPageHandler)
	r.Post("/signup", handlers.SignUpHandler)
	r.Get("/terms", handlers.TermsHandler)
	r.Get("/innlogging", handlers.InnloggingHandler)
	r.Post("/innlogging", handlers.InnloggingHandler)
	r.Get("/logout", handlers.LogoutHandler)

	r.Post("/users", handlers.AddUserHandler)

	r.Post("/users/assign-role", handlers.AssignRoleToUserHandler)
	r.Get("/users/roles", handlers.GetUserRolesHandler)

	r.Post("/users/add-payment-method", handlers.AddPaymentMethodHandler)
	r.Get("/users/payment-methods", handlers.GetUserPaymentMethodsHandler)

	// Admin routes
	r.Get("/admin", handlers.AdminPageHandler)
	r.Get("/api/admin/users", handlers.GetUsersAPIHandler)
	r.Post("/api/admin/events/update-time", handlers.UpdateEventTimeHandler)
	r.Post("/api/admin/freeze-requests/approve", handlers.ApproveFreezeRequestHandler)
	r.Post("/api/admin/freeze-requests/reject", handlers.RejectFreezeRequestHandler)
	r.Route("/api/admin/settings", func(r chi.Router) {
		r.Get("/", handlers.AdminSettingsHandler)
		r.Post("/", handlers.AdminSettingsHandler)
	})

	// Event routes
	r.Get("/api/events", handlers.GetAllEventsHandler)
	r.Post("/api/events", handlers.CreateEventHandler)

	// Test data routes (for development)
	r.Post("/api/shuffle-test-data", handlers.ShuffleTestDataHandler)
	r.Post("/api/shuffle-memberships", handlers.ShuffleMembershipsHandler)
	r.Post("/api/shuffle-user-klippekort", handlers.ShuffleUserKlippekortHandler)
	r.Post("/api/shuffle-all-test-data", handlers.ShuffleAllTestDataHandler)
	r.Post("/api/setup-test-data", handlers.SetupTestDataHandler)

	// Membership and klippekort routes (for compatibility, redirects to elev routes)
	r.Get("/klippekort", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/elev/klippekort", http.StatusMovedPermanently)
	})
	r.Get("/medlemskap", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/elev/medlemskap", http.StatusMovedPermanently)
	})
	r.Post("/medlemskap/recommendations", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/elev/medlemskap/recommendations", http.StatusMovedPermanently)
	})
	r.Post("/api/membership-recommendations", handlers.MembershipRecommendationsHandler)

	// Dashboard component routes (HTMX endpoints)
	r.Get("/api/user/klippekort", handlers.UserKlippekortHandler)
	r.Get("/api/user/membership", handlers.UserMembershipHandler)

	// Membership management API routes
	r.Post("/api/membership/freeze", handlers.FreezeMembershipHandler)
	r.Post("/api/membership/cancel-freeze", handlers.CancelFreezeRequestHandler)
	r.Post("/api/membership/unfreeze", handlers.UnfreezeMembershipHandler)
	r.Post("/api/membership/add", handlers.AddMembershipHandler)
	r.Post("/api/membership/change", handlers.ChangeMembershipHandler)
	r.Post("/api/membership/remove", handlers.RemoveMembershipHandler)

	// Elev dashboard routes
	r.Get("/elev", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/elev/hjem", http.StatusTemporaryRedirect)
	})
	r.Get("/elev/hjem", handlers.ElevDashboardHandler)
	r.Get("/elev/timeplan", handlers.ElevTimeplanHandler)
	r.Get("/elev/klippekort", handlers.KlippekortPageHandler)
	r.Get("/elev/medlemskap", handlers.MembershipSelectorHandler)
	r.Post("/elev/medlemskap/recommendations", handlers.MembershipRecommendationsHandler)
	r.Get("/elev/min-profil", handlers.MinProfilHandler)
	r.Post("/elev/min-profil", handlers.MinProfilHandler)
	r.Get("/elev/testdata", handlers.TestDataPageHandler)

	log.Println("Serving on http://localhost:8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
