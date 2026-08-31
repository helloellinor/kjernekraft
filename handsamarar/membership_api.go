package handsamarar

import (
	"encoding/json"
	"net/http"
)

// FreezeMembership handles membership freeze requests
func (a *App) FreezeMembership(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	userID := int64(user.ID)
	err := a.DB.UpdateMembershipStatus(userID, "freeze_requested")
	if err != nil {
		http.Error(w, "Could not freeze membership", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Forespørsel om frysing er sendt!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CancelFreezeRequest handles cancellation of freeze requests
func (a *App) CancelFreezeRequest(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	userID := int64(user.ID)
	err := a.DB.UpdateMembershipStatus(userID, "active")
	if err != nil {
		http.Error(w, "Could not cancel freeze request", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Forespørsel om frysing er trukket tilbake!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UnfreezeMembership handles membership unfreeze requests
func (a *App) UnfreezeMembership(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	// UnfreezeMembership og ikkje UpdateMembershipStatus: det er han som
	// skuver utlaupet fram med den tidi medlemskapet stod frose. Sjå
	// database/database.go.
	userID := int64(user.ID)
	err := a.DB.UnfreezeMembership(userID)
	if err != nil {
		http.Error(w, "Could not unfreeze membership", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Medlemskapet er reaktivert!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
