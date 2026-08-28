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
	// Aatte veker yoga med Leon — éin serie.
	for v := 0; v < 8; v++ {
		timar = append(timar, prøvetime("Yoga", "Leon", maandag.AddDate(0, 0, 7*v), 60))
	}
	// Same timen, men i fjor — skal ikkje med.
	timar = append(timar, prøvetime("Yoga", "Leon", maandag.AddDate(-1, 0, 0), 60))
	// Ein annan lærar paa same timen er ein annan serie.
	timar = append(timar, prøvetime("Yoga", "Kari", maandag.AddDate(0, 0, 14).Add(-time.Hour), 60))

	seriar := GrupperTimar(timar, no)

	if len(seriar) != 2 {
		t.Fatalf("venta 2 seriar, fekk %d", len(seriar))
	}
	// Kari 17:00 stend fyre Leon 18:00 — same dagen, klokka avgjer.
	if seriar[0].Laerar != "Kari" || seriar[1].Laerar != "Leon" {
		t.Fatalf("venta Kari fyrst og Leon so, fekk %q og %q", seriar[0].Laerar, seriar[1].Laerar)
	}
	if len(seriar[1].Timar) != 8 {
		t.Fatalf("venta 8 timar i Leon-serien, fekk %d", len(seriar[1].Timar))
	}
	if seriar[1].Klokke != "18:00" || seriar[1].Vekedag != time.Monday || seriar[1].Lengd != 60 {
		t.Fatalf("serien les feil: %+v", seriar[1])
	}
	// Timane i serien stend etter dato.
	for i := 1; i < len(seriar[1].Timar); i++ {
		if seriar[1].Timar[i].StartTime.Before(seriar[1].Timar[i-1].StartTime) {
			t.Fatal("timane i serien er ikkje sorterte etter dato")
		}
	}
	if seriar[1].VekedagNykel() != "timeplan.monday" {
		t.Fatalf("feil vekedagnykel: %s", seriar[1].VekedagNykel())
	}
}

func TestRegelIDenAvgjerGrupperingi(t *testing.T) {
	sona, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Fatal(err)
	}
	no := time.Date(2026, 8, 26, 12, 0, 0, 0, sona)
	maandag := time.Date(2026, 8, 31, 18, 0, 0, 0, sona)

	// Tvo timar med same serie-id høyrer saman jamvel um ein av deim
	// hadde vikar — id-en er sanninga, ikkje samantreffet.
	a := prøvetime("Yoga", "Leon", maandag, 60)
	a.SerieID = 42
	b := prøvetime("Yoga", "Vikar-Kari", maandag.AddDate(0, 0, 7), 60)
	b.SerieID = 42

	seriar := GrupperTimar([]models.Event{b, a}, no)

	if len(seriar) != 1 {
		t.Fatalf("venta 1 serie, fekk %d", len(seriar))
	}
	if seriar[0].SerieID != 42 {
		t.Fatalf("venta serie-id 42, fekk %d", seriar[0].SerieID)
	}
	// Serien ser ut som den næraste komande timen sin.
	if seriar[0].Laerar != "Leon" {
		t.Fatalf("venta at serien les læraren or den fyrste timen, fekk %q", seriar[0].Laerar)
	}
}
