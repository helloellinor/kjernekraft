package database

import (
	"database/sql"
	"fmt"
	"kjernekraft/models"
	"log"

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
		attendees TEXT
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

	log.Println("Migrering fullført: alle tabeller oppretta.")
	return nil
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
	query := "SELECT id, title, description, start_time, end_time, location FROM events WHERE 1=1"
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
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// CreateEvent creates a new event in the database
func (db *Database) CreateEvent(event models.Event) (int64, error) {
	res, err := db.Conn.Exec(
		"INSERT INTO events (title, description, start_time, end_time, location, organizer) VALUES (?, ?, ?, ?, ?, ?)",
		event.Title, event.Description, event.StartTime, event.EndTime, event.Location, event.Organizer,
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
	rows, err := db.Conn.Query("SELECT id, title, description, start_time, end_time, location, organizer FROM events")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.Organizer); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}
