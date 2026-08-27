package handlers

import (
	"fmt"
	"net/http"

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

// Fane er ei fane i ei fanerekkje.
type Fane struct {
	Bolk string
	Namn string
}

// Byteval er eit medlemskap du kan byte til, sett frå det du har.
type Byteval struct {
	ID       int
	Namn     string
	Pris     string // ferdig formatert, «590 kr/md»
	Setning  string // kva som endrar seg
	Mindre   string // den delen av setninga som er ei innsparing
	Stengd   bool
	Grunn    string
	VegNamn  string
	VegLenke string
}

// kroneTal gjer øre om til heile kroner.
func kroneTal(ore int) int { return ore / 100 }

// MedlemskapetHandler syner medlemskapet, byta og belastningane.
func MedlemskapetHandler(w http.ResponseWriter, r *http.Request) {
	lang := GetLanguageFromRequest(r)
	user := GetUserFromSession(r)
	if user == nil {
		http.Redirect(w, r, "/innlogging", http.StatusSeeOther)
		return
	}

	noverande, err := DB.GetUserMembership(int64(user.ID))
	if err != nil {
		noverande = nil
	}

	kvalifisert := Kvalifisert(user)
	alle, err := DB.GetAllMemberships()
	if err != nil {
		alle = nil
	}

	overstyrte, err := DB.Produktnamn("medlemskap")
	if err != nil {
		overstyrte = nil
	}

	byte := byteval(lang, alle, noverande, kvalifisert, overstyrte)

	// Namnet på det du har, kjem same vegen som namna i lista.
	noverandeNamn := ""
	if noverande != nil {
		noverandeNamn = Namn(overstyrte[noverande.Membership.ID], lang,
			MedlemskapNamn(lang, noverande.Membership))
	}

	faner := []Fane{
		{Bolk: "medlemskapet", Namn: t(lang, "medlemskapet.tab_medlemskapet")},
		{Bolk: "byt", Namn: t(lang, "medlemskapet.tab_byt")},
	}

	renderPage(w, r, "pages/medlemskapet", map[string]interface{}{
		"Title":         t(lang, "medlemskapet.title"),
		"CurrentPage":   "medlemskap",
		"Lang":          lang,
		"CSRFToken":     CSRFToken(r),
		"IsAdmin":       sessionIsAdmin(r),
		"UserName":      sessionUserName(r),
		"Faner":         faner,
		"Merkelapp":     t(lang, "medlemskapet.title"),
		"Noverande":     noverande,
		"NoverandeNamn": noverandeNamn,
		"Byteval":       byte,
	})
}

// byteval set kvart medlemskap opp mot det brukaren har.
//
// Prisen aleine seier lite når du alt betaler for noko. Det som tel er
// skilnaden: «100 kr mindre i månaden, mot 12 månaders binding».
func byteval(lang string, alle []models.Membership, noverande *models.MembershipWithDetails,
	kvalifisert bool, overstyrte map[int]map[string]string) []Byteval {
	var ut []Byteval
	for _, m := range alle {
		if noverande != nil && m.ID == noverande.Membership.ID {
			continue // det du alt har, er ikkje eit byte
		}

		v := Byteval{
			ID:   m.ID,
			Namn: Namn(overstyrte[m.ID], lang, MedlemskapNamn(lang, m)),
			Pris: fmt.Sprintf(t(lang, "medlemskapet.per_month"), kroneTal(m.Price)),
		}

		// Bindinga er halve svaret på om eit byte er verdt det.
		binding := t(lang, "medlemskapet.no_binding")
		if m.CommitmentMonths > 0 {
			binding = fmt.Sprintf(t(lang, "medlemskapet.binding_months"), m.CommitmentMonths)
		}

		if noverande != nil {
			skilnad := kroneTal(m.Price) - kroneTal(noverande.Membership.Price)
			switch {
			case skilnad < 0:
				v.Mindre = fmt.Sprintf(t(lang, "medlemskapet.less_per_month"), -skilnad)
				v.Setning = ", " + binding
			case skilnad > 0:
				v.Setning = fmt.Sprintf(t(lang, "medlemskapet.more_per_month"), skilnad) + ", " + binding
			default:
				v.Setning = fmt.Sprintf(t(lang, "medlemskapet.same_price"), binding)
			}
		} else {
			v.Setning = binding
		}

		// Studentprisen står der, men stengd, og grunnen er ikkje ei
		// feilmelding: ho er ei lenkje til staden der du kan gjere noko
		// med det. Ei rad som berre er borte, lærer deg ingenting.
		if m.IsStudentSenior && !kvalifisert {
			v.Stengd = true
			v.Grunn = t(lang, "medlemskapet.student_locked")
			v.VegNamn = t(lang, "medlemskapet.set_student_status")
			v.VegLenke = "/elev/min-profil"
		}

		ut = append(ut, v)
	}
	return ut
}
