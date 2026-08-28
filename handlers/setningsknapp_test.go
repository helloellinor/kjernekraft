package handlers

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

// Knappen som lagar timen har lege paa ei lina for seg sjølv tvo gonger:
// fyrst som eit tridje jamstelt barn i spalta, so i ein eigen fot med ei
// haarlina yver. Baae gongene saag han paalimd ut, av di ei lina for seg
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

	i := strings.Index(html, `<p class="setning"`)
	if i < 0 {
		t.Fatal("setningi fanst ikkje")
	}
	slutt := strings.Index(html[i:], "</p>")
	if slutt < 0 {
		t.Fatal("setningi vart aldri stengd")
	}
	setning := html[i : i+slutt]

	if !strings.Contains(setning, `type="submit"`) {
		t.Error("knappen stend utanfor setningi")
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
