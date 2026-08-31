package database

import (
	"database/sql"
	"time"
)

// A clip is spent when someone *was there*, not when they signed up.
//
// Until now a clip was never spent at all: remaining_klipp was only
// written by the purchase and by the test-data shuffle. You could do
// reformer all autumn and still have ten clips left.
//
// Why the check-in and not the signup: the house already distinguishes the
// two — see NettFrammott, "a signup alone says she meant to go". A clip
// pays for a class you *got*, so it follows the check-in. Cancel, or fail
// to turn up, and it costs nothing; undo the check-in and the clip comes
// back (see FjernFrammote).
//
// What a class costs is not a list in the code. Group classes in the hall
// are covered by the membership; Reformer and personal training are sold
// as klippekort. The link between them is the name: the class's class_type
// against the package's category. So the category is found by lookup and
// not by a switch — add a new klippekort category and it works at once.

// MigrerKlippbruk adds the columns the check-in needs.
//
//	klipp_kort_id  the card the clip was taken from. NULL = not clipped.
//	               Knowing *which card* is what lets an undone check-in
//	               put the clip back in the right place; without it you
//	               would guess, and a card that expired meanwhile would
//	               get it.
//	skulda         1 when someone was checked in with no clip to take.
//	               The class was given and should not fall out of the
//	               accounts just because the card was empty at the door.
func MigrerKlippbruk(db *sql.DB) error {
	for _, k := range []struct{ namn, def string }{
		{"klipp_kort_id", "ALTER TABLE event_signups ADD COLUMN klipp_kort_id INTEGER"},
		{"skulda", "ALTER TABLE event_signups ADD COLUMN skulda INTEGER NOT NULL DEFAULT 0"},
	} {
		var finst bool
		if err := db.QueryRow(
			"SELECT COUNT(*) FROM pragma_table_info('event_signups') WHERE name=?", k.namn).
			Scan(&finst); err != nil {
			return err
		}
		if finst {
			continue
		}
		if _, err := db.Exec(k.def); err != nil {
			return err
		}
	}
	return nil
}

// klippKategori gives the klippekort category a class type should be paid
// with, and false when the class costs no clips.
//
// The comparison is loose on purpose: admins write class_type freely, so
// "Reformer", "reformer" and " Reformer " are the same class.
func klippKategori(q rader, timetype string) (string, bool) {
	if timetype == "" {
		return "", false
	}
	var kat string
	err := q.QueryRow(`
		SELECT category FROM klippekort_packages
		WHERE LOWER(TRIM(category)) = LOWER(TRIM(?))
		LIMIT 1`, timetype).Scan(&kat)
	if err != nil {
		return "", false
	}
	return kat, true
}

// rader is the small part of *sql.DB and *sql.Tx used here, so the same
// code runs inside and outside a transaction.
type rader interface {
	QueryRow(string, ...any) *sql.Row
}

// brukKlipp tek eit klipp for timen, inne i transaksjonen krysset gjeng i.
//
// Kortet som gjeng ut fyrst vert nytta fyrst. Elles kunde eit kort med
// kort frist liggja urørt og gaa ut medan klippi vart tekne av eit kort
// som varer til neste sumar.
func brukKlipp(tx *sql.Tx, eventID, userID int64, nå time.Time) error {
	var timetype string
	if err := tx.QueryRow(
		`SELECT COALESCE(class_type, '') FROM events WHERE id = ?`, eventID).
		Scan(&timetype); err != nil {
		return err
	}

	kat, kostar := klippKategori(tx, timetype)
	if !kostar {
		// Group class: the membership covers it.
		return nil
	}

	// If the clip is already taken, it is taken. A PT session is clipped when
	// it is *set up* (see BokPrivatTime) and not at the door, so without this
	// line it would cost two: one for the booking and one for the check-in.
	// The check is general because it should be — clipping the same signup
	// twice is wrong whichever way it happens.
	var alt sql.NullInt64
	if err := tx.QueryRow(`
		SELECT klipp_kort_id FROM event_signups
		WHERE event_id = ? AND user_id = ?`, eventID, userID).Scan(&alt); err == nil && alt.Valid {
		return nil
	}

	var kortID int64
	err := tx.QueryRow(`
		SELECT uk.id
		FROM user_klippekort uk
		JOIN klippekort_packages kp ON kp.id = uk.package_id
		WHERE uk.user_id = ? AND kp.category = ?
		  AND uk.is_active = TRUE AND uk.remaining_klipp > 0
		  AND uk.expiry_date > ?
		ORDER BY uk.expiry_date ASC
		LIMIT 1`, userID, kat, veggtekst(nå)).Scan(&kortID)

	if err == sql.ErrNoRows {
		// No clip to take. She is at the door and the class is starting; we write
		// the debt and raise it for the admins rather than stopping her here.
		_, err := tx.Exec(`
			UPDATE event_signups SET skulda = 1
			WHERE event_id = ? AND user_id = ?`, eventID, userID)
		return err
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
		UPDATE user_klippekort SET remaining_klipp = remaining_klipp - 1
		WHERE id = ? AND remaining_klipp > 0`, kortID); err != nil {
		return err
	}
	_, err = tx.Exec(`
		UPDATE event_signups SET klipp_kort_id = ?, skulda = 0
		WHERE event_id = ? AND user_id = ?`, kortID, eventID, userID)
	return err
}

// gjevAttKlipp puts the clip back when a check-in is undone.
func gjevAttKlipp(tx *sql.Tx, eventID, userID int64) error {
	var kortID sql.NullInt64
	if err := tx.QueryRow(`
		SELECT klipp_kort_id FROM event_signups
		WHERE event_id = ? AND user_id = ?`, eventID, userID).Scan(&kortID); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	if kortID.Valid {
		// Capped at total_klipp: a card must not be able to give back more clips
		// than it had. Only possible if somebody edited the database by hand, but
		// a card with 11 of 10 is worse than a lost clip.
		if _, err := tx.Exec(`
			UPDATE user_klippekort
			SET remaining_klipp = MIN(remaining_klipp + 1, total_klipp)
			WHERE id = ?`, kortID.Int64); err != nil {
			return err
		}
	}

	_, err := tx.Exec(`
		UPDATE event_signups SET klipp_kort_id = NULL, skulda = 0
		WHERE event_id = ? AND user_id = ?`, eventID, userID)
	return err
}
