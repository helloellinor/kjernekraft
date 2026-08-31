package database

import (
	"database/sql"
	"strings"
	"time"
)

// The group.
//
// A class can be open to some and not to all. Reformer is the example: the
// machines are few, people have to be trained before they may use them, and
// the class should therefore only show for those who are.
//
// Until now there was only private_user_id — one class, one person. That is
// right for PT and wrong for this: opening nineteen reformer classes to
// twelve people would be two hundred and twenty choices made by hand, and
// one forgotten is a person who did not get to the class. The group is that
// decision taken once.
//
// A group is *not* a permission. `roles` in this database is admin and
// teacher — what you may *do* with the studio. A group is who you belong
// to. Mix them, and "give access to reformer" and "make an administrator"
// become the same action with the same form, and the day someone misclicks
// the difference is large.

// MigrerGrupper set upp gruppone og kolonna som knyter ein time til ei.
func MigrerGrupper(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS grupper (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		namn TEXT NOT NULL UNIQUE
	)`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS gruppemedlem (
		gruppe_id INTEGER NOT NULL REFERENCES grupper(id) ON DELETE CASCADE,
		user_id   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		PRIMARY KEY (gruppe_id, user_id)
	)`); err != nil {
		return err
	}

	var finst bool
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='gruppe_id'").
		Scan(&finst); err != nil {
		return err
	}
	if !finst {
		if _, err := db.Exec("ALTER TABLE events ADD COLUMN gruppe_id INTEGER"); err != nil {
			return err
		}
	}
	// Timeplanen spør etter dette for kvar veke ein teiknar.
	if _, err := db.Exec(
		"CREATE INDEX IF NOT EXISTS idx_events_gruppe ON events(gruppe_id)"); err != nil {
		return err
	}
	_, err := db.Exec(
		"CREATE INDEX IF NOT EXISTS idx_gruppemedlem_user ON gruppemedlem(user_id)")
	return err
}

// Gruppe er ei gruppe, med talet på medlemer i seg.
type Gruppe struct {
	ID     int64
	Namn   string
	Medlem int
}

// Grupper gjev alle gruppone, med medlemstalet.
func (db *Database) Grupper() ([]Gruppe, error) {
	rows, err := db.Conn.Query(`
		SELECT g.id, g.namn, COUNT(m.user_id)
		FROM grupper g LEFT JOIN gruppemedlem m ON m.gruppe_id = g.id
		GROUP BY g.id, g.namn ORDER BY g.namn`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []Gruppe
	for rows.Next() {
		var g Gruppe
		if err := rows.Scan(&g.ID, &g.Namn, &g.Medlem); err != nil {
			return nil, err
		}
		ut = append(ut, g)
	}
	return ut, rows.Err()
}

// LagGruppe lagar ei gruppe. Same namnet to gonger er den same gruppa.
func (db *Database) LagGruppe(namn string) (int64, error) {
	namn = strings.TrimSpace(namn)
	if namn == "" {
		return 0, sql.ErrNoRows
	}
	res, err := db.Conn.Exec("INSERT OR IGNORE INTO grupper (namn) VALUES (?)", namn)
	if err != nil {
		return 0, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var id int64
		err := db.Conn.QueryRow("SELECT id FROM grupper WHERE namn = ?", namn).Scan(&id)
		return id, err
	}
	return res.LastInsertId()
}

// SlettGruppe removes the group.
//
// The classes that pointed at it become open to everyone again, rather than
// standing pointing at something that does not exist — a class nobody can
// see because its group was deleted is a class that is gone without anyone
// saying so.
func (db *Database) SlettGruppe(gruppeID int64) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE events SET gruppe_id = NULL WHERE gruppe_id = ?", gruppeID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM gruppemedlem WHERE gruppe_id = ?", gruppeID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM grupper WHERE id = ?", gruppeID); err != nil {
		return err
	}
	return tx.Commit()
}

// SettGruppemedlem switches membership off or on. It survives being taken
// twice the same way, like the permission switch beside it.
func (db *Database) SettGruppemedlem(gruppeID, userID int64, på bool) error {
	if !på {
		_, err := db.Conn.Exec(
			"DELETE FROM gruppemedlem WHERE gruppe_id = ? AND user_id = ?", gruppeID, userID)
		return err
	}
	_, err := db.Conn.Exec(
		"INSERT OR IGNORE INTO gruppemedlem (gruppe_id, user_id) VALUES (?, ?)", gruppeID, userID)
	return err
}

// SetSerieGruppe knyter alle komande timar i serien til ei gruppe.
// Null opnar dei for alle att.
func (db *Database) SetSerieGruppe(serieID, gruppeID int64, frå time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET gruppe_id = NULLIF(?, 0) WHERE serie_id = ? AND end_time > ?",
		gruppeID, serieID, veggtekst(frå))
	return err
}

// KannSjåTimen says whether the user may see and sign up for the class.
//
// The same condition as synlegFor, written once more in Go because signup
// does not go through a schedule query. Let the two drift and it leaks:
// visibility is not security.
func (db *Database) KannSjåTimen(eventID, userID int64) (bool, error) {
	var eigar, gruppe sql.NullInt64
	if err := db.Conn.QueryRow(
		"SELECT private_user_id, gruppe_id FROM events WHERE id = ?", eventID).
		Scan(&eigar, &gruppe); err != nil {
		return false, err
	}
	if eigar.Valid && eigar.Int64 != userID {
		return false, nil
	}
	if !gruppe.Valid {
		return true, nil
	}
	var med int
	if err := db.Conn.QueryRow(
		"SELECT COUNT(*) FROM gruppemedlem WHERE gruppe_id = ? AND user_id = ?",
		gruppe.Int64, userID).Scan(&med); err != nil {
		return false, err
	}
	return med > 0, nil
}

// SettGruppePåTime ties one class to a group. Null opens it again.
func (db *Database) SettGruppePåTime(eventID, gruppeID int64) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET gruppe_id = NULLIF(?, 0) WHERE id = ?", gruppeID, eventID)
	return err
}

// GruppeFinst says whether the group exists. A class pointing at a group
// that does not is a class nobody can see.
func (db *Database) GruppeFinst(gruppeID int64) (bool, error) {
	var n int
	err := db.Conn.QueryRow("SELECT COUNT(*) FROM grupper WHERE id = ?", gruppeID).Scan(&n)
	return n > 0, err
}
