package handlers

import (
	"strings"
	"testing"
	"time"

	"kjernekraft/models"
)

// Helsinga sa to timar for seint.
//
// Tidene i databasen er klokka på veggen — «2026-08-27 16:30:00» — og
// drivaren merkjer dei UTC av di strengen ikkje ber noka sone. Helsinga
// rekna dei om til Oslo og la difor to timar på: «Sest 18:30» om ein
// time som gjekk 16:30, og ingen var påmeld noko halv sju.
//
// Prøva byggjer tidi slik drivaren gjev henne, og krev at helsinga
// seier den same klokka som lista under.
func TestHelsingaSegjerDenSameKlokkaSomTimen(t *testing.T) {
	oslo, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Skip("ingi sonefil:", err)
	}
	gamal := OsloLoc
	OsloLoc = oslo
	defer func() { OsloLoc = gamal }()

	// Slik drivaren les «2026-08-27 16:30:00»: rett klokke, feil merkelapp.
	start := time.Date(2026, 8, 27, 16, 30, 0, 0, time.UTC)
	naa := time.Date(2026, 8, 27, 9, 0, 0, 0, oslo)

	fekk := Helsing("nn", "Ellinor Linnea Stokke", &models.Event{StartTime: start}, naa)

	if !strings.Contains(fekk, "16:30") {
		t.Errorf("helsinga sa %q, venta klokka 16:30", fekk)
	}
	if strings.Contains(fekk, "18:30") {
		t.Errorf("helsinga flytte timen to timar: %q", fekk)
	}
	// Og fyrenamnet, ikkje heile namnet.
	if strings.Contains(fekk, "Stokke") {
		t.Errorf("helsinga nytta heile namnet: %q", fekk)
	}
}

// Ein time seinare same døgeret skal framleis lesast som «i dag», og
// ikkje skuvast yver midnatt av ei umrekning.
func TestSeinTimeIDagVertIkkjeIMorgon(t *testing.T) {
	oslo, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Skip("ingi sonefil:", err)
	}
	gamal := OsloLoc
	OsloLoc = oslo
	defer func() { OsloLoc = gamal }()

	start := time.Date(2026, 8, 27, 23, 0, 0, 0, time.UTC) // 23:00 paa veggen
	naa := time.Date(2026, 8, 27, 18, 0, 0, 0, oslo)

	fekk := Helsing("nn", "Ellinor", &models.Event{StartTime: start}, naa)
	if !strings.Contains(fekk, "23:00") {
		t.Errorf("helsinga sa %q, venta 23:00 i dag", fekk)
	}
}
