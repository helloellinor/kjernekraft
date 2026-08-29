package handlers

import (
	"kjernekraft/handlers/config"
	"kjernekraft/handlers/modules"
	"kjernekraft/models"
	"log"
	"net/http"
	"time"
)

// klargjerKlippekort reknar felta kortmalen og dashboard-briefingen
// treng frå éi og same liste. Heimesida lasta korta, talde klippa og
// sende så ei ny henting for å teikna dei same korta.
func klargjerKlippekort(klippekort []models.KlippekortWithDetails, now time.Time) int {
	klippAtt := 0
	for i := range klippekort {
		k := &klippekort[i]
		if k.TotalKlipp > 0 {
			k.ProgressPercentage = (k.RemainingKlipp * 100) / k.TotalKlipp
		}
		k.KlipteHol = models.KlipteHolAv(k.TotalKlipp-k.RemainingKlipp, k.TotalKlipp)
		k.DaysUntilExpiry = int(k.ExpiryDate.Sub(now).Hours() / 24)
		k.IsExpiring = k.DaysUntilExpiry <= 30 && k.DaysUntilExpiry > 0
		if k.RemainingKlipp > 0 && k.ExpiryDate.After(now) {
			klippAtt += k.RemainingKlipp
		}
	}
	return klippAtt
}

// UserKlippekortHandler provides HTMX endpoint for user's klippekort display
func UserKlippekortHandler(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	userID := int64(user.ID)
	klippekort, err := DB.GetUserKlippekort(userID)
	if err != nil {
		http.Error(w, "Could not fetch user klippekort", http.StatusInternalServerError)
		log.Printf("Error fetching klippekort for user %d: %v", userID, err)
		return
	}

	klargjerKlippekort(klippekort, config.GetInstance().GetCurrentTime())

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

// UserSignupsHandler provides HTMX endpoint for the classes a user is signed up for
func UserSignupsHandler(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	lang := GetLanguageFromRequest(r)
	naa := config.GetInstance().GetCurrentTime()

	paamelde, err := EnrolledSessions(int64(user.ID), lang, naa)
	if err != nil {
		log.Printf("paameldingar for %d: %v", user.ID, err)
		http.Error(w, "Could not fetch user signups", http.StatusInternalServerError)
		return
	}

	teiknFragmentFraa(w, "pages/dashboard", "signed_up_classes_module", map[string]interface{}{
		"EnrolledSessions": paamelde,
		"Lang":             lang,
		"CSRFToken":        CSRFToken(r),
		"IsAdmin":          sessionIsAdmin(r),
		"UserName":         sessionUserName(r),
	})
}

// LedigPlassHandler teiknar lista yver timar med ledig plass i dag og i
// morgon. Ho hentar seg att naar du melder deg paa eller av, so timen
// flyt yver i «Paameld» utan at sida lastar paa nytt.
func LedigPlassHandler(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	lang := GetLanguageFromRequest(r)
	naa := config.GetInstance().GetCurrentTime()

	ledige, err := AvailableSessions(int64(user.ID), lang, naa)
	if err != nil {
		log.Printf("ledig plass for %d: %v", user.ID, err)
		http.Error(w, "Could not fetch open classes", http.StatusInternalServerError)
		return
	}

	teiknFragmentFraa(w, "pages/dashboard", "ledig_plass_module", map[string]interface{}{
		"AvailableSessions": ledige,
		"Lang":              lang,
		"CSRFToken":         CSRFToken(r),
		"IsAdmin":           sessionIsAdmin(r),
		"UserName":          sessionUserName(r),
	})
}

// HeimehovudHandler teiknar helsingi og briefingen om att.
//
// Han finst av di dei tvo er dei einaste tingi paa heimesida som segjer
// noko om paameldingane *i ord* — «Sest i morgon tidleg, Solfrid» og «du
// stend paa tri timar denne veka». Listone under dei hev alltid henta seg
// sjølve att; desse tvo stod att med det dei sa daa sida kom, og laug
// difor i nett det augneblinket du melde deg paa.
//
// Reknestykket bur i `Heimehovudet` og ikkje her, so sida og
// oppfriskingi ikkje kann segja kvar sitt.
func HeimehovudHandler(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	lang := GetLanguageFromRequest(r)
	naa := config.GetInstance().GetCurrentTime()

	paamelde, err := EnrolledSessions(int64(user.ID), lang, naa)
	if err != nil {
		log.Printf("paameldingar for %d: %v", user.ID, err)
	}

	klippAtt := 0
	if kort, feil := DB.GetUserKlippekort(int64(user.ID)); feil != nil {
		log.Printf("klippekort for %d: %v", user.ID, feil)
	} else {
		klippAtt = klargjerKlippekort(kort, naa)
	}

	ferdig, feil := DB.NettFrammott(int64(user.ID), naa, 30*time.Minute)
	if feil != nil {
		log.Printf("nett frammøtt for %d: %v", user.ID, feil)
	}

	helsingTittel, briefing := Heimehovudet(lang, user.Name, paamelde, klippAtt, ferdig != nil, naa)

	teiknFragmentFraa(w, "pages/dashboard", "heimehovud_module", map[string]interface{}{
		"HelsingTittel": helsingTittel,
		"Briefing":      briefing,
		"Lang":          lang,
		"CSRFToken":     CSRFToken(r),
		"IsAdmin":       sessionIsAdmin(r),
		"UserName":      user.Name,
	})
}
