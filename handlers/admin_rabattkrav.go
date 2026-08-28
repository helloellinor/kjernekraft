package handlers

import (
	"net/http"
	"strconv"

	"kjernekraft/handlers/config"
)

// RabattkravHandler svarar paa eit krav um studentrabatt.
//
// Eitt kall for baae svari: det er den same handlingi — nokon i studioet
// hev sett paa beviset og sagt ja eller nei — og eit endepunkt per svar
// hadde vore tvo namn paa ei avgjerd.
func RabattkravHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	kravID, err := strconv.ParseInt(r.URL.Query().Get("krav_id"), 10, 64)
	if err != nil || kravID == 0 {
		http.Error(w, "Invalid krav_id", http.StatusBadRequest)
		return
	}
	svar := r.URL.Query().Get("svar")
	if svar != "ja" && svar != "nei" {
		http.Error(w, "Invalid svar (expected ja or nei)", http.StatusBadRequest)
		return
	}

	if err := AdminDB.AvgjerRabattkrav(kravID, svar == "ja",
		config.GetInstance().GetCurrentTime()); err != nil {
		http.Error(w, t(GetLanguageFromRequest(r), "admin.rabatt_borte"), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
}
