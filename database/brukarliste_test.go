package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Ein brukar utan adresse skal ikkje ta heile administrasjonssida.
//
// GetAllUsers las kolonnone raatt, og eit einaste NULL gav «converting
// NULL to string is unsupported» — som vart 500 på /admin, og då kom
// ingen inn nokon stad. Feilen laag der heile tidi; det trongst berre ein
// brukar utan adresse fyre han synte seg.
func TestBrukarUtanAdresseTekIkkjeLista(t *testing.T) {
	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "prøve.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if _, err := conn.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT, birthdate TEXT, email TEXT, phone TEXT,
		address TEXT, postal_code TEXT, city TEXT, country TEXT,
		newsletter_subscription BOOLEAN, terms_accepted BOOLEAN)`); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`CREATE TABLE brukarloyve (user_id INTEGER, loyve_id INTEGER)`); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`CREATE TABLE loyve (id INTEGER PRIMARY KEY, name TEXT)`); err != nil {
		t.Fatal(err)
	}
	// Ein heil brukar og ein utan noko som helst av det valfrie.
	if _, err := conn.Exec(`INSERT INTO users
		(name, birthdate, email, phone, address, postal_code, city, country,
		 newsletter_subscription, terms_accepted) VALUES
		('Heil', '1990-01-01', 'heil@do.me', '900', 'Vegen 1', '0001', 'Oslo', 'NO', 1, 1),
		('Tom', NULL, 'tom@do.me', NULL, NULL, NULL, NULL, NULL, NULL, NULL)`); err != nil {
		t.Fatal(err)
	}

	db := &Database{Conn: conn}
	folk, err := db.GetAllUsers()
	if err != nil {
		t.Fatalf("lista feila paa ein brukar utan adresse: %v", err)
	}
	if len(folk) != 2 {
		t.Fatalf("fekk %d brukarar, venta 2", len(folk))
	}
	if folk[1].Address != "" || folk[1].Phone != "" {
		t.Errorf("det som manglar skulde vorte tomt, ikkje %q/%q", folk[1].Address, folk[1].Phone)
	}
	if folk[0].Address != "Vegen 1" {
		t.Errorf("den heile brukaren misste adressa: %q", folk[0].Address)
	}
}
