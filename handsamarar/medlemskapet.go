package handsamarar

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"kjernekraft/models"
)

// Medlemskapet mitt.
//
// Belastningane står attmed medlemskapet og ikkje i ei eiga fane: dei to
// svarar på det same spørsmålet — kva dette kostar meg — og ein les dei
// saman. Byte er noko anna, og har si eiga fane.
//
// Før låg det på to stader: «Medlemskapet mitt» synte kva du hadde, og
// «Behandle medlemskap» spurde deg kor lenge du ville binde deg og tilrådde
// eit. Det andre er rett for ein som ikkje har noko. For ein som alt har
// eit medlemskap, er det å tilrå frå null — han veit kva han har, og vil
// vite kva som er annleis.
//
// Difor er det éi side med tre faner: det du har, det du kan byte til, og
// det du har betalt.

// SwitchOption er eit medlemskap du kan byte til, sett frå det du har.
type SwitchOption struct {
	ID       int
	Name     string
	Price    string // ferdig formatert, «590 kr/md»
	Summary  string // kva som endrar seg
	Savings  string // den delen av setninga som er ei innsparing
	Disabled bool
	Reason   string
	LinkText string
	LinkHref string
}

// kronerFromOre gjer øre om til heile kroner.
func kronerFromOre(ore int) int { return ore / 100 }

// MembershipPage syner medlemskapet, byta og belastningane.
func (a *App) MembershipPage(w http.ResponseWriter, r *http.Request) {
	lang := GetLanguageFromRequest(r)
	nå := a.Nå()
	user := brukaren(r)

	noverande, err := a.DB.GetUserMembership(int64(user.ID))
	if err != nil {
		noverande = nil
	}

	kvalifisert := Kvalifisert(user)
	alle, err := a.DB.GetAllMemberships()
	if err != nil {
		alle = nil
	}

	overstyrte, err := a.DB.Produktnamn("medlemskap")
	if err != nil {
		overstyrte = nil
	}

	byte := switchOptions(lang, alle, noverande, kvalifisert, overstyrte)

	// Bindinga.
	//
	// Sida sa båe delar før: «Bindinga har 13 månader att» og «Du er
	// ikkje bunden». Grunnen var at setninga las
	// `MonthsUntilBindingEnd`, som berre vert rekna ut i ein *annan*
	// handsamar (htmx-modulen på heimesida) og difor var null her, medan
	// bytelista las `CommitmentMonths` frå planen — 13 for eit
	// medlemskap som heiter «12-måneder», og med ein bindingsdato som
	// alt er gått ut.
	//
	// Bunden er du om det står ein bindingsdato *fram i tid*. Ikkje kva
	// planen ein gong sa, og ikkje eit tal nokon andre rekna ut.
	bunden := false
	maanaderAtt := 0
	if noverande != nil && noverande.BindingEnd != nil {
		slutt := *noverande.BindingEnd
		if slutt.After(nå) {
			bunden = true
			maanaderAtt = (slutt.Year()-nå.Year())*12 + int(slutt.Month()-nå.Month())
			if slutt.Day() < nå.Day() {
				maanaderAtt--
			}
			if maanaderAtt < 1 {
				maanaderAtt = 1 // mindre enn ein månad er framleis bunden
			}
		}
	}

	// Namnet på det du har, kjem same vegen som namna i lista.
	noverandeNamn := ""
	if noverande != nil {
		noverandeNamn = Namn(overstyrte[noverande.ID], lang,
			MedlemskapNamn(lang, noverande.Membership))
	}

	// An assigned membership is not something you switch away from.
	// 	//
	// 	// Black follows the permission and has no row in user_memberships;
	// 	// "switch to" would point at an agreement that does not exist, and what
	// 	// you *could* switch to is all poorer and costs money. The tab stood
	// 	// there offering a teacher the chance to downgrade themselves to
	// 	// something they would have to pay for.
	tildelt := noverande != nil && noverande.Tildelt

	var medlemSidan *time.Time
	if noverande != nil {
		if d, feil := a.DB.MedlemSidan(int64(user.ID)); feil != nil {
			log.Printf("medlem sidan for %d: %v", user.ID, feil)
		} else if !d.IsZero() {
			medlemSidan = &d
		}
	}

	faner := []Tab{
		{Key: "medlemskapet", Name: t(lang, "medlemskapet.tab_medlemskapet")},
	}
	if !tildelt {
		faner = append(faner, Tab{Key: "byt", Name: t(lang, "medlemskapet.tab_byt")})
	}
	rekkja := fanerekkje(r, "faneark-medlemskapet", "fane",
		t(lang, "medlemskapet.title"), faner)

	renderPage(w, r, "pages/medlemskapet", sidedata(r, SidaMedlemskap, t(lang, "medlemskapet.title"), map[string]any{
		"Faner":         rekkja,
		"Noverande":     noverande,
		"NoverandeNamn": noverandeNamn,
		"Bunden":        bunden,
		"MedlemSidan":   medlemSidan,
		"MaanaderAtt":   maanaderAtt,
		// Kann du seia upp? `CanCancel` på modellen vart aldri sett av
		// nokon — feltet stod som `false` for alle, og difor teikna sida
		// aldri uppseiingsknappen. Ingen såg at adressa han gjekk til
		// heller ikkje fanst.
		//
		// Regelen er den same som bindingi elles: er du bunden, kann du
		// ikkje seia upp; er bindingi ute, kann du. Og noko som alt er
		// sagt upp kann ikkje seiast upp ein gong til.
		// … og noko som fylgjer eit løyve kann ikkje seiast upp i det
		// heile: det finst ingi rad aa seia upp, og knappen hadde bode
		// deg noko huset ikkje kann halda.
		"KannSeiaUpp":   noverande != nil && !bunden && noverande.Status != "cancelled" && !tildelt,
		"Tildelt":       tildelt,
		"SwitchOptions": byte,
	}))
}

// switchOptions set kvart medlemskap opp mot det brukaren har.
//
// Prisen aleine seier lite når du alt betaler for noko. Det som tel er
// skilnaden: «100 kr mindre i månaden, mot 12 månaders binding».
func switchOptions(lang string, alle []models.Membership, noverande *models.MembershipWithDetails,
	kvalifisert bool, overstyrte map[int]map[string]string) []SwitchOption {
	var ut []SwitchOption
	for _, m := range alle {
		if noverande != nil && m.ID == noverande.ID {
			continue // det du alt har, er ikkje eit byte
		}

		v := SwitchOption{
			ID:    m.ID,
			Name:  Namn(overstyrte[m.ID], lang, MedlemskapNamn(lang, m)),
			Price: fmt.Sprintf(t(lang, "medlemskapet.per_month"), kronerFromOre(m.Price)),
		}

		// Bindinga er halve svaret på om eit byte er verdt det.
		binding := t(lang, "medlemskapet.no_binding")
		if m.CommitmentMonths > 0 {
			binding = fmt.Sprintf(t(lang, "medlemskapet.binding_months"), m.CommitmentMonths)
		}

		if noverande != nil {
			skilnad := kronerFromOre(m.Price) - kronerFromOre(noverande.Price)
			switch {
			case skilnad < 0:
				v.Savings = fmt.Sprintf(t(lang, "medlemskapet.less_per_month"), -skilnad)
				v.Summary = ", " + binding
			case skilnad > 0:
				v.Summary = fmt.Sprintf(t(lang, "medlemskapet.more_per_month"), skilnad) + ", " + binding
			default:
				v.Summary = fmt.Sprintf(t(lang, "medlemskapet.same_price"), binding)
			}
		} else {
			v.Summary = binding
		}

		// Studentprisen står der, men stengd, og grunnen er ikkje ei
		// feilmelding: ho er ei lenkje til staden der du kan gjere noko
		// med det. Ei rad som berre er borte, lærer deg ingenting.
		if m.IsStudentSenior && !kvalifisert {
			v.Disabled = true
			v.Reason = t(lang, "medlemskapet.student_locked")
			v.LinkText = t(lang, "medlemskapet.set_student_status")
			v.LinkHref = "/elev/min-profil"
		}

		ut = append(ut, v)
	}
	return ut
}
