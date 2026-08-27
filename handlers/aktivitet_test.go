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
func TestAktivitetsbolkenTeiknarSeg(t *testing.T) {
	naa := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	perDag := map[string]int{
		"2026-08-24": 1,
		"2026-08-25": 2,
		"2026-08-26": 5, // over toppen: skal klemmast til trinn 4
		"2026-07-14": 1,
	}
	a := NyAktivitet("nn", perDag, naa, 26)

	if a.Totalt != 9 {
		t.Errorf("totalt %d, venta 9", a.Totalt)
	}
	if len(a.Bjelkar) != 6 {
		t.Errorf("%d bjelkar, venta seks maanadar", len(a.Bjelkar))
	}
	if len(a.Rutar) == 0 {
		t.Fatal("ingen rutor i varmekartet")
	}

	// Trinnet er klemt: fem timar paa ein dag er det same trinnet som
	// fire, av di fleire trinn enn auga kann skilja ikkje er meir.
	var toppen int
	for _, r := range a.Rutar {
		if r.Niva > toppen {
			toppen = r.Niva
		}
	}
	if toppen != 4 {
		t.Errorf("høgste trinnet var %d, venta 4", toppen)
	}

	// Ingen rute skal liggja etter i dag.
	for _, r := range a.Rutar {
		if r.X > a.Breidd || r.Y > a.Hogd {
			t.Errorf("ruta %.1f,%.1f ligg utanfor kartet %.1f×%.1f", r.X, r.Y, a.Breidd, a.Hogd)
		}
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
		"Lang": "nn", "Aktivitet": a,
	}); err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	for _, vil := range []string{"varmekart", "bjelkerad", "niv-4", "varmeskala"} {
		if !strings.Contains(html, vil) {
			t.Errorf("bolken teikna ikkje %s", vil)
		}
	}
	// Tali skal kunna lesast utan farge: fargetrinni ligg under 3:1 mot
	// grunnen, og daa er samandraget og merkelappane avlastinga.
	if !strings.Contains(html, a.Samandrag) {
		t.Error("samandraget stend ikkje i bolken")
	}
	if strings.Count(html, "<title>") < len(a.Bjelkar) {
		t.Error("ikkje kvar bjelke hev ein merkelapp")
	}
}
