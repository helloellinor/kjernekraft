package handsamarar

import (
	"encoding/json"
	"kjernekraft/handsamarar/config"
	"net/http"
	"strconv"
)

// AddMembership handles adding a membership to a user
func (a *App) AddMembership(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	// Get membership ID from form
	membershipID, ok := skjemaTal(w, r, "membership_id")
	if !ok {
		return
	}

	userID := int64(user.ID)
	err := a.DB.AddUserMembership(userID, membershipID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Medlemskap lagt til!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ChangeMembership handles changing a user's membership
func (a *App) ChangeMembership(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	// Get new membership ID from form
	membershipID, ok := skjemaTal(w, r, "membership_id")
	if !ok {
		return
	}

	userID := int64(user.ID)

	// Check if membership change is allowed
	canChange, reason := a.DB.CanChangeMembership(userID, membershipID, config.GetInstance().GetCurrentTime())
	if !canChange {
		svarFeilMelding(w, r, http.StatusBadRequest, reason)
		return
	}

	if err := a.DB.ChangeUserMembership(userID, membershipID); err != nil {
		svarFeilMelding(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Eit byte er ferdig naar du ser det du hev bytt til.
	//
	// Handsamaren svara med JSON fyrr, og skjemaet i malen er eit vanleg
	// `<form method="post">`. Det tyder at nettlesaren *navigerte* til
	// /api/membership/change og teikna svaret: du trykte «Byt til
	// Månadskort» og hamna paa ei kvit sida med
	// `{"success":true,"message":"Medlemskap endret!"}`. Gjekk det ikkje,
	// fekk du ei like naki side med grunnen paa bokmaal. Ingen veg
	// attende utan aa trykkja attende.
	//
	// No er svaret ei adressa. `HX-Redirect` er det htmx ventar; ein
	// vanleg 303 er det nettlesaren gjer utan skript. Baae fører til
	// medlemskapsfana, som er der det nye medlemskapet stend.
	sida := "/elev/medlemskap?fane=medlemskapet"
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", sida)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, sida, http.StatusSeeOther)
}

// RemoveMembership handles removing/cancelling a user's membership
func (a *App) RemoveMembership(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	userID := int64(user.ID)
	err := a.DB.RemoveUserMembership(userID)
	if err != nil {
		http.Error(w, "Could not remove membership", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Medlemskap avsluttet!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CanChangeMembership checks if a user can change to a specific membership
func (a *App) CanChangeMembership(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	// Get membership ID from query parameter
	membershipIDStr := r.URL.Query().Get("membership_id")
	membershipID, err := strconv.ParseInt(membershipIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid membership ID", http.StatusBadRequest)
		return
	}

	userID := int64(user.ID)
	canChange, reason := a.DB.CanChangeMembership(userID, membershipID, config.GetInstance().GetCurrentTime())
	if !canChange {
		http.Error(w, reason, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// PurchaseKlippekort handles purchasing klippekort packages
func (a *App) PurchaseKlippekort(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	// Get package ID from form
	packageIDStr := r.FormValue("package_id")
	packageID, err := strconv.ParseInt(packageIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid package ID", http.StatusBadRequest)
		return
	}

	userID := int64(user.ID)
	err = a.DB.PurchaseKlippekort(userID, packageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Klippekort kjøpt!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
