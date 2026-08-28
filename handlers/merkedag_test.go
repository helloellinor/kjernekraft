package handlers

import (
	"strings"
	"testing"
	"time"

	"kjernekraft/models"
)

// Dagen paa klokka skal skifta med maalet.
//
// Han gjorde det ikkje. `dagKort2` var eit fast kart yver nynorske ord —
// «MÅNDAG», «TYSDAG» — og merket teiknar seg med det ordet kva enn sida
// elles stend paa. Det var den mest synlege staden i huset aa gløyma ei
// umsetjing: merket stend ein gong per time, so ei vika med tjuge timar
// sa «MÅNDAG» tjuge gonger midt i ei engelsk sida.
//
// Prøva ser etter at dei tri maali gjev tri ulike ord der dei *skal* vera
// ulike. Ho kunde ha samanlikna med ein fasit — «LAURDAG» — men daa hadde
// ho prøvd umsetjingsfila og ikkje koda: skifter nokon ordet i locales,
// skal prøva fylgja med og ikkje falla. Det ho held fast er at ordet kjem
// *gjenom* `t` og ikkje utanum honom.
// Den globale umsetjingi les «locales» i høve til arbeidsmappa, og under
// prøvor er ho `handlers/`. `loadTranslations` tegjer naar ei fil ikkje
// let seg lesa, so `t` gjev nykelen attende i staden for ordet — og ei
// prøva som samanliknar tvo nyklar med kvarandre stend grøn av di baae
// er like gale. Difor peikar me henne rett fyrst.
func sikraUmsetjingar(t_ *testing.T) {
	t_.Helper()
	locOnce.Do(func() {})
	if localization == nil || len(localization.languages) == 0 {
		localization = &Localization{
			languages: make(map[string]map[string]interface{}),
			basePath:  "../locales",
		}
		localization.loadTranslations()
	}
	if t("nn", "timeplan.saturday") == "timeplan.saturday" {
		t_.Fatal("umsetjingarne let seg ikkje lesa fraa ../locales")
	}
}

func TestDagenPaaKlokkaSkifterMedMaalet(t_ *testing.T) {
	sikraUmsetjingar(t_)

	// Ein laurdag: nynorsk, bokmaal og engelsk skriv honom ulikt, so
	// han skil dei tri fraa kvarandre i éin dag.
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
		// Det skal vera det umsette ordet, med versalar lagde paa her —
		// `.merkedag` ber ingen `text-transform`.
		vil := strings.ToUpper(t(lang, "timeplan.saturday"))
		if s.DayAbbrev != vil {
			t_.Errorf("%s: dagen paa klokka er %q, venta %q", lang, s.DayAbbrev, vil)
		}
		sett[lang] = s.DayAbbrev

		// Dokka ber den same dagen i klartekst, og han skal ogso vera
		// umsett — baade dagen og maanaden.
		if !strings.Contains(s.DatoTekst, t(lang, "timeplan.saturday")) {
			t_.Errorf("%s: dokkedatoen %q ber ikkje den umsette dagen", lang, s.DatoTekst)
		}
		if !strings.Contains(s.DatoTekst, t(lang, "timeplan.month_august")) {
			t_.Errorf("%s: dokkedatoen %q ber ikkje den umsette maanaden", lang, s.DatoTekst)
		}
	}

	// Og det avgjerande: dei tri er ikkje det same ordet. Var dei det,
	// hadde eit fast kart stade att ein stad — prøva yver hadde gjenge
	// grøn um `t` gav den same strengen for alle tri.
	if sett["nn"] == sett["en"] {
		t_.Errorf("nynorsk og engelsk gjev det same ordet (%q) — noko er hardkoda", sett["nn"])
	}
	if sett["nn"] == sett["nb"] {
		t_.Errorf("nynorsk og bokmaal gjev det same ordet (%q) for laurdag", sett["nn"])
	}
}
