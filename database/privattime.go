package database

import (
	"database/sql"
	"time"

	"kjernekraft/models"
)

// The private class.
//
// Personal training is a class with one participant. It existed as a
// klippekort category and as something people bought, but there was no way
// to *set one up*: every class in the schedule was visible to everyone, so
// a PT session had to be either out there for the whole house to see or
// not there at all.
//
// private_user_id is the whole distinction: NULL is an ordinary class, set
// is a class only that one person sees and only that one person can sign
// up for.
//
// Enforced in two places, on purpose:
//
//   - *visibility* in the queries that build the schedule. If you cannot
//     see the class, you cannot click it.
//   - *access* in SignupUserForEvent. Visibility is not security — someone
//     guessing an id can POST themselves onto a class they never saw, and
//     the kiosk signs people up outside the schedule too. The check has to
//     live where the signup actually happens.

// MigrerPrivatTime adds the column if it is not there.
func MigrerPrivatTime(db *sql.DB) error {
	var finst bool
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='private_user_id'").
		Scan(&finst); err != nil {
		return err
	}
	if finst {
		return nil
	}
	if _, err := db.Exec("ALTER TABLE events ADD COLUMN private_user_id INTEGER"); err != nil {
		return err
	}
	// Timeplanen spør etter dette for kvar veke ein teiknar.
	_, err := db.Exec(
		"CREATE INDEX IF NOT EXISTS idx_events_private ON events(private_user_id)")
	return err
}

// synlegFor is the condition that lets through ordinary classes and the
// private ones belonging to this viewer. It takes the viewer's id.
//
// It is a constant rather than text written out by hand each time, because
// a schedule query that forgets it leaks a PT session to the whole house —
// and that is not something you notice by reading the list. Since groups
// arrived it covers two things: the class belonging to one person, and the
// class belonging to a group. Both are "who gets to see this", and a query
// that remembers one and forgets the other leaks just the same.
//
// It takes the viewer's id *twice*. Not pretty, but SQL cannot bind the
// same position in two places, and it is the safe failure to make: forget
// the second argument and the query fails with
const synlegFor = `(e.private_user_id IS NULL OR e.private_user_id = ?)
	AND (e.gruppe_id IS NULL OR EXISTS (
		SELECT 1 FROM gruppemedlem gm
		WHERE gm.gruppe_id = e.gruppe_id AND gm.user_id = ?))`

// SettPrivatTime gjer ein time privat for ein brukar, eller open att
// når userID er 0.
func (db *Database) SettPrivatTime(eventID, userID int64) error {
	if userID == 0 {
		_, err := db.Conn.Exec(
			`UPDATE events SET private_user_id = NULL WHERE id = ?`, eventID)
		return err
	}
	_, err := db.Conn.Exec(
		`UPDATE events SET private_user_id = ? WHERE id = ?`, userID, eventID)
	return err
}

// EventsSynlegeFor gjev alle timane `sjaaarID` hev lov aa sjå.
//
// GetAllEvents heiter framleis det han gjer og gjev alt — administrasjonen
// skal sjå PT-øktene med. Denne er den ein rettar mot brukaren.
func (db *Database) EventsSynlegeFor(sjaårID int64) ([]models.Event, error) {
	rows, err := db.Conn.Query(`SELECT `+eventKolonnar+eventFrå+`
		WHERE `+synlegFor+`
		ORDER BY e.start_time ASC, e.id ASC`, sjaårID, sjaårID)
	if err != nil {
		return nil, err
	}
	return timane(rows)
}

// PTKategori is the klippekort category a private session is paid with.
// It is a constant here because it ties three things together: the
// class_type a PT session gets when set up, the package the clip is taken
// from, and what the price is called in the list. Let them drift and the
// session costs nothing without anyone noticing.
const PTKategori = "Personlig Trening"

// BokPrivatTime signs the one person up for their session and takes the
// clip immediately.
//
// Otherwise the clip follows the check-in at the door rather than the
// signup — a class you did not get should not cost (see klippbruk.go). A
// PT session is not the same transaction: the class is set aside for one
// person, and the *appointment* is the goods. So it is clipped when it is
// set up.
//
// And it never goes negative. The door writes a debt when the card is
// empty — that class has already been given, so it stays in the accounts —
// but here it has not been given yet. With no clips the session is set up
// anyway and no debt written: it is an appointment to be settled another
// way, not a number to follow someone around.
//
// The clip is recorded on the signup (klipp_kort_id), so it returns by
// itself if the session is cancelled — and so the door sees it is already
// taken and does not take another.
func (db *Database) BokPrivatTime(eventID, userID int64, nå time.Time) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// The signup first: the clip is recorded on it, so it has to exist.
	// UNIQUE(user_id, event_id) makes OR IGNORE the whole check — if already
	// signed up, the signup stands as it is, and the clip below sees for
	// itself whether it is already taken.
	if _, err := tx.Exec(`
		INSERT OR IGNORE INTO event_signups (event_id, user_id, signup_date)
		VALUES (?, ?, ?)`, eventID, userID, veggtekst(nå)); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE events SET current_enrolment = (
			SELECT COUNT(*) FROM event_signups WHERE event_id = ?
		) WHERE id = ?`, eventID, eventID); err != nil {
		return err
	}

	// Kortet som gjeng ut fyrst vert nytta fyrst — same regelen som i
	// døri, og av same grunnen: eit kort med kort frist skal ikkje liggja
	// urørt og gaa ut medan klippi vert tekne av eit som varer lenger.
	// Er klippet alt teke for denne paameldingi, er det teke. Same
	// prøva som i døri, og av same grunnen.
	var alt sql.NullInt64
	if err := tx.QueryRow(`
		SELECT klipp_kort_id FROM event_signups
		WHERE event_id = ? AND user_id = ?`, eventID, userID).Scan(&alt); err == nil && alt.Valid {
		return tx.Commit()
	}

	var kortID int64
	err = tx.QueryRow(`
		SELECT uk.id
		FROM user_klippekort uk
		JOIN klippekort_packages kp ON kp.id = uk.package_id
		WHERE uk.user_id = ? AND kp.category = ?
		  AND uk.is_active = TRUE AND uk.remaining_klipp > 0
		  AND uk.expiry_date > ?
		ORDER BY uk.expiry_date ASC
		LIMIT 1`, userID, PTKategori, veggtekst(nå)).Scan(&kortID)
	if err == sql.ErrNoRows {
		// Ingen klipp. Økta stend; ingi skuld vert skrivi.
		return tx.Commit()
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
		UPDATE user_klippekort SET remaining_klipp = remaining_klipp - 1
		WHERE id = ? AND remaining_klipp > 0`, kortID); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE event_signups SET klipp_kort_id = ?, skulda = 0
		WHERE event_id = ? AND user_id = ?`, kortID, eventID, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// MigrerLaga adds the column saying who set the class up.
//
// teacher_name is free text — the name of whoever *holds* the class — and
// that is not the same question. When someone gives up a PT session the
// notice should go to whoever set it up, and a name is not a recipient:
// two people can share one, and a teacher need not be a user. Hence an id.
//
// Classes that predate the column have NULL, which is an honest answer:
// nobody knows. The notice then goes to the admins alone.
func MigrerLaga(db *sql.DB) error {
	var finst bool
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='created_by'").
		Scan(&finst); err != nil {
		return err
	}
	if finst {
		return nil
	}
	_, err := db.Exec("ALTER TABLE events ADD COLUMN created_by INTEGER REFERENCES users(id)")
	return err
}

// SettLagaAv skriv kven som sette timen upp.
func (db *Database) SettLagaAv(eventID, userID int64) error {
	_, err := db.Conn.Exec(`UPDATE events SET created_by = ? WHERE id = ?`, userID, eventID)
	return err
}

// ErPrivatTime says whether the class is set aside for one person.
//
// A query rather than a field on the model: private_user_id is a condition
// in every query that builds the schedule, not something read out with the
// class. Hanging it on the model would mean fetching it in twenty queries
// to read it in one.
func (db *Database) ErPrivatTime(eventID int64) (bool, error) {
	var eigar sql.NullInt64
	if err := db.Conn.QueryRow(
		`SELECT private_user_id FROM events WHERE id = ?`, eventID).Scan(&eigar); err != nil {
		return false, err
	}
	return eigar.Valid && eigar.Int64 > 0, nil
}

// AvlysPrivatTime is what happens when the one person gives up their
// session: the clip comes back, the signup goes, and the notice is written
// in the same transaction.
//
// The clip comes back because that is the rule the house already has — "if
// she cancels, or does not turn up, it costs nothing" (klippbruk.go). A PT
// session is clipped at booking rather than at the door, so without this it
// would be the one class in the house that cost something you did not get.
// If the notice is too short for that to be reasonable, that is a decision
// a human takes — and the notice is exactly what puts it in front of one.
//
// All in one transaction: a notice about a cancellation that then rolled
// back sends somebody off after a class that is still running.
func (db *Database) AvlysPrivatTime(eventID, userID int64, nå time.Time) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var tittel string
	var start string
	var laga sql.NullInt64
	if err := tx.QueryRow(
		`SELECT title, start_time, created_by FROM events WHERE id = ?`, eventID).
		Scan(&tittel, &start, &laga); err != nil {
		return err
	}
	timestart, _ := time.Parse("2006-01-02 15:04:05", start)

	// Klippet fyrst, medan paameldingi som ber det framleis finst.
	if err := gjevAttKlipp(tx, eventID, userID); err != nil {
		return err
	}

	res, err := tx.Exec(
		`DELETE FROM event_signups WHERE event_id = ? AND user_id = ?`, eventID, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	if _, err := tx.Exec(`
		UPDATE events SET current_enrolment = (
			SELECT COUNT(*) FROM event_signups WHERE event_id = ?
		) WHERE id = ?`, eventID, eventID); err != nil {
		return err
	}

	if err := db.LagMelding(tx, MeldingAvlystPrivat, eventID, userID, laga,
		tittel, timestart, nå); err != nil {
		return err
	}
	return tx.Commit()
}
