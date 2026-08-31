package database

import (
	"database/sql"
	"log"
	"time"
)

// The discount claim.
//
// A student card is something the user tells you, and the studio sees it at
// the desk. The checkbox in the profile set the discount directly — or
// rather: it set nothing, because the profile handler never read the field
// at all. Both halves are wrong the same way: the surface gave an answer it
// had no grounds for.
//
// Now the checkbox is a *claim*. It stands waiting until somebody at the
// studio has seen the card, and only then does the discount apply. And it
// expires, because a student card does: three years is the length of a
// bachelor's, and anyone still studying shows the card again.
//
// Senior is not here. Age follows from the birth date — a number the system
// already has — and nobody should ask permission to have turned 67.
const RabattÅr = 3

// The states a claim can be in.
const (
	RabattVentar   = "ventar"
	RabattGodkjend = "godkjend"
	RabattAvvist   = "avvist"
)

// Rabattkrav is one claim, with whoever asked for it.
type Rabattkrav struct {
	ID      int64
	UserID  int64
	Namn    string
	Epost   string
	Stoda   string
	BedeUm  time.Time
	Avgjord time.Time
	GjengUt time.Time
}

// Ventar says whether the claim is still waiting for an answer.
func (r Rabattkrav) Ventar() bool { return r.Stoda == RabattVentar }

// Godkjend says whether the claim was accepted.
func (r Rabattkrav) Godkjend() bool { return r.Stoda == RabattGodkjend }

// migrerRabattkrav lagar tabellen og kolonna som ber utgangsdatoen.
func migrerRabattkrav(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS rabattkrav (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id),
		stoda TEXT NOT NULL DEFAULT 'ventar',
		bede_um DATETIME NOT NULL,
		avgjord DATETIME
	)`); err != nil {
		return err
	}

	// The expiry date lives on the user and not on the claim: it is the user
	// 	// who is a student, and it is the user we ask when the prices are shown.
	// 	// The claim is the story of how they became one.
	var finst bool
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='student_senior_gjeng_ut'").
		Scan(&finst); err == nil && !finst {
		if _, err := db.Exec("ALTER TABLE users ADD COLUMN student_senior_gjeng_ut DATETIME"); err != nil {
			return err
		}
		log.Println("La til student_senior_gjeng_ut paa users.")
	}
	return nil
}

// LagRabattkrav puts a claim on hold. If the user already has one waiting,
// it stays — ticking twice is not two claims.
func (db *Database) LagRabattkrav(userID int64, nå time.Time) error {
	var finst int
	if err := db.Conn.QueryRow(
		"SELECT COUNT(*) FROM rabattkrav WHERE user_id = ? AND stoda = ?",
		userID, RabattVentar).Scan(&finst); err != nil {
		return err
	}
	if finst > 0 {
		return nil
	}
	_, err := db.Conn.Exec(
		"INSERT INTO rabattkrav (user_id, stoda, bede_um) VALUES (?, ?, ?)",
		userID, RabattVentar, veggtekst(nå))
	return err
}

// TrekkRabattkrav withdraws what is waiting, and takes the discount with
// it if it had already been granted. Unticking is saying "I am not a
// student any more", and then the price should not stand.
func (db *Database) TrekkRabattkrav(userID int64) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM rabattkrav WHERE user_id = ? AND stoda = ?",
		userID, RabattVentar); err != nil {
		return err
	}
	if _, err := tx.Exec(
		"UPDATE users SET student_senior = 0, student_senior_gjeng_ut = NULL WHERE id = ?",
		userID); err != nil {
		return err
	}
	return tx.Commit()
}

// SisteRabattkrav gives the user's most recent claim, if any.
func (db *Database) SisteRabattkrav(userID int64) (*Rabattkrav, error) {
	var k Rabattkrav
	var avgjord sql.NullTime
	err := db.Conn.QueryRow(`
		SELECT id, user_id, stoda, bede_um, avgjord
		FROM rabattkrav WHERE user_id = ?
		ORDER BY bede_um DESC, id DESC LIMIT 1`, userID).
		Scan(&k.ID, &k.UserID, &k.Stoda, &k.BedeUm, &avgjord)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	k.Avgjord = avgjord.Time
	return &k, nil
}

// VentandeRabattkrav gives those waiting for an answer, oldest first —
// whoever has waited longest has waited longest.
func (db *Database) VentandeRabattkrav() ([]Rabattkrav, error) {
	rows, err := db.Conn.Query(`
		SELECT k.id, k.user_id, u.name, u.email, k.stoda, k.bede_um
		FROM rabattkrav k JOIN users u ON u.id = k.user_id
		WHERE k.stoda = ? ORDER BY k.bede_um`, RabattVentar)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []Rabattkrav
	for rows.Next() {
		var k Rabattkrav
		if err := rows.Scan(&k.ID, &k.UserID, &k.Namn, &k.Epost, &k.Stoda, &k.BedeUm); err != nil {
			return nil, err
		}
		ut = append(ut, k)
	}
	return ut, rows.Err()
}

// AvgjerRabattkrav svarar på eit krav.
//
// Godkjenning og utgangsdato gjeng i same økti som stoda på kravet:
// gjekk dei kvar for seg og den andre feila, hadde kravet stade som
// godkjent medan brukaren ikkje hadde rabatten — og ingen hadde sett det,
// av di kravet var burte or køen.
func (db *Database) AvgjerRabattkrav(kravID int64, godkjend bool, nå time.Time) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int64
	if err := tx.QueryRow("SELECT user_id FROM rabattkrav WHERE id = ? AND stoda = ?",
		kravID, RabattVentar).Scan(&userID); err != nil {
		return err
	}

	stoda := RabattAvvist
	if godkjend {
		stoda = RabattGodkjend
	}
	if _, err := tx.Exec("UPDATE rabattkrav SET stoda = ?, avgjord = ? WHERE id = ?",
		stoda, veggtekst(nå), kravID); err != nil {
		return err
	}

	if godkjend {
		ut := nå.AddDate(RabattÅr, 0, 0)
		if _, err := tx.Exec(
			"UPDATE users SET student_senior = 1, student_senior_gjeng_ut = ? WHERE id = ?",
			veggtekst(ut), userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// StudentrabattFor gives the state of the user's discount: whether they
// have it, and for how much longer.
func (db *Database) StudentrabattFor(userID int64) (bool, time.Time, error) {
	var har bool
	var ut sql.NullTime
	err := db.Conn.QueryRow(
		"SELECT COALESCE(student_senior, 0), student_senior_gjeng_ut FROM users WHERE id = ?",
		userID).Scan(&har, &ut)
	if err != nil {
		return false, time.Time{}, err
	}
	return har, ut.Time, nil
}
