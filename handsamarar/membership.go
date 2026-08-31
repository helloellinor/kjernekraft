package handsamarar

import (
	"kjernekraft/database"
	"kjernekraft/handsamarar/config"
	"kjernekraft/models"
	"log"
	"net/http"
)

// KlippekortPage serves the klippekort two-step selection page
func (a *App) KlippekortPage(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	// Get language from cookies/request (using new system)
	lang := GetLanguageFromRequest(r)

	// Pakkorne kjem or basen, ikkje or malen. Sjå klippekjop.go.
	pakkar, err := a.DB.GetAllKlippekortPackages()
	if err != nil {
		log.Printf("klippekortpakkar: %v", err)
	}

	// The briefing. The page has the same shape as the membership page now —
	// 	// title, a line saying how things stand, then the tabs — and the line
	// 	// here answers what a list of cards does not answer by itself: what
	// 	// expires first, and how much is left in total.
	kort, err := a.DB.GetUserKlippekort(int64(user.ID))
	if err != nil {
		log.Printf("klippekort for %d: %v", user.ID, err)
	}
	// Same reckoning the home page and the card list do. Four of its six
	// lines were pasted here, so "expires soon" existed in two places
	// that could drift apart.
	klippAtt := klargjerKlippekort(kort, config.GetInstance().GetCurrentTime())

	// Two tabs, the same two as the membership page: what you have, and what
	// 	// you can get. The charges stand beside the cards and not in a tab of
	// 	// their own — they answer the same question, and you read them
	// 	// together.
	// 	//
	// 	// The key "kjop-klipp" is the one the link from the home page already
	// 	// points at (?fane=kjop-klipp&fill=…): the choice lives in the URL, and
	// 	// the server reads it, so that link opens the right tab with no
	// 	// JavaScript involved at all.
	faner := []Tab{
		{Key: "korta", Name: t(lang, "dashboard.my_punch_cards")},
		{Key: "kjop-klipp", Name: t(lang, "klippekort.buy")},
	}

	data := sidedata(r, SidaKlippekort, "Klippekort", map[string]any{
		"Categories": Categories(pakkar),
		"Faner": fanerekkje(r, "faneark-klippekort", "fane",
			t(lang, "navigation.punch_cards"), faner),
		"Nærast":   models.NærastUtløp(kort),
		"KlippAtt": klippAtt,
		"User":     user,
	})

	renderPage(w, r, "pages/klippekort", data)
}

func (a *App) MinProfil(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	// Handle POST request for profile updates
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Update user data from form
		user.Name = r.FormValue("name")
		user.Email = r.FormValue("email")
		user.Phone = r.FormValue("phone")
		user.Address = r.FormValue("address")
		user.PostalCode = r.FormValue("postal_code")
		user.City = r.FormValue("city")
		user.Country = r.FormValue("country")
		user.Birthdate = r.FormValue("birthdate")

		// Update user in database
		err = a.DB.UpdateUser(user)
		if err != nil {
			http.Error(w, "Could not update profile", http.StatusInternalServerError)
			return
		}

		// The student-discount checkbox is a *claim*, not a discount.
		// 		//
		// 		// It was not read here at all: the form carried the field, the dock
		// 		// said "saved", and nothing happened — the surface promised a
		// 		// discount it neither gave nor asked for. Now the tick stands as a
		// 		// claim until somebody at the studio has seen the card (see
		// 		// database/rabattkrav.go).
		// 		//
		// 		// Unticking is "I am not a student any more", and it takes both the
		// 		// waiting claim and any discount already granted.
		bedd := r.FormValue("student_senior") != ""
		har, _, err := a.DB.StudentrabattFor(int64(user.ID))
		if err == nil {
			if bedd && !har {
				if err := a.DB.LagRabattkrav(int64(user.ID), config.GetInstance().GetCurrentTime()); err != nil {
					log.Printf("rabattkrav: %v", err)
				}
			} else if !bedd {
				if err := a.DB.TrekkRabattkrav(int64(user.ID)); err != nil {
					log.Printf("rabattkrav trekt: %v", err)
				}
			}
		}

		// Update session with new user data
		err = SetUserInSession(w, r, user)
		if err != nil {
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		// Redirect to avoid re-submission on refresh
		http.Redirect(w, r, "/elev/min-profil?updated=true", http.StatusSeeOther)
		return
	}

	// Handle GET request
	showSuccess := r.URL.Query().Get("updated") == "true"

	// The state of the student discount. The template carried .StudentSenior
	// 	// already, but the data map never set it — so the tick stood empty
	// 	// whatever happened, and whoever had the discount saw nowhere that they
	// 	// had it.
	rabattHar, rabattTil, _ := a.DB.StudentrabattFor(int64(user.ID))
	krav, _ := a.DB.SisteRabattkrav(int64(user.ID))
	rabattVentar := krav != nil && krav.Ventar()
	rabattAvvist := krav != nil && krav.Stoda == database.RabattAvvist

	data := sidedata(r, SidaProfil, "Min profil", map[string]any{
		"User":        user,
		"ID":          user.ID,
		"Name":        user.Name,
		"Email":       user.Email,
		"JoinDate":    "1. januar 2024", // TODO: Add join date to user model
		"Phone":       user.Phone,
		"Address":     user.Address,
		"PostalCode":  user.PostalCode,
		"City":        user.City,
		"Country":     user.Country,
		"Birthdate":   user.Birthdate,
		"ShowSuccess": showSuccess,
		// The tick stands for what you have *asked for* as much as for what you
		// 		// have got: a claim that is waiting must not look as though you
		// 		// changed your mind.
		"StudentSenior": rabattHar || rabattVentar,
		"RabattVentar":  rabattVentar,
		"RabattAvvist":  rabattAvvist,
		"RabattTil":     rabattTil,
	})

	renderPage(w, r, "pages/min-profil", data)
}

// TestDataPage shows the test-data tool page.
func (a *App) TestDataPage(w http.ResponseWriter, r *http.Request) {
	renderPage(w, r, "pages/testdata", sidedata(r, SidaTestdata, "Testdata", nil))
}
