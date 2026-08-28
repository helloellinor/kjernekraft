package handlers

import (
	"net/http"
	"time"

	"kjernekraft/models"
)

// merkeProvor lagar timane verkstaden syner. Dei er ikkje teikningar av
// merket: dei gjeng gjenom NyFramsyning som alt anna, so ser du noko
// rart her, ser ein brukar det same i timeplanen.
func merkeVika() (time.Time, time.Time, []models.Event) {
	// Ei fast vika, so prøvone ikkje skifter medan ein ser paa deim.
	// Maandagen er «i dag», og klokka er ti paa tolv.
	maandag := time.Date(2026, 8, 24, 0, 0, 0, 0, OsloLoc)
	naa := maandag.Add(11*time.Hour + 50*time.Minute)
	iDag := maandag.Format("2006-01-02")

	slag := []struct {
		dag           int
		fraa, til     string
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

	titlar := []struct{ tittel, laerar, rom string }{
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
			maandag.AddDate(0, 0, s.dag).Format("2006-01-02")+" "+s.fraa, OsloLoc)
		slutt, _ := time.ParseInLocation("2006-01-02 15:04",
			maandag.AddDate(0, 0, s.dag).Format("2006-01-02")+" "+s.til, OsloLoc)
		ut = append(ut, models.Event{
			ID:               900 + i,
			Title:            titlar[i].tittel,
			TeacherName:      titlar[i].laerar,
			RoomName:         titlar[i].rom,
			StartTime:        start,
			EndTime:          slutt,
			Capacity:         s.plassar,
			CurrentEnrolment: s.teke,
		})
	}
	_ = iDag
	return maandag, naa, ut
}

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

	maandag, naa, hendingar := merkeVika()
	merkeprovor := BuildSessions(lang, hendingar, naa)
	vikedagar := make([]time.Time, 7)
	for i := range vikedagar {
		vikedagar[i] = maandag.AddDate(0, 0, i)
	}

	renderPage(w, r, "pages/arket", map[string]interface{}{
		"Title":       "Arket",
		"CurrentPage": "arket",
		"Lang":        lang,
		"CSRFToken":   CSRFToken(r),
		"IsAdmin":     sessionIsAdmin(r),
		"UserName":    sessionUserName(r),
		"Merkeprovor": merkeprovor,
		// Vika gjeng gjenom KlemVika som i timeplanen, so verkstaden
		// syner rutenetet med den same koden — ikkje ei etterlikning.
		"ClassRows": BuildWeekRows(lang, hendingar, naa, maandag),
		"WeekDays": []string{
			t(lang, "timeplan.monday"), t(lang, "timeplan.tuesday"),
			t(lang, "timeplan.wednesday"), t(lang, "timeplan.thursday"),
			t(lang, "timeplan.friday"), t(lang, "timeplan.saturday"),
			t(lang, "timeplan.sunday"),
		},
		"WeekDates": vikedagar,
		"Today":     maandag.Format("2006-01-02"),
	})
}
