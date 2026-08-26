package handlers

import "net/http"

// ArketHandler syner verkstaden: kvar komponent i stilarket, utanfor
// arbeidet sitt, ljos og myrk attmed kvarandre.
//
// Sida teiknar komponentane med den same koden som sidone nyttar, so ho
// kann ikkje verta usann utan at sidone vert det med det same. Det er
// heile skilnaden fraa ei stilbok som skildrar noko: denne syner det.
//
// Ho ligg i utviklingsbolken av rutaren. Ein verkstad er ikkje noko
// studioet syner nokon, og difor er teksti der inne heller ikkje umsett.
func ArketHandler(w http.ResponseWriter, r *http.Request) {
	lang := GetLanguageFromRequest(r)
	renderPage(w, r, "pages/arket", map[string]interface{}{
		"Title":       "Arket",
		"CurrentPage": "arket",
		"Lang":        lang,
		"CSRFToken":   CSRFToken(r),
		"IsAdmin":     sessionIsAdmin(r),
		"UserName":    sessionUserName(r),
	})
}
