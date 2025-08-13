package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"kjernekraft/database"
	"kjernekraft/handlers"
)

func main() {
	dbConn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	db := &database.Database{Conn: dbConn}
	handlers.DB = db
	handlers.AdminDB = db

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	r.Post("/events", handlers.AddEventHandler)
	r.Post("/users", handlers.AddUserHandler)

	r.Post("/users/assign-role", handlers.AssignRoleToUserHandler)
	r.Get("/users/roles", handlers.GetUserRolesHandler)

	r.Post("/users/add-payment-method", handlers.AddPaymentMethodHandler)
	r.Get("/users/payment-methods", handlers.GetUserPaymentMethodsHandler)

	// Admin routes
	r.Get("/admin", handlers.AdminPageHandler)
	r.Get("/api/admin/users", handlers.GetUsersAPIHandler)

	log.Println("Serving on http://localhost:8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
