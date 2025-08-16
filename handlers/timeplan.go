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

	// Get user signups for these events
	if len(weekEvents) > 0 {
		eventIDs := make([]int64, len(weekEvents))
		for i, event := range weekEvents {
			eventIDs[i] = int64(event.ID)
		}
		
		userSignups, err := DB.GetUserSignupsForEvents(int64(user.ID), eventIDs)
		if err != nil {
			// Log error but don't fail the request
			// Just continue without signup information
		} else {
			// Update events with signup information
			for i := range weekEvents {
				weekEvents[i].IsUserSignedUp = userSignups[int64(weekEvents[i].ID)]
			}
		}
	}

	// Get language from cookies/request (using new system)
	lang := GetLanguageFromRequest(r)
	loc := GetLocalization()

	// Group events by day
	eventsByDay := make(map[string][]models.Event)
	weekdays := []string{
		loc.T(lang, "timeplan.monday"),
		loc.T(lang, "timeplan.tuesday"),
		loc.T(lang, "timeplan.wednesday"),
		loc.T(lang, "timeplan.thursday"),
		loc.T(lang, "timeplan.friday"),
		loc.T(lang, "timeplan.saturday"),
		loc.T(lang, "timeplan.sunday"),
	}
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
		weekTitle = loc.T(lang, "timeplan.this_week")
	} else if weekOffset == 1 {
		weekTitle = loc.T(lang, "timeplan.next_week")
	} else {
		weekTitle = loc.T(lang, "timeplan.week") + " " + strconv.Itoa(targetWeek)
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
