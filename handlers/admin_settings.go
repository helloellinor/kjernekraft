package handlers

import (
	"encoding/json"
	"kjernekraft/handlers/config"
	"net/http"
)

// AdminSettingsHandler handles admin settings management
func AdminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getSettings(w, r)
	case "POST":
		updateSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getSettings returns current settings
func getSettings(w http.ResponseWriter, r *http.Request) {
	settings := config.GetInstance()
	
	response := map[string]interface{}{
		"timezone": settings.GetTimezone(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// updateSettings updates application settings
func updateSettings(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Timezone string `json:"timezone"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	settings := config.GetInstance()
	
	if request.Timezone != "" {
		if err := settings.SetTimezone(request.Timezone); err != nil {
			http.Error(w, "Invalid timezone: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Settings updated successfully",
	})
}