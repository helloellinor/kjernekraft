package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// AddMembershipHandler handles adding a membership to a user
func AddMembershipHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get membership ID from form
	membershipIDStr := r.FormValue("membership_id")
	membershipID, err := strconv.ParseInt(membershipIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid membership ID", http.StatusBadRequest)
		return
	}

	userID := int64(user.ID)
	err = DB.AddUserMembership(userID, membershipID)
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

// ChangeMembershipHandler handles changing a user's membership
func ChangeMembershipHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get new membership ID from form
	membershipIDStr := r.FormValue("membership_id")
	membershipID, err := strconv.ParseInt(membershipIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid membership ID", http.StatusBadRequest)
		return
	}

	userID := int64(user.ID)
	err = DB.ChangeUserMembership(userID, membershipID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Medlemskap endret!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RemoveMembershipHandler handles removing/cancelling a user's membership
func RemoveMembershipHandler(w http.ResponseWriter, r *http.Request) {
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
	err := DB.RemoveUserMembership(userID)
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