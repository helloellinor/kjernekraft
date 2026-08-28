package database

import (
	"database/sql"
	"fmt"
	"kjernekraft/models"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Database struct {
	Conn *sql.DB
}

// DBPathEnv segjer kvar basefila ligg. Stigen var fast — «./kjernekraft.db»
// — og det tydde at prøvorne skreiv i ei fil som laag att millom
// køyringarne. Fyrste gongen gjekk dei; andre gongen fall dei paa
// «e-post er allerede i bruk», av di brukaren fraa fyrre køyringi
// framleis stod der.
const DBPathEnv = "KJERNEKRAFT_DB"

// Connect opnar basen. Stigen kjem or KJERNEKRAFT_DB naar han er sett.
func Connect() (*sql.DB, error) {
	path := os.Getenv(DBPathEnv)
	if path == "" {
		path = "./kjernekraft.db"
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	log.Printf("Kopla til SQLite-databasen (%s).", path)
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
	// Rommet er ein ressurs, ikkje ein tekststreng. Salen tek 18 og
	// Reformer tek 4, og det er den skilnaden heile studioet dreiar um:
	// eit medlemskap gjeld salen, reformeren vert seld for seg.
	//
	// Fyrr laag rommet som fri tekst i events.location, og kapasiteten
	// var eit tal nokon skreiv inn for haand per time — med 20 som
	// utgangspunkt, uansett rom. Det er ei innbjoding til aa selja
	// fjortan plassar som ikkje finst.
	roomsTableSQL := `
	CREATE TABLE IF NOT EXISTS rooms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		capacity INTEGER NOT NULL,
		active BOOLEAN DEFAULT TRUE
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
	if _, err := db.Exec(roomsTableSQL); err != nil {
		return err
	}

	// Dei røynlege romi i Storgata 23.
	if _, err := db.Exec(`INSERT OR IGNORE INTO rooms (name, capacity) VALUES ('Salen', 18), ('Reformer', 4)`); err != nil {
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

	// Student- eller honnørbevis. Studioet gjev 20 % rabatt til den som
	// hev det, og det er brukaren som fortel at han hev det — studioet
	// ser beviset i resepsjonen. Alderen kjem av fødselsdagen; ho treng
	// ingen kolonne.
	var rabattKolonne bool
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='student_senior'").Scan(&rabattKolonne); err == nil && !rabattKolonne {
		if _, err := db.Exec("ALTER TABLE users ADD COLUMN student_senior BOOLEAN DEFAULT FALSE"); err != nil {
			return err
		}
		log.Println("La til student_senior paa users.")
	}

	// Rommet paa ein time. Timane som fanst fyrr peika paa rom gjenom
	// fri tekst i `location`; dei vert kopla yver der namnet stemmer.
	var romKolonne bool
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='room_id'").Scan(&romKolonne); err == nil && !romKolonne {
		if _, err := db.Exec("ALTER TABLE events ADD COLUMN room_id INTEGER REFERENCES rooms(id)"); err != nil {
			return err
		}
		if _, err := db.Exec(`UPDATE events SET room_id = (SELECT id FROM rooms WHERE rooms.name = events.location) WHERE room_id IS NULL`); err != nil {
			return err
		}
		log.Println("La til room_id paa events og kopla dei mot romi.")
	}

	// Regelen attum timane. «Kvar veke i aatte veker» vart skrive inn
	// som aatte sjølvstendige rader, og kva som høyrde saman fanst
	// berre som eit samantreff av like felt. No ber kvar time regelen
	// sin; dei gamle radene vert kopla saman etter det same
	// samantreffet — same time, lærar, rom, vekedag og klokkeslett —
	// éin gong, her, og regelen fær det minste time-id-et sitt som
	// namn.
	// Feilen her vart svelgd fyrr — `err == nil && !regelKolonne`. Svara
	// ikkje spurningi, hoppa migreringa over kolonna og sa ingen ting,
	// og so fall kvart uppslag som les e.rule_id med «no such column»
	// ein heilt annan stad. Ei migrering som ikkje kann prøva om ho
	// trengst, skal stogga.
	var regelKolonne bool
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='rule_id'").Scan(&regelKolonne); err != nil {
		return err
	} else if !regelKolonne {
		if err := leggTilRegelKolonne(db); err != nil {
			return err
		}
		log.Println("La til rule_id paa events og kopla timane til reglane sine.")
	}

	// Frammøtet. Sjaa frammote.go.
	if err := MigrerFrammote(db); err != nil {
		return err
	}
	return nil
}

// leggTilRegelKolonne gjer kolonna og etterfyllinga til éi økt.
//
// Dei stod kvar for seg fyrr, og ALTER TABLE var alt lagra då
// etterfyllinga gjekk. Datt tenaren midt i, fanst kolonna — so vakti
// yver hoppa migreringa neste gong — medan resten av timane stod att
// med rule_id NULL for alltid. Dei hamna då i den regellause gruppa,
// der kvar endring på regelen er eit stille ingen ting.
func leggTilRegelKolonne(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("ALTER TABLE events ADD COLUMN rule_id INTEGER"); err != nil {
		return err
	}
	if err := kopleTimarTilReglar(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// kopleTimarTilReglar gjev kvar gamal time ein regel. Klokka vert lesi
// som ho stend: den lagra tidi er veggklokka i heile huset, og skal
// korkje reknast um her eller nokon annan stad.
func kopleTimarTilReglar(tx *sql.Tx) error {
	rows, err := tx.Query(`SELECT id, title, COALESCE(teacher_name,''), COALESCE(location,''),
		COALESCE(room_id, 0), start_time, end_time FROM events`)
	if err != nil {
		return err
	}
	defer rows.Close()

	grupper := map[string][]int64{}
	for rows.Next() {
		var id, romID int64
		var tittel, laerar, stad string
		var start, slutt time.Time
		if err := rows.Scan(&id, &tittel, &laerar, &stad, &romID, &start, &slutt); err != nil {
			return err
		}
		st := start
		nykel := fmt.Sprintf("%s|%s|%s|%d|%d|%s|%d",
			tittel, laerar, stad, romID, st.Weekday(), st.Format("15:04"),
			int(slutt.Sub(start).Minutes()))
		grupper[nykel] = append(grupper[nykel], id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	// Lesinga lyt vera ferdig fyre skrivinga tek til: ei økt er éin
	// samband, og eit ope svar held han.
	if err := rows.Close(); err != nil {
		return err
	}

	for _, idar := range grupper {
		regel := idar[0]
		for _, id := range idar {
			if id < regel {
				regel = id
			}
		}
		for _, id := range idar {
			if _, err := tx.Exec("UPDATE events SET rule_id = ? WHERE id = ?", regel, id); err != nil {
				return err
			}
		}
	}
	return nil
}

// RoomConflict gjev den fyrste timen som alt ligg i rommet og krossar
// tidsrommet, eller nil.
//
// Tvo tidsrom krossa kvarandre naar den eine byrjar fyre den andre
// endar og endar etter at den andre byrja. Det er heile prøva; ho vert
// ofte skrivi som fire tilfelle, og daa gløymer ein eitt av deim.
// veggtekst skriv eit tidspunkt slik tidene står i events-tabellen:
// klokka på veggen, utan sone.
//
// Drivaren skriv ein time.Time som «2026-08-27 17:00:00+00:00», medan
// radene som alt låg der står som «2026-08-27 17:00:00». To format i
// same kolonna, og SQLite samanliknar dei som tekst — det gjekk godt
// berre av di sona står *etter* klokkeslettet. Ein einaste tid skriven
// med norsk sone hadde velta det: date('2026-08-27 00:30:00+02:00') er
// 26. august, og timen hadde hamna på feil dag i vekelista.
//
// Difor gjeng kvar tid gjennom denne på veg inn og på veg ut att som
// grense i ei spurning. Éin skrivemåte i kolonna, og samanlikningane
// tyder det dei ser ut til å tyde.
func veggtekst(t time.Time) string { return t.Format("2006-01-02 15:04:05") }

func (db *Database) RoomConflict(romID int64, start, slutt time.Time) (*models.Event, error) {
	var e models.Event
	err := db.Conn.QueryRow(`
		SELECT id, title, COALESCE(teacher_name, ''), start_time, end_time
		FROM events
		WHERE room_id = ? AND start_time < ? AND end_time > ?
		ORDER BY start_time LIMIT 1`,
		romID, veggtekst(slutt), veggtekst(start)).Scan(&e.ID, &e.Title, &e.TeacherName, &e.StartTime, &e.EndTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// RoomConflictUtanRegel er den same prøva, men blind for regelen sine
// eigne timar.
//
// Flytter ein heile regelen til eit nytt klokkeslett, står dei gamle
// radene framleis i det gamle sporet medan prøva gjeng. Utan unntaket
// hadde regelen kollidert med seg sjølv og ingen ting late seg flytta.
func (db *Database) RoomConflictUtanRegel(romID, ruleID int64, start, slutt time.Time) (*models.Event, error) {
	var e models.Event
	err := db.Conn.QueryRow(`
		SELECT id, title, COALESCE(teacher_name, ''), start_time, end_time
		FROM events
		WHERE room_id = ? AND COALESCE(rule_id, 0) <> ?
		  AND start_time < ? AND end_time > ?
		ORDER BY start_time LIMIT 1`,
		romID, ruleID, veggtekst(slutt), veggtekst(start)).
		Scan(&e.ID, &e.Title, &e.TeacherName, &e.StartTime, &e.EndTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// RoomConflictUtanTime er den same prøva, blind for éin einskild time.
//
// Ho er syskenet til RoomConflictUtanRegel, og skilnaden er rekkjevidda.
// Flytter ein *heile* regelen, lyt prøva vera blind for alle timane hans.
// Flytter ein éin einskild time — Leon er sjuk den eine tysdagen, og
// timen gjeng ein time seinare — lyt ho vera blind berre for den eine:
// dei andre utslagi av regelen stend framleis, og eit av dei er nett det
// ein kann koma til aa flytta seg oppi. Var ho blind for heile regelen
// her, kunde tvo utslag av den same regelen leggja seg oppaa kvarandre i
// det same rommet utan at nokon sa fraa.
func (db *Database) RoomConflictUtanTime(romID, eventID int64, start, slutt time.Time) (*models.Event, error) {
	var e models.Event
	err := db.Conn.QueryRow(`
		SELECT id, title, COALESCE(teacher_name, ''), start_time, end_time
		FROM events
		WHERE room_id = ? AND id <> ?
		  AND start_time < ? AND end_time > ?
		ORDER BY start_time LIMIT 1`,
		romID, eventID, veggtekst(slutt), veggtekst(start)).
		Scan(&e.ID, &e.Title, &e.TeacherName, &e.StartTime, &e.EndTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// TimeRom gjev rommet timen gjeng i. Null tyder ikkje noko rom — timane
// som vart lagde inn fyre rommi vart ein ressurs hev det, og dei kann
// ikkje kollidera med noko.
//
// GetEventByID hentar han ikkje: ho les ni kolonnor og room_id er ikkje
// ei av deim. Ei spurning som hentar ein heil time for aa lesa eitt tal
// er dessutan meir enn flyttinga treng.
func (db *Database) TimeRom(eventID int64) (int64, error) {
	var romID sql.NullInt64
	err := db.Conn.QueryRow("SELECT room_id FROM events WHERE id = ?", eventID).Scan(&romID)
	if err != nil {
		return 0, err
	}
	return romID.Int64, nil
}

// GetRooms gjev romi studioet hev, med kapasiteten deira.
func (db *Database) GetRooms() ([]models.Room, error) {
	rows, err := db.Conn.Query(`SELECT id, name, capacity FROM rooms WHERE active ORDER BY capacity DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var romi []models.Room
	for rows.Next() {
		var r models.Room
		if err := rows.Scan(&r.ID, &r.Name, &r.Capacity); err != nil {
			return nil, err
		}
		romi = append(romi, r)
	}
	return romi, rows.Err()
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
	query := "SELECT e.id, e.title, COALESCE(e.description, ''), e.start_time, e.end_time, COALESCE(e.location, ''), COALESCE(e.class_type, ''), COALESCE(e.teacher_name, ''), COALESCE(NULLIF(e.capacity, 0), r.capacity, 0), e.current_enrolment, COALESCE(e.color, ''), COALESCE(r.name, e.location, '') FROM events e LEFT JOIN rooms r ON r.id = e.room_id WHERE 1=1"
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
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.ClassType, &event.TeacherName, &event.Capacity, &event.CurrentEnrolment, &event.Color, &event.RoomName); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// CreateEvent creates a new event in the database
func (db *Database) CreateEvent(event models.Event) (int64, error) {
	res, err := db.Conn.Exec(
		`INSERT INTO events (title, description, start_time, end_time, location, room_id,
			organizer, class_type, teacher_name, capacity, current_enrolment, color, rule_id)
		 VALUES (?, ?, ?, ?, ?, NULLIF(?, 0), ?, ?, ?, ?, ?, ?, NULLIF(?, 0))`,
		event.Title, event.Description, veggtekst(event.StartTime), veggtekst(event.EndTime), event.Location, event.RoomID,
		event.Organizer, event.ClassType, event.TeacherName, event.Capacity, event.CurrentEnrolment, event.Color,
		event.RuleID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// LagRegel skriv alle timane i ein ny regel i éi økt.
//
// Regelen finst ikkje som ei eigi rad — han er talet timane hans ber
// saman — og talet vart fyrr henta med eit eige `MAX(rule_id) + 1` fyre
// innskrivinga tok til. To administratorar som la inn kvar sin time i
// same augneblinken fekk då det same talet, og dei to urelaterte
// reglane vart éin: eit lærarbyte på den eine skreiv seg inn på den
// andre òg. Talet vert henta inni økta no, og økta held det til alle
// timane står der.
//
// Innskrivinga er dessutan alt eller ingen ting. Feila den femte av åtte
// vekene fyrr, stod dei fire fyrste att i basen medan svaret var ein
// feil — og den som prøvde ein gong til fekk dei fire ein gong til.
func (db *Database) LagRegel(timar []models.Event) (ruleID int64, ider []int64, err error) {
	if len(timar) == 0 {
		return 0, nil, nil
	}
	tx, err := db.Conn.Begin()
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback()

	if err := tx.QueryRow("SELECT COALESCE(MAX(rule_id), 0) + 1 FROM events").Scan(&ruleID); err != nil {
		return 0, nil, err
	}

	for _, e := range timar {
		res, err := tx.Exec(
			`INSERT INTO events (title, description, start_time, end_time, location, room_id,
				organizer, class_type, teacher_name, capacity, current_enrolment, color, rule_id)
			 VALUES (?, ?, ?, ?, ?, NULLIF(?, 0), ?, ?, ?, ?, ?, ?, ?)`,
			e.Title, e.Description, veggtekst(e.StartTime), veggtekst(e.EndTime), e.Location, e.RoomID,
			e.Organizer, e.ClassType, e.TeacherName, e.Capacity, e.CurrentEnrolment, e.Color, ruleID,
		)
		if err != nil {
			return 0, nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return 0, nil, err
		}
		ider = append(ider, id)
	}
	if err := tx.Commit(); err != nil {
		return 0, nil, err
	}
	return ruleID, ider, nil
}

// UpdateRuleTeacher byter lærar paa alle komande timar i regelen. Det
// som alt er halde stend som det var — historia skriv seg ikkje um.
func (db *Database) UpdateRuleTeacher(ruleID int64, laerar string, fraa time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET teacher_name = ? WHERE rule_id = ? AND end_time > ?",
		laerar, ruleID, veggtekst(fraa),
	)
	return err
}

// UpdateEventTeacher set vikar paa éin einskild time — tannlækjardagen.
// Regelen stend urørd; det er nett denne dagen som fær eit anna namn.
func (db *Database) UpdateEventTeacher(eventID int64, laerar string) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET teacher_name = ? WHERE id = ?",
		laerar, eventID,
	)
	return err
}

// UpdateRuleDescription set skildringi paa alle komande timar i
// regelen — det er regelen som hev ei skildring, timane arvar henne.
func (db *Database) UpdateRuleDescription(ruleID int64, tekst string, fraa time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET description = ? WHERE rule_id = ? AND end_time > ?",
		tekst, ruleID, veggtekst(fraa),
	)
	return err
}

// UpdateRuleTitle byter namn paa alle komande timar i regelen.
//
// Namnet er det du leitar etter i lista, og til no laut ein leggja
// regelen ned og laga honom paa nytt for aa retta ein skrivefeil i
// honom — som tok med seg paameldingane.
func (db *Database) UpdateRuleTitle(ruleID int64, tittel string, fraa time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET title = ? WHERE rule_id = ? AND end_time > ?",
		tittel, ruleID, veggtekst(fraa),
	)
	return err
}

// UpdateRuleRoom flytter alle komande timar i regelen til eit anna rom.
//
// `location` fylgjer med. Han er namnet paa rommet skrive inn i rada, og
// er det gamle namnet naar rommet er eit anna — timeplanen les rom-namnet
// gjenom join-en, men fleire eldre stader les framleis `location`.
func (db *Database) UpdateRuleRoom(ruleID, romID int64, namn string, fraa time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET room_id = NULLIF(?, 0), location = ? WHERE rule_id = ? AND end_time > ?",
		romID, namn, ruleID, veggtekst(fraa),
	)
	return err
}

// UtvidRegel legg fleire timar til ein regel som alt finst.
//
// LagRegel tek eit nytt regelnamn kvar gong; ho kann ikkje brukast til
// aa forlengja ein serie, av di dei nye timane daa hadde vorte ein
// *annan* regel med det same namnet, og lista hadde synt tvo rader der
// det er éin time.
func (db *Database) UtvidRegel(ruleID int64, timar []models.Event) ([]int64, error) {
	if len(timar) == 0 {
		return nil, nil
	}
	tx, err := db.Conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var ider []int64
	for _, e := range timar {
		res, err := tx.Exec(
			`INSERT INTO events (title, description, start_time, end_time, location, room_id,
				organizer, class_type, teacher_name, capacity, current_enrolment, color, rule_id)
			 VALUES (?, ?, ?, ?, ?, NULLIF(?, 0), ?, ?, ?, ?, ?, ?, ?)`,
			e.Title, e.Description, veggtekst(e.StartTime), veggtekst(e.EndTime), e.Location, e.RoomID,
			e.Organizer, e.ClassType, e.TeacherName, e.Capacity, 0, e.Color, ruleID,
		)
		if err != nil {
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		ider = append(ider, id)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ider, nil
}

// SisteITimeregel gjev den siste komande timen i regelen, heil.
//
// GetFutureEventsByRule les fire kolonnor — nok til aa flytta ei rad,
// for lite til aa laga ei ny. Forlengjer ein serien, lyt den nye timen
// arva alt det den gamle bar: namn, lærar, rom, plassar, skildring.
func (db *Database) SisteITimeregel(ruleID int64, fraa time.Time) (*models.Event, error) {
	var e models.Event
	err := db.Conn.QueryRow(`
		SELECT id, title, COALESCE(description, ''), start_time, end_time,
		       COALESCE(location, ''), COALESCE(organizer, ''), COALESCE(class_type, ''),
		       COALESCE(teacher_name, ''), COALESCE(capacity, 0), COALESCE(color, ''),
		       COALESCE(room_id, 0)
		FROM events WHERE rule_id = ? AND end_time > ?
		ORDER BY start_time DESC LIMIT 1`, ruleID, veggtekst(fraa)).
		Scan(&e.ID, &e.Title, &e.Description, &e.StartTime, &e.EndTime,
			&e.Location, &e.Organizer, &e.ClassType, &e.TeacherName, &e.Capacity,
			&e.Color, &e.RoomID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// UpdateRuleCapacity set plassane paa alle komande timar i regelen.
//
// Null tyder «ingi eigi» og gjev rommet ordet attende — difor NULLIF.
// Skreiv me 0 raatt, hadde timen havt null plassar, og COALESCE-en som
// hentar rommet sitt tal hadde aldri sett noko aa henta.
func (db *Database) UpdateRuleCapacity(ruleID int64, plassar int, fraa time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET capacity = NULLIF(?, 0) WHERE rule_id = ? AND end_time > ?",
		plassar, ruleID, veggtekst(fraa),
	)
	return err
}

// PaameldeYver gjev den fyrste komande timen i regelen som hev fleire
// paamelde enn `plassar`, um nokon hev det.
//
// Set ein plassane ned under det som alt er selt, er ikkje spursmaalet
// kva basen toler — det er kven som misser plassen sin. Difor vert det
// spurt fyre, og svaret ber datoen og talet med seg.
func (db *Database) PaameldeYver(ruleID int64, plassar int, fraa time.Time) (*models.Event, error) {
	var e models.Event
	err := db.Conn.QueryRow(`
		SELECT id, title, start_time, current_enrolment
		FROM events
		WHERE rule_id = ? AND end_time > ? AND current_enrolment > ?
		ORDER BY start_time LIMIT 1`, ruleID, veggtekst(fraa), plassar).
		Scan(&e.ID, &e.Title, &e.StartTime, &e.CurrentEnrolment)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// SettVikarFleire set same vikaren paa eit knippe timar i éi økt.
func (db *Database) SettVikarFleire(ider []int64, laerar string) error {
	if len(ider) == 0 {
		return nil
	}
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range ider {
		if _, err := tx.Exec("UPDATE events SET teacher_name = ? WHERE id = ?", laerar, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// AvlysFleire avlyser eit knippe timar i éi økt.
//
// Same grunnen som FlyttFleire: gjekk dei kvar for seg og den fjerde
// feila, var tri avlyste og resten ikkje, medan flata sa «kunde ikkje».
// Anten gjeng heile knippet, eller ingen ting gjer det.
func (db *Database) AvlysFleire(ider []int64) error {
	if len(ider) == 0 {
		return nil
	}
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range ider {
		if _, err := tx.Exec("DELETE FROM events WHERE id = ?", id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetFutureEventsByRule gjev dei komande timane i ein regel, etter dato.
func (db *Database) GetFutureEventsByRule(ruleID int64, fraa time.Time) ([]models.Event, error) {
	rows, err := db.Conn.Query(
		`SELECT id, start_time, end_time, COALESCE(room_id, 0)
		 FROM events WHERE rule_id = ? AND end_time > ? ORDER BY start_time`,
		ruleID, veggtekst(fraa),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timar []models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.StartTime, &e.EndTime, &e.RoomID); err != nil {
			return nil, err
		}
		timar = append(timar, e)
	}
	return timar, rows.Err()
}

// FlyttEvent set ny start og slutt paa éin time. Tidene gjeng inn som
// time.Time og ut att som veggtekst, same vegen som CreateEvent — so
// kolonna held éin skrivemåte.
func (db *Database) FlyttEvent(eventID int64, start, slutt time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET start_time = ?, end_time = ? WHERE id = ?",
		veggtekst(start), veggtekst(slutt), eventID,
	)
	return err
}

// FlyttFleire flytter alle timane i eit knippe i éi økt.
//
// Ein regel er éin ting, og å flytta han er ei handling. Gjekk
// oppdateringane kvar for seg og den fjerde feila, låg tri timar på det
// nye klokkeslettet og resten på det gamle — og flata sa berre «kunne
// ikkje lagra» og sette feltet attende, so ingen visste at rada var
// delt. Anten flytter heile regelen seg, eller ingen ting gjer det.
type Flytting struct {
	EventID      int64
	Start, Slutt time.Time
}

func (db *Database) FlyttFleire(flyttingar []Flytting) error {
	if len(flyttingar) == 0 {
		return nil
	}
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, f := range flyttingar {
		if _, err := tx.Exec(
			"UPDATE events SET start_time = ?, end_time = ? WHERE id = ?",
			veggtekst(f.Start), veggtekst(f.Slutt), f.EventID,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetAllEvents fetches all events from the database
func (db *Database) GetAllEvents() ([]models.Event, error) {
	rows, err := db.Conn.Query("SELECT e.id, e.title, COALESCE(e.description, ''), e.start_time, e.end_time, COALESCE(e.location, ''), COALESCE(e.organizer, ''), COALESCE(e.class_type, ''), COALESCE(e.teacher_name, ''), COALESCE(NULLIF(e.capacity, 0), r.capacity, 0), e.current_enrolment, COALESCE(e.color, ''), COALESCE(r.name, e.location, ''), COALESCE(e.rule_id, 0), COALESCE(e.room_id, 0), COALESCE(e.capacity, 0), COALESCE(r.capacity, 0) FROM events e LEFT JOIN rooms r ON r.id = e.room_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.Organizer, &event.ClassType, &event.TeacherName, &event.Capacity, &event.CurrentEnrolment, &event.Color, &event.RoomName, &event.RuleID, &event.RoomID, &event.EigenPlassar, &event.RoomCapacity); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// GetTodaysEvents fetches events for today
func (db *Database) GetTodaysEvents() ([]models.Event, error) {
	query := `
		SELECT e.id, e.title, COALESCE(e.description, ''), e.start_time, e.end_time, COALESCE(e.location, ''), COALESCE(e.organizer, ''), COALESCE(e.class_type, ''), COALESCE(e.teacher_name, ''), COALESCE(NULLIF(e.capacity, 0), r.capacity, 0), e.current_enrolment, COALESCE(e.color, ''), COALESCE(r.name, e.location, '')
		FROM events e LEFT JOIN rooms r ON r.id = e.room_id 
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
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.Organizer, &event.ClassType, &event.TeacherName, &event.Capacity, &event.CurrentEnrolment, &event.Color, &event.RoomName); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// GetThisWeeksEvents fetches events for the current week
func (db *Database) GetThisWeeksEvents() ([]models.Event, error) {
	query := `
		SELECT e.id, e.title, COALESCE(e.description, ''), e.start_time, e.end_time, COALESCE(e.location, ''), COALESCE(e.organizer, ''), COALESCE(e.class_type, ''), COALESCE(e.teacher_name, ''), COALESCE(NULLIF(e.capacity, 0), r.capacity, 0), e.current_enrolment, COALESCE(e.color, ''), COALESCE(r.name, e.location, '')
		FROM events e LEFT JOIN rooms r ON r.id = e.room_id 
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
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime, &event.Location, &event.Organizer, &event.ClassType, &event.TeacherName, &event.Capacity, &event.CurrentEnrolment, &event.Color, &event.RoomName); err != nil {
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

	// Kapasiteten kjem av rommet. Timen kann setja henne lægre — ein
	// workshop i salen med ti plassar — men han kann ikkje setja henne
	// høgre enn rommet. Difor NULLIF: 0 tyder «ikkje sett».
	query := `
		SELECT e.id, e.title, COALESCE(e.description, ''), e.start_time, e.end_time,
		       COALESCE(e.location, ''), COALESCE(e.organizer, ''),
		       COALESCE(e.class_type, ''), COALESCE(e.teacher_name, ''),
		       COALESCE(NULLIF(e.capacity, 0), r.capacity, 0) AS plassar,
		       e.current_enrolment, e.color,
		       COALESCE(r.id, 0), COALESCE(r.name, e.location), COALESCE(r.capacity, 0)
		FROM events e
		LEFT JOIN rooms r ON r.id = e.room_id
		WHERE DATE(e.start_time) >= DATE(?)
		AND DATE(e.start_time) <= DATE(?)
		ORDER BY e.start_time ASC
	`
	rows, err := db.Conn.Query(query, mondayDate.Format("2006-01-02"), sundayDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.StartTime, &event.EndTime,
			&event.Location, &event.Organizer, &event.ClassType, &event.TeacherName,
			&event.Capacity, &event.CurrentEnrolment, &event.Color,
			&event.RoomID, &event.RoomName, &event.RoomCapacity); err != nil {
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

// MedlemskapFor gjev dei medlemskapi ein brukar kann velja.
//
// Studioet gjev 20 % til den som hev student- eller honnørbevis, og hev
// difor eigne planar for det. Fyrr stod alle side um side, og den som
// ikkje hadde bevis laut lesa seg forbi helvti av lista for aa finna
// sine eigne. Det er arbeid lagt paa lesaren for noko systemet alt veit.
func (db *Database) MedlemskapFor(kvalifisert bool) ([]models.Membership, error) {
	alle, err := db.GetAllMemberships()
	if err != nil {
		return nil, err
	}
	var ut []models.Membership
	for _, m := range alle {
		if m.IsStudentSenior && !kvalifisert {
			continue
		}
		ut = append(ut, m)
	}
	return ut, nil
}

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

	query := `SELECT id, name, COALESCE(birthdate, ''), email, COALESCE(phone, ''),
	                 COALESCE(address, ''), COALESCE(postal_code, ''), COALESCE(city, ''),
	                 COALESCE(country, ''), newsletter_subscription, terms_accepted,
	                 COALESCE(student_senior, 0)
	          FROM users WHERE id = ?`

	err := db.Conn.QueryRow(query, userID).Scan(
		&user.ID, &user.Name, &user.Birthdate, &user.Email, &user.Phone, &user.Address,
		&user.PostalCode, &user.City, &user.Country,
		&user.NewsletterSubscription, &user.TermsAccepted, &user.StudentSenior,
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
// SignupUserForEvent melder ein brukar paa ein time.
//
// Tvo ting var gale her. Han las kapasiteten beinveges or events-rada,
// og etter at rommet vart ein ressurs stend det 0 der naar timen ikkje
// set si eigi — so `7 >= 0` var sant og *kvar* paamelding svara «event
// is full». Han lyt rekna ut den same kapasiteten som resten av huset.
//
// Og han las fyrst og skreiv etterpaa, utan transaksjon. Reformeren hev
// fire plassar; tvo som trykkjer paa den siste samstundes kom baae
// gjenom kontrollen og båe vart melde paa. Talet er ikkje stort nok til
// at ein kann sjaa burt fraa det.
func (db *Database) SignupUserForEvent(userID, eventID int64) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM event_signups WHERE user_id = ? AND event_id = ?`,
		userID, eventID).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return fmt.Errorf("brukaren er alt paameld denne timen")
	}

	// Same utrekningi som i GetEventsForWeek: timen si eigi kapasitet um
	// ho er sett, elles rommet si.
	var paameldte, plassar int
	if err := tx.QueryRow(`
		SELECT e.current_enrolment, COALESCE(NULLIF(e.capacity, 0), r.capacity, 0)
		FROM events e LEFT JOIN rooms r ON r.id = e.room_id
		WHERE e.id = ?`, eventID).Scan(&paameldte, &plassar); err != nil {
		return err
	}
	if plassar <= 0 {
		return fmt.Errorf("timen hev ingi kapasitet sett")
	}
	if paameldte >= plassar {
		return fmt.Errorf("timen er full")
	}

	if _, err := tx.Exec(
		`INSERT INTO event_signups (user_id, event_id, signup_date) VALUES (?, ?, ?)`,
		userID, eventID, time.Now()); err != nil {
		return err
	}

	// Vilkoret stend i UPDATE-en med, so tvo samstundes paameldingar
	// ikkje kann koma forbi kvarandre jamvel um kontrollen yver skulde
	// sleppa baae gjenom.
	res, err := tx.Exec(`
		UPDATE events SET current_enrolment = current_enrolment + 1
		WHERE id = ? AND current_enrolment < COALESCE(NULLIF(capacity, 0),
			(SELECT capacity FROM rooms WHERE rooms.id = events.room_id), 0)`, eventID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("timen er full")
	}

	return tx.Commit()
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
	// role_requirements og attendees vart henta og skanna inn i eit
	// kart og ei liste. SQLite kann ikkje det, so spurningen feila —
	// kvar gong, sidan alltid. Ingen saag det, av di kallaren berre
	// sjekka `err == nil` og lét det vera. Ingen av felti vert nytta av
	// nokon som kallar; dei er ute.
	query := `
		SELECT e.id, e.title, COALESCE(e.description, ''), e.start_time, e.end_time,
		       COALESCE(e.location, ''), COALESCE(e.organizer, ''),
		       COALESCE(e.class_type, ''), COALESCE(e.teacher_name, ''),
		       COALESCE(NULLIF(e.capacity, 0), r.capacity, 0),
		       e.current_enrolment, COALESCE(e.color, ''),
		       COALESCE(r.name, e.location, '')
		FROM events e
		INNER JOIN event_signups es ON e.id = es.event_id
		LEFT JOIN rooms r ON r.id = e.room_id
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
			&event.ID, &event.Title, &event.Description,
			&event.StartTime, &event.EndTime, &event.Location, &event.Organizer,
			&event.ClassType, &event.TeacherName,
			&event.Capacity, &event.CurrentEnrolment, &event.Color,
			&event.RoomName,
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

// Person er ein medlem slik han stend i folkelista.
//
// Tri fakta, ikkje sju: kven, kva han held, og naar han sist var her.
// Fødselsdag og telefon er ikkje noko ein skannar ei liste for — dei
// høyrer til naar rada er open.
type Person struct {
	ID          int
	Namn        string
	Epost       string
	Telefon     string
	Fodd        string
	Roller      string
	Medlemskap  string
	MedlemStod  string
	MedlemPris  int
	KlippAtt    int
	KlippTotalt int
	SistHer     *time.Time
	TimarIAar   int
	TrengSvar   bool // frysing som ventar
	ErLaerar    bool
	ErAdmin     bool
}

// FolkOversyn hentar alle medlemene med det ein treng for aa kjenna
// deim att og sjaa kven som treng noko.
//
// Sorteringi er ikkje alfabetisk. Ho er *kven som treng deg*: fyrst dei
// med ei frysing som ventar svar, so dei som ikkje hev vore her paa
// lenge, so resten. Ei alfabetisk liste er rett naar ein leitar; men
// den som leitar brukar søkjefeltet, og den som *ser* paa lista vil
// vita kva som krev noko av henne.
func (db *Database) FolkOversyn() ([]Person, error) {
	rows, err := db.Conn.Query(`
		SELECT u.id, u.name, u.email, COALESCE(u.phone, ''), COALESCE(u.birthdate, ''),
		       COALESCE((SELECT GROUP_CONCAT(r.name, ', ') FROM user_roles ur
		                 JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = u.id), ''),
		       COALESCE(m.name, ''), COALESCE(um.status, ''), COALESCE(m.price, 0),
		       COALESCE((SELECT SUM(uk.remaining_klipp) FROM user_klippekort uk WHERE uk.user_id = u.id), 0),
		       COALESCE((SELECT SUM(uk.total_klipp) FROM user_klippekort uk WHERE uk.user_id = u.id), 0),
		       (SELECT MAX(e.start_time) FROM event_signups es
		        JOIN events e ON e.id = es.event_id
		        WHERE es.user_id = u.id AND e.start_time < CURRENT_TIMESTAMP),
		       COALESCE((SELECT COUNT(*) FROM event_signups es
		                 JOIN events e ON e.id = es.event_id
		                 WHERE es.user_id = u.id AND e.start_time < CURRENT_TIMESTAMP
		                 AND e.start_time > datetime('now', '-1 year')), 0)
		FROM users u
		LEFT JOIN user_memberships um ON um.user_id = u.id AND um.status != 'cancelled'
		LEFT JOIN memberships m ON m.id = um.membership_id
		ORDER BY u.name COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []Person
	for rows.Next() {
		var p Person
		// MAX() gjev ein streng ut or SQLite, ikkje ei tid — aggregatet
		// misser typen som kolonna hadde. Difor lyt han tolkast her.
		var sist sql.NullString
		if err := rows.Scan(&p.ID, &p.Namn, &p.Epost, &p.Telefon, &p.Fodd, &p.Roller,
			&p.Medlemskap, &p.MedlemStod, &p.MedlemPris,
			&p.KlippAtt, &p.KlippTotalt, &sist, &p.TimarIAar); err != nil {
			return nil, err
		}
		if sist.Valid && sist.String != "" {
			for _, mal := range []string{
				"2006-01-02 15:04:05-07:00", "2006-01-02 15:04:05Z07:00",
				time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z",
			} {
				if t0, err := time.Parse(mal, sist.String); err == nil {
					p.SistHer = &t0
					break
				}
			}
		}
		p.TrengSvar = p.MedlemStod == "freeze_requested"
		p.ErLaerar = harRolla(p.Roller, RollaLaerar)
		p.ErAdmin = harRolla(p.Roller, RollaAdmin)
		ut = append(ut, p)
	}
	return ut, rows.Err()
}
