package handlers

import (
	"kjernekraft/handlers/config"
	"kjernekraft/models"
	"net/http"
	"strconv"
	"time"
)

// ElevTimeplanHandler serves the Elev timeplan (schedule) page
func ElevTimeplanHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user := GetUserFromSession(r)
	if user == nil {
		http.Redirect(w, r, "/innlogging", http.StatusTemporaryRedirect)
		return
	}
	
	settings := config.GetInstance()
	now := settings.GetCurrentTime()

	// Parse week offset from query parameter
	weekOffset := 0
	if weekParam := r.URL.Query().Get("week"); weekParam != "" {
		if parsedWeek, err := strconv.Atoi(weekParam); err == nil {
			weekOffset = parsedWeek
		}
	}
	
	// Prevent navigating to past weeks
	if weekOffset < 0 {
		weekOffset = 0
	}

	// Get filter parameters
	teacherFilter := r.URL.Query().Get("teacher")
	classFilter := r.URL.Query().Get("class")

	// Calculate the target week's Monday
	monday := now.AddDate(0, 0, -int(now.Weekday())+1)
	if now.Weekday() == time.Sunday {
		monday = monday.AddDate(0, 0, -7)
	}
	targetMonday := monday.AddDate(0, 0, weekOffset*7)

	// Get events for the target week
	weekEvents, err := DB.GetEventsForWeek(targetMonday)
	if err != nil {
		http.Error(w, "Could not fetch week's events", http.StatusInternalServerError)
		return
	}

	// Apply filters
	if teacherFilter != "" || classFilter != "" {
		var filteredEvents []models.Event
		for _, event := range weekEvents {
			if teacherFilter != "" && event.TeacherName != teacherFilter {
				continue
			}
			if classFilter != "" && event.Title != classFilter {
				continue
			}
			// Only show events that users can sign up for (not full and in the future)
			if event.CurrentEnrolment >= event.Capacity && event.StartTime.Before(now) {
				continue
			}
			filteredEvents = append(filteredEvents, event)
		}
		weekEvents = filteredEvents
	}

	// Group events by day
	eventsByDay := make(map[string][]models.Event)
	weekdays := []string{"Mandag", "Tirsdag", "Onsdag", "Torsdag", "Fredag", "Lørdag", "Søndag"}
	weekDates := make([]time.Time, 7)

	for i := 0; i < 7; i++ {
		weekDates[i] = targetMonday.AddDate(0, 0, i)
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

	// Calculate week title
	var weekTitle string
	_, targetWeek := targetMonday.ISOWeek()
	
	if weekOffset == 0 {
		weekTitle = "Denne uka"
	} else if weekOffset == 1 {
		weekTitle = "Uka som kommer"
	} else {
		weekTitle = "Uke " + strconv.Itoa(targetWeek)
	}

	// Get distinct teachers and class types for filters
	teachers, err := DB.GetDistinctTeachers()
	if err != nil {
		teachers = []string{} // Continue with empty list if error
	}
	
	classTypes, err := DB.GetDistinctClassTypes()
	if err != nil {
		classTypes = []string{} // Continue with empty list if error
	}

	// Get language from cookies/request (using new system)
	lang := GetLanguageFromRequest(r)

	data := map[string]interface{}{
		"Title":        "Timeplan",
		"WeekTitle":    weekTitle,
		"WeekNumber":   targetWeek,
		"WeekOffset":   weekOffset,
		"WeekDays":     weekdays,
		"WeekDates":    weekDates,
		"EventsByDay":  eventsByDay,
		"Today":        now.Format("2006-01-02"),
		"Teachers":     teachers,
		"ClassTypes":   classTypes,
		"SelectedTeacher": teacherFilter,
		"SelectedClass":   classFilter,
		"CanGoBack":    weekOffset > 0,
		"IsAdmin":      false, // TODO: Implement proper role checking
		"ExternalCSS":  []string{"/static/css/event-card.css"},
		"CurrentPage":  "timeplan",
		"UserName":     user.Name,
		"User":         user,
		"Lang":         lang,
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
