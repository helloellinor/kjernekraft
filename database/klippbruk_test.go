package database

import (
	"testing"
	"time"
)

// Oppsettet: ein reformer-time, eit reformer-kort med klipp paa, og ein
// gruppetime som ikkje skal kosta noko.
func klippoppsett(t *testing.T, db *Database, timetype string, klipp int) (eventID, userID int64) {
	t.Helper()
	userID = lagBrukar(t, db, "Anna"+timetype)

	naa := time.Now()
	res, err := db.Conn.Exec(`
		INSERT INTO events (title, start_time, end_time, class_type, capacity)
		VALUES ('Time', ?, ?, ?, 4)`,
		veggtekst(naa), veggtekst(naa.Add(time.Hour)), timetype)
	if err != nil {
		t.Fatalf("timen: %v", err)
	}
	eventID, _ = res.LastInsertId()

	if _, err := db.Conn.Exec(`
		INSERT INTO klippekort_packages (id, name, category, klipp_count, price, price_per_session, description, valid_days, active)
		VALUES (1, '10 klipp Reformer', 'Reformer', 10, 100, 10, '', 180, TRUE)`); err != nil {
		t.Fatalf("pakka: %v", err)
	}
	if klipp > 0 {
		if _, err := db.Conn.Exec(`
			INSERT INTO user_klippekort (user_id, package_id, total_klipp, remaining_klipp, expiry_date, purchase_date, is_active)
			VALUES (?, 1, 10, ?, ?, ?, TRUE)`,
			userID, klipp, veggtekst(naa.AddDate(0, 6, 0)), veggtekst(naa)); err != nil {
			t.Fatalf("kortet: %v", err)
		}
	}
	if _, err := db.Conn.Exec(
		`INSERT INTO event_signups (user_id, event_id, signup_date) VALUES (?, ?, ?)`,
		userID, eventID, veggtekst(naa)); err != nil {
		t.Fatalf("paameldingi: %v", err)
	}
	return eventID, userID
}

func klippAtt(t *testing.T, db *Database, userID int64) int {
	t.Helper()
	var n int
	if err := db.Conn.QueryRow(
		`SELECT COALESCE(SUM(remaining_klipp), 0) FROM user_klippekort WHERE user_id = ?`,
		userID).Scan(&n); err != nil {
		t.Fatalf("klippAtt: %v", err)
	}
	return n
}

func skulda(t *testing.T, db *Database, eventID, userID int64) int {
	t.Helper()
	var n int
	if err := db.Conn.QueryRow(
		`SELECT skulda FROM event_signups WHERE event_id = ? AND user_id = ?`,
		eventID, userID).Scan(&n); err != nil {
		t.Fatalf("skulda: %v", err)
	}
	return n
}

// Krysset tek eit klipp. Det var heile mangelen: ein kunde gaa paa
// reformer heile hausten og framleis hava ti klipp att.
func TestKryssetTekEitKlipp(t *testing.T) {
	db := prøvebase(t)
	eventID, userID := klippoppsett(t, db, "reformer", 10)

	if err := db.MerkFrammote(eventID, userID, time.Now()); err != nil {
		t.Fatalf("MerkFrammote: %v", err)
	}
	if got := klippAtt(t, db, userID); got != 9 {
		t.Errorf("etter eitt kryss vil me hava 9 klipp att, fekk %d", got)
	}
}

// Tvo trykk paa den same knappen er ein ting som hender i ein kiosk.
// Det skal kosta eitt klipp, ikkje tvo.
func TestTvoKryssTekBerreEitKlipp(t *testing.T) {
	db := prøvebase(t)
	eventID, userID := klippoppsett(t, db, "reformer", 10)

	naa := time.Now()
	for i := 0; i < 3; i++ {
		if err := db.MerkFrammote(eventID, userID, naa); err != nil {
			t.Fatalf("MerkFrammote gong %d: %v", i+1, err)
		}
	}
	if got := klippAtt(t, db, userID); got != 9 {
		t.Errorf("tri kryss paa same timen skal kosta eitt klipp; fekk %d att", got)
	}
}

// Gruppetimen ligg i medlemskapet og skal ikkje røra kortet.
func TestGruppetimeKostarIngiKlipp(t *testing.T) {
	db := prøvebase(t)
	eventID, userID := klippoppsett(t, db, "yoga", 10)

	if err := db.MerkFrammote(eventID, userID, time.Now()); err != nil {
		t.Fatalf("MerkFrammote: %v", err)
	}
	if got := klippAtt(t, db, userID); got != 10 {
		t.Errorf("yoga skal ikkje kosta klipp; fekk %d att", got)
	}
}

// Store og smaa bokstavar er den same timen. `class_type` vert skrivi
// fritt av administrasjonen, so «Reformer» og «reformer» er ein time.
func TestTimetypeMatcharUtanUmsynTilStorleik(t *testing.T) {
	db := prøvebase(t)
	eventID, userID := klippoppsett(t, db, "  REFORMER ", 10)

	if err := db.MerkFrammote(eventID, userID, time.Now()); err != nil {
		t.Fatalf("MerkFrammote: %v", err)
	}
	if got := klippAtt(t, db, userID); got != 9 {
		t.Errorf("«  REFORMER » er reformer; venta 9 klipp att, fekk %d", got)
	}
}

// Ingen klipp att: ho stend i døri og timen byrjar. Krysset skal gaa
// gjenom, og skuldi skal staa att so administrasjonen ser henne.
func TestUtanKlippVertSkuldiSkrivi(t *testing.T) {
	db := prøvebase(t)
	eventID, userID := klippoppsett(t, db, "reformer", 0)

	if err := db.MerkFrammote(eventID, userID, time.Now()); err != nil {
		t.Fatalf("MerkFrammote: %v", err)
	}
	var kryssa bool
	if err := db.Conn.QueryRow(
		`SELECT attended_at IS NOT NULL FROM event_signups WHERE event_id = ? AND user_id = ?`,
		eventID, userID).Scan(&kryssa); err != nil {
		t.Fatalf("uppslag: %v", err)
	}
	if !kryssa {
		t.Error("krysset skal gaa gjenom jamvel utan klipp")
	}
	if got := skulda(t, db, eventID, userID); got != 1 {
		t.Errorf("skuldi skal vera skrivi; skulda = %d", got)
	}
}

// Feil person kryssa av: klippet kjem attende paa det kortet det vart
// teke fraa.
func TestAngraKryssGjevKlippetAttende(t *testing.T) {
	db := prøvebase(t)
	eventID, userID := klippoppsett(t, db, "reformer", 10)

	if err := db.MerkFrammote(eventID, userID, time.Now()); err != nil {
		t.Fatalf("MerkFrammote: %v", err)
	}
	if got := klippAtt(t, db, userID); got != 9 {
		t.Fatalf("oppsettet: venta 9, fekk %d", got)
	}

	if err := db.FjernFrammote(eventID, userID); err != nil {
		t.Fatalf("FjernFrammote: %v", err)
	}
	if got := klippAtt(t, db, userID); got != 10 {
		t.Errorf("angra kryss skal gjeva klippet attende; fekk %d", got)
	}
}

// Aa angra tvo gonger skal ikkje trylla fram klipp or inkje.
func TestAngraTvoGongerGjevBerreEittKlippAttende(t *testing.T) {
	db := prøvebase(t)
	eventID, userID := klippoppsett(t, db, "reformer", 10)

	naa := time.Now()
	if err := db.MerkFrammote(eventID, userID, naa); err != nil {
		t.Fatalf("MerkFrammote: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := db.FjernFrammote(eventID, userID); err != nil {
			t.Fatalf("FjernFrammote gong %d: %v", i+1, err)
		}
	}
	if got := klippAtt(t, db, userID); got != 10 {
		t.Errorf("kortet hadde ti klipp; det skal aldri hava fleire. Fekk %d", got)
	}
}
