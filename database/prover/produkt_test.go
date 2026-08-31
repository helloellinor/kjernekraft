// Package prover tests the database package from outside.
//
// It needs no bridge: Migrate, Database and Conn are already exported,
// so the test builds a database through the same API the rest of the
// house uses.
package prover

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"kjernekraft/database"
)

func prøvebase(t *testing.T) *database.Database {
	t.Helper()
	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "prøve.db"))
	if err != nil {
		t.Fatalf("kunde ikkje opna basen: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := database.Migrate(conn); err != nil {
		t.Fatalf("migreringi: %v", err)
	}
	return &database.Database{Conn: conn}
}

func lagBrukar(t *testing.T, db *database.Database, namn string) int64 {
	t.Helper()
	res, err := db.Conn.Exec(
		"INSERT INTO users (name, birthdate, email, phone, password) VALUES (?, '1990-01-01', ?, ?, 'x')",
		namn, namn+"@dømet.no", namn)
	if err != nil {
		t.Fatalf("brukaren %s: %v", namn, err)
	}
	id, _ := res.LastInsertId()
	return id
}

// A product that comes out of the database should know its own id.
//
// The query did not select m.id, so it was zero for every member: the
// switch list offered you the plan you already had, and the studio's own
// product name was looked up at key 0 and never found.
func TestMedlemskapetKjennerSinEigenID(t *testing.T) {
	db := prøvebase(t)
	brukar := lagBrukar(t, db, "Ida")

	res, err := db.Conn.Exec(`INSERT INTO memberships
	    (name, price, commitment_months, description, features, active, skjult)
	    VALUES ('Aarskort', 69000, 12, '', '', TRUE, TRUE)`)
	if err != nil {
		t.Fatalf("medlemskapet: %v", err)
	}
	medlemskapID, _ := res.LastInsertId()

	if _, err := db.Conn.Exec(`INSERT INTO user_memberships
	    (user_id, membership_id, status, start_date, renewal_date)
	    VALUES (?, ?, 'active', ?, ?)`,
		brukar, medlemskapID, time.Now(), time.Now().AddDate(0, 1, 0)); err != nil {
		t.Fatalf("medlemskapet aat brukaren: %v", err)
	}

	m, err := db.GetUserMembership(brukar)
	if err != nil || m == nil {
		t.Fatalf("las ikkje medlemskapet att: %v", err)
	}
	if m.Membership.ID != int(medlemskapID) {
		t.Errorf("Membership.ID er %d, venta %d", m.Membership.ID, medlemskapID)
	}
	// Same fella, same spurnad: `skjult` avgjer korleis namnet vert
	// rekna ut, og eit uskrive bool-felt er `false` og ikkje «veit ikkje».
	if !m.Membership.Skjult {
		t.Error("Membership.Skjult er false, men medlemskapet er skjult i basen")
	}
	// Den innbakte UserMembership skal framleis bera sin eigen id, og
	// dei tvo skal ikkje vera det same talet.
	if m.UserMembership.MembershipID != int(medlemskapID) {
		t.Errorf("UserMembership.MembershipID er %d, venta %d", m.UserMembership.MembershipID, medlemskapID)
	}
}

// Same claim for the card. Nobody reads KlippekortPackage.ID today, so
// this is not a fix for something that went wrong — it is so the next
// reader does not inherit the same zero.
func TestKlippekortetKjennerSinEigenID(t *testing.T) {
	db := prøvebase(t)
	brukar := lagBrukar(t, db, "Torunn")

	res, err := db.Conn.Exec(`INSERT INTO klippekort_packages
	    (name, category, klipp_count, price, price_per_session, description, valid_days, active, is_popular)
	    VALUES ('10 klipp', 'Reformer', 10, 250000, 25000, '', 365, TRUE, FALSE)`)
	if err != nil {
		t.Fatalf("pakka: %v", err)
	}
	pakkeID, _ := res.LastInsertId()

	if _, err := db.Conn.Exec(`INSERT INTO user_klippekort
	    (user_id, package_id, total_klipp, remaining_klipp, expiry_date, purchase_date, is_active)
	    VALUES (?, ?, 10, 7, ?, ?, TRUE)`,
		brukar, pakkeID, time.Now().AddDate(0, 6, 0), time.Now()); err != nil {
		t.Fatalf("kortet aat brukaren: %v", err)
	}

	kort, err := db.GetUserKlippekort(brukar)
	if err != nil {
		t.Fatalf("las ikkje korti att: %v", err)
	}
	if len(kort) != 1 {
		t.Fatalf("venta eitt kort, fekk %d", len(kort))
	}
	if kort[0].KlippekortPackage.ID != int(pakkeID) {
		t.Errorf("KlippekortPackage.ID er %d, venta %d", kort[0].KlippekortPackage.ID, pakkeID)
	}
	if kort[0].UserKlippekort.PackageID != int(pakkeID) {
		t.Errorf("UserKlippekort.PackageID er %d, venta %d", kort[0].UserKlippekort.PackageID, pakkeID)
	}
}
