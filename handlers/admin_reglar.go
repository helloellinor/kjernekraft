package handlers

import (
	"log"
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
	teiknFragmentFraa(w, "pages/admin", namn, data)
}

// teiknFragmentFraa teiknar eitt namn or malsettet aat ei sida.
//
// Eit brotstykke som vert teikna or si eigi fil kjenner berre det som
// stend i den fila. Malsettet aat ei sida hev derimot alle layout,
// komponentar og modular i seg, so `timeliste`, `dagmerke` og alt hitt
// finst. Teiknar ein utanum dette, feilar malen fyrst naar nokon ser
// etter — og htmx byter ikkje ut noko naar svaret er ein feil.
func teiknFragmentFraa(w http.ResponseWriter, side, namn string, data map[string]interface{}) {
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
		if err := DB.SaveMembershipRules(&reglar); err != nil {
			http.Error(w, "kunne ikkje lagre reglane", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin#prisar", http.StatusSeeOther)
		return
	}

	reglar, err := DB.GetMembershipRules()
	if err != nil || reglar == nil {
		reglar = &models.MembershipRules{}
	}
	medlemskap, err := DB.GetAllMemberships()
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
