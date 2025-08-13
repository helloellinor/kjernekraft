package main

import (
	"fmt"
	"kjernekraft/database"
	"kjernekraft/models"
	"log"
	"time"
)

func main() {
	// Connect to database
	dbConn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// Run migration
	if err := database.Migrate(dbConn); err != nil {
		log.Fatal(err)
	}

	db := &database.Database{Conn: dbConn}

	// Define class types with colors
	classTypes := map[string]string{
		"yoga":        "#8e44ad",
		"pilates":     "#27ae60",
		"strength":    "#e74c3c",
		"cardio":      "#f39c12",
		"flexibility": "#3498db",
	}

	// Sample teachers
	teachers := []string{
		"Anna Larsen",
		"Erik Nordmann",
		"Kari Solberg",
		"Magnus Haugen",
		"Silje Andersen",
	}

	// Sample class titles
	classTitles := map[string][]string{
		"yoga": {
			"Basic Yoga",
			"Forrest Yoga",
			"Yin Yoga",
			"Vinyasa Flow",
			"Hatha Yoga",
		},
		"pilates": {
			"Klassisk Pilates",
			"Pilates Reformer",
			"Pilates Mat",
			"Power Pilates",
			"Gentle Pilates",
		},
		"strength": {
			"Strength Training",
			"Circuit Training",
			"Functional Training",
			"HIIT Strength",
			"Body Sculpt",
		},
		"cardio": {
			"Spin Class",
			"Zumba",
			"Aerobics",
			"Dance Cardio",
			"Boxing Cardio",
		},
		"flexibility": {
			"Stretching",
			"Flexibility Focus",
			"Mobility Training",
			"Gentle Stretch",
			"Recovery Session",
		},
	}

	// Get current week's Monday
	now := time.Now()
	monday := now.AddDate(0, 0, -int(now.Weekday())+1)
	if now.Weekday() == time.Sunday {
		monday = monday.AddDate(0, 0, -7)
	}

	var events []models.Event

	// Create classes for this week (Week 33 according to problem statement)
	for dayOffset := 0; dayOffset < 7; dayOffset++ {
		currentDay := monday.AddDate(0, 0, dayOffset)
		
		// Skip creating events for past days, but create some for today and future
		if currentDay.Before(now.Truncate(24 * time.Hour)) {
			continue
		}

		// Create 3-5 classes per day
		numClasses := 3 + dayOffset%3 // 3-5 classes
		
		for classIndex := 0; classIndex < numClasses; classIndex++ {
			// Pick a random class type
			classTypeIndex := (dayOffset + classIndex) % len(classTypes)
			var classType string
			var color string
			typeIndex := 0
			for ct, c := range classTypes {
				if typeIndex == classTypeIndex {
					classType = ct
					color = c
					break
				}
				typeIndex++
			}

			// Pick a title for this class type
			titles := classTitles[classType]
			title := titles[classIndex%len(titles)]

			// Pick a teacher
			teacher := teachers[(dayOffset+classIndex)%len(teachers)]

			// Generate time slots throughout the day
			startHour := 7 + (classIndex * 3) // Start at 7:00, 10:00, 13:00, 16:00, 19:00
			if startHour > 19 {
				startHour = 19 // Cap at 19:00
			}
			
			startTime := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), startHour, 0, 0, 0, currentDay.Location())
			endTime := startTime.Add(time.Hour) // 1 hour classes

			// Random capacity and enrollment
			capacity := 15 + (classIndex * 5) // 15, 20, 25, 30, 35
			currentEnrollment := capacity - (dayOffset + classIndex + 1) // Some spaces left
			if currentEnrollment < 0 {
				currentEnrollment = capacity // Full class
			}

			event := models.Event{
				Title:            title,
				Description:      fmt.Sprintf("En %s klasse med %s", classType, teacher),
				StartTime:        startTime,
				EndTime:          endTime,
				Location:         "Studio " + fmt.Sprintf("%d", (classIndex%3)+1),
				Organizer:        "Kjernekraft",
				ClassType:        classType,
				TeacherName:      teacher,
				Capacity:         capacity,
				CurrentEnrolment: currentEnrollment,
				Color:            color,
			}

			events = append(events, event)
		}
	}

	// Insert events into database
	for _, event := range events {
		id, err := db.CreateEvent(event)
		if err != nil {
			log.Printf("Error creating event %s: %v", event.Title, err)
		} else {
			fmt.Printf("Created event %s with ID %d on %s\n", 
				event.Title, id, event.StartTime.Format("2006-01-02 15:04"))
		}
	}

	fmt.Printf("Created %d sample classes for this week\n", len(events))
}