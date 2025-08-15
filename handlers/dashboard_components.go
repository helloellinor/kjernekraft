package handlers

import (
	"kjernekraft/handlers/config"
	"kjernekraft/handlers/modules"
	"log"
	"net/http"
)

// UserKlippekortHandler provides HTMX endpoint for user's klippekort display
func UserKlippekortHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := int64(user.ID)
	klippekort, err := DB.GetUserKlippekort(userID)
	if err != nil {
		http.Error(w, "Could not fetch user klippekort", http.StatusInternalServerError)
		log.Printf("Error fetching klippekort for user %d: %v", userID, err)
		return
	}

	// Calculate additional fields for display
	for i := range klippekort {
		k := &klippekort[i]

		// Calculate progress percentage (remaining klipps)
		if k.TotalKlipp > 0 {
			k.ProgressPercentage = (k.RemainingKlipp * 100) / k.TotalKlipp
		}

		// Calculate days until expiry
		settings := config.GetInstance()
		now := settings.GetCurrentTime()
		k.DaysUntilExpiry = int(k.ExpiryDate.Sub(now).Hours() / 24)
		k.IsExpiring = k.DaysUntilExpiry <= 30 && k.DaysUntilExpiry > 0
	}

	// Get language from request (default to Norwegian bokmål)
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "nb"
	}

	// Create module data
	moduleData, err := modules.NewKlippekortModule(klippekort, lang)
	if err != nil {
		http.Error(w, "Error creating module", http.StatusInternalServerError)
		return
	}

	// Get template manager and render
	tm := GetTemplateManager()
	tmpl, exists := tm.GetTemplate("modules/membership/klippekort")
	if !exists {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(w, "klippekort_module", moduleData); err != nil {
		log.Printf("Error executing klippekort template: %v", err)
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

// UserMembershipHandler provides HTMX endpoint for user's membership display
func UserMembershipHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := int64(user.ID)
	membership, err := DB.GetUserMembership(userID)
	if err != nil {
		http.Error(w, "Could not fetch user membership", http.StatusInternalServerError)
		log.Printf("Error fetching membership for user %d: %v", userID, err)
		return
	}

	// Calculate additional fields if membership exists
	if membership != nil {
		settings := config.GetInstance()
		now := settings.GetCurrentTime()
		membership.DaysUntilRenewal = int(membership.RenewalDate.Sub(now).Hours() / 24)

		// Business logic for what actions are available
		membership.CanPause = membership.Status == "active"

		// Can cancel if no binding period OR if binding period has ended
		if membership.BindingEnd == nil {
			membership.CanCancel = true
		} else {
			membership.CanCancel = now.After(*membership.BindingEnd)
		}
	}

	// Get language from request (default to Norwegian bokmål)
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "nb"
	}

	// Create module data
	moduleData, err := modules.NewMembershipModule(membership, lang)
	if err != nil {
		http.Error(w, "Error creating module", http.StatusInternalServerError)
		return
	}

	// Get template manager and render
	tm := GetTemplateManager()
	tmpl, exists := tm.GetTemplate("modules/membership/membership")
	if !exists {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(w, "membership_module", moduleData); err != nil {
		log.Printf("Error executing membership template: %v", err)
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}