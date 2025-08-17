package handlers

import (
	"encoding/json"
	"kjernekraft/models"
	"net/http"
	"strconv"
)

// UpdateMembershipPriceHandler updates the price of a membership
func UpdateMembershipPriceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add admin authentication check here

	var requestData struct {
		MembershipID int `json:"membership_id"`
		Price        int `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := AdminDB.UpdateMembershipPrice(int64(requestData.MembershipID), requestData.Price); err != nil {
		http.Error(w, "Could not update membership price", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Membership price updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateMembershipHandler creates a new membership
func CreateMembershipHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add admin authentication check here

	var membership models.Membership
	if err := json.NewDecoder(r.Body).Decode(&membership); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set default values
	membership.Active = true

	membershipID, err := AdminDB.CreateMembership(membership)
	if err != nil {
		http.Error(w, "Could not create membership", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":       true,
		"message":       "Membership created successfully",
		"membership_id": membershipID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteMembershipHandler deactivates a membership
func DeleteMembershipHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add admin authentication check here

	membershipIDStr := r.URL.Query().Get("id")
	membershipID, err := strconv.ParseInt(membershipIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid membership ID", http.StatusBadRequest)
		return
	}

	if err := AdminDB.DeactivateMembership(membershipID); err != nil {
		http.Error(w, "Could not deactivate membership", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Membership deactivated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}