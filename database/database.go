package database

import (
	"database/sql"
	"fmt"
	"kjernekraft/models"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	Conn *sql.DB
}

func Connect() (*sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./kjernekraft.db"
	}
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable foreign key constraints
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %v", err)
	}

	// Configure connection pool for better performance
	db.SetMaxOpenConns(25)          // Maximum number of open connections
	db.SetMaxIdleConns(5)           // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
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
	eventSignupsTableSQL := `
	CREATE TABLE IF NOT EXISTS event_signups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		event_id INTEGER NOT NULL,
		signup_date DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (event_id) REFERENCES events(id),
		UNIQUE(user_id, event_id)
	);
	`
	membershipRulesTableSQL := `
	CREATE TABLE IF NOT EXISTS membership_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		allow_upgrades BOOLEAN DEFAULT TRUE,
		combine_binding_periods BOOLEAN DEFAULT TRUE,
		allow_downgrades BOOLEAN DEFAULT FALSE,
		allow_change_during_binding BOOLEAN DEFAULT FALSE,
		default_membership_id INTEGER,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (default_membership_id) REFERENCES memberships(id)
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
	if _, err := db.Exec(eventSignupsTableSQL); err != nil {
		return err
	}
	if _, err := db.Exec(membershipRulesTableSQL); err != nil {
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
	
	// Create performance indexes if they don't exist
	if err := createPerformanceIndexes(db); err != nil {
		log.Printf("Warning: Failed to create some performance indexes: %v", err)
		// Don't fail migration for index creation issues
	}
	
	return nil
}

// createPerformanceIndexes creates indexes for better query performance
func createPerformanceIndexes(db *sql.DB) error {
	indexes := []string{
		// User lookup indexes
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone)",
		
		// Event lookup indexes
		"CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time)",
		"CREATE INDEX IF NOT EXISTS idx_events_location ON events(location)",
		"CREATE INDEX IF NOT EXISTS idx_events_teacher_name ON events(teacher_name)",
		
		// Membership indexes
		"CREATE INDEX IF NOT EXISTS idx_user_memberships_user_id ON user_memberships(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_memberships_status ON user_memberships(status)",
		"CREATE INDEX IF NOT EXISTS idx_user_memberships_renewal_date ON user_memberships(renewal_date)",
		
		// Event signup indexes
		"CREATE INDEX IF NOT EXISTS idx_event_signups_user_id ON event_signups(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_event_signups_event_id ON event_signups(event_id)",
		"CREATE INDEX IF NOT EXISTS idx_event_signups_signup_date ON event_signups(signup_date)",
		
		// Role assignment indexes
		"CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id)",
		
		// Payment method indexes
		"CREATE INDEX IF NOT EXISTS idx_payment_methods_user_id ON payment_methods(user_id)",
		
		// Klippekort indexes
		"CREATE INDEX IF NOT EXISTS idx_user_klippekort_user_id ON user_klippekort(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_klippekort_expiry_date ON user_klippekort(expiry_date)",
		"CREATE INDEX IF NOT EXISTS idx_user_klippekort_active ON user_klippekort(is_active)",
	}
	
	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			// Log but don't fail on index creation errors
			log.Printf("Warning: Failed to create index: %s - %v", indexSQL, err)
		}
	}
	
	log.Println("Performance indexes created/verified")
	return nil
}

// isColumnExistsError checks if the error is due to column already existing
func isColumnExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate column name")
}

// handleQueryError provides standardized error handling for database queries
func handleQueryError(err error, operation string, params ...interface{}) error {
	if err == nil {
		return nil
	}
	
	if err == sql.ErrNoRows {
		return nil // For single entity lookups, nil means not found
	}
	
	// Create a descriptive error message
	var paramStr string
	if len(params) > 0 {
		paramStr = fmt.Sprintf(" with params %v", params)
	}
	
	return fmt.Errorf("database %s failed%s: %v", operation, paramStr, err)
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

	// Create default payment methods for the new user
	if err := db.CreateDefaultPaymentMethods(userID); err != nil {
		// Log the error but don't fail user creation
		log.Printf("Warning: Could not create default payment methods for user %d: %v", userID, err)
	}

	return userID, nil
}

// CreateDefaultPaymentMethods creates two default payment cards for a new user
func (db *Database) CreateDefaultPaymentMethods(userID int64) error {
	// Create first default card (Visa simulation)
	card1Query := `INSERT INTO payment_methods (user_id, provider, provider_id) 
	               VALUES (?, 'stripe', ?)`
	
	_, err := db.Conn.Exec(card1Query, userID, fmt.Sprintf("pm_default_visa_%d", userID))
	if err != nil {
		return err
	}
	
	// Create second default card (Mastercard simulation)
	card2Query := `INSERT INTO payment_methods (user_id, provider, provider_id) 
	               VALUES (?, 'stripe', ?)`
	
	_, err = db.Conn.Exec(card2Query, userID, fmt.Sprintf("pm_default_mastercard_%d", userID))
	return err
}

// SimulateBilling creates a simulated charge entry for a user's default payment method
func (db *Database) SimulateBilling(userID int64, amount int, description, chargeType string) error {
	// Get user's first payment method as default
	var paymentMethodID int
	err := db.Conn.QueryRow("SELECT id FROM payment_methods WHERE user_id = ? LIMIT 1", userID).Scan(&paymentMethodID)
	if err != nil {
		return fmt.Errorf("ingen betalingsmetode funnet for bruker")
	}

	// Create a simulated charge (assuming it succeeds)
	chargeQuery := `INSERT INTO charges (user_id, payment_method_id, amount, currency, status, description, type, charge_date, created_at)
	                VALUES (?, ?, ?, 'NOK', 'succeeded', ?, ?, ?, ?)`
	
	now := time.Now()
	
	_, err = db.Conn.Exec(chargeQuery, userID, paymentMethodID, amount, description, chargeType, now, now)
	return err
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

// AddUserMembership creates a new user membership
func (db *Database) AddUserMembership(userID int64, membershipID int64) error {
	// First, check if user already has an active membership
	existingMembership, _ := db.GetUserMembership(userID)
	if existingMembership != nil {
		return fmt.Errorf("bruker har allerede et aktivt medlemskap")
	}

	// Get membership details for start/end dates
	membership, err := db.GetMembershipByID(membershipID)
	if err != nil {
		return err
	}

	now := time.Now()
	startDate := now.Format("2006-01-02")
	renewalDate := now.AddDate(0, 1, 0).Format("2006-01-02") // Next month
	endDate := now.AddDate(0, membership.CommitmentMonths, 0).Format("2006-01-02")
	bindingEnd := endDate // Binding period same as commitment

	query := `INSERT INTO user_memberships (user_id, membership_id, status, start_date, renewal_date, end_date, binding_end, last_billed, created_at)
	          VALUES (?, ?, 'active', ?, ?, ?, ?, ?, ?)`
	
	_, err = db.Conn.Exec(query, userID, membershipID, startDate, renewalDate, endDate, bindingEnd, startDate, now)
	if err != nil {
		return err
	}

	// Simulate billing for the membership
	description := fmt.Sprintf("Medlemskap: %s", membership.Name)
	err = db.SimulateBilling(userID, membership.Price, description, "medlemskap")
	if err != nil {
		log.Printf("Warning: Could not simulate billing for membership purchase: %v", err)
	}

	return nil
}

// ChangeUserMembership changes a user's membership to a different type
func (db *Database) ChangeUserMembership(userID int64, newMembershipID int64) error {
	// Get membership rules
	rules, err := db.GetMembershipRules()
	if err != nil {
		return err
	}

	// Get current membership
	currentMembership, err := db.GetUserMembership(userID)
	if err != nil {
		return err
	}
	if currentMembership == nil {
		return fmt.Errorf("bruker har ingen aktivt medlemskap")
	}

	// Get new membership details
	newMembership, err := db.GetMembershipByID(newMembershipID)
	if err != nil {
		return err
	}

	now := time.Now()
	renewalDate := now.AddDate(0, 1, 0).Format("2006-01-02")
	
	// Calculate new binding end date
	var newBindingEnd string
	isUpgrade := newMembership.Price > currentMembership.Price
	
	if isUpgrade && rules.CombineBindingPeriods && currentMembership.BindingEnd != nil {
		// For upgrades, combine remaining binding time with new commitment
		remainingMonths := 0
		if now.Before(*currentMembership.BindingEnd) {
			// Calculate remaining months in current binding
			remainingMonths = int(currentMembership.BindingEnd.Sub(now).Hours() / (24 * 30))
			if remainingMonths < 0 {
				remainingMonths = 0
			}
		}
		
		// Add new commitment months to remaining months
		totalMonths := remainingMonths + newMembership.CommitmentMonths
		newBindingEndTime := now.AddDate(0, totalMonths, 0)
		newBindingEnd = newBindingEndTime.Format("2006-01-02")
	} else {
		// For downgrades or if not combining, use standard new commitment
		newEndDate := now.AddDate(0, newMembership.CommitmentMonths, 0)
		newBindingEnd = newEndDate.Format("2006-01-02")
	}

	query := `UPDATE user_memberships 
	          SET membership_id = ?, renewal_date = ?, binding_end = ? 
	          WHERE user_id = ? AND status IN ('active', 'paused', 'freeze_requested')`
	
	_, err = db.Conn.Exec(query, newMembershipID, renewalDate, newBindingEnd, userID)
	return err
}

// RemoveUserMembership deactivates a user's membership
func (db *Database) RemoveUserMembership(userID int64) error {
	query := `UPDATE user_memberships SET status = 'cancelled' WHERE user_id = ? AND status IN ('active', 'paused', 'freeze_requested')`
	_, err := db.Conn.Exec(query, userID)
	return err
}

// GetMembershipByID gets a membership by its ID
func (db *Database) GetMembershipByID(membershipID int64) (*models.Membership, error) {
	query := `SELECT id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active 
	          FROM memberships WHERE id = ?`
	
	var membership models.Membership
	err := db.Conn.QueryRow(query, membershipID).Scan(
		&membership.ID, &membership.Name, &membership.Price, &membership.CommitmentMonths,
		&membership.IsStudentSenior, &membership.IsSpecialOffer, &membership.Description,
		&membership.Features, &membership.Active,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &membership, nil
}

// CanChangeMembership checks if a user can change to a specific membership
func (db *Database) CanChangeMembership(userID int64, newMembershipID int64) (bool, string) {
	// Get membership rules
	rules, err := db.GetMembershipRules()
	if err != nil {
		return false, "Kunne ikke hente medlemskapsregler"
	}

	// Get current membership
	currentMembership, err := db.GetUserMembership(userID)
	if err != nil || currentMembership == nil {
		return false, "Bruker har ingen aktivt medlemskap"
	}

	// Get new membership details
	newMembership, err := db.GetMembershipByID(newMembershipID)
	if err != nil {
		return false, "Ugyldig nytt medlemskap"
	}

	// Check if current membership allows changes (must be active or frozen)
	if currentMembership.Status != "active" && currentMembership.Status != "paused" {
		return false, "Medlemskap må være aktivt eller fryst for å bytte"
	}

	// Check binding period based on rules
	now := time.Now()
	if currentMembership.BindingEnd != nil && now.Before(*currentMembership.BindingEnd) {
		if !rules.AllowChangeDuringBinding {
			return false, "Kan ikke bytte medlemskap under bindingsperiode"
		}
	}

	// Check upgrade vs downgrade based on price
	isUpgrade := newMembership.Price > currentMembership.Price
	isDowngrade := newMembership.Price < currentMembership.Price

	if isUpgrade && !rules.AllowUpgrades {
		return false, "Oppgraderinger er ikke tillatt ifølge gjeldende regler"
	}

	if isDowngrade && !rules.AllowDowngrades {
		return false, "Nedgraderinger er ikke tillatt ifølge gjeldende regler"
	}

	// Check if switching involves adding a discount (would require admin approval)
	if newMembership.IsStudentSenior && !currentMembership.IsStudentSenior {
		// This would require admin approval - for now we block it
		return false, "Bytte til student/senior-rabatt krever godkjenning fra admin"
	}

	return true, ""
}

// PurchaseKlippekort creates a new klippekort for a user or adds to existing one
func (db *Database) PurchaseKlippekort(userID int64, packageID int64) error {
	// Get package details
	var pkg models.KlippekortPackage
	query := `SELECT id, name, category, klipp_count, price, price_per_session, description, valid_days, active, is_popular 
	          FROM klippekort_packages WHERE id = ? AND active = TRUE`
	
	err := db.Conn.QueryRow(query, packageID).Scan(
		&pkg.ID, &pkg.Name, &pkg.Category, &pkg.KlippCount, &pkg.Price,
		&pkg.PricePerSession, &pkg.Description, &pkg.ValidDays, &pkg.Active, &pkg.IsPopular,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("klippekort-pakke ikke funnet")
		}
		return err
	}

	// Check if user already has an active klippekort for this category
	existingQuery := `SELECT uk.id, uk.total_klipp, uk.remaining_klipp, uk.expiry_date 
	                  FROM user_klippekort uk
	                  JOIN klippekort_packages kp ON uk.package_id = kp.id
	                  WHERE uk.user_id = ? AND kp.category = ? AND uk.is_active = TRUE AND uk.expiry_date > datetime('now')
	                  ORDER BY uk.expiry_date DESC
	                  LIMIT 1`
	
	var existingID int
	var totalKlipp, remainingKlipp int
	var expiryDate time.Time
	
	err = db.Conn.QueryRow(existingQuery, userID, pkg.Category).Scan(&existingID, &totalKlipp, &remainingKlipp, &expiryDate)
	
	now := time.Now()
	newExpiryDate := now.AddDate(0, 0, pkg.ValidDays)
	
	if err == sql.ErrNoRows {
		// No existing klippekort, create new one
		insertQuery := `INSERT INTO user_klippekort (user_id, package_id, total_klipp, remaining_klipp, expiry_date, purchase_date, is_active)
		                VALUES (?, ?, ?, ?, ?, ?, TRUE)`
		
		_, err = db.Conn.Exec(insertQuery, userID, packageID, pkg.KlippCount, pkg.KlippCount, newExpiryDate, now)
		if err != nil {
			return err
		}

		// Simulate billing for the klippekort
		description := fmt.Sprintf("Klippekort: %s", pkg.Name)
		err = db.SimulateBilling(userID, pkg.Price, description, "klippekort")
		if err != nil {
			log.Printf("Warning: Could not simulate billing for klippekort purchase: %v", err)
		}

		return nil
	} else if err != nil {
		return err
	}
	
	// Existing klippekort found - add to it
	// Check if adding would exceed maximum allowed (20 by default)
	maxKlipp := 20 // TODO: Make this configurable in admin settings
	newTotal := totalKlipp + pkg.KlippCount
	newRemaining := remainingKlipp + pkg.KlippCount
	
	if newTotal > maxKlipp {
		return fmt.Errorf("kan ikke kjøpe flere klipp. Maksimum %d klipp per kort (du har %d)", maxKlipp, totalKlipp)
	}
	
	// Use the longer expiry date (existing or new)
	finalExpiryDate := expiryDate
	if newExpiryDate.After(expiryDate) {
		finalExpiryDate = newExpiryDate
	}
	
	// Update existing klippekort
	updateQuery := `UPDATE user_klippekort 
	                SET total_klipp = ?, remaining_klipp = ?, expiry_date = ?, package_id = ?
	                WHERE id = ?`
	
	_, err = db.Conn.Exec(updateQuery, newTotal, newRemaining, finalExpiryDate, packageID, existingID)
	if err != nil {
		return err
	}

	// Simulate billing for the additional klippekort
	description := fmt.Sprintf("Klippekort tillegg: %s", pkg.Name)
	err = db.SimulateBilling(userID, pkg.Price, description, "klippekort")
	if err != nil {
		log.Printf("Warning: Could not simulate billing for klippekort purchase: %v", err)
	}

	return nil
}

// Event signup related methods

// GetEventByID fetches a single event by ID
func (db *Database) GetEventByID(eventID int64) (*models.Event, error) {
	var event models.Event
	query := `SELECT id, title, description, start_time, end_time, teacher_name, capacity, current_enrolment, class_type
	          FROM events WHERE id = ?`
	
	err := db.Conn.QueryRow(query, eventID).Scan(
		&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime,
		&event.TeacherName, &event.Capacity, &event.CurrentEnrolment, &event.ClassType,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &event, nil
}

// SignupUserForEvent signs up a user for an event
func (db *Database) SignupUserForEvent(userID, eventID int64) error {
	// Check if user is already signed up
	var exists int
	checkQuery := `SELECT COUNT(*) FROM event_signups WHERE user_id = ? AND event_id = ?`
	err := db.Conn.QueryRow(checkQuery, userID, eventID).Scan(&exists)
	if err != nil {
		return err
	}
	
	if exists > 0 {
		return fmt.Errorf("user already signed up for this event")
	}
	
	// Check if event has capacity
	var currentEnrolment, capacity int
	capacityQuery := `SELECT current_enrolment, capacity FROM events WHERE id = ?`
	err = db.Conn.QueryRow(capacityQuery, eventID).Scan(&currentEnrolment, &capacity)
	if err != nil {
		return err
	}
	
	if currentEnrolment >= capacity {
		return fmt.Errorf("event is full")
	}
	
	// Create signup record
	insertQuery := `INSERT INTO event_signups (user_id, event_id, signup_date) VALUES (?, ?, ?)`
	_, err = db.Conn.Exec(insertQuery, userID, eventID, time.Now())
	if err != nil {
		return err
	}
	
	// Update event enrolment count
	updateQuery := `UPDATE events SET current_enrolment = current_enrolment + 1 WHERE id = ?`
	_, err = db.Conn.Exec(updateQuery, eventID)
	return err
}

// CancelUserSignupForEvent cancels a user's signup for an event
func (db *Database) CancelUserSignupForEvent(userID, eventID int64) error {
	// Check if user is signed up
	var exists int
	checkQuery := `SELECT COUNT(*) FROM event_signups WHERE user_id = ? AND event_id = ?`
	err := db.Conn.QueryRow(checkQuery, userID, eventID).Scan(&exists)
	if err != nil {
		return err
	}
	
	if exists == 0 {
		return fmt.Errorf("user is not signed up for this event")
	}
	
	// Remove signup record
	deleteQuery := `DELETE FROM event_signups WHERE user_id = ? AND event_id = ?`
	_, err = db.Conn.Exec(deleteQuery, userID, eventID)
	if err != nil {
		return err
	}
	
	// Update event enrolment count
	updateQuery := `UPDATE events SET current_enrolment = current_enrolment - 1 WHERE id = ?`
	_, err = db.Conn.Exec(updateQuery, eventID)
	return err
}

// GetUserSignupsForEvents returns a map of event IDs that the user is signed up for
func (db *Database) GetUserSignupsForEvents(userID int64, eventIDs []int64) (map[int64]bool, error) {
	if len(eventIDs) == 0 {
		return make(map[int64]bool), nil
	}
	
	// Build query with placeholders for event IDs
	placeholders := make([]string, len(eventIDs))
	args := make([]interface{}, len(eventIDs)+1)
	args[0] = userID
	
	for i, eventID := range eventIDs {
		placeholders[i] = "?"
		args[i+1] = eventID
	}
	
	query := fmt.Sprintf(
		`SELECT event_id FROM event_signups WHERE user_id = ? AND event_id IN (%s)`,
		strings.Join(placeholders, ","),
	)
	
	rows, err := db.Conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	signups := make(map[int64]bool)
	for rows.Next() {
		var eventID int64
		if err := rows.Scan(&eventID); err != nil {
			return nil, err
		}
		signups[eventID] = true
	}
	
	return signups, rows.Err()
}

// GetUserUpcomingSignups returns all upcoming events that the user is signed up for
func (db *Database) GetUserUpcomingSignups(userID int64) ([]models.Event, error) {
	query := `
		SELECT e.id, e.title, e.description, e.role_requirements, e.start_time, e.end_time, 
		       e.location, e.organizer, e.attendees, e.class_type, e.teacher_name, 
		       e.capacity, e.current_enrolment, e.color
		FROM events e
		INNER JOIN event_signups es ON e.id = es.event_id
		WHERE es.user_id = ? AND e.start_time > ?
		ORDER BY e.start_time ASC
	`
	
	now := time.Now()
	rows, err := db.Conn.Query(query, userID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var events []models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID, &event.Title, &event.Description, &event.RoleRequirements,
			&event.StartTime, &event.EndTime, &event.Location, &event.Organizer,
			&event.Attendees, &event.ClassType, &event.TeacherName,
			&event.Capacity, &event.CurrentEnrolment, &event.Color,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	
	return events, rows.Err()
}

// GetMembershipRules retrieves the current membership rules configuration
func (db *Database) GetMembershipRules() (*models.MembershipRules, error) {
	query := `SELECT id, allow_upgrades, combine_binding_periods, allow_downgrades, 
		allow_change_during_binding, default_membership_id, updated_at 
		FROM membership_rules ORDER BY id DESC LIMIT 1`
	
	var rules models.MembershipRules
	err := db.Conn.QueryRow(query).Scan(
		&rules.ID, &rules.AllowUpgrades, &rules.CombineBindingPeriods,
		&rules.AllowDowngrades, &rules.AllowChangeDuringBinding,
		&rules.DefaultMembershipID, &rules.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		// Return default rules if none exist
		return &models.MembershipRules{
			AllowUpgrades:            true,
			CombineBindingPeriods:    true,
			AllowDowngrades:          false,
			AllowChangeDuringBinding: false,
			DefaultMembershipID:      nil,
		}, nil
	}
	
	return &rules, err
}

// SaveMembershipRules saves or updates the membership rules configuration
func (db *Database) SaveMembershipRules(rules *models.MembershipRules) error {
	// First check if any rules exist
	existingRules, err := db.GetMembershipRules()
	if err != nil {
		return err
	}
	
	if existingRules.ID > 0 {
		// Update existing rules
		query := `UPDATE membership_rules SET 
			allow_upgrades = ?, combine_binding_periods = ?, allow_downgrades = ?,
			allow_change_during_binding = ?, default_membership_id = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`
		_, err = db.Conn.Exec(query, rules.AllowUpgrades, rules.CombineBindingPeriods,
			rules.AllowDowngrades, rules.AllowChangeDuringBinding, 
			rules.DefaultMembershipID, existingRules.ID)
	} else {
		// Insert new rules
		query := `INSERT INTO membership_rules 
			(allow_upgrades, combine_binding_periods, allow_downgrades, 
			 allow_change_during_binding, default_membership_id) 
			VALUES (?, ?, ?, ?, ?)`
		_, err = db.Conn.Exec(query, rules.AllowUpgrades, rules.CombineBindingPeriods,
			rules.AllowDowngrades, rules.AllowChangeDuringBinding, rules.DefaultMembershipID)
	}
	
	return err
}

// UpdateMembershipPrice updates the price of a membership
func (db *Database) UpdateMembershipPrice(membershipID int64, newPrice int) error {
	query := `UPDATE memberships SET price = ? WHERE id = ?`
	_, err := db.Conn.Exec(query, newPrice, membershipID)
	return err
}

// CreateMembership creates a new membership
func (db *Database) CreateMembership(membership models.Membership) (int64, error) {
	query := `INSERT INTO memberships 
		(name, price, commitment_months, is_student_senior, is_special_offer, description, features, active) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	// Convert features to JSON if it's not already
	features := membership.Features
	if features == "" {
		features = "[]"
	}
	
	result, err := db.Conn.Exec(query, 
		membership.Name, 
		membership.Price, 
		membership.CommitmentMonths,
		membership.IsStudentSenior,
		membership.IsSpecialOffer,
		membership.Description,
		features,
		membership.Active)
	
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

// DeactivateMembership deactivates a membership (soft delete)
func (db *Database) DeactivateMembership(membershipID int64) error {
	query := `UPDATE memberships SET active = FALSE WHERE id = ?`
	_, err := db.Conn.Exec(query, membershipID)
	return err
}

// UpdateMembershipDetails updates full membership details
func (db *Database) UpdateMembershipDetails(membership models.Membership) error {
	query := `UPDATE memberships SET 
		name = ?, price = ?, commitment_months = ?, is_student_senior = ?, 
		is_special_offer = ?, description = ?, features = ?
		WHERE id = ?`
	
	_, err := db.Conn.Exec(query,
		membership.Name,
		membership.Price,
		membership.CommitmentMonths,
		membership.IsStudentSenior,
		membership.IsSpecialOffer,
		membership.Description,
		membership.Features,
		membership.ID)
	
	return err
}

// DeleteEvent deletes an event
func (db *Database) DeleteEvent(eventID int64) error {
	_, err := db.Conn.Exec("DELETE FROM events WHERE id = ?", eventID)
	return err
}

// UpdateEvent updates an event's details
func (db *Database) UpdateEvent(event models.Event) error {
	query := `UPDATE events SET 
		title = ?, description = ?, start_time = ?, end_time = ?, location = ?, 
		class_type = ?, teacher_name = ?, capacity = ?, color = ?
		WHERE id = ?`
	
	_, err := db.Conn.Exec(query,
		event.Title, event.Description, event.StartTime, event.EndTime, event.Location,
		event.ClassType, event.TeacherName, event.Capacity, event.Color, event.ID)
	
	return err
}