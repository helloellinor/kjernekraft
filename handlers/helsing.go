package handlers

import (
	"fmt"
	"strings"
	"time"

	"kjernekraft/models"
)

// Helsing er det fyrste ein ser paa heimesida.
//
// «Velkommen, Anna!» segjer ingen ting ein ikkje alt visste. Det ein
// faktisk lurer paa naar ein opnar sida er *naar skal eg dit att* — og
// det veit systemet, av di ein hev meldt seg paa.
//
// Difor: «Vi sest torsdag morgon, Anna!» Same opplysningi som stend i
// paameldingslista lenger nede, men sagd slik ein sjølv ville sagt henne.
func Helsing(lang, namn string, neste *models.Event, naa time.Time) string {
	// Fyrenamnet. «Vi sest torsdag morgon, Anna Larsen!» er ikkje ei
	// helsing, det er ei innkalling.
	fyrenamn := namn
	if i := strings.IndexByte(namn, ' '); i > 0 {
		fyrenamn = namn[:i]
	}

	if neste == nil {
		return fmt.Sprintf("%s, %s!", t(lang, "site.welcome"), fyrenamn)
	}

	start := neste.StartTime.In(OsloLoc)
	naa = naa.In(OsloLoc)

	iDag := naa.Format("2006-01-02")
	iMorgon := naa.AddDate(0, 0, 1).Format("2006-01-02")
	dagen := start.Format("2006-01-02")

	var naar string
	switch {
	case dagen == iDag:
		// I dag treng ein slaget, ikkje dagen.
		naar = start.Format("15:04")
	case dagen == iMorgon:
		// «i morgon morgon» er ikkje noko nokon seier. Morgonen fær si
		// eigi vending; dei andre bolkarne fylgjer etter «i morgon».
		if start.Hour() < 10 {
			naar = t(lang, "greeting.tomorrow_early")
		} else {
			naar = t(lang, "greeting.tomorrow") + " " + tidbolk(lang, start)
		}
	case start.Sub(naa) < 7*24*time.Hour:
		naar = norskeDagar[start.Weekday()] + " " + tidbolk(lang, start)
	default:
		// Lenger fram enn ei vika: daa er dagen viktigare enn tidi.
		naar = fmt.Sprintf("%d. %s", start.Day(), norskeMaanader[start.Month()])
	}

	return fmt.Sprintf(t(lang, "greeting.see_you"), naar, fyrenamn)
}

// tidbolk gjev «morgon», «føremiddag», «ettermiddag» eller «kveld».
//
// Ein seier ikkje «vi sest torsdag 07:15» til nokon — ein seier
// «torsdag morgon». Klokkeslettet stend i timeplanen for den som treng
// det paa minuttet.
func tidbolk(lang string, t0 time.Time) string {
	switch h := t0.Hour(); {
	case h < 10:
		return t(lang, "greeting.morning")
	case h < 12:
		return t(lang, "greeting.forenoon")
	case h < 17:
		return t(lang, "greeting.afternoon")
	default:
		return t(lang, "greeting.evening")
	}
}

// Kvalifisert segjer um brukaren fær sjaa student- og honnørprisane.
//
// Honnør kjem av fødselsdagen — det er eit tal systemet alt hev, og
// ingen skal krysse av for at dei hev vorte 67. Studentbevis er noko
// ein fortel, og studioet ser det i resepsjonen.
func Kvalifisert(u *models.User) bool {
	if u == nil {
		return false
	}
	if u.StudentSenior {
		return true
	}
	if fodd, err := time.Parse("2006-01-02", u.Birthdate); err == nil {
		aar := time.Since(fodd).Hours() / 24 / 365.25
		if aar >= 67 {
			return true
		}
	}
	return false
}
