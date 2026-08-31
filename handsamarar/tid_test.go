package handsamarar

import (
	"testing"
	"time"

	"kjernekraft/models"
)

// Ei lagra tid og eit «no» er ikkje same slaget, og skilnaden er ikkje
// synleg fyrr klokka passerer timen.
//
// Basen held klokka på veggen — «2026-08-27 17:00:00» — og drivaren les
// henne som UTC. Sette ein den verdien opp mot eit verkeleg augneblink,
// låg dei to timar frå kvarandre om sommaren: klokka 18:30 stod ein time
// som gjekk 17:00–18:00 framleis som ledig og ikkje som ferdig.
func TestFerdigTimeErFerdig(t *testing.T) {
	oslo, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Skipf("inga tidssonetabell: %v", err)
	}

	// Slik radi kjem or basen: rett klokke, feil merkelapp.
	start := time.Date(2026, 8, 27, 17, 0, 0, 0, time.UTC)
	slutt := time.Date(2026, 8, 27, 18, 0, 0, 0, time.UTC)
	// Slik tenaren les klokka: eit verkeleg augneblink, halvannan time
	// etter at timen var ute.
	no := time.Date(2026, 8, 27, 19, 30, 0, 0, oslo)

	if !fyre(slutt, no) {
		t.Errorf("timen slutta 18:00 og klokka er 19:30, men han stend ikkje som ferdig")
	}
	if etter(start, no) {
		t.Errorf("timen byrja 17:00 og klokka er 19:30, men han stend som komande")
	}

	// Og ein time som verkeleg ligg framfyre skal framleis gjera det.
	seinare := time.Date(2026, 8, 27, 20, 0, 0, 0, time.UTC)
	if !etter(seinare, no) {
		t.Errorf("timen byrjar 20:00 og klokka er 19:30, men han stend ikkje som komande")
	}
}

// Same prøva gjennom Framsyning, som er det sida faktisk les.
func TestErUmmeFylgjerVeggklokka(t *testing.T) {
	oslo, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Skipf("inga tidssonetabell: %v", err)
	}
	no := time.Date(2026, 8, 27, 19, 30, 0, 0, oslo)
	e := models.Event{
		ID:        1,
		Title:     "Yoga",
		StartTime: time.Date(2026, 8, 27, 17, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 8, 27, 18, 0, 0, 0, time.UTC),
		Capacity:  10,
	}
	f := NewSession("nn", e, no.Format("2006-01-02"), no)
	if !f.IsPast {
		t.Errorf("ErUmme er falsk for ein time som slutta halvannan time sidan")
	}
}
