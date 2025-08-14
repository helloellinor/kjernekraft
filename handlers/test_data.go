package handlers

import (
	"encoding/json"
	"kjernekraft/models"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// ShuffleTestDataHandler provides an endpoint to shuffle test data
func ShuffleTestDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// Define class types with colors and descriptions
	classTypes := []ClassTypeInfo{
		{Type: "yoga", Color: "#8e44ad", Titles: []string{"Basic Yoga", "Forrest Yoga", "Yin Yoga", "Vinyasa Flow", "Hatha Yoga", "Power Yoga", "Restorative Yoga"}},
		{Type: "pilates", Color: "#27ae60", Titles: []string{"Klassisk Pilates", "Pilates Reformer", "Pilates Mat", "Power Pilates", "Gentle Pilates", "Core Pilates", "Pilates Flow"}},
		{Type: "strength", Color: "#e74c3c", Titles: []string{"Strength Training", "Circuit Training", "Functional Training", "HIIT Strength", "Body Sculpt", "Weight Training", "CrossFit"}},
		{Type: "cardio", Color: "#f39c12", Titles: []string{"Spin Class", "Zumba", "Aerobics", "Dance Cardio", "Boxing Cardio", "HIIT Cardio", "Step Aerobics"}},
		{Type: "flexibility", Color: "#3498db", Titles: []string{"Stretching", "Flexibility Focus", "Mobility Training", "Gentle Stretch", "Recovery Session", "Deep Stretch", "Myofascial Release"}},
	}

	// Norwegian teacher names
	teachers := []string{
		"Anna Larsen", "Erik Nordmann", "Kari Solberg", "Magnus Haugen", "Silje Andersen",
		"Lars Eriksen", "Ingrid Johansen", "Ole Hansen", "Nina Olsen", "Bjørn Kristiansen",
		"Hanne Nilsen", "Tor Pedersen", "Astrid Svensson", "Gunnar Andersen", "Lise Berg",
	}

	// Studio locations
	studios := []string{"Studio 1", "Studio 2", "Studio 3", "Hovedstudio", "Yogastudio", "Pilatesstudio"}

	// Time slots (hour of day)
	timeSlots := []int{6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}

	var events []models.Event

	// Generate events for the current week (Monday to Sunday)
	now := time.Now().In(handlers.OsloLoc)
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
	// Random remaining klipp between 0 and total_klipp
	userKlippekort, err := DB.GetUserKlippekort(1)
	if err != nil {
		return err
	}

	for _, klipp := range userKlippekort {
		newRemaining := rand.Intn(klipp.TotalKlipp + 1)
		_, err := DB.Conn.Exec(`UPDATE user_klippekort SET remaining_klipp = ? WHERE id = ?`, 
			newRemaining, klipp.UserKlippekort.ID)
		if err != nil {
			log.Printf("Error updating user klippekort %d: %v", klipp.UserKlippekort.ID, err)
		}
	}

	log.Println("Successfully shuffled user klippekort data")
	return nil
}