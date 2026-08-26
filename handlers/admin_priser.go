package handlers

import (
	"net/http"
	"strconv"
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

		student := r.FormValue("student-"+id) != ""

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

	http.Redirect(w, r, "/admin#prisar", http.StatusSeeOther)
}
