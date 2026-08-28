package handlers

import (
	"net/http"
)

// BetalingHandler handles the payment methods page
func BetalingHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	user := GetUserFromSession(r)
	if user == nil {
		http.Redirect(w, r, "/innlogging", http.StatusSeeOther)
		return
	}

	// Get language from cookies/request (using new system)
	lang := GetLanguageFromRequest(r)

	// Same forma som klippekort- og medlemskapssida: tittel, ei lina som
	// seier kva som stend til, og so fanone. Korta kjem fraa den same
	// funksjonen som brotstykket teiknar, so talet og lista ikkje kann
	// segja kvar sitt.
	kort := betalingskorta(user)

	faner := []Tab{
		{Key: "korta", Name: t(lang, "payments.payment_methods")},
		{Key: "nytt-kort", Name: t(lang, "payments.add_payment_method")},
	}

	data := map[string]interface{}{
		"Tabs":        faner,
		"Label":       t(lang, "payments.title"),
		"KortTal":     len(kort),
		"Title":       "Betaling",
		"CurrentPage": "betaling",
		"UserName":    user.Name,
		"User":        user,
		"Lang":        lang,
		"CSRFToken":   CSRFToken(r),
		"IsAdmin":     sessionIsAdmin(r),
	}

	// Use the new template system
	tm := GetTemplateManager()
	if tmpl, exists := tm.GetTemplate("pages/betaling"); exists {
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
		return
	}

	// If template doesn't exist, return error
	http.Error(w, "Template not found", http.StatusInternalServerError)
}
