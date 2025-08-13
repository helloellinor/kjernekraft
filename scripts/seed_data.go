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

	// Run migration to ensure new tables exist
	if err := database.Migrate(dbConn); err != nil {
		log.Fatal(err)
	}

	// Seed membership data
	memberships := []string{
		`INSERT OR IGNORE INTO memberships (id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active) VALUES 
		(1, '12-måneder', 104000, 12, false, false, 'Vår mest populære medlemskap med 12 måneders binding', '["Ubegrenset gruppeklasser", "Rabatt på workshops", "Kan fryses", "Tilgang til alle lokasjoner"]', true)`,
		
		`INSERT OR IGNORE INTO memberships (id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active) VALUES 
		(2, '6-måneder', 115000, 6, false, false, '6 måneders binding med fleksibilitet', '["Ubegrenset gruppeklasser", "Rabatt på workshops", "Kan fryses", "Tilgang til alle lokasjoner"]', true)`,
		
		`INSERT OR IGNORE INTO memberships (id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active) VALUES 
		(3, 'Ingen binding', 125000, 0, false, false, 'Full fleksibilitet uten binding', '["Ubegrenset gruppeklasser", "Rabatt på workshops", "Kan fryses", "1 måned oppsigelse"]', true)`,
		
		`INSERT OR IGNORE INTO memberships (id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active) VALUES 
		(4, 'Student/Senior 12-måneder', 83000, 12, true, false, 'Studentrabatt på 12-måneder medlemskap', '["20% studentrabatt", "Ubegrenset gruppeklasser", "Rabatt på workshops", "Kan fryses"]', true)`,
		
		`INSERT OR IGNORE INTO memberships (id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active) VALUES 
		(5, 'Student/Senior ingen binding', 104000, 0, true, false, 'Studentrabatt uten binding', '["20% studentrabatt", "Ubegrenset gruppeklasser", "Rabatt på workshops", "Kan fryses"]', true)`,
		
		`INSERT OR IGNORE INTO memberships (id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active) VALUES 
		(6, 'Høsttilbud', 104000, 4, false, true, 'Spesialtilbud for høsten - 12-måneders pris med kun 4 måneders binding', '["12-måneders pris", "Kun 4 måneders binding", "Online videobibliotek", "Gratis mattegjenlegging", "Ubegrenset gruppeklasser"]', true)`,
		
		`INSERT OR IGNORE INTO memberships (id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active) VALUES 
		(7, '2-ukers prøve', 52500, 0, false, false, 'Prøv oss i 2 uker', '["2 ukers ubegrenset tilgang", "Alle gruppeklasser", "Ingen binding", "Engangsbeløp"]', true)`,
		
		`INSERT OR IGNORE INTO memberships (id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active) VALUES 
		(8, 'Månedskort', 150000, 1, false, false, 'Ett måneds full tilgang', '["1 måned ubegrenset tilgang", "Automatisk utløp", "Ingen oppsigelse nødvendig"]', true)`,
	}

	// Seed klippekort packages
	klippekort := []string{
		`INSERT OR IGNORE INTO klippekort_packages (id, name, category, klipp_count, price, price_per_session, description, valid_days, active, is_popular) VALUES 
		(1, '5 klipp Personlig Trening', 'Personlig Trening', 5, 300000, 60000, 'Perfekt for å komme i gang', 180, true, false)`,
		
		`INSERT OR IGNORE INTO klippekort_packages (id, name, category, klipp_count, price, price_per_session, description, valid_days, active, is_popular) VALUES 
		(2, '10 klipp Personlig Trening', 'Personlig Trening', 10, 550000, 55000, 'Mest populære pakke', 365, true, true)`,
		
		`INSERT OR IGNORE INTO klippekort_packages (id, name, category, klipp_count, price, price_per_session, description, valid_days, active, is_popular) VALUES 
		(3, '20 klipp Personlig Trening', 'Personlig Trening', 20, 1000000, 50000, 'Best value for money', 365, true, false)`,
		
		`INSERT OR IGNORE INTO klippekort_packages (id, name, category, klipp_count, price, price_per_session, description, valid_days, active, is_popular) VALUES 
		(4, '5 klipp Reformer', 'Reformer', 5, 200000, 40000, 'Prøv Reformer', 180, true, false)`,
		
		`INSERT OR IGNORE INTO klippekort_packages (id, name, category, klipp_count, price, price_per_session, description, valid_days, active, is_popular) VALUES 
		(5, '10 klipp Reformer', 'Reformer', 10, 375000, 37500, 'Populær Reformer pakke', 365, true, true)`,
		
		`INSERT OR IGNORE INTO klippekort_packages (id, name, category, klipp_count, price, price_per_session, description, valid_days, active, is_popular) VALUES 
		(6, '20 klipp Reformer', 'Reformer', 20, 700000, 35000, 'Best verdi for Reformer', 365, true, false)`,
	}

	// Execute membership inserts
	for _, query := range memberships {
		if _, err := dbConn.Exec(query); err != nil {
			log.Printf("Error inserting membership: %v", err)
		}
	}

	// Execute klippekort inserts
	for _, query := range klippekort {
		if _, err := dbConn.Exec(query); err != nil {
			log.Printf("Error inserting klippekort: %v", err)
		}
	}

	log.Println("Seed data inserted successfully!")
}