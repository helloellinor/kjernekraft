package handlers

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

// Knappen som lagar timen laag som eit tridje jamstelt barn i spalta —
// same luft kring seg som fylgja, og heilt til vinstre under ei setning
// paa femti teikn. Han er slutten paa setningi no, og stend paa rad med
// det systemet reknar ut.
//
// Prøva held dei tvo saman. Skil dei lag att, er knappen attende i det
// tome feltet.
func TestKnappenStendIFotenSamanMedFylgja(t *testing.T) {
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

	i := strings.Index(html, `class="setningsfot"`)
	if i < 0 {
		t.Fatal("foten fanst ikkje")
	}
	slutt := strings.Index(html[i:], "</div>")
	if slutt < 0 {
		t.Fatal("foten vart aldri stengd")
	}
	fot := html[i : i+slutt]

	if !strings.Contains(fot, `class="fylgja"`) {
		t.Error("fylgja stend ikkje i foten")
	}
	if !strings.Contains(fot, `type="submit"`) {
		t.Error("knappen stend ikkje i foten")
	}

	// Og han skal koma etter fylgja: grunnlaget fyrst, handlingi sist.
	if strings.Index(fot, `class="fylgja"`) > strings.Index(fot, `type="submit"`) {
		t.Error("knappen stend fyre fylgja; grunnlaget skal koma fyrst")
	}
}
