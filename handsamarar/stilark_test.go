package handsamarar

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The test reads the *assembled* sheet, not a file. The parts under
// static/css/deler/ are what stands on disk; what the browser gets is their
// sum, and the invariants hold for the sum.
func lesStilarket(t *testing.T) string {
	t.Helper()
	gamal := stilarkMappe
	stilarkMappe = "../static/css/deler"
	defer func() { stilarkMappe = gamal }()
	b, err := byggStilark()
	if err != nil {
		t.Fatalf("kunde ikkje setja saman stilarket: %v", err)
	}
	return string(b)
}

// The dark tokens stand twice in the stylesheet, and have to: one block
// answers the system theme, the other a choice the user made. Two blocks
// with the same list is a list that can split — a colour is fixed in one,
// and then it lies in the other for everyone who chose dark themselves.
func lesBlokk(t *testing.T, css, veljar string) map[string]string {
	t.Helper()

	i := strings.Index(css, veljar)
	if i < 0 {
		t.Fatalf("fann ikkje blokki %q i stilarket", veljar)
	}
	rest := css[i+len(veljar):]
	slutt := strings.Index(rest, "}")
	if slutt < 0 {
		t.Fatalf("blokki %q vart aldri stengd", veljar)
	}

	ut := make(map[string]string)
	for _, linja := range strings.Split(rest[:slutt], "\n") {
		if k, v, ok := strings.Cut(strings.TrimSpace(linja), ":"); ok && strings.HasPrefix(k, "--") {
			ut[strings.TrimSpace(k)] = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(v), ";"))
		}
	}
	if len(ut) == 0 {
		t.Fatalf("blokki %q hadde ingen tokeni", veljar)
	}
	return ut
}

func TestDeiTvoMyrkeBlokkaneErSamde(t *testing.T) {
	css := lesStilarket(t)

	ljos := lesBlokk(t, css, `:root, [data-theme="light"] {`)
	// The system block is called :root now and not
	// 	// :root:not([data-theme="light"]) — an explicit choice wins instead in
	// 	// the block *after* it, which comes later in the sheet and carries the
	// 	// same weight (ARKET §10.4). So we have to search with the media query
	// 	// in front: ":root {" alone appears in several places.
	systemet := lesBlokk(t, css, "@media (prefers-color-scheme: dark) {\n  :root {")
	valet := lesBlokk(t, css, `[data-theme="dark"] {`)
	ljosvalet := lesBlokk(t, css, `[data-theme="light"] {`)

	// Light by choice has to say the same as light by default. Same trap as
	// 	// the two dark blocks: two lists with the same content, and a chance to
	// 	// fix only one.
	for nykel, verd := range ljosvalet {
		hitt, finst := ljos[nykel]
		if !finst {
			t.Errorf("%s stend i det valde ljose temaet, men ikkje i standarden", nykel)
			continue
		}
		if hitt != verd {
			t.Errorf("%s er «%s» som standard og «%s» naar brukaren vel ljost", nykel, hitt, verd)
		}
	}

	for nykel, verd := range systemet {
		hitt, finst := valet[nykel]
		if !finst {
			t.Errorf("%s stend i systemblokki, men ikkje i valblokki — den som vel myrkt sjølv fær han ikkje", nykel)
			continue
		}
		if hitt != verd {
			t.Errorf("%s er «%s» naar systemet er myrkt og «%s» naar brukaren vel myrkt", nykel, verd, hitt)
		}
	}
	for nykel := range valet {
		if _, finst := systemet[nykel]; !finst {
			t.Errorf("%s stend i valblokki, men ikkje i systemblokki", nykel)
		}
	}

	// A dark token without a light counterpart has nowhere to fall back to.
	var utan []string
	for nykel := range valet {
		if _, finst := ljos[nykel]; !finst {
			utan = append(utan, nykel)
		}
	}
	if len(utan) > 0 {
		sort.Strings(utan)
		t.Errorf("myrke tokeni utan ljost motstykke: %s", strings.Join(utan, ", "))
	}
}

// Verkstaden er sida som skal syna at ingen ting lyg. Ho lyt difor
// sjølv teikna seg — og ho lyt teikna båe temaom, av di det er heile
// grunnen til at ho finst.
func TestVerkstadenTeiknarSeg(t *testing.T) {
	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
		basePath:  "templates",
	}
	tm.loadTemplates()

	mal, ok := tm.GetTemplate("pages/arket")
	if !ok {
		t.Fatal("malen pages/arket vart ikkje lasta")
	}

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "base", map[string]interface{}{
		"Title": "Arket",
		"Lang":  "nn",
	}); err != nil {
		t.Fatalf("teikning: %v", err)
	}

	html := ut.String()
	for _, tema := range []string{`data-theme="light"`, `data-theme="dark"`} {
		if !strings.Contains(html, tema) {
			t.Errorf("verkstaden teikna ingi %s-prøva", tema)
		}
	}

	// Every sample should appear twice — once in each theme.
	prova := regexp.MustCompile(`class="verkstad-prova"`)
	if n := len(prova.FindAllString(html, -1)); n < 2 || n%2 != 0 {
		t.Errorf("prøvone stend ikkje parvis: fann %d", n)
	}
}

// Every part has to close what it opens.
//
// The sheet is spliced from 31 files, and one brace too many in one of them
// is not a fault you see: the browser corrects itself by throwing it away,
// so everything looks right — until the day somebody wraps the file in a
// media query, and then the rules after it stop applying.
// 82-medlemskapet-mitt.css carried one at the bottom.
func TestKvarDelLukkarDetHoOpnar(t *testing.T) {
	gamal := stilarkMappe
	stilarkMappe = "../static/css/deler"
	defer func() { stilarkMappe = gamal }()

	oppf, err := os.ReadDir(stilarkMappe)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range oppf {
		if o.IsDir() || !strings.HasSuffix(o.Name(), ".css") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(stilarkMappe, o.Name()))
		if err != nil {
			t.Fatal(err)
		}
		s := string(b)
		if opne, stengde := strings.Count(s, "{"), strings.Count(s, "}"); opne != stengde {
			t.Errorf("%s opnar %d og stengjer %d", o.Name(), opne, stengde)
		}
	}
}

// Kva glas eit merke ber er eit token, ikkje ein veljar — og det er ei
// avgjerd som lett vert gjord om att gale.
//
// Fyrste freistnaden skreiv `[data-theme="dark"] .form-glas { fill: … }`.
// Det brast av di regelen i media-blokka hev høgare spesifisitet enn
// den for det ljose temaet, so det myrke glaset vann ogso inni ei ljos
// ramma. Tokenfila hev aldri det problemet, av di ho set verdet *på*
// tema-elementet og let arven gjera valet.
//
// Difor: `--glasmaal` skal setjast i alle fire tema-blokkene, og
// `.form-glas` skal lesa han — ikkje velja gradient med ein veljar.
func TestGlasetErEitTokenOgIkkjeEinVeljar(t *testing.T) {
	ark := stilarkFråProva(t)

	if n := strings.Count(ark, "--glasmaal:"); n != 4 {
		t.Errorf("--glasmaal er sett %d gonger, venta fire (dei fire tema-blokkene)", n)
	}
	if !strings.Contains(ark, "fill: var(--glasmaal)") {
		t.Error(".form-glas les ikkje --glasmaal")
	}
	if strings.Contains(ark, `data-theme="dark"] .form-glas`) ||
		strings.Contains(ark, `data-theme="light"] .form-glas`) {
		t.Error("glaset vert vald med ein etterkomar-veljar att — sjaa kommentaren yver")
	}
}
