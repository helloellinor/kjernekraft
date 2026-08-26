package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"kjernekraft/models"
)

// Klippekortprisane.
//
// Same forma og same handsaminga som medlemskapsprisane: skjemaet sender
// alle pakkane, og tenaren skriv berre dei radene som faktisk er ulike.
// Sjå docs/KORREKTUREN.md.
func AdminKlippeprisarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		pakkar, err := AdminDB.GetAllKlippekortPackages()
		if err != nil {
			pakkar = nil
		}
		teiknFragment(w, "admin_klippekort_prisar", map[string]interface{}{
			"Lang":      GetLanguageFromRequest(r),
			"CSRFToken": CSRFToken(r),
			"Pakkar":    pakkar,
		})
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "ugyldig skjema", http.StatusBadRequest)
		return
	}

	alle, err := DB.GetAllKlippekortPackages()
	if err != nil {
		http.Error(w, "kunne ikkje hente pakkane", http.StatusInternalServerError)
		return
	}

	tal := func(nykel string, standard int) int {
		if s := r.FormValue(nykel); s != "" {
			if n, err := strconv.Atoi(s); err == nil && n >= 0 {
				return n
			}
		}
		return standard
	}

	for _, p := range alle {
		id := strconv.Itoa(p.ID)

		if r.FormValue("slett-"+id) != "" {
			if err := DB.DeactivateKlippekortPackage(int64(p.ID)); err != nil {
				http.Error(w, "kunne ikkje slette", http.StatusInternalServerError)
				return
			}
			continue
		}

		namn := r.FormValue("namn-" + id)
		if namn == "" {
			namn = p.Name
		}
		klipp := tal("klipp-"+id, p.KlippCount)
		pris := tal("pris-"+id, p.Price/100) * 100
		dagar := tal("dagar-"+id, p.ValidDays)

		if namn == p.Name && klipp == p.KlippCount && pris == p.Price && dagar == p.ValidDays {
			continue
		}

		p.Name, p.KlippCount, p.Price, p.ValidDays = namn, klipp, pris, dagar
		if err := DB.UpdateKlippekortPackage(p); err != nil {
			http.Error(w, "kunne ikkje lagre", http.StatusInternalServerError)
			return
		}
	}

	// Nye rader.
	for nykel, verdiar := range r.Form {
		if !strings.HasPrefix(nykel, "namn-ny") || len(verdiar) == 0 {
			continue
		}
		namn := reintNamn(verdiar[0])
		if namn == "" {
			continue
		}
		suffiks := strings.TrimPrefix(nykel, "namn-")
		if _, err := DB.CreateKlippekortPackage(models.KlippekortPackage{
			Name:       namn,
			Category:   r.FormValue("kategori-" + suffiks),
			KlippCount: tal("klipp-"+suffiks, 10),
			Price:      tal("pris-"+suffiks, 0) * 100,
			ValidDays:  tal("dagar-"+suffiks, 120),
			Active:     true,
		}); err != nil {
			http.Error(w, "kunne ikkje opprette", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/admin#prisar", http.StatusSeeOther)
}
