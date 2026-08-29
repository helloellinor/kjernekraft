package handlers

import (
	"encoding/json"
	"kjernekraft/models"
	"net/http"
)

// GetMembershipRulesHandler returns the current membership rules configuration
func GetMembershipRulesHandler(w http.ResponseWriter, r *http.Request) {
	rules, err := DB.GetMembershipRules()
	if err != nil {
		http.Error(w, "Could not retrieve membership rules", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

// SaveMembershipRulesHandler saves the membership rules configuration
func SaveMembershipRulesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add admin authentication check here

	var rules models.MembershipRules
	if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := DB.SaveMembershipRules(&rules); err != nil {
		http.Error(w, "Could not save membership rules", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Membership rules saved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
