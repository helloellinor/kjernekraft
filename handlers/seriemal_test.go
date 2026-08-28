package handlers

import (
	"os"
	"strings"
	"testing"
	"time"

	"kjernekraft/models"
)

// Ein time som ikkje ber nokon serie skal ikkje by fram felt som ikkje
// kan lagrast.
//
// Rada stod med rule_id="0" fyrr, og felti stod opne. Endra ein noko der,
// gjekk oppdateringa mot `WHERE rule_id = 0`, råka ingi rad og svara 200
// — flata sa «lagra», og verdet var borte ved neste lasting. No står
// felti stengde, og vinket seier kvifor.
func TestRegelUtanNamnStengjerFelti(t *testing.T) {
	// Malsettet vert leita fram fraa arbeidskatalogen, og under `go
	// test` er han pakken sin. Steg upp til rota fyrst.
	gamal, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(gamal)

	tm := GetTemplateManager()
	tm.ReloadTemplates()
	mal, ok := tm.GetTemplate("pages/admin")
	if !ok {
		t.Fatal("malsettet for administrasjonssida lét seg ikkje lasta")
	}

	timar := []models.Event{{
		ID: 7, Title: "Yoga", TeacherName: "Leon", RoomName: "Salen",
		StartTime: time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC),
		Capacity:  12,
	}}

	for _, p := range []struct {
		namn    string
		serieID int64
		stengd  bool
	}{
		{"utan serie", 0, true},
		{"med serie", 42, false},
	} {
		t.Run(p.namn, func(t *testing.T) {
			serie := GrupperTimar(timar, time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC))
			if len(serie) != 1 {
				t.Fatalf("fekk %d seriar, venta 1", len(serie))
			}
			serie[0].SerieID = p.serieID

			var ut strings.Builder
			if err := mal.ExecuteTemplate(&ut, "admin_class_management", map[string]interface{}{
				"Lang": "nn", "Timereglar": serie, "Teachers": []string{"Leon"},
				"CSRFToken": "x", "IsAdmin": true, "UserName": "prøve",
			}); err != nil {
				t.Fatalf("malen feila: %v", err)
			}

			html := ut.String()
			// Felti ber skjemanamn no, ikkje krokklassor: heile
			// serien gjeng i eitt kall gjenom dokka, so kvart felt
			// lyt ha eit `name` tenaren kjenner att.
			stengd := lestengd(t, html, `name="laerar"`) &&
				lestengd(t, html, `name="klokke"`) &&
				lestengd(t, html, `name="minutt"`) &&
				lestengd(t, html, `name="skildring"`)
			if stengd != p.stengd {
				t.Errorf("stengde felt = %v, venta %v", stengd, p.stengd)
			}
			if p.stengd && !strings.Contains(html, t2("nn", "admin.serie_none_hint")) {
				t.Errorf("vinket um at timen ikkje hev nokon serie stend ikkje der")
			}
		})
	}
}

func t2(lang, nykel string) string { return t(lang, nykel) }

// lestengd ser etter `disabled` inni nett den taggen klassa høyrer til,
// og ikkje kvar som helst paa sida.
func lestengd(t *testing.T, html, klasse string) bool {
	i := strings.Index(html, klasse)
	if i < 0 {
		t.Fatalf("fann ikkje %s i det som vart teikna", klasse)
	}
	slutt := strings.Index(html[i:], ">")
	if slutt < 0 {
		t.Fatalf("taggen kring %s er ikkje lukka", klasse)
	}
	return strings.Contains(html[i:i+slutt], "disabled")
}
