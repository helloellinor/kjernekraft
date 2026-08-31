package database

import (
	"database/sql"
	"time"

	"kjernekraft/models"
)

// The invisible membership.
//
// A membership could only be two things: active because someone bought
// it, or absent. Teachers and developers get full access without buying
// anything and without standing in the price list where people choose —
// hence "invisible": it exists as a membership, but is not something you
// can pick.
//
// It is *derived from the permission*, not a row in user_memberships.
// That is the decision here, and the reason is what happens when the
// permission goes: a stored row would sit there giving a former teacher
// free access until someone cleared it by hand. A permission taken away
// takes the membership with it in the same moment, because it never was
// anything but the permission.
//
// It also means it cannot be frozen, cancelled or changed — there is
// nothing to change. The template hides those buttons, and Tildelt is the
// flag it looks at.

// SvartMedlemskap er namnet i basen.
const SvartMedlemskap = "Black"

// MigrerSvartMedlemskap adds the hidden flag and sets up Black.
//
//	skjult = TRUE   the membership is not something you can pick. It stays
//	                out of the price list and the selectors, but is still a
//	                full membership for whoever has it. The flag is
//	                "hidden" and not "visible" so that the default — what
//	                nobody has said anything about — is visible.
func MigrerSvartMedlemskap(db *sql.DB) error {
	var finst bool
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('memberships') WHERE name='skjult'").
		Scan(&finst); err != nil {
		return err
	}
	if !finst {
		if _, err := db.Exec(
			"ALTER TABLE memberships ADD COLUMN skjult BOOLEAN NOT NULL DEFAULT FALSE"); err != nil {
			return err
		}
	}

	// Black kostar ingen ting og bind ingen. Han er ein tilgang, ikkje
	// ein avtale.
	if _, err := db.Exec(`
		INSERT INTO memberships (name, price, commitment_months, is_student_senior,
		                         is_special_offer, description, features, active, skjult)
		SELECT ?, 0, 0, FALSE, FALSE,
		       'Full tilgang. Fylgjer løyvet, og kann ikkje kjøpast.',
		       '["Ubegrensa gruppetimar","Ubegrensa reformer","Fylgjer løyvet"]',
		       TRUE, TRUE
		WHERE NOT EXISTS (SELECT 1 FROM memberships WHERE name = ?)`,
		SvartMedlemskap, SvartMedlemskap); err != nil {
		return err
	}

	// If Black already existed — as it did in databases that had the flag
	// before it was renamed — it was not reinserted, and would have stood in
	// the price list with the new column's default. It must be hidden
	// whichever way it got in.
	_, err := db.Exec(
		`UPDATE memberships SET skjult = TRUE WHERE name = ?`, SvartMedlemskap)
	return err
}

// HarFriMedlemskap says whether the user gets Black.
//
// Two sources, different on purpose:
//
//   - the teacher permission, which an admin grants through the interface
//   - the developer list, which lives in a file on the server and cannot
//     be written over the network at all (see utviklar.go)
//
// A permission an admin can give themselves, and one they cannot.
func (db *Database) HarFriMedlemskap(userID int64) (bool, error) {
	var tal int
	if err := db.Conn.QueryRow(`
		SELECT COUNT(*) FROM brukarloyve ur
		JOIN loyve r ON r.id = ur.loyve_id
		WHERE ur.user_id = ? AND r.name = ?`, userID, LøyveLærar).Scan(&tal); err != nil {
		return false, err
	}
	if tal > 0 {
		return true, nil
	}
	return db.ErUtviklarID(userID)
}

// svartMedlemskapFor builds the derived membership.
//
// The dates are not invented to look like a purchase: it starts the day
// the user was created, and it does not renew — it lasts as long as the
// permission. RenewalDate a year out is only so that pages computing
// "days until renewal" do not show something odd.
func (db *Database) svartMedlemskapFor(userID int64) (*models.MembershipWithDetails, error) {
	var m models.Membership
	// skjult has to be here. Without it Black came back with the flag unset,
	// and MedlemskapNamn — which uses that very flag to see that a membership
	// is not a binding product — computed its name from the binding instead.
	// With zero months of binding, "Black" became "Månadskort" on screen
	// while the database said Black all along.
	err := db.Conn.QueryRow(`
		SELECT id, name, price, commitment_months, is_student_senior,
		       is_special_offer, description, features, active, skjult
		FROM memberships WHERE name = ?`, SvartMedlemskap).
		Scan(&m.ID, &m.Name, &m.Price, &m.CommitmentMonths, &m.IsStudentSenior,
			&m.IsSpecialOffer, &m.Description, &m.Features, &m.Active, &m.Skjult)
	if err != nil {
		return nil, err
	}

	// Same grunnen som i MedlemSidan: kolonna beint, ikkje eit uttrykk.
	start, err := db.MedlemSidan(userID)
	if err != nil || start.IsZero() {
		start = time.Now()
	}

	ut := &models.MembershipWithDetails{Membership: m}
	ut.UserMembership = models.UserMembership{
		UserID:       int(userID),
		MembershipID: m.ID,
		Status:       "active",
		StartDate:    start,
		RenewalDate:  time.Now().AddDate(1, 0, 0),
		LastBilled:   start,
		CreatedAt:    start,
	}
	// Nothing to cancel or freeze: there is no row to change.
	ut.CanCancel = false
	ut.CanPause = false
	ut.Tildelt = true
	return ut, nil
}

// SynkFriMedlemskap stops the billing when someone is promoted.
//
// Showing Black on screen is not enough. If an active agreement sits in
// user_memberships, that is what gets billed — the card is only the
// picture. A teacher who got the permission yesterday and a charge today
// has paid for something they get free.
//
// So: permission arrives, agreement steps down. It is *ended*, not
// deleted, so the database still records what it was and when it stopped —
// accounting needs to be able to answer for a charge already made.
//
// A one-way gate on purpose. If the permission goes away again the old
// agreement does not return by itself: it was ended, and what is ended has
// to be bought again. Reviving an agreement and starting to charge for it
// unasked is worse than letting people choose.
func (db *Database) SynkFriMedlemskap(userID int64) (bool, error) {
	fri, err := db.HarFriMedlemskap(userID)
	if err != nil || !fri {
		return false, err
	}

	res, err := db.Conn.Exec(`
		UPDATE user_memberships
		SET status = 'cancelled', end_date = CURRENT_TIMESTAMP
		WHERE user_id = ?
		  AND status IN ('active', 'paused', 'freeze_requested')`, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// MedlemSidan gives the day the account was created.
//
// "Member since" is not the same as "this agreement started". Someone who
// joined in 2019 and changed membership last year has been a member since
// 2019; take the date from user_memberships and it resets every time
// somebody switches plan, and the card lies about something people are
// proud of.
//
// It is also the only source that *exists* for an assigned membership:
// Black has no row in user_memberships at all.
func (db *Database) MedlemSidan(userID int64) (time.Time, error) {
	// Without COALESCE in the query, on purpose. SQLite stores time as text,
	// and the driver converts it to time.Time *only* when the column is
	// declared DATETIME. A COALESCE is an expression rather than a column, so
	// the value came back as a string and the scan failed with "unsupported
	// Scan". The fallback is done in Go instead.
	var når sql.NullTime
	if err := db.Conn.QueryRow(
		`SELECT created_at FROM users WHERE id = ?`, userID).Scan(&når); err != nil {
		return time.Time{}, err
	}
	if !når.Valid {
		return time.Time{}, nil
	}
	return når.Time, nil
}
