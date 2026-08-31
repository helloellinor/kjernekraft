package handsamarar

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"kjernekraft/database"
	"kjernekraft/handsamarar/config"
	"kjernekraft/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CreateClass creates a new class/event
func (a *App) CreateClass(w http.ResponseWriter, r *http.Request) {
	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	var classData struct {
		Title          string `json:"title"`
		ClassType      string `json:"class_type"`
		TeacherName    string `json:"teacher_name"`
		Location       string `json:"location"`
		RoomID         int64  `json:"room_id"`
		Weekday        int    `json:"weekday"` // 0 = sundag, som time.Weekday
		StartTime      string `json:"start_time"`
		EndTime        string `json:"end_time"`
		Capacity       int    `json:"capacity"`
		Color          string `json:"color"`
		Description    string `json:"description"`
		IsRecurring    bool   `json:"is_recurring"`
		RecurringWeeks int    `json:"recurring_weeks"`
		// Sett = ei privat økt for den eine. Sjå database/privattime.go.
		PrivateUserID int64 `json:"private_user_id"`
		// The one person's address, when the form asked for it instead of making
		// someone find a name in a list of everyone. It is looked up here: the
		// browser should not hold a register of who exists, and an id from the
		// browser is an id someone can type.
		PrivateEmail string `json:"private_email"`
		// Gruppa timen er open for. Null er open for alle.
		GruppeID int64 `json:"gruppe_id"`
		// Kor mange vikor fram fyrste timen skal liggja. Null er den
		// fyrste komande gongen vekedagen kjem. Lesaren reknar talet av
		// eit vekenummer (`veketal.js`); her er det berre vikor.
		StartWeekOffset int `json:"start_week_offset"`
	}

	if err := json.NewDecoder(r.Body).Decode(&classData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if classData.Weekday < 0 || classData.Weekday > 6 {
		http.Error(w, "Invalid weekday", http.StatusBadRequest)
		return
	}

	// Adressa fyrst, so resten av prøvingi ser éin id same kva vegen
	// han kom. Ho er den eine staden namnet vert til eit tal.
	if strings.TrimSpace(classData.PrivateEmail) != "" {
		var id int64
		err := a.DB.Conn.QueryRow(
			`SELECT id FROM users WHERE lower(email) = lower(?)`,
			strings.TrimSpace(classData.PrivateEmail)).Scan(&id)
		if err == sql.ErrNoRows {
			http.Error(w, t(GetLanguageFromRequest(r), "admin.unknown_email"), http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "Could not look up user", http.StatusInternalServerError)
			return
		}
		classData.PrivateUserID = id
	}

	// A private session is set aside for a *named* person. If they do not
	// exist, it is set aside for nobody and nobody can see it — a class
	// sitting in the database that never reaches a screen. Better to say so
	// here.
	if classData.PrivateUserID > 0 {
		var finst int
		if err := a.DB.Conn.QueryRow(
			`SELECT COUNT(*) FROM users WHERE id = ?`, classData.PrivateUserID).Scan(&finst); err != nil {
			http.Error(w, "Could not check user", http.StatusInternalServerError)
			return
		}
		if finst == 0 {
			http.Error(w, "Ukjend brukar for privat time", http.StatusBadRequest)
			return
		}
		// Personal training is one participant. Not "unless someone says
		// otherwise" — always: it is not a starting point, it is what a private
		// session *is*. The form disables the places field when you pick private,
		// but a POST is not a form, and a session for one with eighteen places is
		// not a session for one.
		classData.Capacity = 1
		// And it has to say what it costs. The clip is found by the class's
		// class_type meeting the package's category — see klippbruk.go — so a
		// private session without a type is a session that costs nothing without
		// anyone having said so. The form sends an empty type, because the type is
		// not something you write there.
		if strings.TrimSpace(classData.ClassType) == "" {
			classData.ClassType = database.PTKategori
		}
	}

	// The group has to exist. A class pointing at a group that is not there
	// is a class nobody can see, and nothing would have said so.
	if classData.GruppeID > 0 {
		finst, err := a.DB.GruppeFinst(classData.GruppeID)
		if err != nil || !finst {
			http.Error(w, "Ukjend gruppe", http.StatusBadRequest)
			return
		}
	}

	startTime, err := time.Parse("15:04", classData.StartTime)
	if err != nil {
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse("15:04", classData.EndTime)
	if err != nil {
		http.Error(w, "Invalid end time format", http.StatusBadRequest)
		return
	}

	// A series has a weekday, not a date. The first occurrence is the next
	// time that day comes round — today, if the hour has not already passed.
	// The times are built in UTC because the stored time is the wall clock
	// (see GrupperTimar); the configured clock says what it is now.
	no := config.GetInstance().GetCurrentTime()
	fram := (classData.Weekday - int(no.Weekday()) + 7) % 7
	if fram == 0 && startTime.Hour()*60+startTime.Minute() <= no.Hour()*60+no.Minute() {
		fram = 7
	}
	fyrste := no.AddDate(0, 0, fram)

	// "From week N". The browser has already turned the number into how many
	// weeks ahead it is, because a week that has passed does not exist — ask
	// for week 2 in week 51 and you mean the coming week 2. Here it is only
	// weeks to add.
	//
	// The cap is a year. A number above that is not a week further out, it is
	// a number somebody sent us.
	if classData.StartWeekOffset > 0 {
		if classData.StartWeekOffset > 53 {
			http.Error(w, "Invalid start week", http.StatusBadRequest)
			return
		}
		fyrste = fyrste.AddDate(0, 0, 7*classData.StartWeekOffset)
	}

	startDateTime := time.Date(fyrste.Year(), fyrste.Month(), fyrste.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	endDateTime := time.Date(fyrste.Year(), fyrste.Month(), fyrste.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

	// Create events based on recurring settings
	weeksToCreate := 1
	if classData.IsRecurring {
		weeksToCreate = classData.RecurringWeeks
	}

	// What is made here is a *series*, not a class — the classes are its
	// occurrences. A single class is a series with one.
	//
	// All the weeks are built and checked first, and written after. Checking
	// and writing used to interleave, so a collision in the fifth week left
	// the first four in the database while the answer was an error — and
	// whoever tried again got those four a second time.
	var timar []models.Event

	for week := 0; week < weeksToCreate; week++ {
		weekOffset := time.Duration(week) * 7 * 24 * time.Hour
		weekStartTime := startDateTime.Add(weekOffset)
		weekEndTime := endDateTime.Add(weekOffset)

		event := models.Event{
			Title:            classData.Title,
			Description:      classData.Description,
			StartTime:        weekStartTime,
			EndTime:          weekEndTime,
			Location:         classData.Location,
			Organizer:        "Kjernekraft",
			ClassType:        classData.ClassType,
			TeacherName:      classData.TeacherName,
			Capacity:         classData.Capacity,
			CurrentEnrolment: 0,
			Color:            classData.Color,
			RoomID:           int(classData.RoomID),
		}

		// The conflict was already shown while typing, but it has to be checked
		// here too: the page may have stood open while somebody else added
		// something, and a check that only exists in the browser is not a check.
		if classData.RoomID > 0 {
			if kollisjon, err := a.DB.RoomConflict(classData.RoomID, weekStartTime, weekEndTime); err == nil && kollisjon != nil {
				http.Error(w, fmt.Sprintf("%s %s–%s: %s",
					t(GetLanguageFromRequest(r), "admin.room_busy"),
					veggklokka(kollisjon.Start).Format("2.1. 15:04"),
					veggklokka(kollisjon.Slutt).Format("15:04"),
					kollisjon.Tittel), http.StatusConflict)
				return
			}
		}

		timar = append(timar, event)
	}

	_, createdEventIDs, err := a.DB.LagSerie(timar)
	if err != nil {
		http.Error(w, "Could not create event", http.StatusInternalServerError)
		return
	}

	// Who set the class up. Not the same as who holds it (teacher_name is
	// free text) — it is who the notice goes to when somebody gives up a
	// session.
	if uid, ok := sessionUserID(r); ok {
		for _, id := range createdEventIDs {
			if err := a.DB.SettLagaAv(id, uid); err != nil {
				log.Printf("laga-av for time %d: %v", id, err)
			}
		}
	}

	// Økta vert sett av etter at ho er laga. Alle utslagi av serien
	// høyrer den same personen til: eit PT-kjøp på aatte vekor er aatte
	// timar med same namnet på.
	if classData.PrivateUserID > 0 {
		for _, id := range createdEventIDs {
			if err := a.DB.SettPrivatTime(id, classData.PrivateUserID); err != nil {
				log.Printf("privat time %d: %v", id, err)
				http.Error(w, "Could not mark class private", http.StatusInternalServerError)
				return
			}
			// The session is booked the moment it is set up, and a clip goes with it —
			// if there is one. If there is not, the session stands anyway and no debt
			// is written: see BokPrivatTime.
			//
			// An error here does not undo the classes already made. They are in the
			// database, and answering with an error would invite a retry — and then
			// they are there twice. What may be missing is the clip, and that is
			// something the admins can see and fix.
			if err := a.DB.BokPrivatTime(id, classData.PrivateUserID, no); err != nil {
				log.Printf("klipp for privat time %d: %v", id, err)
			}
		}
	}

	// Gruppa gjeld heile serien av same grunnen: er reformer-timen open
	// for dei upplærde, er han det kvar veke og ikkje berre den fyrste.
	if classData.GruppeID > 0 {
		for _, id := range createdEventIDs {
			if err := a.DB.SettGruppePåTime(id, classData.GruppeID); err != nil {
				log.Printf("gruppetime %d: %v", id, err)
				http.Error(w, "Could not set group", http.StatusInternalServerError)
				return
			}
		}
	}

	response := map[string]interface{}{
		"success":        true,
		"message":        "Class(es) created successfully",
		"event_ids":      createdEventIDs,
		"events_created": len(createdEventIDs),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// vekedagssteg gives the number of days forward to the target day.
//
// Always forward, and always counted from the day the class itself falls
// on. Two reasons: an occurrence moved on its own should also land on the
// target, and a step backwards could put a coming class in the past, where
// nobody can sign up and nobody sees it. Monday to Sunday is six days
// forward, not one back.
func vekedagssteg(frå time.Weekday, mål int) int {
	return (mål - int(frå) + 7) % 7
}

// DeleteClass deletes a class/event
func (a *App) DeleteClass(w http.ResponseWriter, r *http.Request) {
	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	// Extract class ID from URL path
	// Expected format: /api/admin/class/{id}
	path := r.URL.Path
	classIDStr := path[len("/api/admin/class/"):]

	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	if err := a.DB.DeleteEvent(classID); err != nil {
		http.Error(w, "Could not delete class", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Class deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateClass updates a class/event
func (a *App) UpdateClass(w http.ResponseWriter, r *http.Request) {
	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	// Extract class ID from URL path
	path := r.URL.Path
	classIDStr := path[len("/api/admin/class/"):]

	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	var updateData struct {
		Title       string `json:"title"`
		ClassType   string `json:"class_type"`
		TeacherName string `json:"teacher_name"`
		Location    string `json:"location"`
		Date        string `json:"date"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		Capacity    int    `json:"capacity"`
		Color       string `json:"color"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Parse date and times (similar to create)
	classDate, err := time.Parse("2006-01-02", updateData.Date)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse("15:04", updateData.StartTime)
	if err != nil {
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse("15:04", updateData.EndTime)
	if err != nil {
		http.Error(w, "Invalid end time format", http.StatusBadRequest)
		return
	}

	startDateTime := time.Date(classDate.Year(), classDate.Month(), classDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, classDate.Location())
	endDateTime := time.Date(classDate.Year(), classDate.Month(), classDate.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, classDate.Location())

	event := models.Event{
		ID:          int(classID),
		Title:       updateData.Title,
		Description: updateData.Description,
		StartTime:   startDateTime,
		EndTime:     endDateTime,
		Location:    updateData.Location,
		ClassType:   updateData.ClassType,
		TeacherName: updateData.TeacherName,
		Capacity:    updateData.Capacity,
		Color:       updateData.Color,
	}

	if err := a.DB.UpdateEvent(event); err != nil {
		http.Error(w, "Could not update class", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Class updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
