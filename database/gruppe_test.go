package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func gruppeDB(t *testing.T) *Database {
	t.Helper()
	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "prøve.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := Migrate(conn); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`INSERT INTO users (id, name, birthdate, email, phone, password) VALUES
		(1, 'Med', '1990-01-01', 'med@do.me', '901', 'x'),
		(2, 'Utan', '1990-01-01', 'utan@do.me', '902', 'x')`); err != nil {
		t.Fatal(err)
	}
	return &Database{Conn: conn}
}

func lagGruppetime(t *testing.T, db *Database, tittel string, gruppe interface{}) int64 {
	t.Helper()
	res, err := db.Conn.Exec(`INSERT INTO events
		(title, start_time, end_time, capacity, current_enrolment, gruppe_id)
		VALUES (?, '2030-09-07 18:00:00', '2030-09-07 19:00:00', 10, 0, ?)`, tittel, gruppe)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()
	return id
}

// Ein time utan gruppe er open for alle; ein med gruppe er berre for dei
// som er med.
func TestGruppetimenSynerSegBerreForGruppa(t *testing.T) {
	db := gruppeDB(t)
	reformer, err := db.LagGruppe("Reformer")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SettGruppemedlem(reformer, 1, true); err != nil {
		t.Fatal(err)
	}
	lagGruppetime(t, db, "Yoga", nil)
	lagGruppetime(t, db, "Reformer", reformer)

	for _, p := range []struct {
		brukar int64
		vil    int
	}{{1, 2}, {2, 1}} {
		timar, err := db.EventsSynlegeFor(p.brukar)
		if err != nil {
			t.Fatal(err)
		}
		if len(timar) != p.vil {
			var namn []string
			for _, e := range timar {
				namn = append(namn, e.Title)
			}
			t.Errorf("brukar %d saag %v, venta %d timar", p.brukar, namn, p.vil)
		}
	}
}

// Synlegheit er ikkje tryggleik: den som gissar eit id skal ikkje koma
// inn heller.
func TestUtanforGruppaKannIkkjeMeldaSegPå(t *testing.T) {
	db := gruppeDB(t)
	reformer, _ := db.LagGruppe("Reformer")
	db.SettGruppemedlem(reformer, 1, true)
	timeID := lagGruppetime(t, db, "Reformer", reformer)

	if err := db.SignupUserForEvent(2, timeID); err == nil {
		t.Error("ein utanfor gruppa vart meld paa")
	}
	if err := db.SignupUserForEvent(1, timeID); err != nil {
		t.Errorf("ein i gruppa kom ikkje inn: %v", err)
	}
}

// Det same spursmaalet, spurt av flata fyre ho teiknar knappen.
func TestKannSjåTimenFylgjerGruppa(t *testing.T) {
	db := gruppeDB(t)
	reformer, _ := db.LagGruppe("Reformer")
	db.SettGruppemedlem(reformer, 1, true)
	open := lagGruppetime(t, db, "Yoga", nil)
	stengd := lagGruppetime(t, db, "Reformer", reformer)

	for _, p := range []struct {
		time   int64
		brukar int64
		vil    bool
	}{{open, 1, true}, {open, 2, true}, {stengd, 1, true}, {stengd, 2, false}} {
		fekk, err := db.KannSjåTimen(p.time, p.brukar)
		if err != nil {
			t.Fatal(err)
		}
		if fekk != p.vil {
			t.Errorf("KannSjaaTimen(%d, %d) = %v, venta %v", p.time, p.brukar, fekk, p.vil)
		}
	}
}

// Ei sletta gruppe opnar timane sine att.
//
// Stod dei att og peika på ei gruppe som ikkje finst, var dei timar
// ingen kunde sjå — burte, utan at nokon hadde sagt at dei skulde burt.
func TestSlettaGruppeOpnarTimaneAtt(t *testing.T) {
	db := gruppeDB(t)
	reformer, _ := db.LagGruppe("Reformer")
	db.SettGruppemedlem(reformer, 1, true)
	lagGruppetime(t, db, "Reformer", reformer)

	if err := db.SlettGruppe(reformer); err != nil {
		t.Fatal(err)
	}
	timar, err := db.EventsSynlegeFor(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(timar) != 1 {
		t.Errorf("fekk %d timar etter slettingi, venta 1 open", len(timar))
	}
}

// Same namnet to gonger er den same gruppa, ikkje tvo.
func TestSameNamnErSameGruppa(t *testing.T) {
	db := gruppeDB(t)
	ein, err := db.LagGruppe("Reformer")
	if err != nil {
		t.Fatal(err)
	}
	tvo, err := db.LagGruppe("Reformer")
	if err != nil {
		t.Fatal(err)
	}
	if ein != tvo {
		t.Errorf("fekk tvo gruppor (%d og %d) av eitt namn", ein, tvo)
	}
	g, _ := db.Grupper()
	if len(g) != 1 {
		t.Errorf("fekk %d gruppor, venta 1", len(g))
	}
}

// Den private timen og gruppa lever attmed kvarandre.
func TestPrivatTimeOgGruppeLeverAttmedKvarandre(t *testing.T) {
	db := gruppeDB(t)
	reformer, _ := db.LagGruppe("Reformer")
	db.SettGruppemedlem(reformer, 1, true)
	db.SettGruppemedlem(reformer, 2, true)

	// Ein PT-time som høyrer brukar 1 til.
	res, err := db.Conn.Exec(`INSERT INTO events
		(title, start_time, end_time, capacity, current_enrolment, private_user_id)
		VALUES ('PT', '2030-09-07 18:00:00', '2030-09-07 19:00:00', 1, 0, 1)`)
	if err != nil {
		t.Fatal(err)
	}
	pt, _ := res.LastInsertId()

	if kann, _ := db.KannSjåTimen(pt, 1); !kann {
		t.Error("eigaren saag ikkje sin eigen PT-time")
	}
	// Brukar 2 er med i gruppa, men PT-timen er ikkje hennar.
	if kann, _ := db.KannSjåTimen(pt, 2); kann {
		t.Error("ein annan saag PT-timen")
	}
}
