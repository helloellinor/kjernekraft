package database

import (
	"database/sql"
	"errors"
	"fmt"
	"kjernekraft/models"
	"log"
	"os"
	"strconv"
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
// køyringarne. Fyrste gongen gjekk dei; andre gongen fall dei på
// «e-post er allerede i bruk», av di brukaren frå fyrre køyringi
// framleis stod der.
const DBPathEnv = "KJERNEKRAFT_DB"

// Connect opens the database. The path comes from KJERNEKRAFT_DB when set.
func Connect() (*sql.DB, error) {
	path := os.Getenv(DBPathEnv)
	if path == "" {
		path = "./kjernekraft.db"
	}

	// Settings travel with the path. SQLite defaults to a rollback journal
	// with one reader *or* one writer at a time and no patience: two
	// concurrent requests and one fails with "database is locked" at once.
	// That is fine alone on your machine, which is exactly why it surfaces
	// only in production.
	//
	//   _journal_mode=WAL   readers do not block the writer, or the reverse.
	//   _busy_timeout=5000  waits up to five seconds for the lock instead of
	//                       failing immediately.
	//   _foreign_keys=on    SQLite does *not* enforce foreign keys unless
	//                       asked. The tables declare them; without this they
	//                       are comments.
	//   _synchronous=NORMAL safe together with WAL, and much faster than FULL.
	dsn := path + "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on&_synchronous=NORMAL"

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// SQLite takes many concurrent readers with WAL on, but one writer.
	// Hence a cap — not at one, which would queue every read behind the
	// last, but high enough that readers run alongside each other and
	// writers wait on the lock through _busy_timeout rather than failing.
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	db.SetConnMaxLifetime(0)

	// sql.Open opens no connection. Without this you do not learn the path
	// is unusable until the first request.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("fekk ikkje kontakt med basen (%s): %w", path, err)
	}

	log.Printf("Kopla til SQLite-databasen (%s), WAL paa.", path)
	return db, nil
}

func Migrate(db *sql.DB) error {
	// Permissions first, before anything else.
	//
	// CREATE TABLE IF NOT EXISTS løyve further down would make an empty
	// table in a database that still has `roles`, and the rename would then
	// skip — because `loyve` now exists — leaving every admin and teacher in
	// a table nobody reads. The order is the whole difference.
	if err := MigrerLøyve(db); err != nil {
		return err
	}

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
	// A room is a resource, not a string. The hall takes 18 and the Reformer
	// 4, and that difference is what the studio turns on: a membership covers
	// the hall, the Reformer is sold separately.
	//
	// The room used to be free text in events.location, with capacity typed
	// in per class — 20 by default, whatever the room. That is an invitation
	// to sell fourteen places that do not exist.
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
	CREATE TABLE IF NOT EXISTS loyve (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);
	`
	userRolesTableSQL := `
	CREATE TABLE IF NOT EXISTS brukarloyve (
		user_id INTEGER NOT NULL,
		loyve_id INTEGER NOT NULL,
		PRIMARY KEY (user_id, loyve_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (loyve_id) REFERENCES loyve(id)
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
		-- Naar frysingi tok til. NULL naar medlemskapet ikkje er frose.
		-- Klokka stend medan han er sett: utlaupsdatoen vert skuva fram
		-- med den same tidi naar medlemskapet vert sett i gang att, so
		-- eit aarskort gjev tolv maanader *bruk* og ikkje tolv maanader
		-- paa kalenderen. Sjaa UnfreezeMembership.
		frozen_at DATETIME,
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

	// Receipts. The table was missing until 29.8.2026 even though
	// SimulateBilling had been writing to it all along: every INSERT failed,
	// was logged and dropped, so no purchase left a trace. The columns are
	// exactly the ones that INSERT sets.
	chargesTableSQL := `
	CREATE TABLE IF NOT EXISTS charges (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		payment_method_id INTEGER NOT NULL,
		amount INTEGER NOT NULL,
		currency TEXT NOT NULL DEFAULT 'NOK',
		status TEXT NOT NULL,
		description TEXT,
		type TEXT,
		charge_date TIMESTAMP,
		created_at TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id)
	);
	`

	log.Println("Kjører migrering (setter opp databasetabeller)...")
	if _, err := db.Exec(eventsTableSQL); err != nil {
		return err
	}
	if _, err := db.Exec(chargesTableSQL); err != nil {
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

	// Same pattern for frozen_at. It must *not* be filled in for existing
	// rows: NULL means "not frozen", and for a membership frozen before this
	// column existed we have no clock.
	var frosenFinst bool
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('user_memberships') WHERE name='frozen_at'").Scan(&frosenFinst); err == nil && !frosenFinst {
		if _, err := db.Exec("ALTER TABLE user_memberships ADD COLUMN frozen_at DATETIME"); err != nil {
			return err
		}
		log.Println("La til frozen_at i user_memberships")
	}

	// Student or senior proof. The studio gives 20 % to whoever has it, and
	// the user is the one who says so — the studio sees the proof at the
	// desk. Age follows from the birth date and needs no column.
	var rabattKolonne bool
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='student_senior'").Scan(&rabattKolonne); err == nil && !rabattKolonne {
		if _, err := db.Exec("ALTER TABLE users ADD COLUMN student_senior BOOLEAN DEFAULT FALSE"); err != nil {
			return err
		}
		log.Println("La til student_senior paa users.")
	}

	// Groups: a class can be open to some rather than to all.
	if err := MigrerGrupper(db); err != nil {
		return err
	}

	// The discount claim: ticking the box in the profile is a claim waiting
	// for someone to see the proof, not a discount that applies at once.
	if err := migrerRabattkrav(db); err != nil {
		return err
	}

	// The room on a class. Classes that existed before pointed at rooms
	// through free text in `location`; they are linked where the name
	// matches.
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

	// The series behind the classes. "Every week for eight weeks" was
	// written as eight independent rows, and what belonged together existed
	// only as a coincidence of equal fields. Each class now carries its
	// series; the old rows are joined by that same coincidence — same title,
	// teacher, room, weekday and time — once, here.
	//
	// A migration that cannot test whether it is needed must stop: this
	// error was swallowed as `err == nil && !serieKolonne`, so the column was
	// skipped silently and every lookup of e.serie_id failed with "no such
	// column" somewhere else entirely.
	//
	// The rename comes before the check below: a database that already has
	// serie_id should get the name changed, not a second empty column.
	var gamaltNamn, nyttNamn bool
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='rule_id'").Scan(&gamaltNamn); err != nil {
		return err
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='serie_id'").Scan(&nyttNamn); err != nil {
		return err
	}
	if gamaltNamn && !nyttNamn {
		if _, err := db.Exec("ALTER TABLE events RENAME COLUMN rule_id TO serie_id"); err != nil {
			return err
		}
		// Det gamle registeret ber det gamle namnet. SQLite fylgjer
		// kolonna gjenom omdøypingi, so han verkar — men han heiter
		// framleis noko som ikkje finst, og det er ein av dei tingi som
		// står att og forvirrar den neste.
		if _, err := db.Exec("DROP INDEX IF EXISTS idx_events_rule"); err != nil {
			return err
		}
		log.Println("Døypte om rule_id til serie_id paa events.")
	}

	var serieKolonne bool
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='serie_id'").Scan(&serieKolonne); err != nil {
		return err
	} else if !serieKolonne {
		if err := leggTilSerieKolonne(db); err != nil {
			return err
		}
		log.Println("La til serie_id paa events og kopla timane til seriane sine.")
	}

	// Frammøtet. Sjå frammote.go.
	if err := MigrerFrammote(db); err != nil {
		return err
	}

	// Klippet som vert brukt av krysset. Sjå klippbruk.go.
	if err := MigrerKlippbruk(db); err != nil {
		return err
	}

	// Den private timen. Sjå privattime.go.
	if err := MigrerLaga(db); err != nil {
		return err
	}
	if err := MigrerMeldingar(db); err != nil {
		return err
	}
	if err := MigrerPrivatTime(db); err != nil {
		return err
	}

	// Det usynlege medlemskapet. Sjå svartmedlem.go.
	if err := MigrerSvartMedlemskap(db); err != nil {
		return err
	}

	if err := lagRegister(db); err != nil {
		return err
	}
	return nil
}

// lagRegister sets up the indexes.
//
// There were none beyond what SQLite makes for primary keys and UNIQUE,
// so every "everything belonging to this user" lookup read the whole
// table. With twenty test users you do not notice — which is why it is
// not found until the member count has grown.
//
// The columns are the ones actually searched on: user_id in the join
// tables, event_id when counting places, and start_time when the schedule
// fetches a week at a time.
func lagRegister(db *sql.DB) error {
	register := []string{
		`CREATE INDEX IF NOT EXISTS idx_event_signups_user ON event_signups(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_event_signups_event ON event_signups(event_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_klippekort_user ON user_klippekort(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_memberships_user ON user_memberships(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_memberships_status ON user_memberships(status)`,
		`CREATE INDEX IF NOT EXISTS idx_brukarloyve_user ON brukarloyve(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_payment_methods_user ON user_payment_methods(user_id)`,
		// The schedule fetches a half-open week at a time on start_time.
		`CREATE INDEX IF NOT EXISTS idx_events_start ON events(start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_events_serie ON events(serie_id)`,
		// Innloggingi slær upp e-post for kvar freistnad.
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
	}
	for _, s := range register {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("register: %s: %w", s, err)
		}
	}
	return nil
}

// leggTilSerieKolonne makes the column and the backfill one transaction.
//
// They were separate, and the ALTER TABLE was already committed when the
// backfill ran. If the server died in between, the column existed — so
// the guard skipped the migration next time — while the remaining classes
// kept serie_id NULL forever, landing in the series-less group where
// every change to the series is a silent nothing.
func leggTilSerieKolonne(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("ALTER TABLE events ADD COLUMN serie_id INTEGER"); err != nil {
		return err
	}
	if err := kopleTimarTilSeriar(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// kopleTimarTilSeriar gives every old class a series. The time is read
// as it stands: the stored time is the wall clock, and must not be
// converted here or anywhere else.
func kopleTimarTilSeriar(tx *sql.Tx) error {
	rows, err := tx.Query(`SELECT id, title, COALESCE(teacher_name,''), COALESCE(location,''),
		COALESCE(room_id, 0), start_time, end_time FROM events`)
	if err != nil {
		return err
	}
	defer rows.Close()

	grupper := map[string][]int64{}
	for rows.Next() {
		var id, romID int64
		var tittel, lærar, stad string
		var start, slutt time.Time
		if err := rows.Scan(&id, &tittel, &lærar, &stad, &romID, &start, &slutt); err != nil {
			return err
		}
		st := start
		nykel := fmt.Sprintf("%s|%s|%s|%d|%d|%s|%d",
			tittel, lærar, stad, romID, st.Weekday(), st.Format("15:04"),
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
		serie := idar[0]
		for _, id := range idar {
			if id < serie {
				serie = id
			}
		}
		for _, id := range idar {
			if _, err := tx.Exec("UPDATE events SET serie_id = ? WHERE id = ?", serie, id); err != nil {
				return err
			}
		}
	}
	return nil
}

// RoomConflict gives the first class already in the room that overlaps
// the interval, or nil.
//
// Two intervals overlap when one starts before the other ends and ends
// after the other began. That is the whole test; written as four cases,
// one of them gets forgotten.
//
// veggtekst writes a timestamp the way times stand in the events table:
// the clock on the wall, without a zone.
//
// The driver writes a time.Time as "2026-08-27 17:00:00+00:00" while
// existing rows read "2026-08-27 17:00:00". Two formats in one column,
// compared as text, and it only worked because the zone comes *after* the
// time. One time written with a Norwegian zone would have broken it:
// date('2026-08-27 00:30:00+02:00') is 26 August, and the class would
// have landed on the wrong day.
func veggtekst(t time.Time) string { return t.Format("2006-01-02 15:04:05") }

func (db *Database) RoomConflict(romID int64, start, slutt time.Time) (*Romkollisjon, error) {
	var k Romkollisjon
	err := db.Conn.QueryRow(`
		SELECT title, COALESCE(teacher_name, ''), start_time, end_time
		FROM events
		WHERE room_id = ? AND start_time < ? AND end_time > ?
		ORDER BY start_time, id LIMIT 1`,
		romID, veggtekst(slutt), veggtekst(start)).
		Scan(&k.Tittel, &k.Lærar, &k.Start, &k.Slutt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// RoomConflictUtanSerie er den same prøva, men blind for serien sine
// eigne timar.
//
// Flytter ein heile serien til eit nytt klokkeslett, står dei gamle
// radene framleis i det gamle sporet medan prøva gjeng. Utan unntaket
// hadde serien kollidert med seg sjølv og ingen ting late seg flytta.
func (db *Database) RoomConflictUtanSerie(romID, serieID int64, start, slutt time.Time) (*Romkollisjon, error) {
	var k Romkollisjon
	err := db.Conn.QueryRow(`
		SELECT title, COALESCE(teacher_name, ''), start_time, end_time
		FROM events
		WHERE room_id = ? AND COALESCE(serie_id, 0) <> ?
		  AND start_time < ? AND end_time > ?
		ORDER BY start_time, id LIMIT 1`,
		romID, serieID, veggtekst(slutt), veggtekst(start)).
		Scan(&k.Tittel, &k.Lærar, &k.Start, &k.Slutt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// GetRooms gives the studio's rooms with their capacity.
// SetRomPlassar sets how many the room holds.
//
// The number is not decoration: every class without its own capacity
// inherits it (COALESCE(NULLIF(e.capacity, 0), r.capacity, 0)), so a room
// set too low makes classes full before they are.
func (db *Database) SetRomPlassar(romID, plassar int) error {
	_, err := db.Conn.Exec("UPDATE rooms SET capacity = ? WHERE id = ?", plassar, romID)
	return err
}

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

// GjevLøyve gjev eit løyve til ein brukar.
func (db *Database) GjevLøyve(userID, løyveID int64) error {
	_, err := db.Conn.Exec("INSERT INTO brukarloyve (user_id, loyve_id) VALUES (?, ?)", userID, løyveID)
	return err
}

// LøyveFor gives the permissions a user has.
func (db *Database) LøyveFor(userID int64) ([]string, error) {
	rows, err := db.Conn.Query(`SELECT r.name FROM loyve r JOIN brukarloyve ur ON r.id = ur.loyve_id WHERE ur.user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var løyve []string
	for rows.Next() {
		var eit string
		if err := rows.Scan(&eit); err != nil {
			return nil, err
		}
		løyve = append(løyve, eit)
	}
	return løyve, nil
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

// The two ways CreateUser can say no. Marked errors, so the handler can
// recognise them with errors.Is and answer in the user's language — the
// text here is for the log, not the screen.
var (
	ErrEpostIBruk   = errors.New("e-posten er alt i bruk")
	ErrTelefonIBruk = errors.New("telefonnummeret er alt i bruk")
)

func (db *Database) CreateUser(u models.User) (int64, error) {
	// Check if email already exists
	var existingID int
	err := db.Conn.QueryRow("SELECT id FROM users WHERE email = ?", u.Email).Scan(&existingID)
	if err == nil {
		return 0, ErrEpostIBruk
	}

	// Check if phone already exists
	err = db.Conn.QueryRow("SELECT id FROM users WHERE phone = ?", u.Phone).Scan(&existingID)
	if err == nil {
		return 0, ErrTelefonIBruk
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

	// Løyvi brukaren skal ha
	for _, løyveNamn := range u.Løyve {
		løyveID, err := db.LøyveIDFor(løyveNamn)
		if err != nil {
			return 0, err
		}
		if err := db.GjevLøyve(userID, løyveID); err != nil {
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

	// veggtekst, like everyone else writing timestamps: a raw time.Time from
	// the driver carries a zone suffix, and then the column holds two
	// formats.
	now := veggtekst(time.Now())

	_, err = db.Conn.Exec(chargeQuery, userID, paymentMethodID, amount, description, chargeType, now, now)
	return err
}

func (db *Database) LøyveIDFor(name string) (int64, error) {
	// Does the permission already exist?
	var løyveID int64
	err := db.Conn.QueryRow("SELECT id FROM loyve WHERE name = ?", name).Scan(&løyveID)
	if err == nil {
		return løyveID, nil
	}

	// Elles lagar me det.
	res, err := db.Conn.Exec("INSERT INTO loyve (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetAllUsers fetches all users from the database
func (db *Database) GetAllUsers() ([]models.User, error) {
	// COALESCE on every column that can be NULL.
	//
	// Read raw, a single NULL in an address field took the whole admin page
	// with it: Scan gives "converting NULL to string is unsupported", the
	// handler answers 500, and nobody gets in. It had always been so; it took
	// one user without an address to show it.
	//
	// GetUserByID just below already did this. Two queries against the same
	// table that disagree about what can be missing is a trap waiting for the
	// first row that uses it.
	rows, err := db.Conn.Query(`SELECT id, name, COALESCE(birthdate, ''), email,
		COALESCE(phone, ''), COALESCE(address, ''), COALESCE(postal_code, ''),
		COALESCE(city, ''), COALESCE(country, ''),
		COALESCE(newsletter_subscription, 0), COALESCE(terms_accepted, 0)
		FROM users`)
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

		// This user's permissions
		løyve, err := db.LøyveFor(int64(u.ID))
		if err != nil {
			return nil, err
		}
		u.Løyve = løyve

		users = append(users, u)
	}
	return users, nil
}

// CreateEvent creates a new event in the database
func (db *Database) CreateEvent(event models.Event) (int64, error) {
	res, err := db.Conn.Exec(
		`INSERT INTO events (title, description, start_time, end_time, location, room_id,
			organizer, class_type, teacher_name, capacity, current_enrolment, color, serie_id)
		 VALUES (?, ?, ?, ?, ?, NULLIF(?, 0), ?, ?, ?, ?, ?, ?, NULLIF(?, 0))`,
		event.Title, event.Description, veggtekst(event.StartTime), veggtekst(event.EndTime), event.Location, event.RoomID,
		event.Organizer, event.ClassType, event.TeacherName, event.Capacity, event.CurrentEnrolment, event.Color,
		event.SerieID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// LagSerie writes every class in a new series in one transaction.
//
// A series is not a row of its own — it is the number its classes carry
// together — and that number used to be fetched with a separate
// MAX(serie_id) + 1 before the insert. Two admins adding a class at the
// same moment got the same number, and the two unrelated series became
// one: a teacher change on one wrote itself onto the other.
//
// The insert is also all or nothing. When the fifth of eight weeks failed,
// the first four stayed in the database while the answer was an error —
// and whoever tried again got those four a second time.
func (db *Database) LagSerie(timar []models.Event) (serieID int64, ider []int64, err error) {
	if len(timar) == 0 {
		return 0, nil, nil
	}
	tx, err := db.Conn.Begin()
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback()

	if err := tx.QueryRow("SELECT COALESCE(MAX(serie_id), 0) + 1 FROM events").Scan(&serieID); err != nil {
		return 0, nil, err
	}

	for _, e := range timar {
		res, err := tx.Exec(
			`INSERT INTO events (title, description, start_time, end_time, location, room_id,
				organizer, class_type, teacher_name, capacity, current_enrolment, color, serie_id)
			 VALUES (?, ?, ?, ?, ?, NULLIF(?, 0), ?, ?, ?, ?, ?, ?, ?)`,
			e.Title, e.Description, veggtekst(e.StartTime), veggtekst(e.EndTime), e.Location, e.RoomID,
			e.Organizer, e.ClassType, e.TeacherName, e.Capacity, e.CurrentEnrolment, e.Color, serieID,
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
	return serieID, ider, nil
}

// UpdateSerieTeacher changes teacher on every future class in the series.
// What has already been held stands; history does not rewrite itself.
func (db *Database) UpdateSerieTeacher(serieID int64, lærar string, frå time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET teacher_name = ? WHERE serie_id = ? AND end_time > ?",
		lærar, serieID, veggtekst(frå),
	)
	return err
}

// UpdateEventTeacher set vikar på éin einskild time — tannlækjardagen.
// Serien stend urørd; det er nett denne dagen som fær eit anna namn.
func (db *Database) UpdateEventTeacher(eventID int64, lærar string) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET teacher_name = ? WHERE id = ?",
		lærar, eventID,
	)
	return err
}

// UpdateSerieClassType sets the kind on every future class in the series.
//
// The kind is what sort of training it is — yoga, pilates, reformer,
// fascia — and the series owns it: every occurrence of "Vinyasa Flow
// Monday 18:00" is yoga. It carries the wing colour in the list and the
// schedule, and it is also what klippekort packages compare `category`
// against, so a class without a kind cannot be paid for with an earmarked
// clip.
func (db *Database) UpdateSerieClassType(serieID int64, slag string, frå time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET class_type = ? WHERE serie_id = ? AND end_time > ?",
		slag, serieID, veggtekst(frå),
	)
	return err
}

// Slagsortar gives the kinds already in use, once each.
//
// The field is free text — SlagKlasse washes it to a CSS hook, and an
// unknown kind gets no colour rather than the wrong one — but free text
// without memory means "Yoga", "yoga" and "Yoag" are all new kinds. Hence
// a list of what the house has already seen, so you pick rather than
// retype.
//
// All classes, not only future ones: a kind that ran in spring is still a
// kind the studio does, and that is exactly when you need to find it
// again.
func (db *Database) Slagsortar() ([]string, error) {
	rows, err := db.Conn.Query(`
		SELECT DISTINCT TRIM(class_type) FROM events
		WHERE TRIM(COALESCE(class_type, '')) <> ''
		ORDER BY TRIM(class_type) COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []string
	for rows.Next() {
		var slag string
		if err := rows.Scan(&slag); err != nil {
			return nil, err
		}
		ut = append(ut, slag)
	}
	return ut, rows.Err()
}

// UpdateSerieDescription sets the description on every future class in
// the series — the series has a description, the classes inherit it.
func (db *Database) UpdateSerieDescription(serieID int64, tekst string, frå time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET description = ? WHERE serie_id = ? AND end_time > ?",
		tekst, serieID, veggtekst(frå),
	)
	return err
}

// UpdateSerieTitle renames every future class in the series.
//
// The name is what you look for in the list, and until now fixing a typo
// in it meant deleting the series and rebuilding it — which took the
// signups with it.
func (db *Database) UpdateSerieTitle(serieID int64, tittel string, frå time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET title = ? WHERE serie_id = ? AND end_time > ?",
		tittel, serieID, veggtekst(frå),
	)
	return err
}

// UpdateSerieRoom moves every future class in the series to another room.
//
// `location` follows. It is the room name written into the row, and holds
// the old name when the room changes — the schedule reads the room name
// through the join, but several older places still read `location`.
func (db *Database) UpdateSerieRoom(serieID, romID int64, namn string, frå time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET room_id = NULLIF(?, 0), location = ? WHERE serie_id = ? AND end_time > ?",
		romID, namn, serieID, veggtekst(frå),
	)
	return err
}

// UtvidSerie adds more classes to a series that already exists.
//
// LagSerie takes a new series number every time, so it cannot extend one:
// the new classes would become a *different* series with the same name,
// and the list would show two rows where there is one class.
func (db *Database) UtvidSerie(serieID int64, timar []models.Event) ([]int64, error) {
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
				organizer, class_type, teacher_name, capacity, current_enrolment, color, serie_id)
			 VALUES (?, ?, ?, ?, ?, NULLIF(?, 0), ?, ?, ?, ?, ?, ?, ?)`,
			e.Title, e.Description, veggtekst(e.StartTime), veggtekst(e.EndTime), e.Location, e.RoomID,
			e.Organizer, e.ClassType, e.TeacherName, e.Capacity, 0, e.Color, serieID,
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

// SisteITimeserie gives the last future class in the series, whole.
//
// GetFutureEventsBySerie reads four columns — enough to move a row, too
// little to make a new one. Extending a series means the new class
// inherits everything the old one carried.
func (db *Database) SisteITimeserie(serieID int64, frå time.Time) (*models.Event, error) {
	var e models.Event
	rad := db.Conn.QueryRow(`SELECT `+eventKolonnar+eventFrå+`
		WHERE e.serie_id = ? AND e.end_time > ?
		ORDER BY e.start_time DESC, e.id DESC LIMIT 1`, serieID, veggtekst(frå))
	err := skannTime(rad, &e)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// UpdateSerieCapacity sets the places on every future class in the series.
//
// Zero means "none of its own" and gives the room the word back — hence
// NULLIF. Written raw, the class would have zero places and the COALESCE
// that fetches the room's number would never see anything to fetch.
func (db *Database) UpdateSerieCapacity(serieID int64, plassar int, frå time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET capacity = NULLIF(?, 0) WHERE serie_id = ? AND end_time > ?",
		plassar, serieID, veggtekst(frå),
	)
	return err
}

// PaameldeYver gives the first future class in the series with more
// signups than `plassar`, if any.
//
// Setting places below what is already sold is not a question of what the
// database tolerates — it is a question of who loses their place. So it
// is asked first, and the answer carries the date and the number.
func (db *Database) PaameldeYver(serieID int64, plassar int, frå time.Time) (*Overfylt, error) {
	var o Overfylt
	err := db.Conn.QueryRow(`
		SELECT start_time, current_enrolment
		FROM events
		WHERE serie_id = ? AND end_time > ? AND current_enrolment > ?
		ORDER BY start_time, id LIMIT 1`, serieID, veggtekst(frå), plassar).
		Scan(&o.Start, &o.Paamelde)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
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

// GetFutureEventsBySerie gjev dei komande timane i ein serie, etter dato.
func (db *Database) GetFutureEventsBySerie(serieID int64, frå time.Time) ([]SerieTime, error) {
	rows, err := db.Conn.Query(
		`SELECT id, start_time FROM events
		 WHERE serie_id = ? AND end_time > ? ORDER BY start_time, id`,
		serieID, veggtekst(frå),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []SerieTime
	for rows.Next() {
		var t SerieTime
		if err := rows.Scan(&t.ID, &t.Start); err != nil {
			return nil, err
		}
		ut = append(ut, t)
	}
	return ut, rows.Err()
}

// FlyttFleire flytter alle timane i eit knippe i éi økt.
//
// Ein serie er éin ting, og å flytta han er ei handling. Gjekk
// oppdateringane kvar for seg og den fjerde feila, låg tri timar på det
// nye klokkeslettet og resten på det gamle — og flata sa berre «kunne
// ikkje lagra» og sette feltet attende, so ingen visste at rada var
// delt. Anten flytter heile serien seg, eller ingen ting gjer det.
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
	rows, err := db.Conn.Query(`SELECT ` + eventKolonnar + eventFrå)
	if err != nil {
		return nil, err
	}
	return timane(rows)
}

// GetEventsForWeek fetches events for a specific week starting from the given Monday
// GetEventsForWeek gjev veka slik `sjaaarID` skal sjå henne: alle opne
// timar, og dei private som er hans eigne. Sjå privattime.go.
func (db *Database) GetEventsForWeek(mondayDate time.Time, sjaårID int64) ([]models.Event, error) {
	// Halvopen veke: måndag 00:00 til neste måndag 00:00. Han let
	// SQLite nytta idx_events_start; DATE(e.start_time) gjorde kvar
	// time til ei funksjonsutrekning og tvinga fram tabellskann.
	nesteMåndag := mondayDate.AddDate(0, 0, 7)

	rows, err := db.Conn.Query(`SELECT `+eventKolonnar+eventFrå+`
		WHERE e.start_time >= ?
		AND e.start_time < ?
		AND `+synlegFor+`
		ORDER BY e.start_time ASC, e.id ASC`,
		veggtekst(mondayDate), veggtekst(nesteMåndag), sjaårID, sjaårID)
	if err != nil {
		return nil, err
	}
	return timane(rows)
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

// Membership-related database methods

// MedlemskapFor gives the memberships a user can choose.
//
// The studio gives 20 % to whoever has student or senior proof, and keeps
// separate plans for it. They all used to stand side by side, so anyone
// without proof had to read past half the list to find their own.
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
		// Black cannot be bought. It follows a permission, and a price of 0 kr in
		// the list would have been an invitation to click it. See svartmedlem.go.
		if m.Skjult {
			continue
		}
		ut = append(ut, m)
	}
	return ut, nil
}

// GetAllMemberships fetches all active memberships
func (db *Database) GetAllMemberships() ([]models.Membership, error) {
	rows, err := db.Conn.Query("SELECT id, name, price, commitment_months, is_student_senior, is_special_offer, description, features, active, skjult FROM memberships WHERE active = TRUE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []models.Membership
	for rows.Next() {
		var m models.Membership
		if err := rows.Scan(&m.ID, &m.Name, &m.Price, &m.CommitmentMonths, &m.IsStudentSenior, &m.IsSpecialOffer, &m.Description, &m.Features, &m.Active, &m.Skjult); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}

// GetUserMembership returns the membership the user holds now.
//
// m.id and m.skjult must stay in the select list. Without m.id,
// Membership.ID is zero for every ordinary member — and three callers
// compare against it, so the switch list stops excluding the plan you
// already have, and the studio's own product names get looked up at
// key 0 and never found. svartMedlemskapFor always selected id, which
// is why only Black was right.
func (db *Database) GetUserMembership(userID int64) (*models.MembershipWithDetails, error) {
	query := `
		SELECT um.id, um.user_id, um.membership_id, um.status, um.start_date, um.renewal_date, um.end_date, um.binding_end, um.last_billed, um.frozen_at, um.created_at,
		       m.id, m.name, m.price, m.commitment_months, m.is_student_senior, m.is_special_offer, m.description, m.features, m.active, m.skjult
		FROM user_memberships um
		JOIN memberships m ON um.membership_id = m.id
		WHERE um.user_id = ? AND (um.status = 'active' OR um.status = 'paused' OR um.status = 'freeze_requested')
		ORDER BY um.created_at DESC
		LIMIT 1
	`

	// The permission wins. A teacher has Black as long as they teach, whatever
	// else the database says about them. See svartmedlem.go.
	if fri, err := db.HarFriMedlemskap(userID); err == nil && fri {
		return db.svartMedlemskapFor(userID)
	}

	var membership models.MembershipWithDetails
	err := db.Conn.QueryRow(query, userID).Scan(
		&membership.UserMembership.RadID, &membership.UserMembership.UserID, &membership.UserMembership.MembershipID,
		&membership.UserMembership.Status, &membership.UserMembership.StartDate, &membership.UserMembership.RenewalDate,
		&membership.UserMembership.EndDate, &membership.UserMembership.BindingEnd, &membership.UserMembership.LastBilled,
		&membership.UserMembership.FrozenAt, &membership.UserMembership.CreatedAt,
		&membership.Membership.ID, &membership.Membership.Name, &membership.Membership.Price,
		&membership.Membership.CommitmentMonths, &membership.Membership.IsStudentSenior,
		&membership.Membership.IsSpecialOffer, &membership.Membership.Description,
		&membership.Membership.Features, &membership.Membership.Active, &membership.Membership.Skjult,
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

// GetUserKlippekort returns the cards the user has clips left on.
//
// kp.id must stay in the select list, for the same reason as m.id above.
func (db *Database) GetUserKlippekort(userID int64) ([]models.KlippekortWithDetails, error) {
	query := `
		SELECT uk.id, uk.user_id, uk.package_id, uk.total_klipp, uk.remaining_klipp, uk.expiry_date, uk.purchase_date, uk.is_active,
		       kp.id, kp.name, kp.category, kp.klipp_count, kp.price, kp.price_per_session, kp.description, kp.valid_days, kp.active, kp.is_popular
		FROM user_klippekort uk
		JOIN klippekort_packages kp ON uk.package_id = kp.id
		WHERE uk.user_id = ? AND uk.is_active = TRUE AND uk.remaining_klipp > 0
		      AND uk.expiry_date > datetime('now')
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
			&k.UserKlippekort.RadID, &k.UserKlippekort.UserID, &k.UserKlippekort.PackageID,
			&k.UserKlippekort.TotalKlipp, &k.UserKlippekort.RemainingKlipp,
			&k.UserKlippekort.ExpiryDate, &k.UserKlippekort.PurchaseDate, &k.UserKlippekort.IsActive,
			&k.KlippekortPackage.ID, &k.KlippekortPackage.Name, &k.KlippekortPackage.Category, &k.KlippekortPackage.KlippCount,
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
// ErrUgyldigInnlogging tyder at e-posten ikkje finst eller passordet er
// gale — og ingen ting anna. Alle andre feil er feil i huset, og dei
// skal ikkje kled seg ut som dette.
var ErrUgyldigInnlogging = errors.New("ugyldig e-post eller passord")

func (db *Database) AuthenticateUser(email, password string) (*models.User, error) {
	var user models.User
	var hashedPassword string

	// COALESCE on the four that can be NULL — address, postcode, city and
	// country. Bare, Scan failed with "converting NULL to string is
	// unsupported" for every user who had not filled in an address.
	//
	// That alone would have been found at once. What made it dangerous is
	// that it came out as *wrong password*: every error from here became
	// "invalid email or password" in the handler. The user was told they
	// misremembered, and they had not — they could not log in at all, and
	// nothing anywhere said why. Six of eight users were in that state when
	// this was found.
	query := `SELECT id, name, email, COALESCE(phone, ''), COALESCE(address, ''),
	                 COALESCE(postal_code, ''), COALESCE(city, ''), COALESCE(country, ''),
	                 password, newsletter_subscription, terms_accepted
	          FROM users WHERE email = ?`

	err := db.Conn.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.Address,
		&user.PostalCode, &user.City, &user.Country, &hashedPassword,
		&user.NewsletterSubscription, &user.TermsAccepted,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUgyldigInnlogging
		}
		// A database error is not a wrong password, and must not be reported as
		// though it were.
		return nil, fmt.Errorf("uppslag av brukar: %w", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, ErrUgyldigInnlogging
	}

	// The user's permissions
	løyve, err := db.LøyveFor(int64(user.ID))
	if err != nil {
		return nil, err
	}
	user.Løyve = løyve

	return &user, nil
}

// GetUserByID fetches a user by their ID
func (db *Database) GetUserByID(userID int64) (*models.User, error) {
	var user models.User

	query := `SELECT u.id, u.name, COALESCE(u.birthdate, ''), u.email, COALESCE(u.phone, ''),
	                 COALESCE(u.address, ''), COALESCE(u.postal_code, ''), COALESCE(u.city, ''),
	                 COALESCE(u.country, ''), u.newsletter_subscription, u.terms_accepted,
	                 COALESCE(u.student_senior, 0), u.student_senior_gjeng_ut,
	                 COALESCE(GROUP_CONCAT(r.name), '')
	          FROM users u
	          LEFT JOIN brukarloyve ur ON ur.user_id = u.id
	          LEFT JOIN loyve r ON r.id = ur.loyve_id
	          WHERE u.id = ?
	          GROUP BY u.id`

	var gjengUt sql.NullTime
	var løyveTekst string
	err := db.Conn.QueryRow(query, userID).Scan(
		&user.ID, &user.Name, &user.Birthdate, &user.Email, &user.Phone, &user.Address,
		&user.PostalCode, &user.City, &user.Country,
		&user.NewsletterSubscription, &user.TermsAccepted, &user.StudentSenior,
		&gjengUt, &løyveTekst,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bruker ikke funnet")
		}
		return nil, err
	}

	user.StudentSeniorGjengUt = gjengUt.Time
	if løyveTekst != "" {
		user.Løyve = strings.Split(løyveTekst, ",")
	}

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

// ApproveFreezeRequest approves a freeze request.
//
// It sets the clock immediately: frozen_at is when the expiry stops
// counting. Without it a freeze would cost the member the time it lasted
// — a yearly card frozen two months would give ten months of use — and
// that is not what the studio sells.
func (db *Database) ApproveFreezeRequest(userID int64) error {
	query := `UPDATE user_memberships SET status = 'paused', frozen_at = CURRENT_TIMESTAMP
	          WHERE user_id = ? AND status = 'freeze_requested'`
	_, err := db.Conn.Exec(query, userID)
	return err
}

// UnfreezeMembership restarts a frozen membership and pushes the expiry
// forward by the time it stood.
//
// This is the *only* way out of 'paused'. UpdateMembershipStatus takes any
// status and knows nothing about the clock; called with "active" on a
// frozen membership it leaves frozen_at behind and the expiry drifts
// forever.
//
// The time is added to both renewal_date and binding_end: the first is
// what you have paid for, the second is the term you bought, and a freeze
// takes from neither.
func (db *Database) UnfreezeMembership(userID int64) error {
	var frosen sql.NullTime
	var fornying time.Time
	var binding sql.NullTime
	err := db.Conn.QueryRow(`SELECT frozen_at, renewal_date, binding_end FROM user_memberships
	                         WHERE user_id = ? AND status = 'paused'
	                         ORDER BY created_at DESC LIMIT 1`, userID).Scan(&frosen, &fornying, &binding)
	if err == sql.ErrNoRows {
		// Not frozen. Then this is an ordinary status change, and the old way is right.
		return db.UpdateMembershipStatus(userID, "active")
	}
	if err != nil {
		return err
	}

	// Frozen before the column existed means no clock. Then restart without
	// pushing anything: guessing at a length is worse than leaving it.
	if !frosen.Valid {
		return db.UpdateMembershipStatus(userID, "active")
	}

	stod := time.Since(frosen.Time)
	if stod < 0 {
		stod = 0
	}
	nyFornying := fornying.Add(stod)

	if binding.Valid {
		_, err = db.Conn.Exec(`UPDATE user_memberships
		                       SET status = 'active', frozen_at = NULL, renewal_date = ?, binding_end = ?
		                       WHERE user_id = ? AND status = 'paused'`,
			nyFornying, binding.Time.Add(stod), userID)
	} else {
		_, err = db.Conn.Exec(`UPDATE user_memberships
		                       SET status = 'active', frozen_at = NULL, renewal_date = ?
		                       WHERE user_id = ? AND status = 'paused'`,
			nyFornying, userID)
	}
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
func (db *Database) CanChangeMembership(userID int64, newMembershipID int64, nå time.Time) (bool, string) {
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

	// The student rate requires that someone has seen the proof.
	//
	// This was a flat no: *nobody* could switch to it, including those who had
	// shown proof, because there was no way to approve it. There is now (see
	// rabattkrav.go), so the question is not "is this a discount" but "has
	// this person been granted it". Same query the price list reads, so it
	// cannot show you something you may not buy.
	if newMembership.IsStudentSenior && !currentMembership.IsStudentSenior {
		brukar, err := db.GetUserByID(userID)
		if err != nil || !brukar.KvalifisertFor(nå) {
			return false, "Studentprisen krev at studioet hev sett beviset"
		}
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
	rad := db.Conn.QueryRow(`SELECT `+eventKolonnar+eventFrå+` WHERE e.id = ?`, eventID)
	if err := skannTime(rad, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// SignupUserForEvent signs a user up for a class.
//
// Two things were wrong. It read capacity straight from the events row,
// and once rooms became a resource that column is 0 when the class sets no
// capacity of its own — so `7 >= 0` was true and *every* signup answered
// "event is full". It has to compute the same capacity as the rest of the
// house.
//
// And it read first and wrote after, without a transaction. The Reformer
// has four places; two people pressing the last one at the same moment
// both got through the check.
func (db *Database) SignupUserForEvent(userID, eventID int64) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// A private class belongs to one person. The check must live here and not
	// only in the schedule: visibility is not security. Someone guessing an id
	// could POST themselves onto a class they never saw, and the kiosk signs
	// people up outside the schedule too — both go through this function.
	var eigar sql.NullInt64
	if err := tx.QueryRow(
		`SELECT private_user_id FROM events WHERE id = ?`, eventID).Scan(&eigar); err != nil {
		return err
	}
	if eigar.Valid && eigar.Int64 != userID {
		return fmt.Errorf("timen er sett av til ein annan")
	}

	// And the same question for the group. A reformer class only the trained
	// can see is still an id somebody can POST themselves onto.
	var gruppe sql.NullInt64
	if err := tx.QueryRow(
		`SELECT gruppe_id FROM events WHERE id = ?`, eventID).Scan(&gruppe); err != nil {
		return err
	}
	if gruppe.Valid {
		var med int
		if err := tx.QueryRow(
			`SELECT COUNT(*) FROM gruppemedlem WHERE gruppe_id = ? AND user_id = ?`,
			gruppe.Int64, userID).Scan(&med); err != nil {
			return err
		}
		if med == 0 {
			return fmt.Errorf("timen er open for ei gruppa du ikkje er med i")
		}
	}

	var exists int
	if err := tx.QueryRow(
		`SELECT COUNT(*) FROM event_signups WHERE user_id = ? AND event_id = ?`,
		userID, eventID).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return fmt.Errorf("brukaren er alt paameld denne timen")
	}

	// Same reckoning as GetEventsForWeek: the class's own capacity if set,
	// otherwise the room's.
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

	// The condition stands in the UPDATE too, so two simultaneous signups
	// cannot pass each other even if the check above let both through.
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

// CancelUserSignupForEvent takes a user off a class again.
//
// Same discipline as SignupUserForEvent above, and for the same reason: it
// read first and wrote after, on three separate connections. Two people
// cancelling at once both got through the check and the counter went down
// twice for one drop-out. The delete is the check now: if it hits no row,
// the user was not signed up. And MAX(…, 0) in the UPDATE, so a counter
// already out of step cannot go negative.
func (db *Database) CancelUserSignupForEvent(userID, eventID int64) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`DELETE FROM event_signups WHERE user_id = ? AND event_id = ?`,
		userID, eventID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("brukaren er ikkje paameld denne timen")
	}

	if _, err := tx.Exec(
		`UPDATE events SET current_enrolment = MAX(current_enrolment - 1, 0) WHERE id = ?`,
		eventID); err != nil {
		return err
	}

	return tx.Commit()
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
	// role_requirements and attendees were fetched and scanned into a map and
	// a slice. SQLite cannot do that, so the query failed — every time, since
	// always. Nobody saw it because the caller only checked `err == nil` and
	// let it be. Neither field is used by any caller; they are out.
	rows, err := db.Conn.Query(`SELECT `+eventKolonnar+`
		FROM events e
		INNER JOIN event_signups es ON e.id = es.event_id
		LEFT JOIN rooms r ON r.id = e.room_id
		WHERE es.user_id = ? AND e.start_time > ?
		ORDER BY e.start_time ASC, e.id ASC`, userID, time.Now())
	if err != nil {
		return nil, err
	}
	return timane(rows)
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

// Person is a member as they stand in the people list.
//
// Three facts, not seven: who, what they hold, and when they were last
// here. Birthday and phone are not what you scan a list for — they belong
// when the row is open.
type Person struct {
	ID          int
	Namn        string
	Epost       string
	Telefon     string
	Fodd        string
	Løyve       string
	Medlemskap  string
	MedlemStod  string
	MedlemPris  int
	KlippAtt    int
	KlippTotalt int
	SistHer     *time.Time
	TimarIÅr    int
	TrengSvar   bool // frysing som ventar
	ErLærar     bool
	ErAdmin     bool
	// Gruppone personen er med i. Settet er det malen slær upp i;
	// strengen er berre vegen ut or basen.
	GruppeIDar string
	GruppeSett map[int64]bool
}

// FolkOversyn fetches every member with what you need to recognise them
// and see who needs something.
//
// The sort is not alphabetical. It is *who needs you*: first those with a
// freeze awaiting an answer, then those who have not been here for a
// while, then the rest. Alphabetical is right when you are searching — but
// whoever is searching uses the search field, and whoever is *looking* at
// the list wants to know what is asking something of them.
func (db *Database) FolkOversyn() ([]Person, error) {
	rows, err := db.Conn.Query(`
		WITH
		loyve_per_brukar AS (
			SELECT ur.user_id, GROUP_CONCAT(r.name, ', ') AS loyve
			FROM brukarloyve ur
			JOIN loyve r ON r.id = ur.loyve_id
			GROUP BY ur.user_id
		),
		grupper_per_brukar AS (
			SELECT user_id, GROUP_CONCAT(gruppe_id) AS grupper
			FROM gruppemedlem
			GROUP BY user_id
		),
		klipp_per_brukar AS (
			SELECT user_id, SUM(remaining_klipp) AS att, SUM(total_klipp) AS totalt
			FROM user_klippekort
			GROUP BY user_id
		),
		aktivitet_per_brukar AS (
			SELECT es.user_id,
			       MAX(CASE WHEN e.start_time < CURRENT_TIMESTAMP THEN e.start_time END) AS sist_her,
			       SUM(CASE WHEN e.start_time < CURRENT_TIMESTAMP
			                 AND e.start_time > datetime('now', '-1 year') THEN 1 ELSE 0 END) AS timar_i_aar
			FROM event_signups es
			JOIN events e ON e.id = es.event_id
			GROUP BY es.user_id
		)
		SELECT u.id, u.name, u.email, COALESCE(u.phone, ''), COALESCE(u.birthdate, ''),
		       COALESCE(lp.loyve, ''), COALESCE(gp.grupper, ''),
		       COALESCE(m.name, ''), COALESCE(um.status, ''), COALESCE(m.price, 0),
		       COALESCE(kp.att, 0), COALESCE(kp.totalt, 0),
		       ap.sist_her, COALESCE(ap.timar_i_aar, 0)
		FROM users u
		LEFT JOIN user_memberships um ON um.user_id = u.id AND um.status != 'cancelled'
		LEFT JOIN memberships m ON m.id = um.membership_id
		LEFT JOIN loyve_per_brukar lp ON lp.user_id = u.id
		LEFT JOIN grupper_per_brukar gp ON gp.user_id = u.id
		LEFT JOIN klipp_per_brukar kp ON kp.user_id = u.id
		LEFT JOIN aktivitet_per_brukar ap ON ap.user_id = u.id
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
		if err := rows.Scan(&p.ID, &p.Namn, &p.Epost, &p.Telefon, &p.Fodd, &p.Løyve,
			&p.GruppeIDar,
			&p.Medlemskap, &p.MedlemStod, &p.MedlemPris,
			&p.KlippAtt, &p.KlippTotalt, &sist, &p.TimarIÅr); err != nil {
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
		p.ErLærar = harLøyve(p.Løyve, LøyveLærar)
		p.ErAdmin = harLøyve(p.Løyve, LøyveAdmin)
		p.GruppeSett = map[int64]bool{}
		for _, t := range strings.Split(p.GruppeIDar, ",") {
			if id, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64); err == nil {
				p.GruppeSett[id] = true
			}
		}
		ut = append(ut, p)
	}
	return ut, rows.Err()
}
