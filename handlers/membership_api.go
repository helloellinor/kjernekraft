package handlers

import (
	"encoding/json"
	"net/http"
)

// FreezeMembershipHandler handles membership freeze requests
func FreezeMembershipHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := int64(user.ID)
	err := DB.UpdateMembershipStatus(userID, "freeze_requested")
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

// CancelFreezeRequestHandler handles cancellation of freeze requests
func CancelFreezeRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := int64(user.ID)
	err := DB.UpdateMembershipStatus(userID, "active")
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

// UnfreezeMembershipHandler handles membership unfreeze requests
func UnfreezeMembershipHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := int64(user.ID)
	err := DB.UpdateMembershipStatus(userID, "active")
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