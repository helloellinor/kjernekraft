package handsamarar

import (
	"net/http"
	"strconv"

	"kjernekraft/models"
)

// AdminPrisar lagrar prisane.
//
// Skjemaet sender alle medlemskapa, ikkje berre dei endra: kva som er
// endra, er noko nettlesaren veit og tenaren ikkje treng vite. Han
// samanliknar sjølv og skriv berre dei radene som faktisk er ulike, so
// eit lagre utan endringar ikkje rører databasen.
// Prisrad er ei rad slik ho vert teikna: namnet er skrive av systemet
// eller overstyrt, og fakta er fakta.
type Prisrad struct {
	ID       int
	Namn     string
	Generert bool
	Pris     int
	Binding  int
	Student  bool
}

func (a *App) AdminPrisar(w http.ResponseWriter, r *http.Request) {
	lang := GetLanguageFromRequest(r)

	if r.Method == http.MethodGet {
		alle, err := a.DB.GetAllMemberships()
		if err != nil {
			alle = nil
		}
		overstyrte, err := a.DB.Produktnamn("medlemskap")
		if err != nil {
			overstyrte = nil
		}
		rader := make([]Prisrad, 0, len(alle))
		for _, m := range alle {
			// Det svarte kortet stend ikkje her. Det er ikkje eit
			// produkt: det fylgjer ei rolla, det vert betalt med Karma,
			// og prisen er null av di det ikkje *hev* ein pris. Ei rad
			// i prislista seier at nokon kann setja talet, og det kann
			// dei ikkje. `Skjult` er alt flagget som segjer at
			// medlemskapet ikkje stend i lista der folk vel (sjå
			// database/svartmedlem.go); det gjeld her med.
			if m.Skjult {
				continue
			}
			generert := MedlemskapNamn(lang, m)
			namn := Namn(overstyrte[m.ID], lang, generert)
			rader = append(rader, Prisrad{
				ID: m.ID, Namn: namn, Generert: namn == generert,
				Pris: kronerFromOre(m.Price), Binding: m.CommitmentMonths,
				Student: m.IsStudentSenior,
			})
		}
		teiknFragment(w, "admin_medlemskapsprisar", map[string]interface{}{
			"Lang": lang, "CSRFToken": CSRFToken(r), "Rader": rader,
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "ugyldig skjema", http.StatusBadRequest)
		return
	}

	alle, err := a.DB.GetAllMemberships()
	if err != nil {
		http.Error(w, "kunne ikkje hente medlemskapa", http.StatusInternalServerError)
		return
	}

	for _, m := range alle {
		// Same grunnen som i GET-greini: eit medlemskap som ikkje kann
		// teiknast i lista kann heller ikkje skrivast frå henne. Utan
		// dette hadde eit skjema som sende id-en likevel — handskrive
		// eller frae ei gamal fana — fenge lov aa setja prisen på Svart.
		if m.Skjult {
			continue
		}
		id := strconv.Itoa(m.ID)

		// Slett fyrst: er raden merkt, er det ingen vits i å skrive
		// endringar i han fyrst.
		if r.FormValue("slett-"+id) != "" {
			// Mjuk sletting: medlemskapet vert sett uverksamt. Nokon kan
			// ha det, og då skal historikken deira ikkje forsvinne.
			if err := a.DB.SlettProduktnamn("medlemskap", m.ID); err != nil {
				http.Error(w, "kunne ikkje slette namnet", http.StatusInternalServerError)
				return
			}
			if err := a.DB.DeactivateMembership(int64(m.ID)); err != nil {
				http.Error(w, "kunne ikkje slette", http.StatusInternalServerError)
				return
			}
			continue
		}

		// Namnet bur ikkje i denne tabellen lenger. Er det som står i
		// feltet det same som systemet ville skrive, er det inga
		// overstyring — og tømer du feltet, tek du overstyringa bort.
		namn := reintNamn(r.FormValue("namn-" + id))
		generert := MedlemskapNamn(lang, m)
		if namn == generert {
			namn = ""
		}

		// Prisen står i kroner i skjemaet og i øre i basen. Feltet er det
		// ein les og skriv; øre er det systemet reknar i.
		// -1 means nobody typed anything, so the price stays untouched in
		// øre. A kroner default would round the øre remainder away, and
		// one membership in the database is not whole kroner.
		pris := m.Price
		if kr := prisTal(r, "pris-"+id, -1); kr >= 0 {
			pris = kr * 100
		}
		binding := prisTal(r, "binding-"+id, m.CommitmentMonths)

		// Rabatten er ikkje ei avkryssing lenger, men eit val i setninga:
		// kven medlemskapet gjeld for.
		student := r.FormValue("gjeld-"+id) == "student"

		if err := a.DB.SetProduktnamn("medlemskap", m.ID, lang, namn); err != nil {
			http.Error(w, "kunne ikkje lagre namnet", http.StatusInternalServerError)
			return
		}

		if pris == m.Price && binding == m.CommitmentMonths &&
			student == m.IsStudentSenior {
			continue
		}

		m.Price = pris
		m.CommitmentMonths = binding
		m.IsStudentSenior = student
		if err := a.DB.UpdateMembershipDetails(m); err != nil {
			http.Error(w, "kunne ikkje lagre", http.StatusInternalServerError)
			return
		}
	}

	// Nye rader. Dei kjem frå den same lista og den same dokka som
	// resten — ein legg til der ein endrar, og ikkje på eit eige skjema
	// ein annan stad.
	for suffiks, namn := range nyeRadene(r) {
		pris := prisTal(r, "pris-"+suffiks, 0) * 100
		binding := prisTal(r, "binding-"+suffiks, 0)
		student := r.FormValue("gjeld-"+suffiks) == "student"
		ny := models.Membership{
			Name:             namn,
			Price:            pris,
			CommitmentMonths: binding,
			IsStudentSenior:  student,
			Active:           true,
		}
		id, err := a.DB.CreateMembership(ny)
		if err != nil {
			http.Error(w, "kunne ikkje opprette", http.StatusInternalServerError)
			return
		}
		// Skreiv dei noko anna enn det systemet ville skrive, er det
		// eit namn dei meiner — og det gjeld språket dei står i.
		ny.ID = int(id)
		if namn != MedlemskapNamn(lang, ny) {
			if err := a.DB.SetProduktnamn("medlemskap", int(id), lang, namn); err != nil {
				http.Error(w, "kunne ikkje lagre namnet", http.StatusInternalServerError)
				return
			}
		}
	}

	http.Redirect(w, r, "/admin#prisar", http.StatusSeeOther)
}
