package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// prøvebase gjev ei tom base i ein katalog prøva eig sjølv. Migrate
// lagar heile skjemaet, so prøva treng ikkje kjenna det.
func prøvebase(t *testing.T) *Database {
	t.Helper()

	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "prøve.db"))
	if err != nil {
		t.Fatalf("kunde ikkje opna basen: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if err := Migrate(conn); err != nil {
		t.Fatalf("migreringi: %v", err)
	}
	return &Database{Conn: conn}
}

func lagBrukar(t *testing.T, db *Database, namn string) int64 {
	t.Helper()

	res, err := db.Conn.Exec(
		"INSERT INTO users (name, birthdate, email, phone, password) VALUES (?, '1990-01-01', ?, ?, 'x')",
		namn, namn+"@dømet.no", namn)
	if err != nil {
		t.Fatalf("kunde ikkje laga brukaren %s: %v", namn, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("id: %v", err)
	}
	return id
}

// Knappen kann trykkjast tvo gonger paa rad utan at nokon ser det.
// AssignRoleToUser fall paa ein nykelkrasj i det tilfellet.
func TestRollaTolerAtHoVertSettTvoGonger(t *testing.T) {
	db := prøvebase(t)
	id := lagBrukar(t, db, "Kristina")

	for i := 0; i < 2; i++ {
		if err := db.SettRolla(id, RollaLaerar, true); err != nil {
			t.Fatalf("aa setja rolla gong %d: %v", i+1, err)
		}
	}

	har, err := db.HarRolla(id, RollaLaerar)
	if err != nil {
		t.Fatalf("HarRolla: %v", err)
	}
	if !har {
		t.Error("rolla vart sett tvo gonger og er ikkje der")
	}
}

func TestRollaLetSegTakaAvAtt(t *testing.T) {
	db := prøvebase(t)
	id := lagBrukar(t, db, "Kristina")

	if err := db.SettRolla(id, RollaLaerar, true); err != nil {
		t.Fatalf("paa: %v", err)
	}
	if err := db.SettRolla(id, RollaLaerar, false); err != nil {
		t.Fatalf("av: %v", err)
	}

	har, err := db.HarRolla(id, RollaLaerar)
	if err != nil {
		t.Fatalf("HarRolla: %v", err)
	}
	if har {
		t.Error("rolla vart teki av og er der endaa")
	}

	// Aa taka ei rolla som aldri vart sett skal ikkje vera ein feil:
	// knappen veit ikkje kva basen hadde fyrr han vart trykt.
	if err := db.SettRolla(id, RollaAdmin, false); err != nil {
		t.Errorf("aa taka ei rolla som ikkje var der: %v", err)
	}
}

// Dette er heile poenget med steget: veljarane skal sjaa lærarane, og
// berre deim.
func TestLaerarNamnGjevBerreDeiMedRolla(t *testing.T) {
	db := prøvebase(t)

	kristina := lagBrukar(t, db, "Kristina")
	åse := lagBrukar(t, db, "Åse")
	elev := lagBrukar(t, db, "Bjørn")

	if err := db.SettRolla(kristina, RollaLaerar, true); err != nil {
		t.Fatal(err)
	}
	if err := db.SettRolla(åse, RollaLaerar, true); err != nil {
		t.Fatal(err)
	}
	if err := db.SettRolla(elev, RollaAdmin, true); err != nil {
		t.Fatal(err)
	}

	namn, err := db.LaerarNamn()
	if err != nil {
		t.Fatalf("LaerarNamn: %v", err)
	}

	vent := []string{"Kristina", "Åse"}
	if len(namn) != len(vent) {
		t.Fatalf("fekk %v, venta %v", namn, vent)
	}
	for i := range vent {
		if namn[i] != vent[i] {
			t.Errorf("plass %d: fekk %q, venta %q", i, namn[i], vent[i])
		}
	}
}

// FolkOversyn les rollone som ein samanslegen streng. «teacher» ligg
// inni «teacher, admin», og ei rein delstrengsprøve hadde difor svara
// ja paa rolla «ach» ogso.
func TestHarRollaLesHeileNamnetOgIkkjeEiDelAvDet(t *testing.T) {
	prøvor := []struct {
		roller string
		rolla  string
		vent   bool
	}{
		{"teacher, admin", RollaLaerar, true},
		{"teacher, admin", RollaAdmin, true},
		{"admin", RollaLaerar, false},
		{"", RollaLaerar, false},
		{"teacher", "ach", false},
		{"teacher", "teach", false},
	}

	for _, p := range prøvor {
		if fekk := harRolla(p.roller, p.rolla); fekk != p.vent {
			t.Errorf("harRolla(%q, %q) = %v, venta %v", p.roller, p.rolla, fekk, p.vent)
		}
	}
}
