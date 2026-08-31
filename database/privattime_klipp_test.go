package database

import (
	"testing"
	"time"
)

func ptOppsett(t *testing.T, db *Database, namn string, klipp int) (int64, int64) {
	t.Helper()
	userID := lagBrukar(t, db, namn)
	nå := time.Now()
	res, err := db.Conn.Exec(`
		INSERT INTO events (title, start_time, end_time, class_type, capacity)
		VALUES ('PT', ?, ?, ?, 1)`,
		veggtekst(nå.Add(48*time.Hour)), veggtekst(nå.Add(49*time.Hour)), PTKategori)
	if err != nil {
		t.Fatalf("timen: %v", err)
	}
	eventID, _ := res.LastInsertId()
	db.Conn.Exec(`INSERT OR IGNORE INTO klippekort_packages
		(id, name, category, klipp_count, price, price_per_session, description, valid_days, active)
		VALUES (7, '5 klipp PT', ?, 5, 100, 20, '', 180, TRUE)`, PTKategori)
	if klipp > 0 {
		if _, err := db.Conn.Exec(`INSERT INTO user_klippekort
			(user_id, package_id, total_klipp, remaining_klipp, expiry_date, purchase_date, is_active)
			VALUES (?, 7, 5, ?, ?, ?, TRUE)`,
			userID, klipp, veggtekst(nå.AddDate(0, 6, 0)), veggtekst(nå)); err != nil {
			t.Fatalf("kortet: %v", err)
		}
	}
	return eventID, userID
}

func ptAtt(t *testing.T, db *Database, userID int64) int {
	var n int
	db.Conn.QueryRow(`SELECT COALESCE(SUM(remaining_klipp),0) FROM user_klippekort WHERE user_id = ?`, userID).Scan(&n)
	return n
}

// Klippet fylgjer elles krysset i døri. Ei PT-økt er tinga i det ho er
// sett upp, so ho vert klippa der — og ho gjeng aldri i minus. Tri ting
// skal halda, og alle tri hev ein grunn i BokPrivatTime:
//
//   - med kort: eitt klipp av, og kortet ført på paameldingi
//   - tvo gonger: eitt klipp, ikkje tvo
//   - utan kort: økta stend, ingi skuld
func TestPrivatTimeKlipp(t *testing.T) {
	db := prøvebase(t)
	nå := time.Now()

	// Med kort: eitt klipp gjeng av, og paameldingi ber kortet.
	e1, u1 := ptOppsett(t, db, "MedKort", 5)
	if err := db.BokPrivatTime(e1, u1, nå); err != nil {
		t.Fatalf("booking: %v", err)
	}
	if att := ptAtt(t, db, u1); att != 4 {
		t.Errorf("med kort: venta 4 klipp att, fekk %d", att)
	}
	var kort, skuld int
	db.Conn.QueryRow(`SELECT COALESCE(klipp_kort_id,0), skulda FROM event_signups WHERE event_id=? AND user_id=?`, e1, u1).Scan(&kort, &skuld)
	if kort == 0 {
		t.Error("klippet vart ikkje ført paa paameldingi")
	}
	if skuld != 0 {
		t.Error("skuld skrivi der det var klipp")
	}

	// Tvo gonger tek ikkje tvo klipp.
	if err := db.BokPrivatTime(e1, u1, nå); err != nil {
		t.Fatalf("booking 2: %v", err)
	}
	if att := ptAtt(t, db, u1); att != 4 {
		t.Errorf("dubbelbooking tok eit klipp til: %d att", att)
	}

	// Utan kort: økta stend, ingi skuld, ingen minus.
	e2, u2 := ptOppsett(t, db, "UtanKort", 0)
	if err := db.BokPrivatTime(e2, u2, nå); err != nil {
		t.Fatalf("booking utan kort: %v", err)
	}
	if att := ptAtt(t, db, u2); att != 0 {
		t.Errorf("utan kort: venta 0, fekk %d", att)
	}
	var n, skuld2 int
	db.Conn.QueryRow(`SELECT COUNT(*), COALESCE(MAX(skulda),0) FROM event_signups WHERE event_id=? AND user_id=?`, e2, u2).Scan(&n, &skuld2)
	if n != 1 {
		t.Errorf("økta utan kort gav %d paameldingar", n)
	}
	if skuld2 != 0 {
		t.Error("utan kort vart det skrivi skuld — ho skal ikkje gaa i minus")
	}
}

// Segjer den eine frå seg økta si, skal tri ting henda i lag: klippet
// kjem attende, paameldingi gjeng, og meldingi stend i fana. Alle tri i
// ein transaksjon — ei melding um ei avlysing som vart rulla attende
// sender nokon av garde etter ein time som stend.
func TestAvlystPrivatTimeGjevKlippAttOgMelder(t *testing.T) {
	db := prøvebase(t)
	nå := time.Now()

	eventID, userID := ptOppsett(t, db, "Angrar", 5)
	lærarID := lagBrukar(t, db, "Leon")
	if err := db.SettLagaAv(eventID, lærarID); err != nil {
		t.Fatalf("laga-av: %v", err)
	}
	if err := db.BokPrivatTime(eventID, userID, nå); err != nil {
		t.Fatalf("booking: %v", err)
	}
	if att := ptAtt(t, db, userID); att != 4 {
		t.Fatalf("oppsettet: venta 4 klipp att, fekk %d", att)
	}

	if err := db.AvlysPrivatTime(eventID, userID, nå); err != nil {
		t.Fatalf("avlysing: %v", err)
	}

	if att := ptAtt(t, db, userID); att != 5 {
		t.Errorf("klippet kom ikkje attende: %d att", att)
	}
	var påmeld int
	db.Conn.QueryRow(`SELECT COUNT(*) FROM event_signups WHERE event_id=? AND user_id=?`,
		eventID, userID).Scan(&påmeld)
	if påmeld != 0 {
		t.Errorf("paameldingi stend att")
	}

	meld, err := db.VentandeMeldingar()
	if err != nil {
		t.Fatalf("meldingar: %v", err)
	}
	if len(meld) != 1 {
		t.Fatalf("venta éi melding, fekk %d", len(meld))
	}
	m := meld[0]
	if m.Slag != MeldingAvlystPrivat {
		t.Errorf("slaget var %q", m.Slag)
	}
	if m.FråNamn != "Angrar" {
		t.Errorf("fraa var %q", m.FråNamn)
	}
	if m.TilNamn != "Leon" {
		t.Errorf("meldingi fann ikkje læraren som sette timen upp: %q", m.TilNamn)
	}
	if m.Tittel != "PT" {
		t.Errorf("tittelen var %q", m.Tittel)
	}

	// Handsama tek henne ut or fana, men ho vert ikkje sletta: ei tom
	// fana tyder «ingen ting ventar», ikkje «ingen ting hev hendt».
	if err := db.HandsamaMelding(m.ID, nå); err != nil {
		t.Fatalf("handsaming: %v", err)
	}
	if att, _ := db.VentandeMeldingar(); len(att) != 0 {
		t.Errorf("meldingi stend att i fana")
	}
	var finst int
	db.Conn.QueryRow(`SELECT COUNT(*) FROM meldingar`).Scan(&finst)
	if finst != 1 {
		t.Errorf("meldingi vart sletta i staden for aa verta merkt")
	}
}
