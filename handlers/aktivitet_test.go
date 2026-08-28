package handlers

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
	"time"
)

// Aktivitetsbolken les data ut or ein struktur som vert bygd ein heilt
// annan stad. Prøva teiknar honom med tal ein kjenner, so ein mismatch
// millom bygg og mal fell her og ikkje paa skjermen.
//
// Han var eit varmekart yver kvar dag fyrr. No er han eit aar i 52
// vikeprikkar: lysstyrken segjer kor mange gonger, fargen kva slag.
func TestAktivitetsbolkenTeiknarSeg(t *testing.T) {
	naa := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	// 24.–26. august ligg i den same vika; 14. juli i ei anna.
	perDag := map[string]int{
		"2026-08-24": 1,
		"2026-08-25": 2,
		"2026-08-26": 5,
		"2026-07-14": 1,
	}
	perType := map[string]map[string]int{
		"2026-08-24": {"yoga": 1},
		"2026-08-25": {"yoga": 2},
		"2026-08-26": {"fascia": 5},
		"2026-07-14": {"pilates": 1},
	}
	a := NewActivity("nn", perDag, perType, naa, 52)

	if a.Total != 9 {
		t.Errorf("totalt %d, venta 9", a.Total)
	}
	if len(a.Bars) != 6 {
		t.Errorf("%d bjelkar, venta seks maanadar", len(a.Bars))
	}
	if len(a.Cells) != 52 {
		t.Errorf("%d prikkar, venta 52 — eitt aar", len(a.Cells))
	}

	// Full styrke er tri, ikkje fire: aatte timar paa ei vika er det
	// same trinnet som tri, av di det ikkje finst sterkare enn full.
	var toppen int
	for _, r := range a.Cells {
		if r.Level > toppen {
			toppen = r.Level
		}
	}
	if toppen != 3 {
		t.Errorf("høgste trinnet var %d, venta 3", toppen)
	}

	// Vika med aatte timar: fascia raadde henne (5 mot 3).
	// Vika med éin: pilates, og trinn 1 — tend, men ikkje full.
	var fascia, pilates bool
	for _, r := range a.Cells {
		if r.Slag == "fascia" && r.Level == 3 {
			fascia = true
		}
		if r.Slag == "pilates" && r.Level == 1 {
			pilates = true
		}
	}
	if !fascia {
		t.Error("vika med flest fascia-timar bar ikkje fascia paa full styrke")
	}
	if !pilates {
		t.Error("vika med éin pilates-time var ikkje tend paa fyrste trinnet")
	}

	// Brettet er 13 hol brei og fire rader djupt.
	if a.Width != 13 || a.Height != 4 {
		t.Errorf("brettet er %.0f×%.0f, venta 13×4", a.Width, a.Height)
	}

	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
		basePath:  "templates",
	}
	tm.loadTemplates()
	mal, ok := tm.GetTemplate("pages/dashboard")
	if !ok {
		t.Fatal("malen pages/dashboard vart ikkje lasta")
	}

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "aktivitet_bolk", map[string]interface{}{
		"Lang": "nn", "Activity": a,
	}); err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	for _, vil := range []string{"brett", "hol", "pinne", "bjelkerad", "niv-3", "varmeskala", "slag-fascia"} {
		if !strings.Contains(html, vil) {
			t.Errorf("bolken teikna ikkje %s", vil)
		}
	}
	// Trinnet som ikkje finst lenger skal ikkje koma attende ved eit uhell.
	if strings.Contains(html, "niv-4") {
		t.Error("niv-4 stend framleis; full styrke er tri no")
	}
	// Tali skal kunna lesast utan farge: fargetrinni ligg under 3:1 mot
	// grunnen, og daa er samandraget og merkelappane avlastinga.
	if !strings.Contains(html, a.Summary) {
		t.Error("samandraget stend ikkje i bolken")
	}
	if strings.Count(html, "<title>") < len(a.Bars) {
		t.Error("ikkje kvar bjelke hev ein merkelapp")
	}
}
