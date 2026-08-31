package handsamarar

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
	"time"

	"kjernekraft/models"
)

// Washing and the vocabulary are tested in handlers/prover, which
// reaches Slagi and SlagKlasse from outside. This one stays because it
// renders a template through the package's own TemplateManager.

// Kvar rad i timelista er eit kort med ein venge i slagfargen. Testen
// teiknar malen og ser på klassane, av di det er klassane vengen heng
// på — ikkje på noko Go-felt.
func TestTimekortetBerSlaget(t *testing.T) {
	tm := &TemplateManager{templates: make(map[string]*template.Template), basePath: "templates"}
	tm.loadTemplates()
	mal, ok := tm.GetTemplate("modules/dashboard/timeliste")
	if !ok {
		t.Fatal("malen modules/dashboard/timeliste vart ikkje lasta")
	}

	nå := time.Date(2026, 8, 28, 7, 0, 0, 0, time.Local)
	var rader []Session
	for i, slag := range []string{"yoga", "Pilates", ""} {
		rader = append(rader, NewSession("nn", models.Event{
			Title: "Time", TeacherName: "Nokon", ClassType: slag,
			StartTime: nå.Add(time.Duration(i+2) * time.Hour),
			EndTime:   nå.Add(time.Duration(i+3) * time.Hour),
			Capacity:  20, CurrentEnrolment: 1,
		}, nå.Format("2006-01-02"), nå))
	}

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "timeliste", map[string]any{"Lang": "nn", "Rader": rader}); err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	for _, vil := range []string{
		`class="timerad heimerad kort slag-yoga"`,
		`class="timerad heimerad kort slag-pilates"`,
	} {
		if !strings.Contains(html, vil) {
			t.Errorf("fann ikkje %s i den teikna lista", vil)
		}
	}
	// Ein ukjend type skal ikkje faa ein venge som lyg. Kortet stend,
	// slagklassen fell burt, og CSS-en tek den grå kanten.
	if n := strings.Count(html, `timerad heimerad kort "`); n != 1 {
		t.Errorf("venta éin rad utan slagklasse, fekk %d", n)
	}
}
