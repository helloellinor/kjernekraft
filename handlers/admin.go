package handlers

import (
	"kjernekraft/database"
	"kjernekraft/handlers/modules"
	"kjernekraft/models"
	"log"
	"net/http"
)

var AdminDB *database.Database

type AdminData struct {
	Title          string
	Users          []models.User
	Events         []models.Event
	FreezeRequests []models.FreezeRequest
	Memberships    []models.Membership
	Stats          *modules.AdminStatsModuleData
	Lang           string
	CurrentPage    string
	CSS            string
	JS             string
	ExternalCSS    []string
	ExternalJS     []string
}

func AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	// For now, we'll skip authentication check
	// TODO: Add proper authentication to check if user has admin role

	users, err := AdminDB.GetAllUsers()
	if err != nil {
		http.Error(w, "Kunne ikke hente brukere", http.StatusInternalServerError)
		return
	}

	events, err := AdminDB.GetAllEvents()
	if err != nil {
		http.Error(w, "Kunne ikke hente events", http.StatusInternalServerError)
		return
	}

	freezeRequests, err := AdminDB.GetPendingFreezeRequests()
	if err != nil {
		http.Error(w, "Kunne ikke hente frysingsforespørsler", http.StatusInternalServerError)
		return
	}

	memberships, err := AdminDB.GetAllMemberships()
	if err != nil {
		http.Error(w, "Kunne ikke hente medlemskap", http.StatusInternalServerError)
		return
	}

	// Get language from request (default to Norwegian bokmål)
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "nb"
	}

	// Create admin stats module
	statsModule, err := modules.NewAdminStatsModule(len(users), len(events), len(freezeRequests), lang)
	if err != nil {
		log.Printf("Error creating admin stats module: %v", err)
		http.Error(w, "Error creating admin module", http.StatusInternalServerError)
		return
	}

	data := AdminData{
		Title:          "Admin",
		Users:          users,
		Events:         events,
		FreezeRequests: freezeRequests,
		Memberships:    memberships,
		Stats:          statsModule,
		Lang:           lang,
		CurrentPage:    "admin",
		CSS:            "",
		JS:             "",
		ExternalCSS:    []string{},
		ExternalJS:     []string{},
	}

	// Use template manager instead of inline template
	tm := GetTemplateManager()
	tmpl, exists := tm.GetTemplate("pages/admin")
	if !exists {
		// Try to reload templates in case they weren't loaded
		tm.ReloadTemplates()
		tmpl, exists = tm.GetTemplate("pages/admin")
		if !exists {
			log.Printf("Available templates: %v", tm.GetAvailableTemplates())
			http.Error(w, "Admin template not found", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("Error executing admin template: %v", err)
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

// Stub functions for API endpoints - these need to be implemented
func GetUsersAPIHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func ApproveFreezeRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func RejectFreezeRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
