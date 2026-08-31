package handsamarar

import (
	"bytes"
	"html/template"
	"kjernekraft/database"
	"kjernekraft/handsamarar/modules"
	"strings"
	"testing"
)

func lastMalane(t *testing.T) *TemplateManager {
	t.Helper()

	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
		basePath:  "templates",
	}
	tm.loadTemplates()
	return tm
}

// The people list is the only place a permission can be set. The test
// renders it and checks that both marks are there, and that what is
// switched on looks switched on — a permission sitting in the database
// without showing in the surface is the same fault as the button not
// existing at all.
func TestFolkelistaTeiknarRollemerki(t *testing.T) {
	mal, ok := lastMalane(t).GetTemplate("pages/admin")
	if !ok {
		t.Fatal("malen pages/admin vart ikkje lasta")
	}

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "admin_users_table", map[string]interface{}{
		"Lang": "nn",
		"Folk": []database.Person{
			{ID: 7, Namn: "Kristina", Epost: "kristina@dømet.no", ErLærar: true},
			{ID: 8, Namn: "Bjørn", Epost: "bjørn@dømet.no"},
		},
	}); err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	// Kristina holds the teacher permission; Bjørn does not. Both should have
	// 	// both buttons — a permission you do not have is not something you see
	// 	// somebody lacking, so the list has to show both outcomes.
	for _, vil := range []string{
		`data-brukar="7"`, `data-brukar="8"`,
		`data-loyve="teacher"`, `data-loyve="admin"`,
	} {
		if !strings.Contains(html, vil) {
			t.Errorf("folkelista teikna ikkje %s", vil)
		}
	}

	if n := strings.Count(html, `class="held loyvemerke"`); n != 4 {
		t.Errorf("venta fire rollemerke paa tvo folk, fann %d", n)
	}
	if n := strings.Count(html, `aria-pressed="true"`); n != 1 {
		t.Errorf("berre Kristina hev eit løyve; fann %d merke som var slegne paa", n)
	}
}

// The greeting stands above the sheet, not inside a tab. It answers "how
// big is the house" — and you ask that when you come in, not after picking
// a tab. It used to sit in the overview tab, and then the answer was hidden
// behind a choice.
func TestHelsingaStendYverArketOgIkkjeInniDet(t *testing.T) {
	mal, ok := lastMalane(t).GetTemplate("pages/admin")
	if !ok {
		t.Fatal("malen pages/admin vart ikkje lasta")
	}

	stats := modules.NewAdminStatsModule(6, 12, 2, "nn")

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "content", map[string]interface{}{
		"Lang":  "nn",
		"Stats": stats,
		"Faner": proveFanerekkje("faneark-admin", "meldingar",
			"meldingar", "timeplan", "prisar", "folk", "innstillingar"),
		"Prisfaner": proveFanerekkje("faneark-prisar", "medlemskap",
			"medlemskap", "klippekort", "reglar"),
	}); err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	// Klassa heiter `briefing`, og lenkja inni henne `briefing-lenkje`.
	// Eit laust sok etter «briefing» finn båe, so det lyt vera heile
	// klasseverdet det vert leita etter.
	helsing := strings.Index(html, `class="briefing"`)
	arket := strings.Index(html, `class="faneark"`)
	if helsing < 0 {
		t.Fatal("helsinga vart ikkje teikna i det heile")
	}
	if arket < 0 {
		t.Fatal("fanearket vart ikkje teikna")
	}
	if helsing > arket {
		t.Error("helsinga stend inni eller under fanearket; ho skal staa yver det")
	}

	// It should stand in *one* place. If it was not taken out of the tab when
	// 	// it moved up, it stands twice and counts the same house twice.
	if n := strings.Count(html, `class="briefing"`); n != 1 {
		t.Errorf("helsinga stend %d gonger, ikkje ein", n)
	}

	// Fana som held det som ventar heiter det ho held. Ho er ei lenkja
	// no, og nykelen stend i adressa hennar og ikkje i eit data-felt.
	if !strings.Contains(html, `fane=meldingar`) {
		t.Error("meldingsfana fanst ikkje")
	}
	if strings.Contains(html, `fane=oversyn`) {
		t.Error("den gamle oversynsfana stend att")
	}

	// Talet som ventar er ei lenkja, og ho skal føra dit dei ligg no.
	// Ein emneknagg gjorde det fyrr, og eit skript laut lesa honom.
	if !strings.Contains(html, `href="/admin?fane=meldingar"`) {
		t.Error("lenkja i helsinga peikar ikkje inn i meldingsfana")
	}
	if strings.Contains(html, `href="#meldingar"`) {
		t.Error("helsinga peikar endaa paa ein emneknagg")
	}
}

// The group slider above the people list. It is the same .faner as the
// tabs elsewhere in the house, but it does not switch sections — it filters
// the list — and so it carries data-rolla and not data-bolk.
//
// The test is exactly that difference. bolkveljar.js picks up every
// .fane[data-bolk] inside the nearest .faneark, and the people list *sits*
// inside the admin's faneark: had the slider written data-bolk, pressing
// "Teachers" would have hidden the whole people section instead of
// filtering it, because no section answers to that name.
func TestBunkeskyvarenSiktarOgByterIkkjeBolk(t *testing.T) {
	mal, ok := lastMalane(t).GetTemplate("pages/admin")
	if !ok {
		t.Fatal("malen pages/admin vart ikkje lasta")
	}

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "admin_users_table", map[string]interface{}{
		"Lang": "nn",
		"Folk": []database.Person{
			{ID: 7, Namn: "Kristina", Epost: "kristina@dømet.no", ErLærar: true},
			{ID: 8, Namn: "Bjørn", Epost: "bjørn@dømet.no"},
		},
	}); err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	// Fire bunkar: alle, elevar, lærarar, sjefar. «Alle» ber ein tom
	// verd — det er ikkje eit filter, det er utgangsstoda.
	for _, vil := range []string{
		`data-rolla=""`, `data-rolla="elev"`,
		`data-rolla="laerar"`, `data-rolla="sjef"`,
	} {
		if !strings.Contains(html, vil) {
			t.Errorf("bunkeskyvaren teikna ikkje %s", vil)
		}
	}

	// None of its tabs may carry data-bolk. If they did, bolkveljar.js would
	// 	// take them, and the people section would close itself when you pressed.
	skyvar := html[strings.Index(html, `class="faner rollefaner"`):]
	skyvar = skyvar[:strings.Index(skyvar, "</div>")]
	if strings.Contains(skyvar, "data-bolk") {
		t.Error("bunkeskyvaren ber data-bolk; daa byter han bolk i staden for aa sikta lista")
	}

	// "All" is selected when the page arrives. A filter that is the default
	// 	// state would have made a search for a teacher answer "nobody".
	if n := strings.Count(skyvar, `aria-selected="true"`); n != 1 {
		t.Errorf("venta éi vald fana, fann %d", n)
	}
	alle := skyvar[strings.Index(skyvar, `data-rolla=""`):]
	alle = alle[:strings.Index(alle, "</button>")]
	if !strings.Contains(alle, `aria-selected="true"`) {
		t.Error("det er ikkje «Alle» som er vald naar sida kjem")
	}

	// Lista er det fanone siktar, so ho lyt ha namnet dei peikar på.
	if !strings.Contains(html, `id="folk-liste"`) {
		t.Error("lista fanone peikar paa hev ingen id")
	}
	// Eit filter som gjev ingen ting lyt kunna segja det.
	if !strings.Contains(html, `id="folk-inkje"`) {
		t.Error("setningi for eit tomt filter fanst ikkje")
	}
}
