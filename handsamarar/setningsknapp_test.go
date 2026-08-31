package handsamarar

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

// Knappen som lagar timen har lege på ei lina for seg sjølv tvo gonger:
// fyrst som eit tridje jamstelt barn i spalta, so i ein eigen fot med ei
// haarlina yver. Båe gongene såg han paalimd ut, av di ei lina for seg
// sjølv er ei ny setning — og han er ikkje det. Han er punktumet i den
// setningi som alt stend der.
//
// Prøva held honom inne i setningi. Fell han ut att, stend han aaleine.
func TestKnappenSlutterSetningiOgStendIkkjeAaleine(t *testing.T) {
	tm := &TemplateManager{templates: make(map[string]*template.Template), basePath: "templates"}
	tm.loadTemplates()
	mal, ok := tm.GetTemplate("pages/admin")
	if !ok {
		t.Fatal("malen pages/admin vart ikkje lasta")
	}

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "admin_class_management", map[string]interface{}{
		"Lang": "nn",
	}); err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	if strings.Contains(html, `class="setningsfot"`) {
		t.Error("foten er attende; knappen skal slutta setningi, ikkje staa i eit rom for seg")
	}

	// Regelen: knappen skal stå inni ei setning og ikkje aaleine på ei
	// lina. Prøva såg på den *fyrste* `<p class="setning">` den gongen
	// det berre fanst ei; skjemaet er tvo setningar no — timen, og kven
	// han er open for — og knappen sluttar den siste. Difor gjeng ho
	// andre vegen: finn knappen, gaa attende til avsnittet han stend i,
	// og sjå etter at det avsnittet *er* ei setning med noko meir i seg
	// enn knappen. Det er regelen sagt beint fram, og han fylgjer med um
	// det skulde koma ei tridje setning.
	k := strings.Index(html, `type="submit"`)
	if k < 0 {
		t.Fatal("knappen fanst ikkje")
	}
	i := strings.LastIndex(html[:k], "<p ")
	if i < 0 {
		t.Fatal("knappen stend ikkje i eit avsnitt i det heile")
	}
	slutt := strings.Index(html[i:], "</p>")
	if slutt < 0 {
		t.Fatal("setningi vart aldri stengd")
	}
	setning := html[i : i+slutt]

	if !strings.Contains(setning, `class="setning`) {
		t.Error("knappen stend i eit avsnitt som ikkje er ei setning")
	}
	// Ei setning med berre ein knapp i er ei lina for seg sjølv med ei
	// onnor drakt.
	if !strings.Contains(setning, "<input") && !strings.Contains(setning, "<select") {
		t.Error("knappen stend aaleine i si eigi setning")
	}

	// Fylgja er ikkje ein del av setningi — du skreiv henne ikkje. Ho
	// stend under, og ho stend framleis *fyre* du trykkjer.
	if strings.Contains(setning, `class="fylgja"`) {
		t.Error("fylgja stend inni setningi; ho er rekna, ikkje skriven")
	}
	if !strings.Contains(html, `class="fylgja"`) {
		t.Error("fylgja fanst ikkje i det heile")
	}
	if strings.Index(html, `class="fylgja"`) < strings.Index(html, `type="submit"`) {
		t.Error("fylgja stend fyre knappen i kjelda; ho skal koma etter setningi")
	}
}
