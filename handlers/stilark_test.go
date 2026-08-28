package handlers

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

// Prøva les det *samansette* arket, ikkje ei fil. Delane under
// static/css/deler/ er det som stend paa disken; det nettlesaren fær er
// summen av deim, og det er summen invariantane gjeld for.
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

// Dei myrke tokeni stend tvo gonger i stilarket, og lyt gjera det: den
// eine blokki svarar paa systemtemaet, den hine paa eit val brukaren
// hev teke. To blokkar med den same lista er ei lista som kann skilja
// lag — ein farge vert retta i den eine, og so lyg han i den hine for
// alle som hev valt myrkt sjølve. Prøva ser etter det.
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
	// Systemblokki heiter `:root` no og ikkje `:root:not([data-theme="light"])`
	// — eit uttrykkeleg val vinn i staden i blokki *etter* henne, som kjem
	// seinare i arket og hev same vekt (ARKET §10.4, utført 28.8.2026).
	// Difor lyt me søkja med media-spurnaden framfyre: `:root {` aaleine
	// finst fleire stader i arket.
	systemet := lesBlokk(t, css, "@media (prefers-color-scheme: dark) {\n  :root {")
	valet := lesBlokk(t, css, `[data-theme="dark"] {`)
	ljosvalet := lesBlokk(t, css, `[data-theme="light"] {`)

	// Ljost av val lyt segja det same som ljost av standard. Same fella
	// som dei tvo myrke: tvo lister med det same, og ein sjanse til aa
	// retta berre den eine.
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

	// Eit myrkt token som ikkje hev eit ljost motstykke, hev ingen stad
	// aa falla attende til.
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
// sjølv teikna seg — og ho lyt teikna baae temaom, av di det er heile
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

	// Kvar prøva skal finnast tvo gonger — ein gong i kvart tema.
	prova := regexp.MustCompile(`class="verkstad-prova"`)
	if n := len(prova.FindAllString(html, -1)); n < 2 || n%2 != 0 {
		t.Errorf("prøvone stend ikkje parvis: fann %d", n)
	}
}

// Kvar del lyt lukka det ho opnar.
//
// Arket vert skøytt saman av 31 filer, og ein klamme for mykje i ei av
// deim er ikkje ein feil du ser: nettlesaren rettar seg sjølv ved aa
// kasta han, so alt ser rett ut — til den dagen nokon pakkar fila inn i
// ein media-spurnad, og daa sluttar reglane etter honom aa gjelda.
// `82-medlemskapet-mitt.css` bar ein slik i botnen.
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
