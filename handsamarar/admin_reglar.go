package handsamarar

import (
	"net/http"
	"strconv"

	"kjernekraft/models"
)

// Rules for switching membership. The fragment fetches itself into the
// tab and saves through the same dock as the prices.

// AdminReglar shows the rules (GET) and saves them (POST).
func (a *App) AdminReglar(w http.ResponseWriter, r *http.Request) {
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
		if err := a.DB.SaveMembershipRules(&reglar); err != nil {
			http.Error(w, "kunne ikkje lagre reglane", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin#prisar", http.StatusSeeOther)
		return
	}

	reglar, err := a.DB.GetMembershipRules()
	if err != nil || reglar == nil {
		reglar = &models.MembershipRules{}
	}
	medlemskap, err := a.DB.GetAllMemberships()
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
