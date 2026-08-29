package handlers

import (
	"kjernekraft/handlers/config"
	"kjernekraft/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ElevTimeplanHandler serves the Elev timeplan (schedule) page
func ElevTimeplanHandler(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

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
	targetMonday := VikeMaandag(now, weekOffset)

	// Get events for the target week
	weekEvents, err := DB.GetEventsForWeek(targetMonday, int64(user.ID))
	if err != nil {
		// Feilen vart slukt her. Ein 500 utan grunn i loggen er ein
		// feil ein lyt finna att med gissing.
		log.Printf("vika %s: %v", targetMonday.Format("2006-01-02"), err)
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
			// Filteret samanlikna med *tittelen*, so «yoga» aldri
			// treft noko: ingen time heiter «yoga», dei heiter «Hatha
			// Yoga». Han ser paa slaget no — og «reformer» er ikkje eit
			// slag, det er rommet, som er slik ein tenkjer um det.
			if classFilter != "" {
				if classFilter == "reformer" {
					if !strings.EqualFold(event.RoomName, "Reformer") {
						continue
					}
				} else if !strings.EqualFold(event.ClassType, classFilter) {
					continue
				}
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

	// Kategoriar, ikkje titlar. Lista kom fraa DISTINCT class_type og
	// gav tretten val — «boksing», «spinning», «crossfit» — eit studio
	// som ikkje finst. Ein filtrerer paa kva slag time det er, og
	// studioet hev tri slag: yoga, pilates og fascia. Reformer er ikkje
	// eit slag, det er eit rom, men det er slik ein tenkjer um det, so
	// han stend her og filtrerer paa rommet.
	classTypes := []string{"yoga", "pilates", "reformer", "fascia"}

	data := map[string]interface{}{
		"Title":           "Timeplan",
		"WeekTitle":       weekTitle,
		"WeekNumber":      targetWeek,
		"WeekOffset":      weekOffset,
		"VikorIAaret":     VikorIAaret(targetMonday),
		"WeekDays":        weekdays,
		"WeekDates":       weekDates,
		"EventsByDay":     eventsByDay,
		"ClassRows":       BuildWeekRows(lang, weekEvents, now, targetMonday),
		"Vekeval":         weekOptions(lang, targetWeek, weekOffset),
		"Today":           now.Format("2006-01-02"),
		"Teachers":        teachers,
		"ClassTypes":      classTypes,
		"SelectedTeacher": teacherFilter,
		"SelectedClass":   classFilter,
		"CanGoBack":       weekOffset > 0,
		"ExternalCSS":     []string{},
		"CurrentPage":     "timeplan",
		"UserName":        user.Name,
		"User":            user,
		"Lang":            lang,
		"CSRFToken":       CSRFToken(r),
		"IsAdmin":         sessionIsAdmin(r),
	}

	renderPage(w, r, "pages/timeplan", data)
}
