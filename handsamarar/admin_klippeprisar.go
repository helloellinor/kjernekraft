package handsamarar

import (
	"net/http"
	"strconv"

	"kjernekraft/models"
)

// Klippekortprisane.
//
// Same forma og same handsaminga som medlemskapsprisane: skjemaet sender
// alle pakkane, og tenaren skriv berre dei radene som faktisk er ulike.
// Sjå docs/KORREKTUREN.md.
func (a *App) AdminKlippeprisar(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		lang := GetLanguageFromRequest(r)
		pakkar, err := a.DB.GetAllKlippekortPackages()
		if err != nil {
			pakkar = nil
		}
		overstyrte, err := a.DB.Produktnamn("klippekort")
		if err != nil {
			overstyrte = nil
		}
		type rad struct {
			models.KlippekortPackage
			Namn     string
			Generert bool
		}
		rader := make([]rad, 0, len(pakkar))
		for _, p := range pakkar {
			generert := KlippekortNamn(lang, p)
			namn := Namn(overstyrte[p.ID], lang, generert)
			rader = append(rader, rad{p, namn, namn == generert})
		}
		teiknFragment(w, "admin_klippekort_prisar", map[string]interface{}{
			"Lang":      lang,
			"CSRFToken": CSRFToken(r),
			"Packages":  rader,
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "ugyldig skjema", http.StatusBadRequest)
		return
	}

	lang := GetLanguageFromRequest(r)

	alle, err := a.DB.GetAllKlippekortPackages()
	if err != nil {
		http.Error(w, "kunne ikkje hente pakkane", http.StatusInternalServerError)
		return
	}

	for _, p := range alle {
		id := strconv.Itoa(p.ID)

		if r.FormValue("slett-"+id) != "" {
			if err := a.DB.SlettProduktnamn("klippekort", p.ID); err != nil {
				http.Error(w, "kunne ikkje slette namnet", http.StatusInternalServerError)
				return
			}
			if err := a.DB.DeactivateKlippekortPackage(int64(p.ID)); err != nil {
				http.Error(w, "kunne ikkje slette", http.StatusInternalServerError)
				return
			}
			continue
		}

		namn := reintNamn(r.FormValue("namn-" + id))
		if namn == KlippekortNamn(lang, p) {
			namn = ""
		}
		if err := a.DB.SetProduktnamn("klippekort", p.ID, lang, namn); err != nil {
			http.Error(w, "kunne ikkje lagre namnet", http.StatusInternalServerError)
			return
		}
		klipp := prisTal(r, "klipp-"+id, p.KlippCount)
		pris := prisTal(r, "pris-"+id, p.Price/100) * 100
		dagar := prisTal(r, "dagar-"+id, p.ValidDays)

		if klipp == p.KlippCount && pris == p.Price && dagar == p.ValidDays {
			continue
		}

		p.KlippCount, p.Price, p.ValidDays = klipp, pris, dagar
		if err := a.DB.UpdateKlippekortPackage(p); err != nil {
			http.Error(w, "kunne ikkje lagre", http.StatusInternalServerError)
			return
		}
	}

	// Nye rader.
	for suffiks, namn := range nyeRadene(r) {
		nyPakke := models.KlippekortPackage{
			Name:       namn,
			Category:   r.FormValue("kategori-" + suffiks),
			KlippCount: prisTal(r, "klipp-"+suffiks, 10),
			Price:      prisTal(r, "pris-"+suffiks, 0) * 100,
			ValidDays:  prisTal(r, "dagar-"+suffiks, 120),
			Active:     true,
		}
		nyID, err := a.DB.CreateKlippekortPackage(nyPakke)
		if err != nil {
			http.Error(w, "kunne ikkje opprette", http.StatusInternalServerError)
			return
		}
		nyPakke.ID = int(nyID)
		if namn != KlippekortNamn(lang, nyPakke) {
			if err := a.DB.SetProduktnamn("klippekort", int(nyID), lang, namn); err != nil {
				http.Error(w, "kunne ikkje lagre namnet", http.StatusInternalServerError)
				return
			}
		}
	}

	http.Redirect(w, r, "/admin#prisar", http.StatusSeeOther)
}
