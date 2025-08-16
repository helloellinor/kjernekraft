package handlers

import (
	"kjernekraft/handlers/config"
	"kjernekraft/models"
	"log"
	"net/http"
	"time"
)

var OsloLoc *time.Location

// ElevDashboardHandler serves the Elev dashboard home page
func ElevDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user := GetUserFromSession(r)
	if user == nil {
		http.Redirect(w, r, "/innlogging", http.StatusTemporaryRedirect)
		return
	}
	
	settings := config.GetInstance()
	now := settings.GetCurrentTime()

	// Get today's events
	allTodaysEvents, err := DB.GetTodaysEvents()
	if err != nil {
		http.Error(w, "Could not fetch today's events", http.StatusInternalServerError)
		return
	}

	// Filter out events that have already started
	var upcomingEvents []models.Event
	for _, event := range allTodaysEvents {
		if event.StartTime.After(now) {
			upcomingEvents = append(upcomingEvents, event)
		}
	}

	// Get language from cookies/request (using new system)
	lang := GetLanguageFromRequest(r)

	data := map[string]interface{}{
		"Title":        "Elev Dashboard",
		"TodaysEvents": upcomingEvents,
		"IsAdmin":      false, // TODO: Implement proper role checking
		"ExternalCSS":  []string{"/static/css/event-card.css"},
		"CurrentPage":  "hjem",
		"UserName":     user.Name,
		"User":         user,
		"Lang":         lang,
	}

	// Use the new template system
	tm := GetTemplateManager()
	if tmpl, exists := tm.GetTemplate("pages/dashboard"); exists {
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
		return
	}

	// If template doesn't exist, return error
	http.Error(w, "Template not found", http.StatusInternalServerError)
}
