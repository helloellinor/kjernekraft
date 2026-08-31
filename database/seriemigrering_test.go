package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Etterfyllinga av serie_id gjeng no i same økta som ALTER TABLE, og ei
// økt er eitt samband: lesinga lyt vera ferdig fyre skrivinga tek til.
// Prøva held båe delane — at kolonna kjem, og at kvar time fær regelen
// sin.
func TestRegelKolonneOgEtterfylling(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "prøve.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL, description TEXT,
		start_time DATETIME NOT NULL, end_time DATETIME,
		location TEXT, organizer TEXT, class_type TEXT,
		teacher_name TEXT, capacity INTEGER, current_enrolment INTEGER,
		color TEXT, room_id INTEGER)`); err != nil {
		t.Fatal(err)
	}

	// Tri utslag av same regelen — same time, lærar, rom, vekedag og
	// klokkeslett — og eitt av ein annan.
	rader := [][2]string{
		{"2026-08-24 18:00:00", "2026-08-24 19:00:00"},
		{"2026-08-31 18:00:00", "2026-08-31 19:00:00"},
		{"2026-09-07 18:00:00", "2026-09-07 19:00:00"},
	}
	for _, r := range rader {
		if _, err := db.Exec(`INSERT INTO events (title, teacher_name, location, room_id, start_time, end_time)
			VALUES ('Yoga', 'Leon', 'Salen', 1, ?, ?)`, r[0], r[1]); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO events (title, teacher_name, location, room_id, start_time, end_time)
		VALUES ('Pilates', 'Mo', 'Salen', 1, '2026-08-25 07:15:00', '2026-08-25 08:15:00')`); err != nil {
		t.Fatal(err)
	}

	if err := leggTilSerieKolonne(db); err != nil {
		t.Fatalf("migreringa feila: %v", err)
	}

	var reglar, utan int
	if err := db.QueryRow("SELECT COUNT(DISTINCT serie_id), COUNT(*) FILTER (WHERE serie_id IS NULL) FROM events").
		Scan(&reglar, &utan); err != nil {
		t.Fatal(err)
	}
	if reglar != 2 {
		t.Errorf("fekk %d reglar, venta 2", reglar)
	}
	if utan != 0 {
		t.Errorf("%d timar stend att utan regel", utan)
	}

	// Dei tri like skal bera det same talet, og det skal vera det
	// minste time-id-et deira.
	var tal, minste int64
	if err := db.QueryRow(`SELECT COUNT(DISTINCT serie_id), MIN(serie_id) FROM events WHERE title = 'Yoga'`).
		Scan(&tal, &minste); err != nil {
		t.Fatal(err)
	}
	if tal != 1 || minste != 1 {
		t.Errorf("dei tri yoga-timane fekk %d ulike regelnamn (minste %d), venta 1 og 1", tal, minste)
	}
}

// Ein base som alt hev `serie_id` skal faa namnet skift, ikkje ei ny og
// tom kolonne attaat.
//
// Dette er den farlege vegen: kvar time som alt ligg der ber serienamnet
// sitt i den gamle kolonna, og eit feilsteg her gjev eit hus fullt av
// timar som ikkje høyrer til noko — utan at noko feilar.
func TestGamalKolonneVertDoeyptOgIkkjeLagdTilPåNytt(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "gamal.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Ein base slik han såg ut fyre omdøypingi, med timar i seg.
	if _, err := db.Exec(`CREATE TABLE events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL, description TEXT,
		start_time DATETIME NOT NULL, end_time DATETIME,
		location TEXT, organizer TEXT, class_type TEXT,
		teacher_name TEXT, capacity INTEGER, current_enrolment INTEGER,
		color TEXT, room_id INTEGER, rule_id INTEGER)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO events
		(title, start_time, end_time, teacher_name, rule_id) VALUES
		('Yoga', '2026-09-07 18:00:00', '2026-09-07 19:00:00', 'Leon', 7),
		('Yoga', '2026-09-14 18:00:00', '2026-09-14 19:00:00', 'Leon', 7)`); err != nil {
		t.Fatal(err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migreringa feila: %v", err)
	}

	// Det gamle namnet er burte, det nye er der.
	var gamalt, nytt int
	db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='rule_id'").Scan(&gamalt)
	db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='serie_id'").Scan(&nytt)
	if gamalt != 0 {
		t.Error("rule_id stend der endaa")
	}
	if nytt != 1 {
		t.Fatal("serie_id kom ikkje")
	}

	// Og — det som tel — verdi fylgde med.
	var tal, serie int
	if err := db.QueryRow(
		"SELECT COUNT(*), COALESCE(MIN(serie_id), 0) FROM events WHERE serie_id IS NOT NULL").
		Scan(&tal, &serie); err != nil {
		t.Fatal(err)
	}
	if tal != 2 || serie != 7 {
		t.Errorf("timane misste serien sin: tal=%d serie=%d, venta 2 og 7", tal, serie)
	}
}
