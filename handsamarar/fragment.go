package handsamarar

import (
	"log"
	"net/http"
)

// teiknFragment renders one named template from the admin page's set.
func teiknFragment(w http.ResponseWriter, namn string, data map[string]interface{}) {
	teiknFragmentFrå(w, "pages/admin", namn, data)
}

// teiknFragmentFrå renders one named template from a page's template set.
//
// It must be the page's set, not the fragment's own file: only the page
// carries the layouts, components and modules a fragment may reference.
// Render outside it and the failure is silent — htmx swaps nothing in
// when the response is an error.
func teiknFragmentFrå(w http.ResponseWriter, side, namn string, data map[string]interface{}) {
	tm := GetTemplateManager()
	mal, ok := tm.GetTemplate(side)
	if !ok {
		tm.ReloadTemplates()
		mal, ok = tm.GetTemplate(side)
	}
	if !ok {
		http.Error(w, "malen finst ikkje", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := mal.ExecuteTemplate(w, namn, data); err != nil {
		log.Printf("feil ved teikning av %s or %s: %v", namn, side, err)
		http.Error(w, "feil ved teikning", http.StatusInternalServerError)
	}
}
