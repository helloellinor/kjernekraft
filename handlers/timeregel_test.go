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

// Vikaren skal ikkje verta læraren.
//
// `Laerar` stod som den næraste komande timen sin, og det feltet ber
// seriesetningi — det same feltet `UpdateSerieTeacher` skriv til kvar
// einaste time i rekkja. Gav ein den næraste timen ein vikar, var
// vikaren serien sin lærar ved neste lasting, og neste lagring skreiv
// namnet hans yver heile rekkja. Ein vikar er unnataket, so det er
// fleirtalet som er læraren.
func TestVikarenPaaFyrsteTimenVertIkkjeLaerarenIRekkja(t *testing.T) {
	maandag := time.Date(2026, 9, 7, 18, 0, 0, 0, time.Local)

	var timar []models.Event
	for i := 0; i < 6; i++ {
		e := prøvetime("Yoga", "Leon", maandag.AddDate(0, 0, 7*i), 60)
		e.SerieID = 12
		if i == 0 {
			// Den næraste timen — den som avgjorde alt fyrr.
			e.TeacherName = "Vikar-Kari"
		}
		timar = append(timar, e)
	}

	seriar := GrupperTimar(timar, maandag.AddDate(0, 0, -1))
	if len(seriar) != 1 {
		t.Fatalf("venta éi rekkje, fann %d", len(seriar))
	}
	r := seriar[0]

	if r.Laerar != "Leon" {
		t.Errorf("læraren i rekkja er %q; ein vikar paa éin time av seks skal ikkje taka rekkja", r.Laerar)
	}

	// Og lista skal syna baae: rekkja er ikkje éin lærar heile vegen,
	// og det er nett det ein treng aa sjaa.
	if len(r.Laerarar) != 2 || r.Laerarar[0] != "Leon" || r.Laerarar[1] != "Vikar-Kari" {
		t.Errorf("namnelista er %v; venta læraren fyrst og so vikaren", r.Laerarar)
	}
}

// Ei rekkje utan vikar ber eitt namn, og det er læraren.
func TestRekkjaUtanVikarBerEittNamn(t *testing.T) {
	maandag := time.Date(2026, 9, 7, 18, 0, 0, 0, time.Local)

	var timar []models.Event
	for i := 0; i < 3; i++ {
		e := prøvetime("Yoga", "Leon", maandag.AddDate(0, 0, 7*i), 60)
		e.SerieID = 12
		timar = append(timar, e)
	}

	r := GrupperTimar(timar, maandag.AddDate(0, 0, -1))[0]
	if r.Laerar != "Leon" || len(r.Laerarar) != 1 {
		t.Errorf("venta berre Leon, fekk %q %v", r.Laerar, r.Laerarar)
	}
}
