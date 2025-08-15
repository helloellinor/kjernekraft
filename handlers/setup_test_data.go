package handlers

import (
	"kjernekraft/models"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SetupTestDataHandler creates test users and data for development
func SetupTestDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := setupTestData()
	if err != nil {
		log.Printf("Error setting up test data: %v", err)
		http.Error(w, "Failed to setup test data", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Test data setup completed successfully!"))
}

func setupTestData() error {
	// Create test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	testUser := models.User{
		Name:                   "Anna Larsen",
		Birthdate:              "1990-05-15",
		Email:                  "anna@example.com",
		Phone:                  "+47 12345678",
		Address:                "Testveien 123",
		PostalCode:             "0123",
		City:                   "Oslo",
		Country:                "Norge",
		Password:               string(hashedPassword),
		NewsletterSubscription: true,
		TermsAccepted:          true,
		Roles:                  []string{"user"},
	}

	// Check if user already exists
	existingUser, _ := DB.AuthenticateUser(testUser.Email, "password123")
	if existingUser != nil {
		log.Println("Test user already exists, skipping user creation")
		return setupMembershipAndKlippekortData(int64(existingUser.ID))
	}

	userID, err := DB.CreateUser(testUser)
	if err != nil {
		return err
	}

	log.Printf("Created test user with ID: %d", userID)

	return setupMembershipAndKlippekortData(userID)
}

func setupMembershipAndKlippekortData(userID int64) error {
	// Create membership types if they don't exist
	_, err := DB.Conn.Exec(`INSERT OR IGNORE INTO memberships (id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active) VALUES 
		(1, 'Standard Medlemskap', 99900, 12, false, false, 'Vårt mest populære medlemskap', '["Ubegrenset tilgang", "Gruppetimer", "Rabatt på personlig trening"]', true),
		(2, 'Premium Medlemskap', 149900, 6, false, false, 'All-inclusive pakke', '["Alt fra Standard", "Klippekort inkludert", "Gratis workshops"]', true),
		(3, 'Student Medlemskap', 69900, 0, true, true, 'Spesialpris for studenter', '["Grunnleggende tilgang", "Utvalgte gruppetimer"]', true)
	`)
	if err != nil {
		log.Printf("Warning: Could not create membership types: %v", err)
	}

	// Create klippekort packages if they don't exist
	_, err = DB.Conn.Exec(`INSERT OR IGNORE INTO klippekort_packages (id, name, category, klipp_count, price, price_per_session, description, valid_days, active, is_popular) VALUES 
		(1, '5 Klipp - Gruppetimer', 'Gruppetimer Sal', 5, 49900, 9980, 'Perfekt for å komme i gang', 90, true, false),
		(2, '10 Klipp - Gruppetimer', 'Gruppetimer Sal', 10, 89900, 8990, 'Vårt mest populære tilbud', 120, true, true),
		(3, '20 Klipp - Gruppetimer', 'Gruppetimer Sal', 20, 159900, 7995, 'Best verdi for dedikerte utøvere', 180, true, false),
		(4, '5 Klipp - Reformer', 'Reformer/Apparatus', 5, 74900, 14980, 'Introduksjon til apparattrening', 90, true, false),
		(5, '10 Klipp - Reformer', 'Reformer/Apparatus', 10, 139900, 13990, 'Regelmessig apparattrening', 120, true, true)
	`)
	if err != nil {
		log.Printf("Warning: Could not create klippekort packages: %v", err)
	}

	// Check if user already has membership
	existingMembership, _ := DB.GetUserMembership(userID)
	if existingMembership == nil {
		// Create user membership
		now := time.Now()
		renewalDate := now.AddDate(0, 1, 0) // Next month
		bindingEnd := now.AddDate(1, 0, 0)  // One year binding
		lastBilled := now.AddDate(0, 0, -15) // Billed 15 days ago

		_, err = DB.Conn.Exec(`INSERT INTO user_memberships (user_id, membership_id, status, start_date, renewal_date, binding_end, last_billed) VALUES (?, 1, 'active', ?, ?, ?, ?)`,
			userID, now, renewalDate, bindingEnd, lastBilled)
		if err != nil {
			log.Printf("Warning: Could not create user membership: %v", err)
		}
	}

	// Check if user already has klippekort
	existingKlippekort, _ := DB.GetUserKlippekort(userID)
	if len(existingKlippekort) == 0 {
		// Create user klippekort
		now := time.Now()
		expiryDate := now.AddDate(0, 4, 0) // Expires in 4 months

		_, err = DB.Conn.Exec(`INSERT INTO user_klippekort (user_id, package_id, total_klipp, remaining_klipp, expiry_date, purchase_date) VALUES 
			(?, 2, 10, 7, ?, ?),
			(?, 4, 5, 3, ?, ?)`,
			userID, expiryDate, now,
			userID, expiryDate, now)
		if err != nil {
			log.Printf("Warning: Could not create user klippekort: %v", err)
		}
	}

	log.Printf("Setup test data completed for user ID: %d", userID)
	return nil
}