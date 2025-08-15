package database

import (
	"database/sql"
	"fmt"
	"kjernekraft/models"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	Conn *sql.DB
}

func Connect() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./kjernekraft.db")
	if err != nil {
		return nil, err
	}

	log.Println("Kopla til SQLite-databasen.")
	return db, nil
}

func Migrate(db *sql.DB) error {
	eventsTableSQL := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		role_requirements TEXT,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		location TEXT,
		organizer TEXT,
		attendees TEXT,
		class_type TEXT DEFAULT '',
		teacher_name TEXT DEFAULT '',
		capacity INTEGER DEFAULT 0,
		current_enrolment INTEGER DEFAULT 0,
		color TEXT DEFAULT ''
	);
	`
	usersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		birthdate TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		phone TEXT NOT NULL UNIQUE,
		address TEXT,
		postal_code TEXT,
		city TEXT,
		country TEXT,
		password TEXT NOT NULL,
		newsletter_subscription BOOLEAN DEFAULT FALSE,
		terms_accepted BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	rolesTableSQL := `
	CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);
	`
	userRolesTableSQL := `
	CREATE TABLE IF NOT EXISTS user_roles (
		user_id INTEGER NOT NULL,
		role_id INTEGER NOT NULL,
		PRIMARY KEY (user_id, role_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (role_id) REFERENCES roles(id)
	);
	`
	paymentMethodsTableSQL := `
	CREATE TABLE IF NOT EXISTS payment_methods (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		provider TEXT NOT NULL,
		provider_id TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`
	userPaymentMethodsTableSQL := `
	CREATE TABLE IF NOT EXISTS user_payment_methods (
		user_id INTEGER NOT NULL,
		payment_method_id INTEGER NOT NULL,
		PRIMARY KEY (user_id, payment_method_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id)
	);
	`
	membershipsTableSQL := `
	CREATE TABLE IF NOT EXISTS memberships (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		price INTEGER NOT NULL,
		commitment_months INTEGER DEFAULT 0,
		is_student_senior BOOLEAN DEFAULT FALSE,
		is_special_offer BOOLEAN DEFAULT FALSE,
		description TEXT,
		features TEXT,
		active BOOLEAN DEFAULT TRUE
	);
	`
	userMembershipsTableSQL := `
	CREATE TABLE IF NOT EXISTS user_memberships (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		membership_id INTEGER NOT NULL,
		status TEXT DEFAULT 'active',
		start_date DATETIME NOT NULL,
		renewal_date DATETIME NOT NULL,
		end_date DATETIME,
		binding_end DATETIME,
		last_billed DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (membership_id) REFERENCES memberships(id)
	);
	`
	klippekortPackagesTableSQL := `
	CREATE TABLE IF NOT EXISTS klippekort_packages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		category TEXT NOT NULL,
		klipp_count INTEGER NOT NULL,
		price INTEGER NOT NULL,
		price_per_session INTEGER NOT NULL,
		description TEXT,
		valid_days INTEGER DEFAULT 365,
		active BOOLEAN DEFAULT TRUE,
		is_popular BOOLEAN DEFAULT FALSE
	);
	`
	userKlippekortTableSQL := `
	CREATE TABLE IF NOT EXISTS user_klippekort (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		package_id INTEGER NOT NULL,
		total_klipp INTEGER NOT NULL,
		remaining_klipp INTEGER NOT NULL,
		expiry_date DATETIME NOT NULL,
		purchase_date DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN DEFAULT TRUE,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (package_id) REFERENCES klippekort_packages(id)
	);
	`

	log.Println("Kjører migrering (setter opp databasetabeller)...")
	if _, err := db.Exec(eventsTableSQL); err != nil {
		return err
	}
	if _, err := db.Exec(usersTableSQL); err != nil {
		return err
	}
	if _, err := db.Exec(rolesTableSQL); err != nil {
		return err
	}
	if _, err := db.Exec(userRolesTableSQL); err != nil {
		return err
	}
	if _, err := db.Exec(paymentMethodsTableSQL); err != nil {
		return err
	}
	if _, err := db.Exec(userPaymentMethodsTableSQL); err != nil {
		return err
	}
	if _, err := db.Exec(membershipsTableSQL); err != nil {
		return err
	}
	if _, err := db.Exec(userMembershipsTableSQL); err != nil {
		return err
	}
	if _, err := db.Exec(klippekortPackagesTableSQL); err != nil {
		return err
	}
	if _, err := db.Exec(userKlippekortTableSQL); err != nil {
		return err
	}

	log.Println("Migrering fullført: alle tabeller oppretta.")
	
	// Check if last_billed column exists and add it if missing
	var columnExists bool
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('user_memberships') WHERE name='last_billed'").Scan(&columnExists)
	if err == nil && !columnExists {
		_, err = db.Exec("ALTER TABLE user_memberships ADD COLUMN last_billed DATETIME")
		if err != nil {
			return err
		}
		
		// Update existing rows with a default value
		_, err = db.Exec("UPDATE user_memberships SET last_billed = CURRENT_TIMESTAMP WHERE last_billed IS NULL")
		if err != nil {
			return err
		}
		
		log.Println("Added last_billed column to user_memberships table")
	}
	
	return nil
}

// isColumnExistsError checks if the error is due to column already existing
func isColumnExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate column name")
}

// AddRole adds a new role to the roles table
func (db *Database) AddRole(name string) (int64, error) {
	res, err := db.Conn.Exec("INSERT INTO roles (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// AssignRoleToUser links a role to a user
func (db *Database) AssignRoleToUser(userID, roleID int64) error {
	_, err := db.Conn.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID)
	return err
}

// GetUserRoles fetches all roles for a user
func (db *Database) GetUserRoles(userID int64) ([]string, error) {
	rows, err := db.Conn.Query(`SELECT r.name FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// AddPaymentMethod adds a payment method for a user
func (db *Database) AddPaymentMethod(userID int64, provider, providerID string) (int64, error) {
	res, err := db.Conn.Exec("INSERT INTO payment_methods (user_id, provider, provider_id) VALUES (?, ?, ?)", userID, provider, providerID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// AssignPaymentMethodToUser links a payment method to a user
func (db *Database) AssignPaymentMethodToUser(userID, paymentMethodID int64) error {
	_, err := db.Conn.Exec("INSERT INTO user_payment_methods (user_id, payment_method_id) VALUES (?, ?)", userID, paymentMethodID)
	return err
}

// GetUserPaymentMethods fetches all payment methods for a user
func (db *Database) GetUserPaymentMethods(userID int64) ([]struct{ Provider, ProviderID string }, error) {
	rows, err := db.Conn.Query(`SELECT pm.provider, pm.provider_id FROM payment_methods pm JOIN user_payment_methods upm ON pm.id = upm.payment_method_id WHERE upm.user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var methods []struct{ Provider, ProviderID string }
	for rows.Next() {
		var m struct{ Provider, ProviderID string }
		if err := rows.Scan(&m.Provider, &m.ProviderID); err != nil {
			return nil, err
		}
		methods = append(methods, m)
	}
	return methods, nil
}

// CreateUser inserts a new user into the users table
func (db *Database) CreateUser(u models.User) (int64, error) {
	// Check if email already exists
	var existingID int
	err := db.Conn.QueryRow("SELECT id FROM users WHERE email = ?", u.Email).Scan(&existingID)
	if err == nil {
		return 0, fmt.Errorf("e-post er allerede i bruk")
	}
	
	// Check if phone already exists
	err = db.Conn.QueryRow("SELECT id FROM users WHERE phone = ?", u.Phone).Scan(&existingID)
	if err == nil {
		return 0, fmt.Errorf("telefonnummer er allerede i bruk")
	}

	res, err := db.Conn.Exec(
		`INSERT INTO users (name, birthdate, email, phone, address, postal_code, city, country, password, newsletter_subscription, terms_accepted) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.Name, u.Birthdate, u.Email, u.Phone, u.Address, u.PostalCode, u.City, u.Country, u.Password, u.NewsletterSubscription, u.TermsAccepted,
	)
	if err != nil {
		return 0, err
	}
	userID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Assign roles to user
	for _, roleName := range u.Roles {
		roleID, err := db.GetOrCreateRole(roleName)
		if err != nil {
			return 0, err
		}
		if err := db.AssignRoleToUser(userID, roleID); err != nil {
			return 0, err
		}
	}

	return userID, nil
}

func (db *Database) GetOrCreateRole(name string) (int64, error) {
	// Try to get existing role first
	var roleID int64
	err := db.Conn.QueryRow("SELECT id FROM roles WHERE name = ?", name).Scan(&roleID)
	if err == nil {
		return roleID, nil
	}

	// Create new role if it doesn't exist
	res, err := db.Conn.Exec("INSERT INTO roles (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetAllUsers fetches all users from the database
func (db *Database) GetAllUsers() ([]models.User, error) {
	rows, err := db.Conn.Query("SELECT id, name, birthdate, email, phone, address, postal_code, city, country, newsletter_subscription, terms_accepted FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Birthdate, &u.Email, &u.Phone, &u.Address, &u.PostalCode, &u.City, &u.Country, &u.NewsletterSubscription, &u.TermsAccepted); err != nil {
			return nil, err
		}

		// Get roles for this user
		roles, err := db.GetUserRoles(int64(u.ID))
		if err != nil {
			return nil, err
		}
		u.Roles = roles

		users = append(users, u)
	}
	return users, nil
}

func (db *Database) GetFilteredEvents(startDate, endDate, location string) ([]models.Event, error) {
	query := "SELECT id, title, description, start_time, end_time, location, class_type, teacher_name, capacity, current_enrolment, color FROM events WHERE 1=1"
	var args []interface{}

	if startDate != "" {
		query += " AND start_time >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND end_time <= ?"
		args = append(args, endDate)
	}
	if location != "" {
		query += " AND location = ?"
		args = append(args, location)
	}

	rows, err := db.Conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.ClassType, &event.TeacherName, &event.Capacity, &event.CurrentEnrolment, &event.Color); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// CreateEvent creates a new event in the database
func (db *Database) CreateEvent(event models.Event) (int64, error) {
	res, err := db.Conn.Exec(
		"INSERT INTO events (title, description, start_time, end_time, location, organizer, class_type, teacher_name, capacity, current_enrolment, color) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		event.Title, event.Description, event.StartTime, event.EndTime, event.Location, event.Organizer, event.ClassType, event.TeacherName, event.Capacity, event.CurrentEnrolment, event.Color,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateEventTime updates the start and end time of an event
func (db *Database) UpdateEventTime(eventID int64, startTime, endTime string) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET start_time = ?, end_time = ? WHERE id = ?",
		startTime, endTime, eventID,
	)
	return err
}

// GetAllEvents fetches all events from the database
func (db *Database) GetAllEvents() ([]models.Event, error) {
	rows, err := db.Conn.Query("SELECT id, title, description, start_time, end_time, location, organizer, class_type, teacher_name, capacity, current_enrolment, color FROM events")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.Organizer, &event.ClassType, &event.TeacherName, &event.Capacity, &event.CurrentEnrolment, &event.Color); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// GetTodaysEvents fetches events for today
func (db *Database) GetTodaysEvents() ([]models.Event, error) {
	query := `
		SELECT id, title, description, start_time, end_time, location, organizer, class_type, teacher_name, capacity, current_enrolment, color 
		FROM events 
		WHERE DATE(start_time) = DATE('now', 'localtime')
		ORDER BY start_time ASC
	`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.Organizer, &event.ClassType, &event.TeacherName, &event.Capacity, &event.CurrentEnrolment, &event.Color); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// GetThisWeeksEvents fetches events for the current week
func (db *Database) GetThisWeeksEvents() ([]models.Event, error) {
	query := `
		SELECT id, title, description, start_time, end_time, location, organizer, class_type, teacher_name, capacity, current_enrolment, color 
		FROM events 
		WHERE DATE(start_time) >= DATE('now', 'weekday 0', '-6 days', 'localtime') 
		AND DATE(start_time) <= DATE('now', 'weekday 0', 'localtime')
		ORDER BY start_time ASC
	`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.Organizer, &event.ClassType, &event.TeacherName, &event.Capacity, &event.CurrentEnrolment, &event.Color); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// GetEventsForWeek fetches events for a specific week starting from the given Monday
func (db *Database) GetEventsForWeek(mondayDate time.Time) ([]models.Event, error) {
	// Calculate the Sunday of the same week
	sundayDate := mondayDate.AddDate(0, 0, 6)
	
	query := `
		SELECT id, title, description, start_time, end_time, location, organizer, class_type, teacher_name, capacity, current_enrolment, color 
		FROM events 
		WHERE DATE(start_time) >= DATE(?) 
		AND DATE(start_time) <= DATE(?)
		ORDER BY start_time ASC
	`
	rows, err := db.Conn.Query(query, mondayDate.Format("2006-01-02"), sundayDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.Organizer, &event.ClassType, &event.TeacherName, &event.Capacity, &event.CurrentEnrolment, &event.Color); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// GetDistinctTeachers fetches all distinct teacher names from events
func (db *Database) GetDistinctTeachers() ([]string, error) {
	query := `SELECT DISTINCT teacher_name FROM events WHERE teacher_name != '' ORDER BY teacher_name`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []string
	for rows.Next() {
		var teacher string
		if err := rows.Scan(&teacher); err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}
	return teachers, nil
}

// GetDistinctClassTypes fetches all distinct class titles from events
func (db *Database) GetDistinctClassTypes() ([]string, error) {
	query := `SELECT DISTINCT title FROM events WHERE title != '' ORDER BY title`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classTypes []string
	for rows.Next() {
		var classType string
		if err := rows.Scan(&classType); err != nil {
			return nil, err
		}
		classTypes = append(classTypes, classType)
	}
	return classTypes, nil
}

// Membership-related database methods

// GetAllMemberships fetches all active memberships
func (db *Database) GetAllMemberships() ([]models.Membership, error) {
	rows, err := db.Conn.Query("SELECT id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active FROM memberships WHERE active = TRUE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []models.Membership
	for rows.Next() {
		var m models.Membership
		if err := rows.Scan(&m.ID, &m.Name, &m.Price, &m.CommitmentMonths, &m.IsStudentSenior, &m.IsSpecialOffer, &m.Description, &m.Features, &m.Active); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}

// GetUserMembership fetches a user's current membership
func (db *Database) GetUserMembership(userID int64) (*models.MembershipWithDetails, error) {
	query := `
		SELECT um.id, um.user_id, um.membership_id, um.status, um.start_date, um.renewal_date, um.end_date, um.binding_end, um.last_billed, um.created_at,
		       m.name, m.price, m.commitment_months, m.is_student_senior, m.is_special_offer, m.description, m.features, m.active
		FROM user_memberships um
		JOIN memberships m ON um.membership_id = m.id
		WHERE um.user_id = ? AND (um.status = 'active' OR um.status = 'paused' OR um.status = 'freeze_requested')
		ORDER BY um.created_at DESC
		LIMIT 1
	`
	
	var membership models.MembershipWithDetails
	err := db.Conn.QueryRow(query, userID).Scan(
		&membership.UserMembership.ID, &membership.UserMembership.UserID, &membership.UserMembership.MembershipID,
		&membership.UserMembership.Status, &membership.UserMembership.StartDate, &membership.UserMembership.RenewalDate,
		&membership.UserMembership.EndDate, &membership.UserMembership.BindingEnd, &membership.UserMembership.LastBilled, &membership.UserMembership.CreatedAt,
		&membership.Membership.Name, &membership.Membership.Price, &membership.Membership.CommitmentMonths,
		&membership.Membership.IsStudentSenior, &membership.Membership.IsSpecialOffer, &membership.Membership.Description,
		&membership.Membership.Features, &membership.Membership.Active,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active membership found
		}
		return nil, err
	}
	
	return &membership, nil
}

// UpdateMembershipStatus updates the status of a user's membership
func (db *Database) UpdateMembershipStatus(userID int64, status string) error {
	query := `UPDATE user_memberships SET status = ? WHERE user_id = ? AND (status = 'active' OR status = 'paused' OR status = 'freeze_requested')`
	_, err := db.Conn.Exec(query, status, userID)
	return err
}

// Klippekort-related database methods

// GetAllKlippekortPackages fetches all active klippekort packages grouped by category
func (db *Database) GetAllKlippekortPackages() ([]models.KlippekortPackage, error) {
	rows, err := db.Conn.Query("SELECT id, name, category, klipp_count, price, price_per_session, description, valid_days, active, is_popular FROM klippekort_packages WHERE active = TRUE ORDER BY category, price")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []models.KlippekortPackage
	for rows.Next() {
		var p models.KlippekortPackage
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.KlippCount, &p.Price, &p.PricePerSession, &p.Description, &p.ValidDays, &p.Active, &p.IsPopular); err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}
	return packages, nil
}

// GetUserKlippekort fetches all active klippekort for a user
func (db *Database) GetUserKlippekort(userID int64) ([]models.KlippekortWithDetails, error) {
	query := `
		SELECT uk.id, uk.user_id, uk.package_id, uk.total_klipp, uk.remaining_klipp, uk.expiry_date, uk.purchase_date, uk.is_active,
		       kp.name, kp.category, kp.klipp_count, kp.price, kp.price_per_session, kp.description, kp.valid_days, kp.active, kp.is_popular
		FROM user_klippekort uk
		JOIN klippekort_packages kp ON uk.package_id = kp.id
		WHERE uk.user_id = ? AND uk.is_active = TRUE AND uk.expiry_date > datetime('now')
		ORDER BY uk.expiry_date ASC
	`
	
	rows, err := db.Conn.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var klippekort []models.KlippekortWithDetails
	for rows.Next() {
		var k models.KlippekortWithDetails
		if err := rows.Scan(
			&k.UserKlippekort.ID, &k.UserKlippekort.UserID, &k.UserKlippekort.PackageID,
			&k.UserKlippekort.TotalKlipp, &k.UserKlippekort.RemainingKlipp,
			&k.UserKlippekort.ExpiryDate, &k.UserKlippekort.PurchaseDate, &k.UserKlippekort.IsActive,
			&k.KlippekortPackage.Name, &k.KlippekortPackage.Category, &k.KlippekortPackage.KlippCount,
			&k.KlippekortPackage.Price, &k.KlippekortPackage.PricePerSession, &k.KlippekortPackage.Description,
			&k.KlippekortPackage.ValidDays, &k.KlippekortPackage.Active, &k.KlippekortPackage.IsPopular,
		); err != nil {
			return nil, err
		}
		klippekort = append(klippekort, k)
	}
	return klippekort, nil
}

// Authentication methods

// AuthenticateUser verifies user credentials and returns user info if valid
func (db *Database) AuthenticateUser(email, password string) (*models.User, error) {
	var user models.User
	var hashedPassword string
	
	query := `SELECT id, name, email, phone, address, postal_code, city, country, password, newsletter_subscription, terms_accepted
	          FROM users WHERE email = ?`
	
	err := db.Conn.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.Address,
		&user.PostalCode, &user.City, &user.Country, &hashedPassword,
		&user.NewsletterSubscription, &user.TermsAccepted,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ugyldig e-post eller passord")
		}
		return nil, err
	}
	
	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("ugyldig e-post eller passord")
	}
	
	// Get user roles
	roles, err := db.GetUserRoles(int64(user.ID))
	if err != nil {
		return nil, err
	}
	user.Roles = roles
	
	return &user, nil
}

// GetUserByID fetches a user by their ID
func (db *Database) GetUserByID(userID int64) (*models.User, error) {
	var user models.User
	
	query := `SELECT id, name, email, phone, address, postal_code, city, country, newsletter_subscription, terms_accepted
	          FROM users WHERE id = ?`
	
	err := db.Conn.QueryRow(query, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.Address,
		&user.PostalCode, &user.City, &user.Country,
		&user.NewsletterSubscription, &user.TermsAccepted,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bruker ikke funnet")
		}
		return nil, err
	}
	
	// Get user roles
	roles, err := db.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
	
	return &user, nil
}

// GetPendingFreezeRequests returns all memberships with freeze_requested status
func (db *Database) GetPendingFreezeRequests() ([]models.FreezeRequest, error) {
	query := `
		SELECT um.id, um.user_id, um.status, um.start_date, um.renewal_date, um.end_date, um.binding_end, um.last_billed, um.created_at,
		       u.name, u.email, u.phone,
		       m.name, m.price, m.commitment_months
		FROM user_memberships um
		JOIN users u ON um.user_id = u.id
		JOIN memberships m ON um.membership_id = m.id
		WHERE um.status = 'freeze_requested'
		ORDER BY um.created_at DESC`
	
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var requests []models.FreezeRequest
	for rows.Next() {
		var req models.FreezeRequest
		err := rows.Scan(
			&req.MembershipID, &req.UserID, &req.Status, &req.StartDate, &req.RenewalDate, &req.EndDate, &req.BindingEnd, &req.LastBilled, &req.CreatedAt,
			&req.UserName, &req.UserEmail, &req.UserPhone,
			&req.MembershipName, &req.MembershipPrice, &req.CommitmentMonths,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	
	return requests, nil
}

// ApproveFreezeRequest approves a freeze request by setting status to 'paused'
func (db *Database) ApproveFreezeRequest(userID int64) error {
	query := `UPDATE user_memberships SET status = 'paused' WHERE user_id = ? AND status = 'freeze_requested'`
	_, err := db.Conn.Exec(query, userID)
	return err
}

// RejectFreezeRequest rejects a freeze request by setting status back to 'active'
func (db *Database) RejectFreezeRequest(userID int64) error {
	query := `UPDATE user_memberships SET status = 'active' WHERE user_id = ? AND status = 'freeze_requested'`
	_, err := db.Conn.Exec(query, userID)
	return err
}

// UpdateUser updates user profile information
func (db *Database) UpdateUser(user *models.User) error {
	query := `UPDATE users SET name = ?, email = ?, phone = ?, address = ?, postal_code = ?, city = ?, country = ?, birthdate = ? 
	          WHERE id = ?`
	_, err := db.Conn.Exec(query, user.Name, user.Email, user.Phone, user.Address, user.PostalCode, user.City, user.Country, user.Birthdate, user.ID)
	return err
}