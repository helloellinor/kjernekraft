package handlers

import (
	"encoding/json"
	"fmt"
	"kjernekraft/handlers/config"
	"kjernekraft/models"
	"log"
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
		// Sett = ei privat økt for den eine. Sjaa database/privattime.go.
		PrivateUserID int64 `json:"private_user_id"`
		// Gruppa timen er open for. Null er open for alle.
		GruppeID int64 `json:"gruppe_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&classData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if classData.Weekday < 0 || classData.Weekday > 6 {
		http.Error(w, "Invalid weekday", http.StatusBadRequest)
		return
	}

	// Ei privat økt er sett av til ein *namngjeven* person. Finst han
	// ikkje, er økta sett av til ingen, og ingen kann sjaa henne — ho
	// hadde vorte ein time som ligg i basen og aldri kjem paa nokon
	// skjerm. Betre aa segja fraa her.
	if classData.PrivateUserID > 0 {
		var finst int
		if err := AdminDB.Conn.QueryRow(
			`SELECT COUNT(*) FROM users WHERE id = ?`, classData.PrivateUserID).Scan(&finst); err != nil {
			http.Error(w, "Could not check user", http.StatusInternalServerError)
			return
		}
		if finst == 0 {
			http.Error(w, "Ukjend brukar for privat time", http.StatusBadRequest)
			return
		}
		// Personlig Trening er éin deltakar. Set ingen noko anna, er
		// det talet me meiner.
		if classData.Capacity == 0 {
			classData.Capacity = 1
		}
	}

	// Gruppa lyt finnast. Ein time som peikar paa ei gruppe som ikkje er
	// der, er ein time ingen kann sjaa — og ingen ting hadde sagt fraa.
	if classData.GruppeID > 0 {
		finst, err := AdminDB.GruppeFinst(classData.GruppeID)
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

	// Serien hev ein vekedag, ikkje ein dato. Fyrste utslaget er neste
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

	// Det som vert laga her er ein *serie*, ikkje ein time — timane er
	// utslagi hans. Ein einskildtime er ein serie med eitt utslag.
	//
	// Alle vekene vert bygde og prøvde fyrst, og skrivne etterpå. Fyrr
	// gjekk prøva og skrivinga om kvarandre, so ein kollisjon i femte
	// veka lét dei fire fyrste stå att i basen medan svaret var ein
	// feil — og den som prøvde ein gong til fekk dei fire ein gong til.
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

		timar = append(timar, event)
	}

	_, createdEventIDs, err := AdminDB.LagSerie(timar)
	if err != nil {
		http.Error(w, "Could not create event", http.StatusInternalServerError)
		return
	}

	// Økta vert sett av etter at ho er laga. Alle utslagi av serien
	// høyrer den same personen til: eit PT-kjøp paa aatte vekor er aatte
	// timar med same namnet paa.
	if classData.PrivateUserID > 0 {
		for _, id := range createdEventIDs {
			if err := AdminDB.SettPrivatTime(id, classData.PrivateUserID); err != nil {
				log.Printf("privat time %d: %v", id, err)
				http.Error(w, "Could not mark class private", http.StatusInternalServerError)
				return
			}
		}
	}

	// Gruppa gjeld heile serien av same grunnen: er reformer-timen open
	// for dei upplærde, er han det kvar veke og ikkje berre den fyrste.
	if classData.GruppeID > 0 {
		for _, id := range createdEventIDs {
			if err := AdminDB.SettGruppePaaTime(id, classData.GruppeID); err != nil {
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

// vekedagssteg gjev talet paa dagar fram til maaldagen.
//
// Alltid framover, og alltid rekna av dagen timen sjølv stend paa. Tvo
// grunnar: eit utslag som er flutt for seg sjølv skal ogso hamna paa
// maalet, og eit steg bakover kunde ha lagt ein komande time i fortidi
// — der ingen kann melda seg paa honom og ingen ser honom meir.
// Maandag til sundag er difor seks dagar fram, ikkje éin attende.
func vekedagssteg(fraa time.Weekday, maal int) int {
	return (maal - int(fraa) + 7) % 7
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
