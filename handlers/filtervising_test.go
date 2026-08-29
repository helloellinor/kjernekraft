package handlers

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

var (
	// Ogso dei siterte formene: `url("#id")` og `url( #id )` er like
	// gyldig CSS, og ein formaterar som set hermeteikn hadde elles gjort
	// visingi usynleg for prøva utan aa gjera henne usynleg for
	// nettlesaren.
	visingRe = regexp.MustCompile(`url\(\s*["']?#([A-Za-z0-9_-]+)["']?\s*\)`)
	// Blank fyre, so `data-time-id="7"` ikkje vert lese som `id="7"` —
	// eit attributnamn som *endar* paa id hadde elles registrert falske
	// ankerfeste som ei dinglande vising kunde gøyma seg attum.
	idRe    = regexp.MustCompile(`\sid="([A-Za-z0-9_-]+)"`)
	komm    = regexp.MustCompile(`(?s)/\*.*?\*/`)
	malkomm = regexp.MustCompile(`(?s)\{\{/\*.*?\*/\}\}|<!--.*?-->`)
)

// stilarkFraaProva er lesStilarket (stilark_test.go) utan
// CSS-kommentarane: ei forklaring som nemner ei gamal vising er ikkje
// ei vising.
func stilarkFraaProva(t *testing.T) string {
	t.Helper()
	return komm.ReplaceAllString(lesStilarket(t), " ")
}

// alleMalar les alle malfilone som éin streng, med kommentarane tekne
// burt: ein id som berre stend i ein kommentar finst ikkje, og ei
// vising som berre stend i ein kommentar viser ingen stad.
func alleMalar(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	err := filepath.WalkDir("templates", func(stig string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(stig, ".html") {
			return nil
		}
		innhald, err := os.ReadFile(stig)
		if err != nil {
			return err
		}
		b.Write(innhald)
		b.WriteByte('\n')
		return nil
	})
	if err != nil {
		t.Fatalf("las ikkje malane: %v", err)
	}
	return malkomm.ReplaceAllString(b.String(), " ")
}

// Ei vising til eit filter som ikkje finst tek elementet med seg.
//
// Det er det som gjer denne feilen verd ei prøva. Ein kunde venta at
// filteret berre vart hoppa yver — men `filter: url(#finst-ikkje)` er
// ein *feil*, og nettlesaren teiknar daa ikkje elementet i det heile.
// Ingen ting i konsollen, ingi raud line: forma er berre burte.
//
// `.ruteflate` — sjølve flata i det digitale vindauga paa merket — bar
// `filter: url(#rutedjup)`, og det filteret hev aldri vore skrive. So
// vart flata aldri maala. Att stod den myrke yverleppa og talet, og
// vindauga saag ikkje senka ut av di det ikkje fanst nokon botn aa sjaa
// ned i. Han saag ut som han var *nesten* rett, som er den verste
// slags feil aa leita etter.
//
// Prøva tek kvar `url(#…)` og krev at ho peikar paa noko som verkeleg
// finst — og kva «finst» tyder, kjem an paa kvar visingi stend. Den
// teikna sida lyt finna id-en *paa sida*: eit filter som stend i ei
// malfil denne sida ikkje teiknar, hjelper ikkje nettlesaren som
// teiknar henne. (Fall merke-defs ut or base.html, skulde prøva raudna
// — fyrr talde raafila som «finst», og det hadde ho ikkje merkt.) For
// raatekst og stilark er unionen rett: stilarket er delt yver alle
// sidone, og ein raa mal kann høyra til ei onnor sida enn denne.
func TestKvarFiltervisingPeikarPaaNokoSomFinst(t *testing.T) {
	tm := lastMalane(t)
	mal, ok := tm.GetTemplate("pages/timeplan")
	if !ok {
		t.Fatal("malen pages/timeplan vart ikkje lasta")
	}

	var ut bytes.Buffer
	err := mal.ExecuteTemplate(&ut, "base", map[string]interface{}{
		"Lang": "nn", "Title": "t", "CSRFToken": "x",
		"WeekOffset": 0, "WeekNumber": 35, "WeekTitle": "veke 35",
		"VikorIAaret": 52, "CanGoBack": false,
		"WeekDays": []string{}, "WeekDates": []time.Time{},
		"ClassRows": proveRader(), "Today": "", "Teachers": nil,
		"CurrentPage": "timeplan", "UserName": "Solfrid",
	})
	if err != nil {
		t.Fatalf("teikning: %v", err)
	}
	html := ut.String()

	idar := func(tekst string) map[string]bool {
		ut := map[string]bool{}
		for _, m := range idRe.FindAllStringSubmatch(tekst, -1) {
			ut[m[1]] = true
		}
		return ut
	}
	sideIDar := idar(html)
	malane := alleMalar(t)
	alleIDar := idar(malane)
	for id := range sideIDar {
		alleIDar[id] = true
	}

	kjelder := []struct {
		namn  string
		tekst string
		finst map[string]bool
	}{
		{"den teikna sida", html, sideIDar},
		{"raamalane", malane, alleIDar},
		{"stilarket", stilarkFraaProva(t), alleIDar},
	}

	var manglar []string
	for _, k := range kjelder {
		visingar := visingRe.FindAllStringSubmatch(k.tekst, -1)
		// Vaktene i sjølve prøva: ei kjelda utan ei einaste vising er
		// ikkje prøvd — daa hev nokon skrive um korleis merket vert
		// teikna, og prøva stod elles grøn utan aa prøva noko.
		if len(visingar) == 0 {
			t.Errorf("%s hev ingi url(#…) — prøva ser ingen ting der og provar ingen ting", k.namn)
		}
		for _, m := range visingar {
			if !k.finst[m[1]] {
				manglar = append(manglar, k.namn+": url(#"+m[1]+")")
			}
		}
	}
	slices.Sort(manglar)
	manglar = slices.Compact(manglar)

	for _, m := range manglar {
		t.Errorf("%s peikar paa noko som ikkje stend i dokumentet — "+
			"elementet som ber henne vert ikkje teikna i det heile", m)
	}
}

// Og ein tryggleik til: sjølve den ljose underleppa i vindauga — det
// rettingi i f50e471 la til — skal standa i stilarket. At kvar kjelda i
// det heile *ser* visingar, vaktar hovudprøva sjølv no, per kjelda.
func TestFiltervisingsprovaSerNokoIDetHeile(t *testing.T) {
	if !strings.Contains(stilarkFraaProva(t), "ruteljos") {
		t.Error("den ljose underleppa i vindauga er burte or stilarket")
	}
}
