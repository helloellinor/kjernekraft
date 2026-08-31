package handsamarar

import (
	"net/http"
	"time"

	"kjernekraft/models"
)

// merkeProvor lagar timane verkstaden syner. Dei er ikkje teikningar av
// merket: dei gjeng gjenom NyFramsyning som alt anna, so ser du noko
// rart her, ser ein brukar det same i timeplanen.
func merkeVika() (time.Time, time.Time, []models.Event) {
	// A fixed week, so the samples do not change while you look at them.
	// 	// Monday is "today", and the time is ten to twelve.
	måndag := time.Date(2026, 8, 24, 0, 0, 0, 0, OsloLoc)
	nå := måndag.Add(11*time.Hour + 50*time.Minute)
	iDag := måndag.Format("2006-01-02")

	slag := []struct {
		dag           int
		frå, til      string
		teke, plassar int
	}{
		{0, "07:15", "08:15", 7, 18},
		{1, "09:00", "10:00", 11, 18},
		{2, "12:30", "13:15", 4, 18},
		{3, "17:30", "18:30", 18, 18},
		{4, "20:00", "21:30", 9, 18},
		{5, "10:00", "11:00", 6, 8},
		{6, "16:45", "17:45", 16, 18},
	}

	titlar := []struct{ tittel, lærar, rom string }{
		{"Vinyasa Flow", "Gry", "Salen"},
		{"Reformer", "Ida", "Reformer"},
		{"Fascia Flyt", "Torunn", "Salen"},
		{"Yin", "Kristina", "Salen"},
		{"Vinyasa Flow", "Gry", "Salen"},
		{"Reformer", "Ida", "Reformer"},
		{"Yin", "Kristina", "Salen"},
	}

	ut := make([]models.Event, 0, len(slag))
	for i, s := range slag {
		start, _ := time.ParseInLocation("2006-01-02 15:04",
			måndag.AddDate(0, 0, s.dag).Format("2006-01-02")+" "+s.frå, OsloLoc)
		slutt, _ := time.ParseInLocation("2006-01-02 15:04",
			måndag.AddDate(0, 0, s.dag).Format("2006-01-02")+" "+s.til, OsloLoc)
		ut = append(ut, models.Event{
			ID:               900 + i,
			Title:            titlar[i].tittel,
			TeacherName:      titlar[i].lærar,
			RoomName:         titlar[i].rom,
			StartTime:        start,
			EndTime:          slutt,
			Capacity:         s.plassar,
			CurrentEnrolment: s.teke,
		})
	}
	_ = iDag
	return måndag, nå, ut
}

// medlemskortProvor builds the membership cards the workshop shows.
//
// The workshop used to draw a *different* card here: .membership-card with
// .membership-header, .price-info and .dates-info, written by hand in the
// template. None of those classes existed anywhere else — the pages have
// always used the medlemskort template. So the workshop was showing a
// component that does not exist, which is worse than showing nothing: ARKET
// §9 says what you do not draw here drifts, and here the drawing itself had
// drifted.
//
// Now the cards go through the same template the pages use, with the four
// states user_memberships.status can hold, plus the student rate and an
// assigned membership — the only two things that change what the card
// carries without being a state.
func medlemskortProvor(lang string) []map[string]interface{} {
	// Fixed dates, so the samples do not change while you look at them.
	medlemSidan := time.Date(2024, 3, 11, 0, 0, 0, 0, OsloLoc)
	// Neste fornying er alltid ein maanad fram (database.go), og
	// bindingi er start + commitment_months. Verkstaden lyt syna båe
	// greinene av «Gjeld til», so tali her er dei som fell ut av eit
	// kjøp gjort 11. mars.
	nesteFornying := time.Date(2026, 9, 11, 0, 0, 0, 0, OsloLoc)
	aaretUt := time.Date(2027, 3, 11, 0, 0, 0, 0, OsloLoc)

	kort := func(namn, stoda string, pris, binding int, student, tildelt bool) models.MembershipWithDetails {
		m := models.MembershipWithDetails{}
		m.Name = namn
		m.Price = pris
		m.CommitmentMonths = binding
		m.IsStudentSenior = student
		m.Status = stoda
		m.RenewalDate = nesteFornying
		if binding > 0 {
			m.BindingEnd = &aaretUt
		}
		m.Tildelt = tildelt
		return m
	}

	provor := []struct {
		namn     string
		kort     models.MembershipWithDetails
		handling bool
		seiaUpp  bool
	}{
		// Aarskort: bindingi ber «Gjeld til», og ho stend eit år fram.
		{"Klas", kort("Klas", "active", 69000, 12, false, false), true, true},
		{"Klas", kort("Klas", "freeze_requested", 69000, 12, false, false), true, false},
		{"Klas", kort("Klas", "paused", 69000, 12, false, false), true, false},
		// Uppsagt aarskort: bindingi stend att i basen, men kortet varer
		// berre ut det som er betalt. Fornyingi, ikkje bindingi.
		{"Klas", kort("Klas", "cancelled", 69000, 12, false, false), false, false},
		// Maanadskort: ingi binding, so fornyingi ber henne — ein maanad fram.
		{"Maanadskort", kort("Maanadskort", "active", 49000, 0, false, false), false, false},
		{"Klas student", kort("Klas student", "active", 49000, 0, true, false), false, false},
		// Svart fylgjer ei rolla og gjeng ikkje ut: ∞.
		{"Svart", kort("Svart", "active", 0, 0, false, true), true, false},
	}

	ut := make([]map[string]interface{}, 0, len(provor))
	for _, p := range provor {
		ut = append(ut, map[string]interface{}{
			"Medlemskap":  p.kort,
			"Namn":        p.namn,
			"Lang":        lang,
			"UserName":    "Ellinor Linnea",
			"MedlemSidan": &medlemSidan,
			"Handlingar":  p.handling,
			"KannSeiaUpp": p.seiaUpp,
		})
	}
	return ut
}

// Arket shows the workshop: every component in the stylesheet,
// outside its work, light and dark side by side.
//
// The page draws the components with the same code the pages use, so it
// cannot become untrue without the pages becoming untrue at the same
// moment. That is the whole difference from a style book that describes
// something: this one shows it.
//
// It lives in the development group of the router. A workshop is not
// something the studio shows anyone, which is why the text in there is not
// translated either.
func (a *App) Arket(w http.ResponseWriter, r *http.Request) {
	lang := GetLanguageFromRequest(r)

	måndag, nå, hendingar := merkeVika()
	merkeprovor := BuildSessions(lang, hendingar, nå)
	vikedagar := make([]time.Time, 7)
	for i := range vikedagar {
		vikedagar[i] = måndag.AddDate(0, 0, i)
	}

	renderPage(w, r, "pages/arket", sidedata(r, SidaArket, "Arket", map[string]any{
		"Merkeprovor": merkeprovor,
		// Medlemskorti gjeng gjenom den same malen sidone nyttar; sjå
		// medlemskortProvor ovanfor.
		"Medlemskortprovor": medlemskortProvor(lang),
		// Vika gjeng gjenom KlemVika som i timeplanen, so verkstaden
		// syner rutenetet med den same koden — ikkje ei etterlikning.
		"ClassRows": BuildWeekRows(lang, hendingar, nå, måndag),
		"WeekDays": []string{
			t(lang, "timeplan.monday"), t(lang, "timeplan.tuesday"),
			t(lang, "timeplan.wednesday"), t(lang, "timeplan.thursday"),
			t(lang, "timeplan.friday"), t(lang, "timeplan.saturday"),
			t(lang, "timeplan.sunday"),
		},
		"WeekDates": vikedagar,
		"Today":     måndag.Format("2006-01-02"),
	}))
}
