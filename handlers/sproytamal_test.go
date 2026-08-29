package handlers

import (
	"bytes"
	"strings"
	"testing"

	"time"

	"kjernekraft/handlers/modules"
	"kjernekraft/models"
)

// Ein mal som vert sett inn med `innerHTML` kann ikkje bera skriptet sitt.
//
// `innerHTML` køyrer aldri `<script>`. Det er ikkje ein lyte i nokon
// nettlesar — det stend i standarden — og det tyder at eit skript som
// ligg i ein bolk som vert *henta etter* at sida er lasta, aldri gjeng.
//
// Klippekortbolken gjorde nettupp det. Knappen bar
// `onclick="fillUpKlippekort(...)"`, og funksjonen stod i eit `<script>`
// i den same malen. Paa heimesida gjekk det, av di tenaren teiknar
// bolken der og skriptet fylgjer med sida. Paa klippekortsida vert han
// henta med `fetch` og sett inn med `innerHTML`: knappane synte seg,
// funksjonen fanst ikkje, og eit trykk gav ein `ReferenceError` i
// konsollen og ingen ting paa skjermen. «Fyll på» gjorde ingen ting, og
// berre paa den eine sida.
//
// Prøva held tvo ting: bolken ber ikkje skript, og knappen seier kva han
// vil ha gjort med eit data-attributt i staden. Lyttaren bur i
// `static/js/klippfyll.js` og fangar trykket paa `document`, so han bryr
// seg ikkje um naar knappen kom.
func TestKlippekortbolkenBerIkkjeSkriptetSitt(t *testing.T) {
	mal, ok := lastMalane(t).GetTemplate("pages/dashboard")
	if !ok {
		t.Fatal("malen pages/dashboard vart ikkje lasta")
	}

	kort, err := modules.NewKlippekortModule([]models.KlippekortWithDetails{{
		KlippekortPackage: models.KlippekortPackage{
			ID: 1, Name: "10 klipp Reformer", Category: "Reformer", KlippCount: 10,
		},
		UserKlippekort: models.UserKlippekort{
			ID: 1, TotalKlipp: 10, RemainingKlipp: 3,
			ExpiryDate: time.Now().AddDate(0, 2, 0),
		},
		DaysUntilExpiry: 60,
	}}, "nn")
	if err != nil {
		t.Fatalf("klippekortmodul: %v", err)
	}

	var ut bytes.Buffer
	if err := mal.ExecuteTemplate(&ut, "klippekort_module", kort); err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	if strings.Contains(html, "<script") {
		t.Error("klippekortbolken ber eit <script>. Han vert sett inn med " +
			"innerHTML paa klippekortsida, og daa gjeng det aldri. " +
			"Legg koda i ei fil under static/js/ og lyd paa document.")
	}
	if strings.Contains(html, "onclick=") {
		t.Error("klippekortbolken ber ein onclick. Han peikar paa ein " +
			"funksjon som ikkje finst naar bolken er sett inn med innerHTML.")
	}
	if !strings.Contains(html, "data-fyll-kategori=") {
		t.Error("«Fyll på»-knappen seier ikkje kva kategori han gjeld — " +
			"lyttaren i klippfyll.js har ingen ting aa lesa")
	}
}

// Og skriptet lyt faktisk verta sendt. Ein lyttar som ikkje er lasta er
// den same daude knappen, berre eit hakk lenger unna.
func TestKlippfyllVertSendMedSida(t *testing.T) {
	mal, ok := lastMalane(t).GetTemplate("pages/betaling")
	if !ok {
		t.Fatal("malen pages/betaling vart ikkje lasta")
	}
	for _, side := range []string{"hjem", "klippekort", "medlemskap"} {
		var ut bytes.Buffer
		err := mal.ExecuteTemplate(&ut, "base", map[string]interface{}{
			"Lang": "nn", "Title": "t", "CSRFToken": "x",
			"CurrentPage": side, "UserName": "Solfrid",
		})
		if err != nil {
			t.Errorf("%q: %v", side, err)
			continue
		}
		if !strings.Contains(ut.String(), "klippfyll.js") {
			t.Errorf("%q: klippfyll.js vert ikkje send — «Fyll på» er daud der", side)
		}
	}
}
