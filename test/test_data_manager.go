package main

import (
	"database/sql"
	"fmt"
	"kjernekraft/database"
	"kjernekraft/models"
	"log"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestDataManager provides utilities for managing test data
type TestDataManager struct {
	DB *database.Database
}

// NewTestDataManager creates a new test data manager
func NewTestDataManager() (*TestDataManager, error) {
	// Connect to database - use parent directory path when running from test folder
	dbConn, err := sql.Open("sqlite3", "../kjernekraft.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Run migration to ensure tables exist
	if err := database.Migrate(dbConn); err != nil {
		return nil, fmt.Errorf("failed to run migration: %v", err)
	}

	return &TestDataManager{
		DB: &database.Database{Conn: dbConn},
	}, nil
}

// Close closes the database connection
func (tdm *TestDataManager) Close() error {
	return tdm.DB.Conn.Close()
}

// ClearAllEvents removes all events from the database
func (tdm *TestDataManager) ClearAllEvents() error {
	_, err := tdm.DB.Conn.Exec("DELETE FROM events")
	if err != nil {
		return fmt.Errorf("failed to clear events: %v", err)
	}
	log.Println("✅ Cleared all events from database")
	return nil
}

// GenerateRandomizedSchedule creates a new randomized weekly schedule
func (tdm *TestDataManager) GenerateRandomizedSchedule() error {
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
	now := time.Now()
	monday := getStartOfWeek(now)

	log.Printf("📅 Generating events for week starting %s", monday.Format("2006-01-02"))

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
				Description:      fmt.Sprintf("En %s klasse med %s", classInfo.Type, teacher),
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
				Description:      fmt.Sprintf("En %s klasse med %s", classInfo.Type, teacher),
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
		_, err := tdm.DB.CreateEvent(event)
		if err != nil {
			log.Printf("❌ Error creating event %s: %v", event.Title, err)
		} else {
			successCount++
		}
	}

	log.Printf("✅ Successfully created %d new events out of %d total", successCount, len(events))
	return nil
}

// GetEventStats returns statistics about the current events in the database
func (tdm *TestDataManager) GetEventStats() error {
	// Total events
	var totalEvents int
	err := tdm.DB.Conn.QueryRow("SELECT COUNT(*) FROM events").Scan(&totalEvents)
	if err != nil {
		return err
	}

	// Today's events
	var todayEvents int
	err = tdm.DB.Conn.QueryRow("SELECT COUNT(*) FROM events WHERE DATE(start_time) = DATE('now', 'localtime')").Scan(&todayEvents)
	if err != nil {
		return err
	}

	// This week's events
	var weekEvents int
	err = tdm.DB.Conn.QueryRow(`
		SELECT COUNT(*) FROM events 
		WHERE DATE(start_time) >= DATE('now', 'weekday 0', '-6 days', 'localtime') 
		AND DATE(start_time) <= DATE('now', 'weekday 0', 'localtime')
	`).Scan(&weekEvents)
	if err != nil {
		return err
	}

	// Events by class type
	rows, err := tdm.DB.Conn.Query("SELECT class_type, COUNT(*) FROM events GROUP BY class_type ORDER BY COUNT(*) DESC")
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Printf("\n📊 EVENT STATISTICS\n")
	fmt.Printf("==================\n")
	fmt.Printf("📅 Total events: %d\n", totalEvents)
	fmt.Printf("🗓️  Today's events: %d\n", todayEvents)
	fmt.Printf("📆 This week's events: %d\n", weekEvents)
	fmt.Printf("\n🏃 Events by class type:\n")

	for rows.Next() {
		var classType string
		var count int
		if err := rows.Scan(&classType, &count); err != nil {
			return err
		}
		
		emoji := getClassEmoji(classType)
		fmt.Printf("   %s %-12s: %d events\n", emoji, classType, count)
	}

	fmt.Printf("\n")
	return nil
}

// ShuffleTestData combines clearing and regenerating data
func (tdm *TestDataManager) ShuffleTestData() error {
	log.Println("🔄 Starting test data shuffle...")
	
	if err := tdm.ClearAllEvents(); err != nil {
		return err
	}
	
	if err := tdm.GenerateRandomizedSchedule(); err != nil {
		return err
	}
	
	if err := tdm.GetEventStats(); err != nil {
		return err
	}
	
	fmt.Println("🎉 Test data shuffle complete! Refresh your browser to see the new schedule.")
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

func getClassEmoji(classType string) string {
	switch classType {
	case "yoga":
		return "🧘"
	case "pilates":
		return "🤸"
	case "strength":
		return "💪"
	case "cardio":
		return "❤️"
	case "flexibility":
		return "🤲"
	default:
		return "🏃"
	}
}

func main() {
	tdm, err := NewTestDataManager()
	if err != nil {
		log.Fatalf("Failed to create test data manager: %v", err)
	}
	defer tdm.Close()

	if err := tdm.ShuffleTestData(); err != nil {
		log.Fatalf("Failed to shuffle test data: %v", err)
	}
}