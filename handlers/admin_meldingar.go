package handlers

import (
	"net/http"
	"strconv"

	"kjernekraft/handlers/config"
)

// MeldingHandsamaHandler tek ei melding ut or fana.
//
// Ho vert ikkje sletta. Ei tom fana skal tyda «ingen ting ventar» og
// ikkje «ingen ting hev hendt», og den dagen ein e-postsendar kjem, er
// det den same rada han skal sjaa paa — `sendt` er hans kolonne, denne
// er menneska si.
func MeldingHandsamaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	id, err := strconv.ParseInt(r.URL.Query().Get("melding_id"), 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "Invalid melding_id", http.StatusBadRequest)
		return
	}

	if err := AdminDB.HandsamaMelding(id, config.GetInstance().GetCurrentTime()); err != nil {
		http.Error(w, "Could not update melding", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
