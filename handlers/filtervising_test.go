package handlers

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

var (
	visingRe = regexp.MustCompile(`url\(#([A-Za-z0-9_-]+)\)`)
	idRe     = regexp.MustCompile(`id="([A-Za-z0-9_-]+)"`)
	komm     = regexp.MustCompile(`(?s)/\*.*?\*/`)
)

// stilarkFraaProva peikar arksamlaren eit hakk opp, slik dei hine
// arkprøvone gjer: dei gjeng med `handlers/` som arbeidsmappa.
func stilarkFraaProva(t *testing.T) string {
	t.Helper()
	gamal := stilarkMappe
	stilarkMappe = "../static/css/deler"
	defer func() { stilarkMappe = gamal }()
	ark, err := byggStilark()
	if err != nil {
		t.Fatalf("stilarket lét seg ikkje setja saman: %v", err)
	}
	return komm.ReplaceAllString(string(ark), " ")
}

// alleMalar les alle malfilone som éin streng, so id-ar som stend
// bokstavleg i ein mal me ikkje teiknar her, likevel tel som «finst».
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
	return b.String()
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
// Prøva tek kvar `url(#…)` i stilarket og i det merket teiknar, og
// krev at ho peikar paa noko som verkeleg stend i dokumentet.
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

	// Alt som *finst*. Tvo kjelder, av di stilarket er delt yver alle
	// sidone medan denne prøva berre teiknar éi: id-ar som stend
	// bokstavleg i ein mal (gradientane i aktivitetsbolken, til dømes)
	// vert lesne beint or malfilone, og dei som fyrst vert til naar ein
	// mal er teikna (`glas-{{.Form.Ident}}`) kjem fraa den teikna sida.
	finst := map[string]bool{}
	for _, m := range idRe.FindAllStringSubmatch(html, -1) {
		finst[m[1]] = true
	}
	for _, m := range idRe.FindAllStringSubmatch(alleMalar(t), -1) {
		finst[m[1]] = true
	}

	// Alt som vert *vist til* — baade fraa malen og fraa stilarket.
	// Kommentarane i CSS-en er tekne burt fyrst: ei forklaring som
	// nemner ei gamal vising er ikkje ei vising.
	kjelder := map[string]string{
		"malen":     html,
		"stilarket": stilarkFraaProva(t),
	}

	var manglar []string
	for namn, tekst := range kjelder {
		for _, m := range visingRe.FindAllStringSubmatch(tekst, -1) {
			if !finst[m[1]] {
				manglar = append(manglar, namn+": url(#"+m[1]+")")
			}
		}
	}
	sort.Strings(manglar)
	manglar = unike(manglar)

	for _, m := range manglar {
		t.Errorf("%s peikar paa noko som ikkje stend i dokumentet — "+
			"elementet som ber henne vert ikkje teikna i det heile", m)
	}
}

func unike(s []string) []string {
	var ut []string
	for i, v := range s {
		if i == 0 || v != s[i-1] {
			ut = append(ut, v)
		}
	}
	return ut
}

// Og ein tryggleik: prøva yver er berre verd noko um ho *ser* visingar.
// Finn ho ingen, hev nokon skrive um korleis merket vert teikna, og daa
// stend ho grøn utan aa prøva noko.
func TestFiltervisingsprovaSerNokoILDetHeile(t *testing.T) {
	ark := stilarkFraaProva(t)
	tal := len(visingRe.FindAllString(ark, -1))
	if tal == 0 {
		t.Error("fann ingi url(#…) i stilarket — prøva yver prøver ingen ting")
	}
	if !strings.Contains(ark, "ruteljos") {
		t.Error("den ljose underleppa i vindauga er burte or stilarket")
	}
}
