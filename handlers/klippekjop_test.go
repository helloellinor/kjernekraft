package handlers

import (
	"bytes"
	"html/template"
	"kjernekraft/models"
	"regexp"
	"strings"
	"testing"
)

func pakke(id int, kategori string, klipp, pris int, populaer bool) models.KlippekortPackage {
	return models.KlippekortPackage{
		ID: id, Name: kategori, Category: kategori,
		KlippCount: klipp, Price: pris, PricePerSession: pris / klipp,
		ValidDays: 90, Active: true, IsPopular: populaer,
	}
}

func TestKategoriarGrupperarIRekkja(t *testing.T) {
	k := Kategoriar([]models.KlippekortPackage{
		pakke(1, "Gruppetimer Sal", 5, 49900, false),
		pakke(2, "Gruppetimer Sal", 10, 89900, true),
		pakke(3, "Reformer/Apparatus", 5, 74900, false),
	})
	if len(k) != 2 {
		t.Fatalf("%d kategoriar, venta tvo", len(k))
	}
	if len(k[0].Pakkar) != 2 || len(k[1].Pakkar) != 1 {
		t.Errorf("pakkane fordelte seg gale: %d og %d", len(k[0].Pakkar), len(k[1].Pakkar))
	}
}

// Nykelen stend i ein id og i ei adressa. «Reformer/Apparatus» vart
// elles tvo stigsteg i ei adressa.
func TestNykelenTolerAaStaaIEiAdressa(t *testing.T) {
	for inn, ut := range map[string]string{
		"Reformer/Apparatus":  "reformer-apparatus",
		"Gruppetimer Sal":     "gruppetimer-sal",
		"Sjølvøving på Måtte": "sjoelvoeving-paa-maatte",
	} {
		if fekk := nykel(inn); fekk != ut {
			t.Errorf("nykel(%q) = %q, venta %q", inn, fekk, ut)
		}
	}
	if strings.ContainsAny(nykel("A/B C"), "/ ") {
		t.Error("nykelen slepp gjenom teikn som ikkje toler ei adressa")
	}
}

// Sida skal ikkje kunna segja ein pris basen ikkje hev sagt. Prøva
// teiknar henne og ser etter at kvart kronetal som stend der, kjem fraa
// ei pakke ho fekk inn.
func TestSidaSeierBerreDetBasenSeier(t *testing.T) {
	pakkar := []models.KlippekortPackage{
		pakke(1, "Gruppetimer Sal", 5, 49900, false),
		pakke(2, "Gruppetimer Sal", 10, 89900, true),
		pakke(3, "Reformer/Apparatus", 5, 74900, false),
	}

	tm := &TemplateManager{templates: make(map[string]*template.Template), basePath: "templates"}
	tm.loadTemplates()
	mal, ok := tm.GetTemplate("pages/klippekort")
	if !ok {
		t.Fatal("malen pages/klippekort vart ikkje lasta")
	}

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "content", map[string]interface{}{
		"Lang": "nn", "Kategoriar": Kategoriar(pakkar),
	}); err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	// skrivKronor skil tusen med hardt mellomrom (U+00A0) og set eit
	// hardt mellomrom fyre «kr». Ei prøva som leitar etter vanleg
	// mellomrom finn ingen ting og gjeng gjenom av di ho ikkje prøver
	// noko — det hende her, og difor tel me treffi fyrst.
	pris := regexp.MustCompile("([0-9\u00a0]+)\u00a0kr")
	treff := pris.FindAllStringSubmatch(html, -1)
	if len(treff) < 2*len(pakkar) {
		t.Fatalf("fann %d kronetal, venta minst %d — proova ser ikkje det ho skal",
			len(treff), 2*len(pakkar))
	}

	lovlege := map[string]bool{"499": true, "899": true, "749": true,
		"99": true, "89": true, "149": true} // pakkepris og pris per gong
	for _, tr := range treff {
		tal := strings.ReplaceAll(tr[1], "\u00a0", "")
		if !lovlege[tal] {
			t.Errorf("sida seier «%s kr», som ingi pakke hev sagt", tal)
		}
	}

	// Kvar pakke skal kunna kjøpast: id-en er det kjøpet sender.
	for _, p := range pakkar {
		if !strings.Contains(html, `data-pakke="`+itoa(p.ID)+`"`) {
			t.Errorf("pakke %d hev ingen id paa sida — ho kann ikkje kjøpast", p.ID)
		}
	}

	// Det som var gale med sida fyrr skal ikkje koma attende.
	for _, forbode := range []string{"alert(", "confirm(", "🧘", "💪", "Stressmestring", "Steg 1:"} {
		if strings.Contains(html, forbode) {
			t.Errorf("sida hev framleis %q i seg", forbode)
		}
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}
