package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"kjernekraft/models"
)

// AdminPriserHandler lagrar prisane.
//
// Skjemaet sender alle medlemskapa, ikkje berre dei endra: kva som er
// endra, er noko nettlesaren veit og tenaren ikkje treng vite. Han
// samanliknar sjølv og skriv berre dei radene som faktisk er ulike, so
// eit lagre utan endringar ikkje rører databasen.
func AdminPriserHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "ugyldig skjema", http.StatusBadRequest)
		return
	}

	alle, err := DB.GetAllMemberships()
	if err != nil {
		http.Error(w, "kunne ikkje hente medlemskapa", http.StatusInternalServerError)
		return
	}

	for _, m := range alle {
		id := strconv.Itoa(m.ID)

		namn := r.FormValue("namn-" + id)
		if namn == "" {
			namn = m.Name
		}

		// Prisen står i kroner i skjemaet og i øre i basen. Feltet er det
		// ein les og skriv; øre er det systemet reknar i.
		pris := m.Price
		if s := r.FormValue("pris-" + id); s != "" {
			if kr, err := strconv.Atoi(s); err == nil && kr >= 0 {
				pris = kr * 100
			}
		}

		binding := m.CommitmentMonths
		if s := r.FormValue("binding-" + id); s != "" {
			if md, err := strconv.Atoi(s); err == nil && md >= 0 {
				binding = md
			}
		}

		// Rabatten er ikkje ei avkryssing lenger, men eit val i setninga:
		// kven medlemskapet gjeld for.
		student := r.FormValue("gjeld-"+id) == "student"

		if namn == m.Name && pris == m.Price && binding == m.CommitmentMonths &&
			student == m.IsStudentSenior {
			continue
		}

		m.Name = namn
		m.Price = pris
		m.CommitmentMonths = binding
		m.IsStudentSenior = student
		if err := DB.UpdateMembershipDetails(m); err != nil {
			http.Error(w, "kunne ikkje lagre", http.StatusInternalServerError)
			return
		}
	}

	// Nye rader. Dei kjem frå den same lista og den same dokka som
	// resten — ein legg til der ein endrar, og ikkje på eit eige skjema
	// ein annan stad.
	for nykel, verdiar := range r.Form {
		if !strings.HasPrefix(nykel, "namn-ny") || len(verdiar) == 0 {
			continue
		}
		namn := strings.TrimSpace(verdiar[0])
		if namn == "" {
			continue // ein tom rad er ein rad nokon ombestemte seg om
		}
		suffiks := strings.TrimPrefix(nykel, "namn-")

		pris := 0
		if kr, err := strconv.Atoi(r.FormValue("pris-" + suffiks)); err == nil && kr >= 0 {
			pris = kr * 100
		}
		binding := 0
		if md, err := strconv.Atoi(r.FormValue("binding-" + suffiks)); err == nil && md >= 0 {
			binding = md
		}

		if _, err := DB.CreateMembership(models.Membership{
			Name:             namn,
			Price:            pris,
			CommitmentMonths: binding,
			IsStudentSenior:  r.FormValue("gjeld-"+suffiks) == "student",
			Active:           true,
		}); err != nil {
			http.Error(w, "kunne ikkje opprette", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/admin#prisar", http.StatusSeeOther)
}
