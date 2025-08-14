package main

import (
	"kjernekraft/database"
	"log"
)

func main() {
	dbConn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// Create test user membership (user ID 1 with membership ID 1)
	membershipSQL := `
	INSERT OR IGNORE INTO user_memberships (id, user_id, membership_id, status, start_date, renewal_date, binding_end) 
	VALUES (1, 1, 1, 'active', datetime('now', '-30 days'), datetime('now', '+11 months'), datetime('now', '+11 months'))
	`

	// Create test user klippekort (user ID 1 with some klippekort)
	klippekortSQL := []string{
		`INSERT OR IGNORE INTO user_klippekort (id, user_id, package_id, total_klipp, remaining_klipp, expiry_date) 
		VALUES (1, 1, 2, 10, 7, datetime('now', '+6 months'))`,
		
		`INSERT OR IGNORE INTO user_klippekort (id, user_id, package_id, total_klipp, remaining_klipp, expiry_date) 
		VALUES (2, 1, 5, 10, 2, datetime('now', '+3 months'))`,
	}

	// Execute membership insert
	if _, err := dbConn.Exec(membershipSQL); err != nil {
		log.Printf("Error inserting test membership: %v", err)
	} else {
		log.Println("Test membership created for user 1")
	}

	// Execute klippekort inserts
	for i, query := range klippekortSQL {
		if _, err := dbConn.Exec(query); err != nil {
			log.Printf("Error inserting test klippekort %d: %v", i+1, err)
		} else {
			log.Printf("Test klippekort %d created for user 1", i+1)
		}
	}

	log.Println("Test user data created successfully!")
}