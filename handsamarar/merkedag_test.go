package handsamarar

import (
	"strings"
	"testing"
	"time"

	"kjernekraft/models"
)

// The day on the clock should change with the language.
//
// It did not. dagKort2 was a fixed table of Nynorsk words — "MÅNDAG",
// "TYSDAG" — and the mark draws itself with that word whatever the page is
// in. It was the most visible place in the house to forget a translation:
// the mark appears once per class, so a week with twenty classes said
// "MÅNDAG" twenty times in the middle of an English page.
//
// The test checks that the three languages give three different words where
// they *should* differ. It could have compared against a fixture —
// "LAURDAG" — but then it would be testing the translation file and not the
// code: change the word in locales and the test should follow, not fail.
// What it holds fast is that the word comes *through* t and not around it.
//
// The global translation reads "mål" relative to the working directory,
// which under test is handlers/. loadTranslations is silent when a file
// cannot be read, so t returns the key instead of the word — and a test
// comparing two keys stands green because both are equally wrong. So we
// point it right first.
func sikraUmsetjingar(t_ *testing.T) {
	t_.Helper()
	locOnce.Do(func() {})
	if localization == nil || len(localization.languages) == 0 {
		localization = &Localization{
			languages: make(map[string]map[string]interface{}),
			basePath:  "../mål",
		}
		localization.loadTranslations()
	}
	if t("nn", "timeplan.saturday") == "timeplan.saturday" {
		t_.Fatal("umsetjingarne let seg ikkje lesa fraa ../locales")
	}
}

func TestDagenPåKlokkaSkifterMedMaalet(t_ *testing.T) {
	sikraUmsetjingar(t_)

	// A Saturday: Nynorsk, Bokmål and English write it differently, so it
	// 	// separates the three in one day.
	laurdag := time.Date(2026, time.August, 29, 10, 0, 0, 0, time.UTC)
	e := models.Event{
		ID: 1, Title: "Yoga", StartTime: laurdag,
		EndTime: laurdag.Add(time.Hour), Capacity: 10,
	}

	sett := map[string]string{}
	for _, lang := range []string{"nn", "nb", "en"} {
		s := NewSession(lang, e, "", laurdag.Add(-24*time.Hour))

		if s.DayAbbrev == "" {
			t_.Errorf("%s: dagen paa klokka er tom", lang)
			continue
		}
		// Det skal vera det umsette ordet, med versalar lagde på her —
		// `.merkedag` ber ingen `text-transform`.
		vil := strings.ToUpper(t(lang, "timeplan.saturday"))
		if s.DayAbbrev != vil {
			t_.Errorf("%s: dagen paa klokka er %q, venta %q", lang, s.DayAbbrev, vil)
		}
		sett[lang] = s.DayAbbrev

		// Dokka ber den same dagen i klartekst, og han skal ogso vera
		// umsett — båe dagen og maanaden.
		if !strings.Contains(s.DatoTekst, t(lang, "timeplan.saturday")) {
			t_.Errorf("%s: dokkedatoen %q ber ikkje den umsette dagen", lang, s.DatoTekst)
		}
		if !strings.Contains(s.DatoTekst, t(lang, "timeplan.month_august")) {
			t_.Errorf("%s: dokkedatoen %q ber ikkje den umsette maanaden", lang, s.DatoTekst)
		}
	}

	// And the decisive part: the three are not the same word. If they were, a
	// 		// fixed table would still stand somewhere — the test above would pass
	// 		// if t gave the same string for all three.
	if sett["nn"] == sett["en"] {
		t_.Errorf("nynorsk og engelsk gjev det same ordet (%q) — noko er hardkoda", sett["nn"])
	}
	if sett["nn"] == sett["nb"] {
		t_.Errorf("nynorsk og bokmaal gjev det same ordet (%q) for laurdag", sett["nn"])
	}
}
