package database

import (
	"database/sql"
	"strings"
	"time"

	"kjernekraft/models"
)

// Frammøtet.
//
// `event_signups` sa til no berre at du *hadde meldt deg paa*. Det er
// ikkje det same som at du var der, og heile huset visste det: fotnoten
// under oppmøtebrettet sa «Frammøte er ikkje registrert enno», og
// spurningen bak brettet ber ein kommentar um at han skal byta
// `event_signups` mot frammøtet den dagen det finst.
//
// Det finst no. `attended_at` er tidspunktet nokon kryssa av i
// vestibylen — NULL so lenge ingen hev gjort det.

// MigrerFrammote legg til kolonna um ho ikkje er der.
func MigrerFrammote(db *sql.DB) error {
	var finst bool
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('event_signups') WHERE name='attended_at'").
		Scan(&finst); err != nil {
		return err
	}
	if finst {
		return nil
	}
	_, err := db.Exec("ALTER TABLE event_signups ADD COLUMN attended_at DATETIME")
	return err
}

// TimarIVindauget gjev timane kiosken skal syna: dei som byrjar innan
// `fyre`, og dei som gjeng no og ikkje hev slutta for meir enn `etter`
// sidan.
//
// Vindauget er heile personvernet i denne sida. Ho er open — det stend
// ein skjerm i vestibylen og ingen loggar seg paa han — so ho skal aldri
// kunna svara paa «kven trenar her». Ho svarar paa «kven er her no», og
// berre so lenge det er no.
func (db *Database) TimarIVindauget(naa time.Time, fyre, etter time.Duration) ([]models.Event, error) {
	// Kapasiteten er med av di kiosken no kann melda folk *paa*: han
	// skal berre bjoda fram søket naar det finst ledige plassar. Same
	// utrekningi som i SignupUserForEvent — timen sitt eige tal um det
	// er sett, elles rommet sitt.
	rows, err := db.Conn.Query(`
		SELECT e.id, e.title, e.start_time, e.end_time,
		       COALESCE(e.teacher_name, ''), COALESCE(r.name, e.location, ''),
		       e.current_enrolment, COALESCE(NULLIF(e.capacity, 0), r.capacity, 0),
		       COALESCE(e.class_type, '')
		FROM events e
		LEFT JOIN rooms r ON r.id = e.room_id
		WHERE e.start_time <= ? AND e.end_time > ?
		ORDER BY e.start_time ASC`,
		veggtekst(naa.Add(fyre)), veggtekst(naa.Add(-etter)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.Title, &e.StartTime, &e.EndTime,
			&e.TeacherName, &e.RoomName, &e.CurrentEnrolment, &e.Capacity,
			&e.ClassType); err != nil {
			return nil, err
		}
		ut = append(ut, e)
	}
	return ut, rows.Err()
}

// Paameld er ein person paa lista til ein time, med eller utan kryss.
type Paameld struct {
	UserID   int64
	Namn     string
	Frammott *time.Time // nil = ikkje kryssa av
}

// PaameldeTil gjev lista kiosken kryssar av i.
func (db *Database) PaameldeTil(eventID int64) ([]Paameld, error) {
	rows, err := db.Conn.Query(`
		SELECT u.id, u.name, es.attended_at
		FROM event_signups es
		INNER JOIN users u ON u.id = es.user_id
		WHERE es.event_id = ?
		ORDER BY u.name COLLATE NOCASE ASC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []Paameld
	for rows.Next() {
		var p Paameld
		var naar sql.NullTime
		if err := rows.Scan(&p.UserID, &p.Namn, &naar); err != nil {
			return nil, err
		}
		if naar.Valid {
			t := naar.Time
			p.Frammott = &t
		}
		ut = append(ut, p)
	}
	return ut, rows.Err()
}

// MerkFrammote kryssar av. Han er *idempotent* med vilje: trykkjer nokon
// tvo gonger, stend det fyrste tidspunktet — kiosken er ein skjerm mange
// tek paa, og det andre trykket er som oftast eit uhell.
func (db *Database) MerkFrammote(eventID, userID int64, naa time.Time) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		UPDATE event_signups SET attended_at = ?
		WHERE event_id = ? AND user_id = ? AND attended_at IS NULL`,
		veggtekst(naa), eventID, userID)
	if err != nil {
		return err
	}

	// `attended_at IS NULL` gjer krysset eingongs. Rørde me ingi rad, var
	// ho alt kryssa av — og daa skal klippet *ikkje* takast ein gong til.
	// Tvo trykk paa den same knappen i kiosken er ein ting som hender.
	rørde, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rørde == 0 {
		return tx.Commit()
	}

	// Klippet gjeng i den same transaksjonen som krysset. Fell det eine,
	// fell det andre: eit kryss utan klipp er ein gratis time, og eit
	// klipp utan kryss er eit klipp ingen fekk noko for.
	if err := brukKlipp(tx, eventID, userID, naa); err != nil {
		return err
	}
	return tx.Commit()
}

// NettFrammott gjev timen brukaren *var paa* og som slutta for mindre
// enn `innan` sidan — eller nil.
//
// Han krev kryss. Ei paamelding aaleine segjer at ho hadde tenkt seg
// dit, og «Takk for no» til ein som ikkje kom er ei helsing som lyg.
func (db *Database) NettFrammott(userID int64, naa time.Time, innan time.Duration) (*models.Event, error) {
	var e models.Event
	err := db.Conn.QueryRow(`
		SELECT e.id, e.title, e.start_time, e.end_time
		FROM events e
		INNER JOIN event_signups es ON e.id = es.event_id
		WHERE es.user_id = ? AND es.attended_at IS NOT NULL
		  AND e.end_time <= ? AND e.end_time > ?
		ORDER BY e.end_time DESC
		LIMIT 1`,
		userID, veggtekst(naa), veggtekst(naa.Add(-innan))).
		Scan(&e.ID, &e.Title, &e.StartTime, &e.EndTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// FjernFrammote tek krysset bort att: feil person vart kryssa av.
// Paameldingi stend — det er berre krysset som gjekk gale. Kryssar
// nokon same personen paa nytt, fær rada eit nytt tidspunkt, og det er
// rett: det gamle var jo aldri sant.
func (db *Database) FjernFrammote(eventID, userID int64) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		UPDATE event_signups SET attended_at = NULL
		WHERE event_id = ? AND user_id = ? AND attended_at IS NOT NULL`,
		eventID, userID)
	if err != nil {
		return err
	}
	rørde, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rørde == 0 {
		return tx.Commit()
	}

	// Var krysset feil, var klippet det ogso. Det kjem attende paa det
	// kortet det vart teke fraa — difor stend kort-id-et paa paameldingi.
	if err := gjevAttKlipp(tx, eventID, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// SokTilInnsjekk finn folk etter namn, til drop-in i kiosken: nokon
// stend i vestibylen utan aa ha tinga plass, og det er plass att.
//
// Sida er open, so dette søket er den eine staden kiosken kann segja
// noko um folk som *ikkje* er her no. Grensa og minstelengdi paa
// spurninga ligg i handsamaren; her ligg resten av varsemdi: berre
// namn, aldri noko meir, og dei som alt stend paa lista er utelatne —
// dei høyrer til avkryssingi, ikkje søket.
func (db *Database) SokTilInnsjekk(q string, eventID int64, grense int) ([]Paameld, error) {
	monster := "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q) + "%"
	rows, err := db.Conn.Query(`
		SELECT u.id, u.name
		FROM users u
		WHERE u.name LIKE ? ESCAPE '\'
		  AND NOT EXISTS (SELECT 1 FROM event_signups es
		                  WHERE es.event_id = ? AND es.user_id = u.id)
		ORDER BY u.name COLLATE NOCASE ASC
		LIMIT ?`, monster, eventID, grense)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []Paameld
	for rows.Next() {
		var p Paameld
		if err := rows.Scan(&p.UserID, &p.Namn); err != nil {
			return nil, err
		}
		ut = append(ut, p)
	}
	return ut, rows.Err()
}

// VarFrammott segjer um brukaren er kryssa av paa ein time.
func (db *Database) VarFrammott(eventID, userID int64) (bool, error) {
	var n int
	err := db.Conn.QueryRow(`
		SELECT COUNT(*) FROM event_signups
		WHERE event_id = ? AND user_id = ? AND attended_at IS NOT NULL`,
		eventID, userID).Scan(&n)
	return n > 0, err
}
