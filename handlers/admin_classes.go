package handlers

import (
	"encoding/json"
	"fmt"
	"kjernekraft/handlers/config"
	"kjernekraft/models"
	"net/http"
	"strconv"
	"time"
)

// CreateClassHandler creates a new class/event
func CreateClassHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	}

	if err := json.NewDecoder(r.Body).Decode(&classData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if classData.Weekday < 0 || classData.Weekday > 6 {
		http.Error(w, "Invalid weekday", http.StatusBadRequest)
		return
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

	// Regelen hev ein vekedag, ikkje ein dato. Fyrste utslaget er neste
	// gongen den dagen kjem — i dag, um klokka ikkje alt er gjengi.
	// Tidene vert bygde i UTC av di den lagra tidi er veggklokka (sjaa
	// GrupperTimar); klokka i innstillingane segjer kva ho er no.
	no := config.GetInstance().GetCurrentTime()
	fram := (classData.Weekday - int(no.Weekday()) + 7) % 7
	if fram == 0 && startTime.Hour()*60+startTime.Minute() <= no.Hour()*60+no.Minute() {
		fram = 7
	}
	fyrste := no.AddDate(0, 0, fram)

	startDateTime := time.Date(fyrste.Year(), fyrste.Month(), fyrste.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	endDateTime := time.Date(fyrste.Year(), fyrste.Month(), fyrste.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

	// Create events based on recurring settings
	weeksToCreate := 1
	if classData.IsRecurring {
		weeksToCreate = classData.RecurringWeeks
	}

	// Det som vert laga her er ein *regel*, ikkje ein time — timane er
	// utslagi hans. Ein einskildtime er ein regel med eitt utslag.
	ruleID, err := AdminDB.NesteRegelID()
	if err != nil {
		http.Error(w, "Could not create rule", http.StatusInternalServerError)
		return
	}

	var createdEventIDs []int64

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
			RuleID:           ruleID,
		}

		// Konflikten er alt synt medan ein skreiv, men han lyt prøvast
		// her ogso: sida kann ha stade open medan nokon annan la inn
		// noko, og ei kontroll som berre finst i lesaren er ikkje ei
		// kontroll.
		if classData.RoomID > 0 {
			if kollisjon, err := AdminDB.RoomConflict(classData.RoomID, weekStartTime, weekEndTime); err == nil && kollisjon != nil {
				http.Error(w, fmt.Sprintf("%s %s–%s: %s",
					t(GetLanguageFromRequest(r), "admin.room_busy"),
					veggklokka(kollisjon.StartTime).Format("2.1. 15:04"),
					veggklokka(kollisjon.EndTime).Format("15:04"),
					kollisjon.Title), http.StatusConflict)
				return
			}
		}

		eventID, err := AdminDB.CreateEvent(event)
		if err != nil {
			http.Error(w, "Could not create event", http.StatusInternalServerError)
			return
		}
		createdEventIDs = append(createdEventIDs, eventID)
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

// UpdateRuleTeacherHandler byter lærar paa regelen — alle komande
// timar. Det som alt er halde stend som det var.
func UpdateRuleTeacherHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	ruleID, err := strconv.ParseInt(r.URL.Query().Get("rule_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid rule_id", http.StatusBadRequest)
		return
	}
	laerar := r.URL.Query().Get("teacher_name")
	if laerar == "" {
		http.Error(w, "teacher_name is required", http.StatusBadRequest)
		return
	}

	if err := AdminDB.UpdateRuleTeacher(ruleID, laerar, config.GetInstance().GetCurrentTime()); err != nil {
		http.Error(w, "Could not update teacher", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateClassVikarHandler set vikar paa éin einskild time. Regelen
// stend urørd — Leon er hjaa tannlækjaren éin dag, ikkje slutta.
func UpdateClassVikarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	eventID, err := strconv.ParseInt(r.URL.Query().Get("event_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid event_id", http.StatusBadRequest)
		return
	}
	laerar := r.URL.Query().Get("teacher_name")
	if laerar == "" {
		http.Error(w, "teacher_name is required", http.StatusBadRequest)
		return
	}

	if err := AdminDB.UpdateEventTeacher(eventID, laerar); err != nil {
		http.Error(w, "Could not update teacher", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateRuleLengthHandler set ny lengd paa regelen: kvar komande time
// held starten sin og fær ny slutt.
func UpdateRuleLengthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	ruleID, err := strconv.ParseInt(r.URL.Query().Get("rule_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid rule_id", http.StatusBadRequest)
		return
	}
	minutt, err := strconv.Atoi(r.URL.Query().Get("minutt"))
	if err != nil || minutt < 15 || minutt > 240 {
		http.Error(w, "Invalid minutt (expected 15-240)", http.StatusBadRequest)
		return
	}

	timar, err := AdminDB.GetFutureEventsByRule(ruleID, config.GetInstance().GetCurrentTime())
	if err != nil {
		http.Error(w, "Could not fetch classes", http.StatusInternalServerError)
		return
	}
	for _, e := range timar {
		if err := AdminDB.FlyttEvent(int64(e.ID), e.StartTime,
			e.StartTime.Add(time.Duration(minutt)*time.Minute)); err != nil {
			http.Error(w, "Could not update class length", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateRuleDescriptionHandler set skildringi paa regelen — alle
// komande timar. Ho gjeng i kroppen og ikkje i adressa: ei skildring
// er lang og hev linebrot.
func UpdateRuleDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	var data struct {
		RuleID      int64  `json:"rule_id"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if data.RuleID == 0 {
		http.Error(w, "rule_id is required", http.StatusBadRequest)
		return
	}

	if err := AdminDB.UpdateRuleDescription(data.RuleID, data.Description, config.GetInstance().GetCurrentTime()); err != nil {
		http.Error(w, "Could not update description", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateRuleClockHandler flytter regelen til eit nytt klokkeslett:
// kvar komande time held dagen sin og lengdi si, berre klokka skifter.
func UpdateRuleClockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	ruleID, err := strconv.ParseInt(r.URL.Query().Get("rule_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid rule_id", http.StatusBadRequest)
		return
	}
	klokke, err := time.Parse("15:04", r.URL.Query().Get("klokke"))
	if err != nil {
		http.Error(w, "Invalid klokke (expected HH:MM)", http.StatusBadRequest)
		return
	}

	timar, err := AdminDB.GetFutureEventsByRule(ruleID, config.GetInstance().GetCurrentTime())
	if err != nil {
		http.Error(w, "Could not fetch classes", http.StatusInternalServerError)
		return
	}

	for _, e := range timar {
		// Klokka vert lesi og skrivi som ho stend — den lagra tidi er
		// veggklokka (sjaa GrupperTimar), so den nye tidi vert bygd i
		// same sona som timen alt hev.
		lengd := e.EndTime.Sub(e.StartTime)
		dag := e.StartTime
		start := time.Date(dag.Year(), dag.Month(), dag.Day(),
			klokke.Hour(), klokke.Minute(), 0, 0, dag.Location())
		if err := AdminDB.FlyttEvent(int64(e.ID), start, start.Add(lengd)); err != nil {
			http.Error(w, "Could not move class", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteClassHandler deletes a class/event
func DeleteClassHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	if err := AdminDB.DeleteEvent(classID); err != nil {
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

// UpdateClassHandler updates a class/event
func UpdateClassHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	if err := AdminDB.UpdateEvent(event); err != nil {
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
