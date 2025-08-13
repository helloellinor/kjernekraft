package main

import (
	"fmt"
	"kjernekraft/database"
	"kjernekraft/models"
	"log"
)

func main() {
	dbConn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	db := &database.Database{Conn: dbConn}

	// Kjør migrering først
	if err := database.Migrate(dbConn); err != nil {
		log.Fatal(err)
	}

	// Legg til noen testbrukere
	users := []models.User{
		{
			Name:      "Ola Nordmann",
			Birthdate: "1990-01-01",
			Email:     "ola@example.com",
			Phone:     "12345678",
			Roles:     []string{"admin", "reformer"},
		},
		{
			Name:      "Kari Nordmann",
			Birthdate: "1992-02-02",
			Email:     "kari@example.com",
			Phone:     "87654321",
			Roles:     []string{"user", "reformer"},
		},
		{
			Name:      "Per Hansen",
			Birthdate: "1988-03-15",
			Email:     "per@example.com",
			Phone:     "55512345",
			Roles:     []string{"user", "plutocrat"},
		},
		{
			Name:      "Scrumbo Jones",
			Birthdate: "1948-03-15",
			Email:     "scrumbo@example.com",
			Phone:     "55512345",
			Roles:     []string{"user", "plutocrat"},
		},
	}

	for _, u := range users {
		id, err := db.CreateUser(u)
		if err != nil {
			log.Printf("Feil ved oppretting av bruker %s: %v", u.Name, err)
		} else {
			fmt.Printf("Oppretta bruker %s med ID %d\n", u.Name, id)
		}
	}

	// Hent og vis alle brukere
	allUsers, err := db.GetAllUsers()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Alle brukere i databasen:")
	for _, u := range allUsers {
		fmt.Printf("ID: %d, Navn: %s, Roller: %v\n", u.ID, u.Name, u.Roles)
	}
}
