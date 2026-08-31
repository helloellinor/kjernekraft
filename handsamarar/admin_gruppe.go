package handsamarar

import (
	"net/http"
	"strconv"
	"strings"
)

// Gruppone i administrasjonen.
//
// Ei gruppe er kven du høyrer til, ikkje kva du fær gjera. Difor stend
// ho ikkje i det same skjemaet som lærar- og administratorknappane: den
// dagen dei tvo ser like ut, er «gjev tilgang til reformer» og «gjer til
// administrator» den same handlingi for auga som gjer henne.

// LagGruppe lagar ei gruppe.
func (a *App) LagGruppe(w http.ResponseWriter, r *http.Request) {
	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	namn := strings.TrimSpace(r.URL.Query().Get("namn"))
	if namn == "" {
		http.Error(w, "namn is required", http.StatusBadRequest)
		return
	}
	if _, err := a.DB.LagGruppe(namn); err != nil {
		http.Error(w, "Could not create group", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// SlettGruppe tek gruppa burt og opnar timane hennar att.
func (a *App) SlettGruppe(w http.ResponseWriter, r *http.Request) {
	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	id, err := strconv.ParseInt(r.URL.Query().Get("gruppe"), 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "Invalid gruppe", http.StatusBadRequest)
		return
	}
	if err := a.DB.SlettGruppe(id); err != nil {
		http.Error(w, "Could not delete group", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Gruppemedlem slær medlemskapet av eller på.
func (a *App) Gruppemedlem(w http.ResponseWriter, r *http.Request) {
	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	gruppe, err := strconv.ParseInt(r.URL.Query().Get("gruppe"), 10, 64)
	if err != nil || gruppe == 0 {
		http.Error(w, "Invalid gruppe", http.StatusBadRequest)
		return
	}
	brukar, err := strconv.ParseInt(r.URL.Query().Get("brukar"), 10, 64)
	if err != nil || brukar == 0 {
		http.Error(w, "Invalid brukar", http.StatusBadRequest)
		return
	}
	if err := a.DB.SettGruppemedlem(gruppe, brukar, r.URL.Query().Get("paa") == "1"); err != nil {
		http.Error(w, "Could not change membership", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
