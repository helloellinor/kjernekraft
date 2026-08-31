package handsamarar

import (
	"kjernekraft/handsamarar/config"
	"kjernekraft/handsamarar/modules"
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

// UserKlippekort provides HTMX endpoint for user's klippekort display
func (a *App) UserKlippekort(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	userID := int64(user.ID)
	klippekort, err := a.DB.GetUserKlippekort(userID)
	if err != nil {
		http.Error(w, "Could not fetch user klippekort", http.StatusInternalServerError)
		log.Printf("Error fetching klippekort for user %d: %v", userID, err)
		return
	}

	klargjerKlippekort(klippekort, config.GetInstance().GetCurrentTime())

	// Get language from request (default to Norwegian bokmål)
	lang := GetLanguageFromRequest(r)

	// Create module data
	moduleData := modules.NewKlippekortModule(klippekort, lang)

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

// UserSignups provides HTMX endpoint for the classes a user is signed up for
func (a *App) UserSignups(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	lang := GetLanguageFromRequest(r)
	nå := config.GetInstance().GetCurrentTime()

	paamelde, err := a.EnrolledSessions(int64(user.ID), lang, nå)
	if err != nil {
		log.Printf("paameldingar for %d: %v", user.ID, err)
		http.Error(w, "Could not fetch user signups", http.StatusInternalServerError)
		return
	}

	teiknFragmentFrå(w, "pages/dashboard", "signed_up_classes_module", map[string]interface{}{
		"EnrolledSessions": paamelde,
		"Lang":             lang,
		"CSRFToken":        CSRFToken(r),
		"IsAdmin":          sessionIsAdmin(r),
		"UserName":         sessionUserName(r),
	})
}

// LedigPlass teiknar lista yver timar med ledig plass i dag og i
// morgon. Ho hentar seg att når du melder deg på eller av, so timen
// flyt yver i «Påmeld» utan at sida lastar på nytt.
func (a *App) LedigPlass(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	lang := GetLanguageFromRequest(r)
	nå := config.GetInstance().GetCurrentTime()

	ledige, err := a.AvailableSessions(int64(user.ID), lang, nå)
	if err != nil {
		log.Printf("ledig plass for %d: %v", user.ID, err)
		http.Error(w, "Could not fetch open classes", http.StatusInternalServerError)
		return
	}

	teiknFragmentFrå(w, "pages/dashboard", "ledig_plass_module", map[string]interface{}{
		"AvailableSessions": ledige,
		"Lang":              lang,
		"CSRFToken":         CSRFToken(r),
		"IsAdmin":           sessionIsAdmin(r),
		"UserName":          sessionUserName(r),
	})
}

// Heimehovud teiknar helsingi og briefingen om att.
//
// Han finst av di dei tvo er dei einaste tingi på heimesida som segjer
// noko om paameldingane *i ord* — «Sest i morgon tidleg, Solfrid» og «du
// stend på tri timar denne veka». Listone under dei hev alltid henta seg
// sjølve att; desse tvo stod att med det dei sa då sida kom, og laug
// difor i nett det augneblinket du melde deg på.
//
// Reknestykket bur i `Heimehovudet` og ikkje her, so sida og
// oppfriskingi ikkje kann segja kvar sitt.
func (a *App) Heimehovud(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	lang := GetLanguageFromRequest(r)
	nå := config.GetInstance().GetCurrentTime()

	paamelde, err := a.EnrolledSessions(int64(user.ID), lang, nå)
	if err != nil {
		log.Printf("paameldingar for %d: %v", user.ID, err)
	}

	klippAtt := 0
	if kort, feil := a.DB.GetUserKlippekort(int64(user.ID)); feil != nil {
		log.Printf("klippekort for %d: %v", user.ID, feil)
	} else {
		klippAtt = klargjerKlippekort(kort, nå)
	}

	ferdig, feil := a.DB.NettFrammott(int64(user.ID), nå, 30*time.Minute)
	if feil != nil {
		log.Printf("nett frammøtt for %d: %v", user.ID, feil)
	}

	helsingTittel, briefing := Heimehovudet(lang, user.Name, paamelde, klippAtt, ferdig, nå)

	teiknFragmentFrå(w, "pages/dashboard", "heimehovud_module", map[string]interface{}{
		"HelsingTittel": helsingTittel,
		"Briefing":      briefing,
		"Lang":          lang,
		"CSRFToken":     CSRFToken(r),
		"IsAdmin":       sessionIsAdmin(r),
		"UserName":      user.Name,
	})
}
