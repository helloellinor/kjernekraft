package handsamarar

import (
	"net/http"
	"strconv"

	"kjernekraft/handsamarar/config"
)

// Rabattkrav svarar på eit krav um studentrabatt.
//
// Eitt kall for båe svari: det er den same handlingi — nokon i studioet
// hev sett på beviset og sagt ja eller nei — og eit endepunkt per svar
// hadde vore tvo namn på ei avgjerd.
func (a *App) Rabattkrav(w http.ResponseWriter, r *http.Request) {
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

	if err := a.DB.AvgjerRabattkrav(kravID, svar == "ja",
		config.GetInstance().GetCurrentTime()); err != nil {
		http.Error(w, t(GetLanguageFromRequest(r), "admin.rabatt_borte"), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
}
