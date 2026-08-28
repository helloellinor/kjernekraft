package database

import (
	"database/sql"
	"strings"
	"time"
)

// Gruppa.
//
// Ein time kann vera open for somme og ikkje for alle. Reformer er
// døme: apparati er faa, folk lyt vera lærde upp fyre dei fær bruka
// deim, og timen skal difor berre syna seg for dei som er det.
//
// Til no fanst berre `private_user_id` — éin time, éin person. Det er
// rett for PT og gale for dette: skulde nitten reformer-timar opnast
// for tolv personar, var det tvo hundre og tjuge val nokon laut gjera
// for haand, og eitt gløymt var ein person som ikkje kom paa timen.
// Gruppa er den same avgjerdi teki ein gong.
//
// Gruppa er *ikkje* eit løyve. `roles` i denne basen er `admin` og
// `teacher` — det er kva du fær *gjera* med studioet, altso løyve. Ei
// gruppe er kven du høyrer til. Blanda ein dei, vert «gjev tilgang til
// reformer» og «gjer til administrator» den same handlingi med det same
// skjemaet, og den dagen nokon feilklikkar er skilnaden stor.

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

// Gruppe er ei gruppe, med talet paa medlemer i seg.
type Gruppe struct {
	ID     int64
	Namn   string
	Medlem int
	ErMed  bool // om ein viss brukar er med — sett av GruppeneMed
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

// GruppeneMed gjev alle gruppone, og merkjer dei brukaren er med i.
func (db *Database) GruppeneMed(userID int64) ([]Gruppe, error) {
	rows, err := db.Conn.Query(`
		SELECT g.id, g.namn, COUNT(m.user_id),
		       EXISTS (SELECT 1 FROM gruppemedlem x
		               WHERE x.gruppe_id = g.id AND x.user_id = ?)
		FROM grupper g LEFT JOIN gruppemedlem m ON m.gruppe_id = g.id
		GROUP BY g.id, g.namn ORDER BY g.namn`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []Gruppe
	for rows.Next() {
		var g Gruppe
		if err := rows.Scan(&g.ID, &g.Namn, &g.Medlem, &g.ErMed); err != nil {
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

// SlettGruppe tek gruppa burt.
//
// Timane som peika paa henne vert opne for alle att, og ikkje ståande
// og peika paa noko som ikkje finst — ein time ingen kann sjaa av di
// gruppa hans er sletta, er ein time som er burte utan at nokon sa det.
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

// SettGruppemedlem slær medlemskapet av eller paa. Han toler aa gaast
// tvo gonger same vegen, som løyvebrytaren attmed.
func (db *Database) SettGruppemedlem(gruppeID, userID int64, paa bool) error {
	if !paa {
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
func (db *Database) SetSerieGruppe(serieID, gruppeID int64, fraa time.Time) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET gruppe_id = NULLIF(?, 0) WHERE serie_id = ? AND end_time > ?",
		gruppeID, serieID, veggtekst(fraa))
	return err
}

// KannSjaaTimen segjer om brukaren fær sjaa og melda seg paa timen.
//
// Same vilkåret som `synlegFor`, skrive ein gong til i Go av di
// paamelding ikkje gjeng gjenom ei timeplan-spurning. Held dei tvo
// fraa kvarandre, lek det: synlegheit er ikkje tryggleik.
func (db *Database) KannSjaaTimen(eventID, userID int64) (bool, error) {
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

// SettGruppePaaTime knyter éin time til ei gruppe. Null opnar honom att.
func (db *Database) SettGruppePaaTime(eventID, gruppeID int64) error {
	_, err := db.Conn.Exec(
		"UPDATE events SET gruppe_id = NULLIF(?, 0) WHERE id = ?", gruppeID, eventID)
	return err
}

// GruppeFinst segjer om gruppa finst. Ein time som peikar paa ei gruppe
// som ikkje finst, er ein time ingen kann sjaa.
func (db *Database) GruppeFinst(gruppeID int64) (bool, error) {
	var n int
	err := db.Conn.QueryRow("SELECT COUNT(*) FROM grupper WHERE id = ?", gruppeID).Scan(&n)
	return n > 0, err
}
