package handsamarar

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"kjernekraft/models"
)

// Flytane ein brukar gjeng gjenom lyt teikna seg, og dei lyt teikna seg
// utan dialogar.
//
// Ein mal som ikkje parsar syner seg fyrst når nokon opnar sida, og ein
// `confirm()` som kjem snikande attende i ein flyt syner seg aldri i ein
// diff — han ser ut som alt anna javascript. Difor les prøva den ferdige
// HTML-en og ser etter dei fire tingi som ikkje skal finnast der:
// `confirm(`, `alert(`, og dei tvo gamle knappenamni.
func TestFlytaneTeiknarSeg(t *testing.T) {
	gamal, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(gamal)

	tm := GetTemplateManager()
	tm.ReloadTemplates()

	nå := time.Now()
	seinare := nå.AddDate(0, 6, 0)

	medlem := &models.MembershipWithDetails{
		Membership:     models.Membership{ID: 1, Name: "Månedskort", Price: 104000},
		UserMembership: models.UserMembership{Status: "active", RenewalDate: seinare},
	}

	sider := []struct {
		mal  string
		data map[string]interface{}
		vil  []string
	}{
		{"pages/medlemskapet", map[string]interface{}{
			"Lang": "nn", "Title": "t", "CSRFToken": "x",
			"Noverande": medlem, "NoverandeNamn": "Månedskort",
			"Bunden": false, "MaanaderAtt": 0, "KannSeiaUpp": true,
			"Faner":         proveFanerekkje("faneark-medlemskapet", "medlemskapet", "medlemskapet", "byt"),
			"SwitchOptions": nil,
		}, []string{
			`data-handling="freeze"`, `data-handling="remove"`,
			`class="btn-danger"`, `class="merknad"`, `id="medlemssvar"`,
			"/api/membership/remove",
			// Det som ikkje kann gjerast um att krev trykk nummer tvo.
			`data-stadfest=`,
		}},
		{"pages/vilkaar", map[string]interface{}{
			"Lang": "nn", "Title": "t", "CSRFToken": "x", "Terms": "",
		}, []string{`<h1 class="page-title">`}},
		{"pages/innlogging", map[string]interface{}{
			"Lang": "nn", "Title": "t", "CSRFToken": "x",
		}, []string{`class="btn-primary brei"`}},
		{"pages/gloymt-passord", map[string]interface{}{
			"Lang": "nn", "Title": "t", "CSRFToken": "x", "StudioEpost": "a@b.no",
		}, []string{`class="btn-primary brei"`}},
		{"pages/betaling", map[string]interface{}{
			"Lang": "nn", "Title": "t", "CSRFToken": "x",
			"Faner": proveFanerekkje("faneark-betaling", "korta", "korta", "nytt-kort"),
			// Sikti paa ladningane var ein `onchange` som kalla ein
			// funksjon skriptet la paa `window`. No er ho `hx-get` og
			// `name="type"`, og det er *det* som lyt standa der: ei
			// prøva som endaa leita etter «filterCharges» hadde vore
			// grøn den dagen veljaren slutta aa gjera noko.
		}, []string{`id="betalingssvar"`, `class="merknad"`,
			`hx-get="/api/charges"`, `name="type"`}},
		// Rutenetet lyt teiknast med noko i seg. Med ei tom lista gjeng
		// `{{range .ClassRows}}` aldri inn i kroppen sin, og då vert
		// korkje radi, merket eller dagruta prøvd — malen kann be um eit
		// felt som ikkje finst, og prøva stend like grøn. Det hende:
		// `ClassRow.Minutt` heitte `Minute` i malen og ikkje i typen, og
		// timeplanen rende halvvegs ut i det stille.
		//
		// Rada kjem frå `BuildWeekRows` og ikkje frå ein literal, so
		// alt som malen les — merket, urskiva, dei sju dagane — er bygt
		// av den same koden som sida byggjer det med.
		{"pages/timeplan", map[string]interface{}{
			"Lang": "nn", "Title": "t", "CSRFToken": "x",
			"WeekOffset": 0, "WeekNumber": 35, "WeekTitle": "veke 35",
			"VikorIAaret": 52, "CanGoBack": false,
			"WeekDays": []string{}, "WeekDates": []time.Time{},
			"ClassRows": proveRader(), "Today": "", "Teachers": nil,
			"ForrigeVeke": "/elev/timeplan?week=0", "NesteVeke": "/elev/timeplan?week=1",
		}, []string{
			// Vikebladingi og sikti var tvo `window.`-funksjonar som sette
			// `location.href`. No er dei lenkjor og `hx-get` mot det eine
			// stykket som skifter — so det er *det* som lyt standa her.
			// Ei prøva som endaa leita etter «navigateWeek» hadde vore
			// grøn den dagen pilene slutta aa peika nokon stad.
			`id="vekebolken"`, `hx-target="#vekebolken"`, `preload="mouseover"`,
			// Kroppen i rutenetet, ikkje berre skalet kring det.
			"Vinyasa Flow", "Kristina", "60 min", "class=\"timerad"}},
	}

	for _, s := range sider {
		mal, ok := tm.GetTemplate(s.mal)
		if !ok {
			t.Errorf("%s: malen finst ikkje", s.mal)
			continue
		}
		var ut bytes.Buffer
		if err := mal.ExecuteTemplate(&ut, "base", s.data); err != nil {
			t.Errorf("%s: %v", s.mal, err)
			continue
		}
		h := ut.String()
		for _, vil := range s.vil {
			if !strings.Contains(h, vil) {
				t.Errorf("%s: fann ikkje %q", s.mal, vil)
			}
		}
		// Ingen dialog, og ingen streng som ikkje er umsett.
		for _, ulov := range []string{"confirm(", "alert(", "action-btn", "login-btn"} {
			if strings.Contains(h, ulov) {
				t.Errorf("%s: %q stend der framleis", s.mal, ulov)
			}
		}
	}
}

// Betalingsmåtane vert teikna for seg gjenom htmx.
func TestBetalingsmaataneTeiknarSeg(t *testing.T) {
	gamal, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(gamal)

	tm := GetTemplateManager()
	tm.ReloadTemplates()
	mal, ok := tm.GetTemplate("components/common/betalingsmaatar")
	if !ok {
		t.Skip("malen finst ikkje under det namnet")
	}
	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "betalingsmaatar", map[string]interface{}{
		"Lang": "nn",
		"Kort": []map[string]interface{}{
			{"ID": 7, "Merke": "Visa", "Last4": "4242", "IsDefault": false,
				"ExpiryMonth": 4, "ExpiryYear": 2029},
		},
	}); err != nil {
		t.Fatal(err)
	}
	h := ut.String()
	for _, vil := range []string{`data-kort="7"`, `data-handling="fjern"`, `data-handling="standard"`, `data-stadfest=`} {
		if !strings.Contains(h, vil) {
			t.Errorf("fann ikkje %q", vil)
		}
	}
	if strings.Contains(h, "onclick=") {
		t.Error("onclick stend der framleis")
	}
}

// proveRader byggjer éi verkeleg rad i rutenetet: ein time som gjeng
// tysdag, med lærar, rom og plassar, kjørt gjenom `BuildWeekRows` so
// merket og dei sju dagrutone vert til slik sida lagar deim.
func proveRader() []ClassRow {
	nå := time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC) // ein onsdag
	måndag := VikeMåndag(nå, 0)
	// Torsdag, altso *framfyre* «no». Timen laag på tysdag fyrr, og då
	// var han gjengen — og ein gjengen time vert ikkje teikna lenger, so
	// rada fanst ikkje og prøva fann korkje namn eller merke.
	start := måndag.AddDate(0, 0, 3).Add(18 * time.Hour) // torsdag 18:00
	e := models.Event{
		ID: 1, Title: "Vinyasa Flow", TeacherName: "Kristina",
		StartTime: start, EndTime: start.Add(60 * time.Minute),
		ClassType: "yoga", RoomName: "Salen",
		Capacity: 12, CurrentEnrolment: 5,
	}
	return BuildWeekRows("nn", []models.Event{e}, nå, måndag)
}
