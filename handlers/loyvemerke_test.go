package handlers

import (
	"bytes"
	"html/template"
	"kjernekraft/database"
	"kjernekraft/handlers/modules"
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

// Folkelista er den einaste staden eit løyve kann setjast. Prøva teiknar
// henne og ser etter at baae merki stend der, og at det som er slege paa
// ser slege paa ut — eit løyve som ligg i basen utan aa syna seg i flata
// er den same feilen som daa knappen ikkje fanst i det heile.
func TestFolkelistaTeiknarRollemerki(t *testing.T) {
	mal, ok := lastMalane(t).GetTemplate("pages/admin")
	if !ok {
		t.Fatal("malen pages/admin vart ikkje lasta")
	}

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "admin_users_table", map[string]interface{}{
		"Lang": "nn",
		"Folk": []database.Person{
			{ID: 7, Namn: "Kristina", Epost: "kristina@dømet.no", ErLaerar: true},
			{ID: 8, Namn: "Bjørn", Epost: "bjørn@dømet.no"},
		},
	}); err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	// Kristina held lærarløyvet; Bjørn held henne ikkje. Baae skal ha
	// baae knappane — eit løyve ein ikkje hev er ikkje noko ein ser at
	// nokon manglar, so lista lyt syna baae utfalli.
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

// Helsinga stend yver arket, ikkje inni ei fane. Ho svarar paa «kor
// stort er huset» — og det spursmaalet stiller ein naar ein kjem inn,
// ikkje etter at ein hev valt ei fane. Fyrr laag ho i oversynsfana, og
// daa var svaret gøymt bak eit val.
func TestHelsingaStendYverArketOgIkkjeInniDet(t *testing.T) {
	mal, ok := lastMalane(t).GetTemplate("pages/admin")
	if !ok {
		t.Fatal("malen pages/admin vart ikkje lasta")
	}

	stats, err := modules.NewAdminStatsModule(6, 12, 2, "nn")
	if err != nil {
		t.Fatalf("tali: %v", err)
	}

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "content", map[string]interface{}{
		"Lang":  "nn",
		"Stats": stats,
	}); err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	// Klassa heiter `briefing`, og lenkja inni henne `briefing-lenkje`.
	// Eit laust sok etter «briefing» finn baae, so det lyt vera heile
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

	// Ho skal staa *ein* stad. Vart ho ikkje teki ut or fana daa ho
	// vart flutt upp, stend ho tvo gonger og tel same huset tvo gonger.
	if n := strings.Count(html, `class="briefing"`); n != 1 {
		t.Errorf("helsinga stend %d gonger, ikkje ein", n)
	}

	// Fana som held det som ventar heiter det ho held.
	if !strings.Contains(html, `data-bolk="meldingar"`) {
		t.Error("meldingsfana fanst ikkje")
	}
	if strings.Contains(html, `data-bolk="oversyn"`) {
		t.Error("den gamle oversynsfana stend att")
	}

	// Talet som ventar er ei lenkja, og ho skal føra dit dei ligg no.
	if !strings.Contains(html, `href="#meldingar"`) {
		t.Error("lenkja i helsinga peikar ikkje inn i meldingsfana")
	}
}

// Bunkeskyvaren yver folkelista. Han er den same `.faner` som fanone
// elles i huset, men han byter ikkje bolk — han siktar lista — og difor
// ber han `data-rolla` og ikkje `data-bolk`.
//
// Prøva er nettupp den skilnaden. `faner.js` plukkar upp kvar
// `.fane[data-bolk]` inni næraste `.faneark`, og folkelista *stend* inni
// administrasjonen sitt faneark: skreiv skyvaren `data-bolk`, hadde eit
// trykk paa «Lærarar» gøymt heile folkebolken i staden for aa sikta han,
// av di ingen bolk svarar til det namnet. Same fella som prisfana gjekk i
// med `Bolk`/`Namn`.
func TestBunkeskyvarenSiktarOgByterIkkjeBolk(t *testing.T) {
	mal, ok := lastMalane(t).GetTemplate("pages/admin")
	if !ok {
		t.Fatal("malen pages/admin vart ikkje lasta")
	}

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "admin_users_table", map[string]interface{}{
		"Lang": "nn",
		"Folk": []database.Person{
			{ID: 7, Namn: "Kristina", Epost: "kristina@dømet.no", ErLaerar: true},
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

	// Ingen av fanone hans maa bera `data-bolk`. Gjorde dei det, tok
	// faner.js deim, og folkebolken hadde stengt seg naar ein trykte.
	skyvar := html[strings.Index(html, `class="faner rollefaner"`):]
	skyvar = skyvar[:strings.Index(skyvar, "</div>")]
	if strings.Contains(skyvar, "data-bolk") {
		t.Error("bunkeskyvaren ber data-bolk; daa byter han bolk i staden for aa sikta lista")
	}

	// «Alle» er vald naar sida kjem. Eit filter som er utgangsstoda
	// hadde late eit søk etter ein lærar svara «ingen».
	if n := strings.Count(skyvar, `aria-selected="true"`); n != 1 {
		t.Errorf("venta éi vald fana, fann %d", n)
	}
	alle := skyvar[strings.Index(skyvar, `data-rolla=""`):]
	alle = alle[:strings.Index(alle, "</button>")]
	if !strings.Contains(alle, `aria-selected="true"`) {
		t.Error("det er ikkje «Alle» som er vald naar sida kjem")
	}

	// Lista er det fanone siktar, so ho lyt ha namnet dei peikar paa.
	if !strings.Contains(html, `id="folk-liste"`) {
		t.Error("lista fanone peikar paa hev ingen id")
	}
	// Eit filter som gjev ingen ting lyt kunna segja det.
	if !strings.Contains(html, `id="folk-inkje"`) {
		t.Error("setningi for eit tomt filter fanst ikkje")
	}
}
