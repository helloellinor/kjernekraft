package database

import (
	"testing"
	"time"
)

func lagTime(t *testing.T, db *Database, tittel string, naar time.Time, plassar int) int64 {
	t.Helper()
	res, err := db.Conn.Exec(`
		INSERT INTO events (title, start_time, end_time, class_type, capacity)
		VALUES (?, ?, ?, 'reformer', ?)`,
		tittel, veggtekst(naar), veggtekst(naar.Add(time.Hour)), plassar)
	if err != nil {
		t.Fatalf("timen %s: %v", tittel, err)
	}
	id, _ := res.LastInsertId()
	return id
}

// PT-økti skal berre synast for den ho er sett av til.
func TestPrivatTimeSynestBerreForEigaren(t *testing.T) {
	db := prøvebase(t)
	anna := lagBrukar(t, db, "Anna")
	bjørn := lagBrukar(t, db, "Bjørn")

	naa := time.Now()
	lagTime(t, db, "Opa reformer", naa, 4)
	privat := lagTime(t, db, "PT Anna", naa, 1)
	if err := db.SettPrivatTime(privat, anna); err != nil {
		t.Fatalf("SettPrivatTime: %v", err)
	}

	forAnna, err := db.EventsSynlegeFor(anna)
	if err != nil {
		t.Fatalf("EventsSynlegeFor(anna): %v", err)
	}
	if len(forAnna) != 2 {
		t.Errorf("Anna skal sjaa baae timane, saag %d", len(forAnna))
	}

	forBjørn, err := db.EventsSynlegeFor(bjørn)
	if err != nil {
		t.Fatalf("EventsSynlegeFor(bjørn): %v", err)
	}
	if len(forBjørn) != 1 {
		t.Fatalf("Bjørn skal berre sjaa den opne timen, saag %d", len(forBjørn))
	}
	if forBjørn[0].Title != "Opa reformer" {
		t.Errorf("Bjørn saag «%s» — PT-økti lak ut", forBjørn[0].Title)
	}
}

// Timeplanen for vika skal fylgja det same skiljet.
func TestPrivatTimeErUteAvVikaTilAndre(t *testing.T) {
	db := prøvebase(t)
	anna := lagBrukar(t, db, "Anna")
	bjørn := lagBrukar(t, db, "Bjørn")

	// Maandagen i denne vika. handlers.VikeMaandag kann ikkje nyttast
	// her — det hadde vorte ein importring.
	naa := time.Now()
	maandag := time.Date(naa.Year(), naa.Month(), naa.Day(), 0, 0, 0, 0, naa.Location())
	for maandag.Weekday() != time.Monday {
		maandag = maandag.AddDate(0, 0, -1)
	}
	privat := lagTime(t, db, "PT Anna", maandag.Add(10*time.Hour), 1)
	if err := db.SettPrivatTime(privat, anna); err != nil {
		t.Fatalf("SettPrivatTime: %v", err)
	}

	vekaTilAnna, err := db.GetEventsForWeek(maandag, anna)
	if err != nil {
		t.Fatalf("veka til Anna: %v", err)
	}
	if len(vekaTilAnna) != 1 {
		t.Errorf("Anna skal sjaa PT-økti si i vika, saag %d", len(vekaTilAnna))
	}

	vekaTilBjørn, err := db.GetEventsForWeek(maandag, bjørn)
	if err != nil {
		t.Fatalf("veka til Bjørn: %v", err)
	}
	if len(vekaTilBjørn) != 0 {
		t.Errorf("Bjørn skal ikkje sjaa PT-økti; saag %d timar", len(vekaTilBjørn))
	}
}

// Synlegheit er ikkje tryggleik. Bjørn ser ikkje timen, men han kann
// gissa id-et — og daa lyt paameldingi seia nei.
func TestPrivatTimeAvviserPaameldingFraaAndre(t *testing.T) {
	db := prøvebase(t)
	anna := lagBrukar(t, db, "Anna")
	bjørn := lagBrukar(t, db, "Bjørn")

	privat := lagTime(t, db, "PT Anna", time.Now(), 1)
	if err := db.SettPrivatTime(privat, anna); err != nil {
		t.Fatalf("SettPrivatTime: %v", err)
	}

	if err := db.SignupUserForEvent(bjørn, privat); err == nil {
		t.Error("Bjørn kom paa ei PT-økt som ikkje er hans")
	}
	if err := db.SignupUserForEvent(anna, privat); err != nil {
		t.Errorf("Anna skal koma paa si eigi økt: %v", err)
	}
}

// Ein vanleg time skal ikkje verta strengare av at kolonna finst.
func TestOpenTimeErFramleisOpen(t *testing.T) {
	db := prøvebase(t)
	bjørn := lagBrukar(t, db, "Bjørn")
	opa := lagTime(t, db, "Opa reformer", time.Now(), 4)

	if err := db.SignupUserForEvent(bjørn, opa); err != nil {
		t.Errorf("open time skal taka imot kven som helst: %v", err)
	}
}
