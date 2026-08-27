package handlers

import (
	"testing"
	"time"

	"kjernekraft/models"
)

// Ein time paa gjeven vekedag og klokkeslett, n veker fram i tidi.
func prøvetime(tittel, laerar string, start time.Time, lengdMin int) models.Event {
	return models.Event{
		Title:       tittel,
		TeacherName: laerar,
		RoomName:    "Salen",
		StartTime:   start,
		EndTime:     start.Add(time.Duration(lengdMin) * time.Minute),
	}
}

func TestGrupperTimarLesReglaneUtOrRadene(t *testing.T) {
	sona, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Fatal(err)
	}
	no := time.Date(2026, 8, 26, 12, 0, 0, 0, sona)
	maandag := time.Date(2026, 8, 31, 18, 0, 0, 0, sona)

	var timar []models.Event
	// Aatte veker yoga med Leon — éin regel.
	for v := 0; v < 8; v++ {
		timar = append(timar, prøvetime("Yoga", "Leon", maandag.AddDate(0, 0, 7*v), 60))
	}
	// Same timen, men i fjor — skal ikkje med.
	timar = append(timar, prøvetime("Yoga", "Leon", maandag.AddDate(-1, 0, 0), 60))
	// Ein annan lærar paa same timen er ein annan regel.
	timar = append(timar, prøvetime("Yoga", "Kari", maandag.AddDate(0, 0, 14).Add(-time.Hour), 60))

	reglar := GrupperTimar(timar, no)

	if len(reglar) != 2 {
		t.Fatalf("venta 2 reglar, fekk %d", len(reglar))
	}
	// Kari 17:00 stend fyre Leon 18:00 — same dagen, klokka avgjer.
	if reglar[0].Laerar != "Kari" || reglar[1].Laerar != "Leon" {
		t.Fatalf("venta Kari fyrst og Leon so, fekk %q og %q", reglar[0].Laerar, reglar[1].Laerar)
	}
	if len(reglar[1].Timar) != 8 {
		t.Fatalf("venta 8 timar i Leon-regelen, fekk %d", len(reglar[1].Timar))
	}
	if reglar[1].Klokke != "18:00" || reglar[1].Vekedag != time.Monday || reglar[1].Lengd != 60 {
		t.Fatalf("regelen les feil: %+v", reglar[1])
	}
	// Timane i regelen stend etter dato.
	for i := 1; i < len(reglar[1].Timar); i++ {
		if reglar[1].Timar[i].StartTime.Before(reglar[1].Timar[i-1].StartTime) {
			t.Fatal("timane i regelen er ikkje sorterte etter dato")
		}
	}
	if reglar[1].VekedagNykel() != "timeplan.monday" {
		t.Fatalf("feil vekedagnykel: %s", reglar[1].VekedagNykel())
	}
}

func TestRegelIDenAvgjerGrupperingi(t *testing.T) {
	sona, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Fatal(err)
	}
	no := time.Date(2026, 8, 26, 12, 0, 0, 0, sona)
	maandag := time.Date(2026, 8, 31, 18, 0, 0, 0, sona)

	// Tvo timar med same regel-id høyrer saman jamvel um ein av deim
	// hadde vikar — id-en er sanninga, ikkje samantreffet.
	a := prøvetime("Yoga", "Leon", maandag, 60)
	a.RuleID = 42
	b := prøvetime("Yoga", "Vikar-Kari", maandag.AddDate(0, 0, 7), 60)
	b.RuleID = 42

	reglar := GrupperTimar([]models.Event{b, a}, no)

	if len(reglar) != 1 {
		t.Fatalf("venta 1 regel, fekk %d", len(reglar))
	}
	if reglar[0].RegelID != 42 {
		t.Fatalf("venta regel-id 42, fekk %d", reglar[0].RegelID)
	}
	// Regelen ser ut som den næraste komande timen sin.
	if reglar[0].Laerar != "Leon" {
		t.Fatalf("venta at regelen les læraren or den fyrste timen, fekk %q", reglar[0].Laerar)
	}
}
