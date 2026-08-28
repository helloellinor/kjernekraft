package database

import (
	"database/sql"
	"log"
	"time"
)

// Rabattkravet.
//
// Studentbeviset er noko brukaren fortel, og studioet ser det i
// resepsjonen. Avkryssingi i profilen sette rabatten beinveges — eller
// rettare: ho sette ingen ting, av di profilhandsamaren aldri las feltet
// i det heile. Baae delar er gale paa same viset: flata gav eit svar ho
// ikkje hadde grunnlag for.
//
// No er avkryssingi eit *krav*. Det stend og ventar til nokon i studioet
// hev sett beviset, og fyrst daa gjeld rabatten. Og han gjeng ut att, av
// di eit studentbevis gjer det: tri aar er lengdi paa ein bachelor, og
// den som framleis studerer viser beviset ein gong til.
//
// Honnør er ikkje her. Alderen kjem av fødselsdagen — eit tal systemet
// alt hev — og ingen skal be um løyve til aa ha vorte 67.
const RabattAar = 3

// Stodone eit krav kann staa i.
const (
	RabattVentar   = "ventar"
	RabattGodkjend = "godkjend"
	RabattAvvist   = "avvist"
)

// Rabattkrav er eitt krav, med den som bad um det.
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

// Ventar segjer um kravet framleis ligg og ventar paa svar.
func (r Rabattkrav) Ventar() bool { return r.Stoda == RabattVentar }

// Godkjend segjer um kravet vart teke imot.
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

	// Utgangsdatoen bur paa brukaren og ikkje paa kravet: det er
	// brukaren som er student, og det er han me spør naar prisane skal
	// syna seg. Kravet er soga um korleis han vart det.
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

// LagRabattkrav set eit krav paa vent. Hev brukaren eit som ventar frå
// fyrr, vert det staaande — å krysse av to gonger er ikkje to krav.
func (db *Database) LagRabattkrav(userID int64, naa time.Time) error {
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
		userID, RabattVentar, veggtekst(naa))
	return err
}

// TrekkRabattkrav tek attende det som ventar, og tek rabatten med um han
// alt var gjeven. Krysset av seg sjølv er å seia «eg er ikkje student
// lenger», og då skal ikkje prisen stå att.
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

// SisteRabattkrav gjev det ferskaste kravet brukaren hev, um noko.
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

// VentandeRabattkrav gjev dei som ventar paa svar, eldste fyrst — den
// som hev venta lengst hev venta lengst.
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

// AvgjerRabattkrav svarar paa eit krav.
//
// Godkjenning og utgangsdato gjeng i same økti som stoda paa kravet:
// gjekk dei kvar for seg og den andre feila, hadde kravet stade som
// godkjent medan brukaren ikkje hadde rabatten — og ingen hadde sett det,
// av di kravet var burte or køen.
func (db *Database) AvgjerRabattkrav(kravID int64, godkjend bool, naa time.Time) error {
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
		stoda, veggtekst(naa), kravID); err != nil {
		return err
	}

	if godkjend {
		ut := naa.AddDate(RabattAar, 0, 0)
		if _, err := tx.Exec(
			"UPDATE users SET student_senior = 1, student_senior_gjeng_ut = ? WHERE id = ?",
			veggtekst(ut), userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// StudentrabattFor gjev stoda paa rabatten til brukaren: om han hev han,
// og kor lenge til.
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
