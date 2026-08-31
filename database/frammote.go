package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"kjernekraft/models"
)

// Attendance.
//
// event_signups only said that you *had signed up*. That is not the same as
// being there, and the whole house knew it: the footnote under the
// attendance board said "attendance is not recorded yet", and the query
// behind the board carried a comment saying it should swap event_signups
// for attendance the day it existed.
//
// It exists now. attended_at is the moment somebody ticked you off in the
// lobby — NULL as long as nobody has.

// MigrerFrammote adds the column if it is not there.
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

// TimarIVindauget gives the classes the kiosk should show: those starting
// within `fyre`, and those running now that have not ended more than
// `etter` ago.
//
// The window is the whole of the privacy on this page. It is open — there
// is a screen in the lobby and nobody logs into it — so it must never be
// able to answer "who trains here". It answers "who is here now", and only
// while that is now.
func (db *Database) TimarIVindauget(nå time.Time, fyre, etter time.Duration) ([]models.Event, error) {
	// Capacity is included because the kiosk can now sign people *up*: it
	// 	// should only offer the search when there are places left. Same
	// 	// reckoning as in SignupUserForEvent — the class's own number if set,
	// 	// otherwise the room's.
	rows, err := db.Conn.Query(`
		SELECT e.id, e.title, e.start_time, e.end_time,
		       COALESCE(e.teacher_name, ''), COALESCE(r.name, e.location, ''),
		       e.current_enrolment, COALESCE(NULLIF(e.capacity, 0), r.capacity, 0),
		       COALESCE(e.class_type, '')
		FROM events e
		LEFT JOIN rooms r ON r.id = e.room_id
		WHERE e.start_time <= ? AND e.end_time > ?
		ORDER BY e.start_time ASC, e.id ASC`,
		veggtekst(nå.Add(fyre)), veggtekst(nå.Add(-etter)))
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

// Påmeld er ein person på lista til ein time, med eller utan kryss.
type Påmeld struct {
	UserID   int64
	Namn     string
	Frammott *time.Time // nil = ikkje kryssa av
}

// PaameldeTilTimar hentar dei påmelde til fleire timar i éi søkning.
// Innsjekkskjermen viser gjerne fleire timar samtidig; eitt kall per
// time gjorde då sidebygginga til N+1 søkningar.
func (db *Database) PaameldeTilTimar(eventIDs []int64) (map[int64][]Påmeld, error) {
	ut := make(map[int64][]Påmeld, len(eventIDs))
	if len(eventIDs) == 0 {
		return ut, nil
	}

	plasshaldarar := make([]string, len(eventIDs))
	args := make([]any, len(eventIDs))
	for i, id := range eventIDs {
		plasshaldarar[i] = "?"
		args[i] = id
		ut[id] = nil
	}

	rows, err := db.Conn.Query(fmt.Sprintf(`
		SELECT es.event_id, u.id, u.name, es.attended_at
		FROM event_signups es
		INNER JOIN users u ON u.id = es.user_id
		WHERE es.event_id IN (%s)
		ORDER BY es.event_id ASC, u.name COLLATE NOCASE ASC`, strings.Join(plasshaldarar, ", ")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var eventID int64
		var p Påmeld
		var når sql.NullTime
		if err := rows.Scan(&eventID, &p.UserID, &p.Namn, &når); err != nil {
			return nil, err
		}
		if når.Valid {
			t := når.Time
			p.Frammott = &t
		}
		ut[eventID] = append(ut[eventID], p)
	}
	return ut, rows.Err()
}

// MerkFrammote ticks someone off. It is *idempotent* deliberately: press
// twice and the first timestamp stands — the kiosk is a screen many people
// touch, and the second press is usually an accident.
func (db *Database) MerkFrammote(eventID, userID int64, nå time.Time) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		UPDATE event_signups SET attended_at = ?
		WHERE event_id = ? AND user_id = ? AND attended_at IS NULL`,
		veggtekst(nå), eventID, userID)
	if err != nil {
		return err
	}

	// attended_at IS NULL makes the tick one-shot. If we touched no row it was
	// 	// already ticked — and then the clip must *not* be taken again. Two
	// 	// presses on the same kiosk button is a thing that happens.
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
	if err := brukKlipp(tx, eventID, userID, nå); err != nil {
		return err
	}
	return tx.Commit()
}

// NettFrammott gjev timen brukaren *var på* og som slutta for mindre
// enn `innan` sidan — eller nil.
//
// Han krev kryss. Ei paamelding aaleine segjer at ho hadde tenkt seg
// dit, og «Takk for no» til ein som ikkje kom er ei helsing som lyg.
func (db *Database) NettFrammott(userID int64, nå time.Time, innan time.Duration) (bool, error) {
	var eitt int
	err := db.Conn.QueryRow(`
		SELECT 1
		FROM events e
		INNER JOIN event_signups es ON e.id = es.event_id
		WHERE es.user_id = ? AND es.attended_at IS NOT NULL
		  AND e.end_time <= ? AND e.end_time > ?
		LIMIT 1`,
		userID, veggtekst(nå), veggtekst(nå.Add(-innan))).Scan(&eitt)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// FjernFrammote tek krysset bort att: feil person vart kryssa av.
// Paameldingi stend — det er berre krysset som gjekk gale. Kryssar
// nokon same personen på nytt, fær rada eit nytt tidspunkt, og det er
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

	// Var krysset feil, var klippet det ogso. Det kjem attende på det
	// kortet det vart teke frå — difor stend kort-id-et på paameldingi.
	if err := gjevAttKlipp(tx, eventID, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// SokTilInnsjekk finds people by name, for drop-in at the kiosk: somebody
// is standing in the lobby without having booked, and there is room.
//
// The page is open, so this search is the one place the kiosk can say
// anything about people who are *not* here now. The limit and the minimum
// query length live in the handler; the rest of the caution lives here:
// names only, never anything more, and those already on the list are left
// out — they belong to the tick, not the search.
func (db *Database) SokTilInnsjekk(q string, eventID int64, grense int) ([]Påmeld, error) {
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

	var ut []Påmeld
	for rows.Next() {
		var p Påmeld
		if err := rows.Scan(&p.UserID, &p.Namn); err != nil {
			return nil, err
		}
		ut = append(ut, p)
	}
	return ut, rows.Err()
}
