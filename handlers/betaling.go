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

	data := map[string]interface{}{
		"Title":       "Betaling",
		"CurrentPage": "betaling",
		"UserName":    user.Name,
		"User":        user,
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