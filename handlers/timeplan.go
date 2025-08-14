package handlers

import (
	"kjernekraft/handlers/config"
	"kjernekraft/models"
	"net/http"
	"time"
)

// ElevTimeplanHandler serves the Elev timeplan (schedule) page
func ElevTimeplanHandler(w http.ResponseWriter, r *http.Request) {
	// Get this week's events
	weekEvents, err := DB.GetThisWeeksEvents()
	if err != nil {
		http.Error(w, "Could not fetch week's events", http.StatusInternalServerError)
		return
	}

	// Group events by day
	eventsByDay := make(map[string][]models.Event)
	settings := config.GetInstance()
	now := settings.GetCurrentTime()

	// Calculate this week's dates (Monday to Sunday)
	weekdays := []string{"Mandag", "Tirsdag", "Onsdag", "Torsdag", "Fredag", "Lørdag", "Søndag"}
	weekDates := make([]time.Time, 7)

	// Find Monday of this week
	monday := now.AddDate(0, 0, -int(now.Weekday())+1)
	if now.Weekday() == time.Sunday {
		monday = monday.AddDate(0, 0, -7)
	}

	for i := 0; i < 7; i++ {
		weekDates[i] = monday.AddDate(0, 0, i)
		dateKey := weekDates[i].Format("2006-01-02")
		eventsByDay[dateKey] = []models.Event{}
	}

	// Group events by date
	for _, event := range weekEvents {
		dateKey := event.StartTime.Format("2006-01-02")
		if _, exists := eventsByDay[dateKey]; exists {
			eventsByDay[dateKey] = append(eventsByDay[dateKey], event)
		}
	}

	data := map[string]interface{}{
		"Title":        "Timeplan",
		"WeekDays":     weekdays,
		"WeekDates":    weekDates,
		"EventsByDay":  eventsByDay,
		"Today":        now.Format("2006-01-02"),
		"IsAdmin":      false, // TODO: Implement proper role checking
		"ExternalCSS":  []string{"/static/css/event-card.css"},
		"CurrentPage":  "timeplan",
	}

	// Use the new template system
	tm := GetTemplateManager()
	if tmpl, exists := tm.GetTemplate("pages/timeplan"); exists {
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
		return
	}

	// If template doesn't exist, return error
	http.Error(w, "Template not found", http.StatusInternalServerError)
}
