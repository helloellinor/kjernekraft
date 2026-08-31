package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func timeDB(t *testing.T) *Database {
	t.Helper()
	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "timar.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := Migrate(conn); err != nil {
		t.Fatal(err)
	}
	return &Database{Conn: conn}
}

// Skildringi er nullbar i skjemaet, og testdata-såingi skriv ikkje
// kolonnen i det heile. Les ein henne rått, fell heile uppslaget — og
// handsamaren svarar «timen finst ikkje» på ein time som finst.
func TestTimenUtanSkildringLetSegHenta(t *testing.T) {
	db := timeDB(t)
	res, err := db.Conn.Exec(`INSERT INTO events
		(title, start_time, end_time, class_type, teacher_name, capacity, current_enrolment)
		VALUES ('Vinyasa', '2030-09-07 18:00:00', '2030-09-07 19:00:00', 'yoga', 'Kristina', 0, 0)`)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()

	if _, err := db.GetEventByID(id); err != nil {
		t.Fatalf("GetEventByID på ein time utan skildring: %v", err)
	}
}

// Plassane skal tyda det same kva for spurning timen kom or. Kom han
// or GetEventByID fyrr, var talet timen sitt rå felt — 0 når rommet
// eig talet — medan GetAllEvents gav rommet sitt. Same timen, tvo svar.
func TestPlassaneTyderDetSameKvenSomSpyr(t *testing.T) {
	db := timeDB(t)
	var romID int64
	if err := db.Conn.QueryRow(`INSERT INTO rooms (name, capacity) VALUES ('Prøvesalen', 18) RETURNING id`).Scan(&romID); err != nil {
		t.Fatal(err)
	}
	res, err := db.Conn.Exec(`INSERT INTO events
		(title, description, start_time, end_time, room_id, capacity, current_enrolment)
		VALUES ('Vinyasa', '', '2030-09-07 18:00:00', '2030-09-07 19:00:00', ?, 0, 0)`, romID)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()

	ein, err := db.GetEventByID(id)
	if err != nil {
		t.Fatal(err)
	}
	alle, err := db.GetAllEvents()
	if err != nil {
		t.Fatal(err)
	}
	if ein.Capacity != alle[len(alle)-1].Capacity {
		t.Fatalf("GetEventByID gjev %d plassar, GetAllEvents gjev %d — same timen",
			ein.Capacity, alle[len(alle)-1].Capacity)
	}
}
