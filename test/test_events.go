package main

import (
	"fmt"
	"kjernekraft/database"
	"kjernekraft/models"
	"log"
	"time"
)

func main() {
	dbConn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	db := &database.Database{Conn: dbConn}

	// Create some test events
	events := []models.Event{
		{
			Title:       "Kjernekraft Møte",
			Description: "Månedlig møte om kjernekraft",
			StartTime:   time.Now().Add(24 * time.Hour),
			EndTime:     time.Now().Add(26 * time.Hour),
			Location:    "Oslo",
			Organizer:   "Admin",
		},
		{
			Title:       "Atomkraft Workshop",
			Description: "Workshop om atomkraft teknologi",
			StartTime:   time.Now().Add(48 * time.Hour),
			EndTime:     time.Now().Add(52 * time.Hour),
			Location:    "Bergen",
			Organizer:   "Expert Team",
		},
		{
			Title:       "Sikkerhet Seminar",
			Description: "Seminar om kjernekraft sikkerhet",
			StartTime:   time.Now().Add(72 * time.Hour),
			EndTime:     time.Now().Add(76 * time.Hour),
			Location:    "Trondheim",
			Organizer:   "Safety Team",
		},
	}

	for _, event := range events {
		id, err := db.CreateEvent(event)
		if err != nil {
			log.Printf("Feil ved oppretting av event %s: %v", event.Title, err)
		} else {
			fmt.Printf("Oppretta event %s med ID %d\n", event.Title, id)
		}
	}

	// Show all events
	allEvents, err := db.GetAllEvents()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Alle events i databasen:")
	for _, e := range allEvents {
		fmt.Printf("ID: %d, Tittel: %s, Start: %s, Slutt: %s\n",
			e.ID, e.Title, e.StartTime.Format("2006-01-02 15:04"), e.EndTime.Format("2006-01-02 15:04"))
	}
}