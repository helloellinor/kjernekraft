package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Ein base som alt hev `roles` skal faa namni skifte og *behalda* kven
// som er kva.
//
// Dette er den farlege vegen. `CREATE TABLE IF NOT EXISTS loyve` i
// skjemablokka vilde ha laga ein tom tabell, og so hadde umdøypingi
// hoppa yver av di `loyve` no fanst — kvar administrator og lærar hadde
// vorte liggjande att i ein tabell ingen les meir, og flata hadde synt
// eit hus utan administratorar utan aa segja eit ord.
func TestGamleLøyveTabellarVertDoeyptOgHeldFolket(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "gamal.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Basen slik han såg ut fyrr.
	if _, err := db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL,
			birthdate TEXT NOT NULL, email TEXT NOT NULL UNIQUE,
			phone TEXT NOT NULL UNIQUE, password TEXT NOT NULL);
		CREATE TABLE roles (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE);
		CREATE TABLE user_roles (user_id INTEGER NOT NULL, role_id INTEGER NOT NULL,
			PRIMARY KEY (user_id, role_id));
		INSERT INTO users (id, name, birthdate, email, phone, password)
			VALUES (1, 'Kristina', '1990-01-01', 'k@do.me', '901', 'x');
		INSERT INTO roles (id, name) VALUES (1, 'admin'), (2, 'teacher');
		INSERT INTO user_roles (user_id, role_id) VALUES (1, 1), (1, 2);`); err != nil {
		t.Fatal(err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migreringa feila: %v", err)
	}

	// Dei gamle namni er burte, dei nye er der.
	for _, p := range []struct {
		tabell string
		vil    bool
	}{{"roles", false}, {"user_roles", false}, {"loyve", true}, {"brukarloyve", true}} {
		finst, err := harTabell(db, p.tabell)
		if err != nil {
			t.Fatal(err)
		}
		if finst != p.vil {
			t.Errorf("tabellen %s finst = %v, venta %v", p.tabell, finst, p.vil)
		}
	}
	if gamal, _ := harKolonne(db, "brukarloyve", "role_id"); gamal {
		t.Error("role_id stend der endaa")
	}

	// Og — det som tel — Kristina er framleis båe lærar og administrator.
	dbb := &Database{Conn: db}
	løyve, err := dbb.LøyveFor(1)
	if err != nil {
		t.Fatal(err)
	}
	har := map[string]bool{}
	for _, l := range løyve {
		har[l] = true
	}
	if !har[LøyveAdmin] || !har[LøyveLærar] {
		t.Errorf("Kristina misste løyvi sine i umdøypingi: %v", løyve)
	}
}

// Ein base som alt er døypt om skal ikkje faa noko gjort med seg.
func TestUmdoeypingaKannGaaTvoGonger(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "to.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("den andre migreringa feila: %v", err)
	}
	for _, gamal := range []string{"roles", "user_roles"} {
		if finst, _ := harTabell(db, gamal); finst {
			t.Errorf("%s kom attende", gamal)
		}
	}
}
