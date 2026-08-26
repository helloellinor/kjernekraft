package handlers

import (
	"net/http"
	"strconv"

	"kjernekraft/models"
)

// Reglane for medlemskapsbyte.
//
// Fragmentet hentar seg sjølv inn i fana, og lagrar med den same dokka
// som prisane. Det gamle skjemaet lasta reglane med eit JSON-kall i JS
// etter at sida stod der: fyrst såg ein alle avkryssingane tomme, so
// hoppa dei på plass.

// teiknFragment skriv ut éin definisjon frå malsamlinga — ikkje ei heil
// side. Malane for ei side ber alle modulane med seg, so ein kan hente
// éin av dei ut åleine.
func teiknFragment(w http.ResponseWriter, namn string, data map[string]interface{}) {
	tm := GetTemplateManager()
	mal, ok := tm.GetTemplate("pages/admin")
	if !ok {
		tm.ReloadTemplates()
		mal, ok = tm.GetTemplate("pages/admin")
	}
	if !ok {
		http.Error(w, "malen finst ikkje", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := mal.ExecuteTemplate(w, namn, data); err != nil {
		http.Error(w, "feil ved teikning", http.StatusInternalServerError)
	}
}

// AdminReglarHandler syner reglane (GET) og lagrar dei (POST).
func AdminReglarHandler(w http.ResponseWriter, r *http.Request) {
	lang := GetLanguageFromRequest(r)

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "ugyldig skjema", http.StatusBadRequest)
			return
		}
		reglar := models.MembershipRules{
			AllowUpgrades:            r.FormValue("oppgradering") == "ja",
			CombineBindingPeriods:    r.FormValue("kombiner") == "ja",
			AllowDowngrades:          r.FormValue("nedgradering") == "ja",
			AllowChangeDuringBinding: r.FormValue("underbinding") == "ja",
		}
		if s := r.FormValue("standard"); s != "" {
			if id, err := strconv.Atoi(s); err == nil {
				reglar.DefaultMembershipID = &id
			}
		}
		if err := AdminDB.SaveMembershipRules(&reglar); err != nil {
			http.Error(w, "kunne ikkje lagre reglane", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin#prisar", http.StatusSeeOther)
		return
	}

	reglar, err := AdminDB.GetMembershipRules()
	if err != nil || reglar == nil {
		reglar = &models.MembershipRules{}
	}
	medlemskap, err := AdminDB.GetAllMemberships()
	if err != nil {
		medlemskap = nil
	}

	standard := 0
	if reglar.DefaultMembershipID != nil {
		standard = *reglar.DefaultMembershipID
	}

	teiknFragment(w, "admin_membership_rules", map[string]interface{}{
		"Lang":        lang,
		"CSRFToken":   CSRFToken(r),
		"Rules":       reglar,
		"Memberships": medlemskap,
		"StandardID":  standard,
	})
}
