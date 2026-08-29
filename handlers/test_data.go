package handlers

import (
	"encoding/json"
	"kjernekraft/handlers/config"
	"kjernekraft/models"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// ShuffleTestDataHandler provides an endpoint to shuffle test data
func ShuffleTestDataHandler(w http.ResponseWriter, r *http.Request) {
	err := shuffleTestData()
	if err != nil {
		log.Printf("Error shuffling test data: %v", err)
		http.Error(w, "Failed to shuffle test data", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Test data successfully shuffled!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// shuffleTestData clears existing events and generates new randomized test data
func shuffleTestData() error {
	// Clear existing events
	_, err := DB.Conn.Exec("DELETE FROM events")
	if err != nil {
		return err
	}
	log.Println("Cleared existing events from database")

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Kjernekraft Oslo held yoga, pilates og fascia — i salen eller paa
	// reformeren. Det stod tretten slag her, med boksing, spinning og
	// crossfit; eit studio som ikkje finst.
	classTypes := []ClassTypeInfo{
		{Type: "yoga", Color: "", Titles: []string{"Hatha Yoga", "Vinyasa Flow", "Yin Yoga", "Basic Yoga", "Yoga Styrke 45"}},
		{Type: "pilates", Color: "", Titles: []string{"Classical Pilates", "Klassisk Pilates Matte", "Pilates Apparatus", "Self Practice Pilates Apparatus"}},
		{Type: "fascia", Color: "", Titles: []string{"Fascia Movement", "Fascia Flyt"}},
	}

	// Dei som faktisk held timar i Storgata 23.
	teachers := []string{
		"Gry", "Ida", "Kristina", "Carla", "Cyrena",
		"Torunn", "Amanda", "Mariamah Aurora", "Leon", "Veronica",
	}

	// Studio locations
	studios := []string{"Salen", "Reformer"}

	// Time slots (hour of day)
	timeSlots := []int{6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}

	var events []models.Event

	// Generate events for the current week (Monday to Sunday)
	now := time.Now().In(OsloLoc)
	monday := getStartOfWeek(now)

	// Generate 5-12 events per day with randomization
	for dayOffset := 0; dayOffset < 7; dayOffset++ {
		currentDay := monday.AddDate(0, 0, dayOffset)

		// Randomize number of events per day (more events on weekdays)
		var numEvents int
		if currentDay.Weekday() == time.Saturday || currentDay.Weekday() == time.Sunday {
			numEvents = 3 + rand.Intn(5) // 3-7 events on weekends
		} else {
			numEvents = 6 + rand.Intn(7) // 6-12 events on weekdays
		}

		// Track used time slots to avoid conflicts
		usedTimeSlots := make(map[int]bool)

		for eventIndex := 0; eventIndex < numEvents; eventIndex++ {
			// Pick random class type
			classInfo := classTypes[rand.Intn(len(classTypes))]
			title := classInfo.Titles[rand.Intn(len(classInfo.Titles))]

			// Pick random teacher
			teacher := teachers[rand.Intn(len(teachers))]

			// Pick random time slot (avoid conflicts)
			var startHour int
			attempts := 0
			for {
				startHour = timeSlots[rand.Intn(len(timeSlots))]
				if !usedTimeSlots[startHour] || attempts > 10 {
					break
				}
				attempts++
			}
			usedTimeSlots[startHour] = true

			// Random class duration (45min, 60min, or 90min)
			durations := []int{45, 60, 90}
			duration := durations[rand.Intn(len(durations))]

			// Random minutes (0, 15, 30, 45)
			minutes := []int{0, 15, 30, 45}
			startMinute := minutes[rand.Intn(len(minutes))]

			startTime := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), startHour, startMinute, 0, 0, currentDay.Location())
			endTime := startTime.Add(time.Duration(duration) * time.Minute)

			// Random capacity (10-30 people)
			capacity := 10 + rand.Intn(21)

			// Random current enrollment (0 to capacity+5 to sometimes have waiting lists)
			maxEnrollment := capacity + 5
			currentEnrollment := rand.Intn(maxEnrollment + 1)

			// Random studio
			studio := studios[rand.Intn(len(studios))]

			event := models.Event{
				Title:            title,
				Description:      "En " + classInfo.Type + " klasse med " + teacher,
				StartTime:        startTime,
				EndTime:          endTime,
				Location:         studio,
				Organizer:        "Kjernekraft",
				ClassType:        classInfo.Type,
				TeacherName:      teacher,
				Capacity:         capacity,
				CurrentEnrolment: currentEnrollment,
				Color:            classInfo.Color,
			}

			events = append(events, event)
		}
	}

	// Also add some events for next week to show upcoming classes
	nextWeekMonday := monday.AddDate(0, 0, 7)
	for dayOffset := 0; dayOffset < 3; dayOffset++ { // Just first 3 days of next week
		currentDay := nextWeekMonday.AddDate(0, 0, dayOffset)
		numEvents := 2 + rand.Intn(4) // 2-5 events

		for eventIndex := 0; eventIndex < numEvents; eventIndex++ {
			classInfo := classTypes[rand.Intn(len(classTypes))]
			title := classInfo.Titles[rand.Intn(len(classInfo.Titles))]
			teacher := teachers[rand.Intn(len(teachers))]

			startHour := timeSlots[rand.Intn(len(timeSlots))]
			startMinute := []int{0, 30}[rand.Intn(2)]
			duration := []int{60, 90}[rand.Intn(2)]

			startTime := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), startHour, startMinute, 0, 0, currentDay.Location())
			endTime := startTime.Add(time.Duration(duration) * time.Minute)

			capacity := 10 + rand.Intn(21)
			currentEnrollment := rand.Intn(capacity + 3)
			studio := studios[rand.Intn(len(studios))]

			event := models.Event{
				Title:            title,
				Description:      "En " + classInfo.Type + " klasse med " + teacher,
				StartTime:        startTime,
				EndTime:          endTime,
				Location:         studio,
				Organizer:        "Kjernekraft",
				ClassType:        classInfo.Type,
				TeacherName:      teacher,
				Capacity:         capacity,
				CurrentEnrolment: currentEnrollment,
				Color:            classInfo.Color,
			}

			events = append(events, event)
		}
	}

	// Insert events into database
	successCount := 0
	for _, event := range events {
		_, err := DB.CreateEvent(event)
		if err != nil {
			log.Printf("Error creating event %s: %v", event.Title, err)
		} else {
			successCount++
		}
	}

	log.Printf("Successfully shuffled test data: Created %d new events", successCount)
	return nil
}

type ClassTypeInfo struct {
	Type   string
	Color  string
	Titles []string
}

// getStartOfWeek returns the Monday of the current week
func getStartOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	monday := t.AddDate(0, 0, -int(weekday)+1)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// ShuffleMembershipsHandler provides an endpoint to shuffle membership data
func ShuffleMembershipsHandler(w http.ResponseWriter, r *http.Request) {
	err := shuffleMembershipData()
	if err != nil {
		log.Printf("Error shuffling membership data: %v", err)
		http.Error(w, "Failed to shuffle membership data", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Membership data successfully shuffled!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ShuffleUserKlippekortHandler provides an endpoint to shuffle user klippekort
func ShuffleUserKlippekortHandler(w http.ResponseWriter, r *http.Request) {
	err := shuffleUserKlippekortData()
	if err != nil {
		log.Printf("Error shuffling user klippekort: %v", err)
		http.Error(w, "Failed to shuffle user klippekort", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "User klippekort successfully shuffled!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ShuffleAllTestDataHandler provides an endpoint to shuffle all test data
func ShuffleAllTestDataHandler(w http.ResponseWriter, r *http.Request) {
	// Shuffle events
	err := shuffleTestData()
	if err != nil {
		log.Printf("Error shuffling test data: %v", err)
		http.Error(w, "Failed to shuffle test data", http.StatusInternalServerError)
		return
	}

	// Shuffle memberships
	err = shuffleMembershipData()
	if err != nil {
		log.Printf("Error shuffling membership data: %v", err)
		// Continue even if this fails
	}

	// Shuffle user klippekort
	err = shuffleUserKlippekortData()
	if err != nil {
		log.Printf("Error shuffling user klippekort: %v", err)
		// Continue even if this fails
	}

	// Eit aar med fortid, so aktivitetsbrettet hev noko aa syna.
	err = seedFortid()
	if err != nil {
		log.Printf("Error seeding past activity: %v", err)
		// Continue even if this fails
	}

	response := map[string]interface{}{
		"success": true,
		"message": "All test data successfully shuffled!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// shuffleMembershipData shuffles membership names and prices
func shuffleMembershipData() error {
	membershipNames := []string{
		"Basis", "Standard", "Premium", "VIP", "Student", "Senior",
		"Familie", "Duo", "Unlimited", "Flex", "Morning", "Evening",
		"Weekend", "Prøve", "Høst Special", "Vinter Deal", "Sommer Pass",
	}

	descriptions := []string{
		"Perfekt for nybegynnere", "Vårt mest populære tilbud", "All-inclusive pakke",
		"Eksklusive fordeler", "Rabatt for studenter", "Senior rabatt",
		"For hele familien", "For to personer", "Ubegrenset tilgang",
		"Fleksibel løsning", "Morgen-trening", "Kveldstrening",
		"Kun helger", "Prøv oss først", "Spesialtilbud", "Vinter tilbud", "Sommer pakke",
	}

	// Update existing memberships with random names and prices
	query := `UPDATE memberships SET 
		name = ?, 
		description = ?, 
		price = ? 
		WHERE id = ?`

	for i := 1; i <= 10; i++ { // Assume we have 10 memberships
		name := membershipNames[rand.Intn(len(membershipNames))]
		description := descriptions[rand.Intn(len(descriptions))]
		// Random price between 300-1500 kr (30000-150000 øre)
		price := 30000 + rand.Intn(120000)

		_, err := DB.Conn.Exec(query, name, description, price, i)
		if err != nil {
			log.Printf("Error updating membership %d: %v", i, err)
		}
	}

	log.Println("Successfully shuffled membership data")
	return nil
}

// shuffleUserKlippekortData shuffles user's klippekort remaining amounts
func shuffleUserKlippekortData() error {
	// Beint paa tabellen, ikkje gjenom GetUserKlippekort. Den spurningi
	// syner berre kort ein kann bruka, og eit kort som stokkinga sette
	// til null hadde daa vorte usynleg for stokkinga sjølv — det kunde
	// aldri faa klipp att, og sat fast paa null for alltid.
	rows, err := DB.Conn.Query(
		`SELECT id, total_klipp FROM user_klippekort WHERE user_id = ?`, 1)
	if err != nil {
		return err
	}
	defer rows.Close()

	type kort struct{ id, total int }
	var kortet []kort
	for rows.Next() {
		var k kort
		if err := rows.Scan(&k.id, &k.total); err != nil {
			return err
		}
		kortet = append(kortet, k)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, k := range kortet {
		// Minst eitt klipp att. Stokkinga skal gjeva noko ein kann sjaa
		// paa, og eit kort paa null er eit kort som ikkje syner seg.
		newRemaining := rand.Intn(k.total) + 1
		if _, err := DB.Conn.Exec(`UPDATE user_klippekort SET remaining_klipp = ? WHERE id = ?`,
			newRemaining, k.id); err != nil {
			log.Printf("Error updating user klippekort %d: %v", k.id, err)
		}
	}

	log.Println("Successfully shuffled user klippekort data")
	return nil
}

// seedFortid lagar eit aar med *gjengne* timar og melder testbrukaren
// paa deim.
//
// Aktivitetsbrettet tel berre dagar som hev vore. Testdataa var alle
// framover i tid, so brettet stod tomt same kor mykje ein la inn — ein
// kunde ikkje sjaa om fargane verka, av di det ikkje var noko aa farga.
//
// Rytmen er med vilje ujamn: nokre vikor med tri timar, nokre med ein,
// og nokre heilt utan. Eit brett der kvar veke er lik syner ikkje om
// trinni verkar, og eit brett utan hol syner ikkje kva eit hol er.
func seedFortid() error {
	// Testbrukaren. Slaar honom upp direkte: det finst ingen
	// oppslagsfunksjon paa e-post, og seedaren treng berre id-en.
	var brukarID int64
	if err := DB.Conn.QueryRow(
		`SELECT id FROM users WHERE email = ?`, "anna@example.com").Scan(&brukarID); err != nil {
		return err
	}

	// Slaget skiftar gjenom aaret, so alle tri fargane kjem fram.
	slag := []struct {
		typ, tittel, laerar, rom string
	}{
		{"yoga", "Yin Yoga", "Gry", "Salen"},
		{"fascia", "Fascia Flyt", "Leon", "Salen"},
		{"pilates", "Classical Pilates", "Ida", "Salen"},
		{"reformer", "Reformer Flow", "Veronica", "Reformer"},
	}

	// Kor mange timar kvar av dei siste 52 vikone. Null er ei vika ein
	// ikkje var her: brettet skal ha hol i seg.
	rytme := []int{0, 1, 2, 3, 3, 0, 1, 3, 2, 0, 0, 1, 3, 3, 2, 1, 0, 2, 3, 1,
		0, 0, 3, 2, 1, 3, 0, 1, 2, 3, 3, 0, 2, 1, 0, 3, 1, 2, 3, 0,
		1, 3, 2, 0, 3, 1, 0, 2, 3, 1, 2, 3}

	naa := config.GetInstance().GetCurrentTime()
	maandag := VikeMaandag(naa, 0)

	tx, err := DB.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Reint bord fyrst, so gjentekne kall ikkje hopar seg upp.
	if _, err := tx.Exec(`DELETE FROM event_signups WHERE user_id = ?
		AND event_id IN (SELECT id FROM events WHERE title LIKE 'Fortid:%')`, brukarID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM events WHERE title LIKE 'Fortid:%'`); err != nil {
		return err
	}

	for v, tal := range rytme {
		if tal == 0 {
			continue
		}
		vikestart := maandag.AddDate(0, 0, -7*(len(rytme)-1-v))
		sl := slag[v%len(slag)]
		for i := 0; i < tal; i++ {
			// Spreidde utyver vika, og alltid fyre no.
			start := vikestart.AddDate(0, 0, i*2).Add(time.Duration(9+i) * time.Hour)
			if !start.Before(naa) {
				continue
			}
			res, err := tx.Exec(`
				INSERT INTO events (title, start_time, end_time, class_type, teacher_name,
				                    location, capacity, current_enrolment)
				VALUES (?, ?, ?, ?, ?, ?, 12, 1)`,
				"Fortid: "+sl.tittel,
				start.Format("2006-01-02 15:04:05"),
				start.Add(time.Hour).Format("2006-01-02 15:04:05"),
				sl.typ, sl.laerar, sl.rom)
			if err != nil {
				return err
			}
			id, err := res.LastInsertId()
			if err != nil {
				return err
			}
			if _, err := tx.Exec(
				`INSERT INTO event_signups (user_id, event_id, signup_date) VALUES (?, ?, ?)`,
				brukarID, id, start.Add(-24*time.Hour).Format("2006-01-02 15:04:05")); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}
