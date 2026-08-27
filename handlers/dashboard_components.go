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
	lang := GetLanguageFromRequest(r)

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

		// Calculate months until binding end
		if membership.BindingEnd != nil {
			months := now.Month()
			year := now.Year()
			bindingEndMonths := membership.BindingEnd.Month()
			bindingEndYear := membership.BindingEnd.Year()

			totalMonths := (bindingEndYear-year)*12 + int(bindingEndMonths-months)
			if membership.BindingEnd.Day() < now.Day() {
				totalMonths--
			}
			if totalMonths < 0 {
				totalMonths = 0
			}
			membership.MonthsUntilBindingEnd = totalMonths
		}

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
	lang := GetLanguageFromRequest(r)

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

// UserSignupsHandler teiknar lista yver timar du hev meldt deg paa.
//
// Brotstykket vart teikna or si eigi fil fyrr — `modules/dashboard/
// signed-up-classes` — og den fila kjenner korkje `timeliste` eller
// `dagmerke`. Malen feila kvar einaste gong, tenaren svara 500, og
// htmx byter ikkje ut noko naar svaret er ein feil. Difor stod det
// «Lastar paameldingar…» paa heimesida for alltid.
//
// Malsettet aat ei *sida* hev alle komponentar og modular i seg. Difor
// kjem brotstykket derifraa.
func UserSignupsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	lang := GetLanguageFromRequest(r)
	naa := config.GetInstance().GetCurrentTime()

	paamelde, err := PaameldeFramsyningar(int64(user.ID), lang, naa)
	if err != nil {
		log.Printf("paameldingar for %d: %v", user.ID, err)
		http.Error(w, "Could not fetch user signups", http.StatusInternalServerError)
		return
	}

	teiknFragmentFraa(w, "pages/dashboard", "signed_up_classes_module", map[string]interface{}{
		"PaameldeFramsyningar": paamelde,
		"Lang":                 lang,
		"CSRFToken":            CSRFToken(r),
		"IsAdmin":              sessionIsAdmin(r),
		"UserName":             sessionUserName(r),
	})
}

// LedigPlassHandler teiknar lista yver timar med ledig plass i dag og i
// morgon. Ho hentar seg att naar du melder deg paa eller av, so timen
// flyt yver i «Paameld» utan at sida lastar paa nytt.
func LedigPlassHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	lang := GetLanguageFromRequest(r)
	naa := config.GetInstance().GetCurrentTime()

	ledige, err := LedigeFramsyningar(int64(user.ID), lang, naa)
	if err != nil {
		log.Printf("ledig plass for %d: %v", user.ID, err)
		http.Error(w, "Could not fetch open classes", http.StatusInternalServerError)
		return
	}

	teiknFragmentFraa(w, "pages/dashboard", "ledig_plass_module", map[string]interface{}{
		"LedigeFramsyningar": ledige,
		"Lang":               lang,
		"CSRFToken":          CSRFToken(r),
		"IsAdmin":            sessionIsAdmin(r),
		"UserName":           sessionUserName(r),
	})
}
