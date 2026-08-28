package database

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func rabattDB(t *testing.T) *Database {
	t.Helper()
	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "prøve.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	if _, err := conn.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT, email TEXT,
		student_senior BOOLEAN DEFAULT FALSE)`); err != nil {
		t.Fatal(err)
	}
	if err := migrerRabattkrav(conn); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec("INSERT INTO users (id, name, email) VALUES (1, 'Ola', 'ola@do.me')"); err != nil {
		t.Fatal(err)
	}
	return &Database{Conn: conn}
}

var naa = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

// Å krysse av gjev eit krav, ikkje ein rabatt.
//
// Fyrr sette krysset rabatten beinveges — eller rettare: det sette ingen
// ting, av di profilhandsamaren aldri las feltet. Baae delar er gale paa
// same viset: flata avgjorde noko ho ikkje hadde grunnlag for.
func TestKrysetGjevEitKravOgIkkjeEinRabatt(t *testing.T) {
	db := rabattDB(t)

	if err := db.LagRabattkrav(1, naa); err != nil {
		t.Fatal(err)
	}

	har, _, err := db.StudentrabattFor(1)
	if err != nil {
		t.Fatal(err)
	}
	if har {
		t.Error("rabatten gjeld alt — ingen hev sett beviset enno")
	}

	ventande, err := db.VentandeRabattkrav()
	if err != nil {
		t.Fatal(err)
	}
	if len(ventande) != 1 || ventande[0].Namn != "Ola" {
		t.Fatalf("kravet stend ikkje i køen: %+v", ventande)
	}
}

// Å krysse av to gonger er ikkje to krav.
func TestToKryssErIkkjeToKrav(t *testing.T) {
	db := rabattDB(t)
	for i := 0; i < 3; i++ {
		if err := db.LagRabattkrav(1, naa); err != nil {
			t.Fatal(err)
		}
	}
	ventande, _ := db.VentandeRabattkrav()
	if len(ventande) != 1 {
		t.Errorf("fekk %d krav, venta 1", len(ventande))
	}
}

// Godkjenning gjev rabatten, og han gjeng ut etter tri aar.
func TestGodkjenningGjevRabattSomGjengUt(t *testing.T) {
	db := rabattDB(t)
	if err := db.LagRabattkrav(1, naa); err != nil {
		t.Fatal(err)
	}
	ventande, _ := db.VentandeRabattkrav()

	if err := db.AvgjerRabattkrav(ventande[0].ID, true, naa); err != nil {
		t.Fatal(err)
	}

	har, til, err := db.StudentrabattFor(1)
	if err != nil {
		t.Fatal(err)
	}
	if !har {
		t.Fatal("rabatten gjeld ikkje etter godkjenning")
	}
	vil := naa.AddDate(RabattAar, 0, 0)
	if til.Year() != vil.Year() || til.Month() != vil.Month() || til.Day() != vil.Day() {
		t.Errorf("utgangsdatoen er %s, venta %s", til.Format("2006-01-02"), vil.Format("2006-01-02"))
	}

	// Og kravet er ute or køen.
	if att, _ := db.VentandeRabattkrav(); len(att) != 0 {
		t.Errorf("kravet stend framleis og ventar: %+v", att)
	}
}

// Eit avvist krav gjev ingen rabatt, og det står som avvist.
func TestAvvistKravGjevIngenRabatt(t *testing.T) {
	db := rabattDB(t)
	db.LagRabattkrav(1, naa)
	ventande, _ := db.VentandeRabattkrav()

	if err := db.AvgjerRabattkrav(ventande[0].ID, false, naa); err != nil {
		t.Fatal(err)
	}
	if har, _, _ := db.StudentrabattFor(1); har {
		t.Error("rabatten gjeld etter eit nei")
	}
	siste, err := db.SisteRabattkrav(1)
	if err != nil || siste == nil {
		t.Fatalf("kravet er burte: %v", err)
	}
	if siste.Stoda != RabattAvvist {
		t.Errorf("stoda er %q, venta %q", siste.Stoda, RabattAvvist)
	}
}

// Eit krav vert berre avgjort ein gong.
func TestKravetVertAvgjortBerreEinGong(t *testing.T) {
	db := rabattDB(t)
	db.LagRabattkrav(1, naa)
	ventande, _ := db.VentandeRabattkrav()

	if err := db.AvgjerRabattkrav(ventande[0].ID, true, naa); err != nil {
		t.Fatal(err)
	}
	if err := db.AvgjerRabattkrav(ventande[0].ID, false, naa); err == nil {
		t.Error("kravet lét seg avgjera ein gong til")
	}
	// Og det fyrste svaret stend.
	if har, _, _ := db.StudentrabattFor(1); !har {
		t.Error("det andre svaret tok rabatten")
	}
}

// Å ta krysset attende tek baade kravet og rabatten.
func TestKrysetAvTekBaadeKravOgRabatt(t *testing.T) {
	db := rabattDB(t)
	db.LagRabattkrav(1, naa)
	ventande, _ := db.VentandeRabattkrav()
	db.AvgjerRabattkrav(ventande[0].ID, true, naa)

	if err := db.TrekkRabattkrav(1); err != nil {
		t.Fatal(err)
	}
	har, til, _ := db.StudentrabattFor(1)
	if har {
		t.Error("rabatten stend att etter at krysset vart teke")
	}
	if !til.IsZero() {
		t.Errorf("utgangsdatoen stend att: %s", til)
	}
}
