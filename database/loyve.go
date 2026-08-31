package database

import "strings"

// Permissions are the only reasons a user is anything more than a student.
// The names are English because "admin" and "user" were already in the
// database when this file arrived; a Norwegian permission in the same
// column would be worse than two English ones. The visible word comes
// through {{t}}, not from here.
const (
	LøyveAdmin = "admin"
	LøyveLærar = "teacher"
)

// LøyveFinst says whether the name is a permission the system knows.
// Without it a call could write anything into the roles table, and then the
// permission list is no longer a list of anything.
//
// The developer permission is *not* here, deliberately. It is not granted
// by anyone: it is read from a file on the server (see utviklar.go).
// LøyveFinst is the gate the admin surface writes through, so a permission
// listed here is one an admin can give themselves. The developer permission
// gives free run of the house; it should come from whoever owns the
// machine, not whoever owns a button.
func LøyveFinst(løyvet string) bool {
	return løyvet == LøyveAdmin || løyvet == LøyveLærar
}

// SettLøyve turns a permission on or off for a user.
//
// GjevLøyve only does the first, and fails on a key collision when the
// permission is already there. A button you can press twice needs both
// directions, and needs the first direction to survive being taken
// twice.
func (db *Database) SettLøyve(userID int64, løyvet string, på bool) error {
	if !på {
		_, err := db.Conn.Exec(`
			DELETE FROM brukarloyve
			WHERE user_id = ? AND loyve_id = (SELECT id FROM loyve WHERE name = ?)`,
			userID, løyvet)
		return err
	}

	løyveID, err := db.LøyveIDFor(løyvet)
	if err != nil {
		return err
	}
	if _, err := db.Conn.Exec(
		"INSERT OR IGNORE INTO brukarloyve (user_id, loyve_id) VALUES (?, ?)",
		userID, løyveID); err != nil {
		return err
	}

	// Forfremjingi skal gjelda med ein gong, og ho skal gjelda pengane
	// med. Sjå SynkFriMedlemskap i svartmedlem.go.
	_, err = db.SynkFriMedlemskap(userID)
	return err
}

// HarLøyve svarar for ein einskild brukar.
func (db *Database) HarLøyve(userID int64, løyvet string) (bool, error) {
	var tal int
	err := db.Conn.QueryRow(`
		SELECT COUNT(*) FROM brukarloyve ur
		JOIN loyve r ON r.id = ur.loyve_id
		WHERE ur.user_id = ? AND r.name = ?`, userID, løyvet).Scan(&tal)
	return tal > 0, err
}

// LærarNamn gives the names of those with the teacher permission.
//
// The selects used to ask GetDistinctTeachers, which reads SELECT DISTINCT
// teacher_name FROM events. That list is the *history* — who actually held
// a class — and it grows with every typo anyone ever made in the field. It
// is still right in the schedule filter, where you filter over classes that
// have been. It is not right in a select, where you point at someone who
// will hold a class that has not.
func (db *Database) LærarNamn() ([]string, error) {
	rows, err := db.Conn.Query(`
		SELECT u.name FROM users u
		JOIN brukarloyve ur ON ur.user_id = u.id
		JOIN loyve r ON r.id = ur.loyve_id
		WHERE r.name = ?
		ORDER BY u.name COLLATE NOCASE`, LøyveLærar)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []string
	for rows.Next() {
		var namn string
		if err := rows.Scan(&namn); err != nil {
			return nil, err
		}
		ut = append(ut, namn)
	}
	return ut, rows.Err()
}

// harLøyve les den samanslegne løyvestrengen FolkOversyn hentar, so
// lista slepp eit uppslag per person.
func harLøyve(løyve, løyvet string) bool {
	for _, r := range strings.Split(løyve, ",") {
		if strings.TrimSpace(r) == løyvet {
			return true
		}
	}
	return false
}
