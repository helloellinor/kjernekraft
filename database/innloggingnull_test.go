package database

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// Ein brukar utan adresse skal kunna logga inn.
//
// Dette er tridje gongen den same lyten syner seg, og han ser ulik ut
// kvar gong av di han kjem ut ein ny stad:
//
//  1. GetAllUsers las kolonnone raatt -> 500 på /admin.
//     (TestBrukarUtanAdresseTekIkkjeLista, i denne mappa)
//  2. GetUserByID fekk COALESCE og var trygg.
//  3. AuthenticateUser vart gløymd — og der kom han ut som *feil
//     passord*.
//
// Nummer tri var den verste av dei tri, og ikkje av di han var
// vanskelegare: `Scan` fall med «converting NULL to string is
// unsupported», handsamaren gjorde kvar feil um til «ugyldig e-post
// eller passord», og brukaren fekk vita at han hugsa gale. Han hugsa
// rett. Ei feilmelding som peikar på brukaren når det er huset som er
// gale, er ei melding som gjer at ingen leitar der feilen er. Seks av
// aatte brukarar i basen stod slik då det vart funne.
//
// Lærdomen som stend att: kolonnor som *kann* vera NULL lyt lesast som
// um dei er det, kvar einaste stad dei vert lesne — og ein feil frå
// basen skal aldri kle seg ut som eit gale passord.
func TestBrukarUtanAdresseKannLoggaInn(t *testing.T) {
	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "prøve.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if _, err := conn.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL, birthdate TEXT, email TEXT UNIQUE, phone TEXT,
		address TEXT, postal_code TEXT, city TEXT, country TEXT,
		password TEXT NOT NULL,
		newsletter_subscription BOOLEAN DEFAULT 0,
		terms_accepted BOOLEAN DEFAULT 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`CREATE TABLE "loyve" (
		id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE)`); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`CREATE TABLE "brukarloyve" (
		user_id INTEGER NOT NULL, loyve_id INTEGER NOT NULL,
		PRIMARY KEY (user_id, loyve_id))`); err != nil {
		t.Fatal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	// Adresse, postnummer, by og land stend ikkje. Det er den vanlege
	// stoda for ein brukar som er lagd inn ein annan veg enn gjenom
	// registreringsskjemaet.
	if _, err := conn.Exec(
		`INSERT INTO users (name, email, password) VALUES ('Utan Adresse', 'utan@dømet.no', ?)`,
		string(hash)); err != nil {
		t.Fatal(err)
	}

	db := &Database{Conn: conn}

	brukar, err := db.AuthenticateUser("utan@dømet.no", "password123")
	if err != nil {
		t.Fatalf("ein brukar utan adresse kom ikkje inn: %v", err)
	}
	if brukar.Name != "Utan Adresse" {
		t.Errorf("fekk brukaren %q", brukar.Name)
	}
	if brukar.Address != "" {
		t.Errorf("adressa skulde vore tom, var %q", brukar.Address)
	}

	// Og det som *skal* vera gale, skal framleis vera gale — og det skal
	// segja at det er gale passord og ikkje noko anna.
	_, err = db.AuthenticateUser("utan@dømet.no", "heilt feil")
	if !errors.Is(err, ErrUgyldigInnlogging) {
		t.Errorf("gale passord gav %v, venta ErrUgyldigInnlogging", err)
	}
	_, err = db.AuthenticateUser("finstikkje@dømet.no", "password123")
	if !errors.Is(err, ErrUgyldigInnlogging) {
		t.Errorf("ukjend e-post gav %v, venta ErrUgyldigInnlogging", err)
	}
}
