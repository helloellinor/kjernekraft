package handsamarar

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
	// The quoted forms too: url("#id") and url( #id ) are equally valid CSS,
	// 	// and a formatter that adds quotes would otherwise make the reference
	// 	// invisible to the test without making it invisible to the browser.
	// `url(#…)` er filter og fyllingar; `<use href="#…">` er attbruk.
	// Baae peikar på eit anker i dokumentet, og baae teiknar *ingen ting*
	// i det stille om ankeret ikkje finst — eit merke utan indeksar ser
	// ut som eit merke, berre tomt.
	//
	// Berre `href` på <use>: ei vanleg lenkje til `#byt` peikar på eit
	// anker i den *ferdige* sida og ikkje i råmalen, og ho er ikkje ei
	// teikning som kann falla burt.
	visingRe = regexp.MustCompile(`url\(\s*["']?#([A-Za-z0-9_-]+)["']?\s*\)|<use[^>]*href="#([A-Za-z0-9_-]+)"`)
	// Blank fyre, so `data-time-id="7"` ikkje vert lese som `id="7"` —
	// eit attributnamn som *endar* på id hadde elles registrert falske
	// ankerfeste som ei dinglande vising kunde gøyma seg attum.
	idRe    = regexp.MustCompile(`\sid="([A-Za-z0-9_-]+)"`)
	komm    = regexp.MustCompile(`(?s)/\*.*?\*/`)
	malkomm = regexp.MustCompile(`(?s)\{\{/\*.*?\*/\}\}|<!--.*?-->`)
)

// stilarkFråProva er lesStilarket (stilark_test.go) utan
// CSS-kommentarane: ei forklaring som nemner ei gamal vising er ikkje
// ei vising.
func stilarkFråProva(t *testing.T) string {
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

// A reference to a filter that does not exist takes the element with it.
//
// That is what makes this fault worth a test. You might expect the filter
// simply to be skipped — but filter: url(#does-not-exist) is an *error*,
// and the browser then does not draw the element at all. Nothing in the
// console, no red line: the shape is simply gone.
//
// .ruteflate — the surface of the digital window on the mark — carried
// filter: url(#rutedjup), and that filter has never been written. So the
// surface was never painted. What was left was the dark top lip and the
// number, and the window did not look sunken because there was no floor to
// look down into. It looked *almost* right, which is the worst kind of
// fault to hunt.
//
// The test takes every url(#…) and requires it to point at something that
// really exists — and what "exists" means depends on where the reference
// stands. The rendered page has to find the id *on the page*: a filter in a
// template file this page does not render does not help the browser
// rendering it. For raw text and the stylesheet the union is right: the
// sheet is shared across all pages, and a raw template may belong to
// another page.
func TestKvarFiltervisingPeikarPåNokoSomFinst(t *testing.T) {
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
		{"stilarket", stilarkFråProva(t), alleIDar},
	}

	var manglar []string
	for _, k := range kjelder {
		visingar := visingRe.FindAllStringSubmatch(k.tekst, -1)
		// Guards in the test itself: a source with not a single reference is not
		// 		// tested — that means somebody rewrote how the mark is drawn, and the
		// 		// test would otherwise stand green without testing anything.
		if len(visingar) == 0 {
			t.Errorf("%s hev ingi url(#…) — prøva ser ingen ting der og provar ingen ting", k.namn)
		}
		for _, m := range visingar {
			namn := m[1]
			if namn == "" {
				namn = m[2]
			}
			if !k.finst[namn] {
				manglar = append(manglar, k.namn+": #"+namn)
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
//
// Ho stod som `.ruteljos` med eit SVG-filter fyrr. Vindauga er ein runda
// firkant, so lippone er `box-shadow: inset` på `.merkerute` no — men
// *eigenskapen* prøva vaktar er den same: det skal finnast ei ljos leppa
// i nedkanten av vindauga. Difor ser ho etter verkemiddelet og ikkje
// etter klassenamnet.
func TestFiltervisingsprovaSerNokoIDetHeile(t *testing.T) {
	ark := stilarkFråProva(t)
	i := strings.Index(ark, ".merkerute")
	if i < 0 {
		t.Fatal("vindauga (.merkerute) er burte or stilarket")
	}
	blokk := ark[i:]
	if j := strings.Index(blokk, "}"); j > 0 {
		blokk = blokk[:j]
	}
	if !strings.Contains(blokk, "--lippe-ljos") {
		t.Error("den ljose underleppa i vindauga er burte or stilarket")
	}
	if !strings.Contains(blokk, "inset") {
		t.Error("lippone i vindauga er ikkje innskuggar lenger")
	}
}
