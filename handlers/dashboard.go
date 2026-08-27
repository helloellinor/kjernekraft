package handlers

import (
	"kjernekraft/handlers/config"
	"kjernekraft/models"
	"log"
	"net/http"
	"time"
)

var OsloLoc *time.Location

// ElevDashboardHandler serves the Elev dashboard home page
func ElevDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user := GetUserFromSession(r)
	if user == nil {
		http.Redirect(w, r, "/innlogging", http.StatusTemporaryRedirect)
		return
	}

	settings := config.GetInstance()
	now := settings.GetCurrentTime()

	// Timar med ledig plass, i dag og i morgon.
	//
	// Bolken spurde «kva går i dag» før. Men ein time som er full er
	// ikkje eit tilbod, og i dag åleine er for kort: er klokka sju om
	// kvelden, står bolken tom kvar kveld. Han spør «kvar er det plass»
	// no, og då må i morgon vere med.
	//
	// Lista vert bygd av den same funksjonen som brotstykket nyttar, so
	// sida og ei henting midt i økta aldri kan seie kvar sitt.
	lang := GetLanguageFromRequest(r)

	ledige, err := LedigeFramsyningar(int64(user.ID), lang, now)
	if err != nil {
		http.Error(w, "Could not fetch today's events", http.StatusInternalServerError)
		return
	}

	paamelde, err := PaameldeFramsyningar(int64(user.ID), lang, now)
	if err != nil {
		log.Printf("paameldingar for %d: %v", user.ID, err)
	}

	// Den fyrste timen han hev meldt seg paa. Han ber helsingi.
	var neste *models.Event
	if len(paamelde) > 0 {
		neste = &paamelde[0].Event
	}

	// Aktiviteten: eit halvt aar attende, dag for dag.
	fraa := VikeMaandag(now, 0).AddDate(0, 0, -7*25)
	perDag, err := DB.AktivitetPerDag(int64(user.ID), fraa, now)
	if err != nil {
		log.Printf("aktivitet for %d: %v", user.ID, err)
		perDag = map[string]int{}
	}

	// Medlemskapet stend som eitt merke her; forvaltningi ligg paa
	// medlemskapssida. Feilar uppslaget, syner merket «ikkje medlem» —
	// det er ikkje sant, men det er ikkje verre enn ei tom rute, og
	// lenkja ved sida av fører til sanningi.
	medlemskap, err := DB.GetUserMembership(int64(user.ID))
	if err != nil {
		log.Printf("medlemskap for %d: %v", user.ID, err)
	}

	data := map[string]interface{}{
		"Aktivitet":            NyAktivitet(lang, perDag, now, 26),
		"Medlemskap":           medlemskap,
		"Helsing":              Helsing(lang, user.Name, neste, now),
		"Title":                "Elev Dashboard",
		"LedigeFramsyningar":   ledige,
		"PaameldeFramsyningar": paamelde,
		"ExternalCSS":          []string{},
		"CurrentPage":          "hjem",
		"UserName":             user.Name,
		"User":                 user,
		"Lang":                 lang,
		"CSRFToken":            CSRFToken(r),
		"IsAdmin":              sessionIsAdmin(r),
	}

	// Use the new template system
	tm := GetTemplateManager()
	if tmpl, exists := tm.GetTemplate("pages/dashboard"); exists {
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
		return
	}

	// If template doesn't exist, return error
	http.Error(w, "Template not found", http.StatusInternalServerError)
}
