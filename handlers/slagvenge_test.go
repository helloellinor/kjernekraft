package handlers

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"kjernekraft/models"
)

// Slaget skal fylgja timen ut i malen som ein CSS-krok, so vengen paa
// kortet kann bera fargen. Verdet kjem raatt fraa `events.class_type`,
// der administrasjonen skriv fritt.
func TestSlagklasseVaskarKlassetypen(t *testing.T) {
	fasit := map[string]string{
		"yoga":         "slag-yoga",
		"Yoga":         "slag-yoga",
		"  Reformer ":  "slag-reformer",
		"Hot Yoga":     "slag-hotyoga",
		"pilates-matt": "slag-pilatesmatt",
		"":             "",
		"   ":          "",
		"123":          "",
		// Ingen veg ut or klasseattributtet.
		`" onclick="x`: "slag-onclickx",
	}
	fn := getTemplateFuncs()["slagklasse"].(func(string) string)
	for inn, ut := range fasit {
		if fekk := fn(inn); fekk != ut {
			t.Errorf("slagklasse(%q) = %q, venta %q", inn, fekk, ut)
		}
	}
}

// Kvar rad i timelista er eit kort med ein venge i slagfargen. Testen
// teiknar malen og ser paa klassane, av di det er klassane vengen heng
// paa — ikkje paa noko Go-felt.
func TestTimekortetBerSlaget(t *testing.T) {
	tm := &TemplateManager{templates: make(map[string]*template.Template), basePath: "templates"}
	tm.loadTemplates()
	mal, ok := tm.GetTemplate("modules/dashboard/timeliste")
	if !ok {
		t.Fatal("malen modules/dashboard/timeliste vart ikkje lasta")
	}

	naa := time.Date(2026, 8, 28, 7, 0, 0, 0, time.Local)
	var rader []Session
	for i, slag := range []string{"yoga", "Pilates", ""} {
		rader = append(rader, NewSession("nn", models.Event{
			Title: "Time", TeacherName: "Nokon", ClassType: slag,
			StartTime: naa.Add(time.Duration(i+2) * time.Hour),
			EndTime:   naa.Add(time.Duration(i+3) * time.Hour),
			Capacity:  20, CurrentEnrolment: 1,
		}, naa.Format("2006-01-02"), naa))
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

// Kartet fraa slag til farge lyt bu ein stad baae som les det kann sjaa
// det. Det stod i 90-aktiviteten.css so lenge brettet var den einaste
// lesaren; heimesida les det no òg.
func TestSlagfargekartetStendITokenfila(t *testing.T) {
	rot := filepath.Join("..", "static", "css", "deler")
	token, err := os.ReadFile(filepath.Join(rot, "00-token.css"))
	if err != nil {
		t.Fatal(err)
	}
	for _, slag := range []string{"fascia", "yoga", "pilates", "reformer"} {
		if !strings.Contains(string(token), ".slag-"+slag) {
			t.Errorf(".slag-%s stend ikkje i 00-token.css", slag)
		}
	}
	aktivitet, err := os.ReadFile(filepath.Join(rot, "90-aktiviteten.css"))
	if err != nil {
		t.Fatal(err)
	}
	// Regel og ikkje understreng: filen *nemner* `--slagfarge` mange
	// gonger — ho les han — og eitt av dei er i ein kommentar som
	// siterer ei gamal tilordning. Det som ikkje skal stå att er ein
	// `.slag-*`-regel som set han.
	if m := regexp.MustCompile(`(?m)^\.slag-\w+`).FindString(string(aktivitet)); m != "" {
		t.Errorf("kartet stend framleis i 90-aktiviteten.css òg (%s) — daa er det tvo", m)
	}
}
