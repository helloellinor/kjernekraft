package handlers

import (
	"bytes"
	"strings"
	"testing"
)

// Hovudet stend paa kvar side, og skripti som styrer det lyt sendast
// like breidt som det.
//
// Namnemenyen gjorde ikkje det. Han laag i `faner.js`, av di han er ein
// veljar og fanone er veljarar — men `faner.js` vert send til dei sidone
// som *hev* ei fanerekkje, og det er klippekort, medlemskap, betaling,
// admin og arket. Heimesida og timeplanen hev ingi. Der stod difor
// knappen i hovudet med `aria-expanded="false"` og ingen ting som lydde
// paa honom: du trykte paa namnet ditt, og ingen ting hende. Det vart
// meldt som «det tek ei æve fyre menyen kjem» — og det er slik ein daud
// knapp ter seg naar ein ikkje veit at han er daud. Ein ventar, og so
// trykkjer ein att.
//
// Prøva er skrivi som ei *fylgje* og ikkje som ei lista: stend knappen
// der, skal skriptet ogso standa der. Ei lista yver kva sidor som treng
// honom hadde vore den same feilen ein gong til — nokon legg til ei
// side, gløymer lista, og prøva stend grøn.
func TestNamnemenyenHevSkriptetSittPaaKvarSidaHanStendPaa(t *testing.T) {
	tm := lastMalane(t)

	// `pages/betaling` av di han teiknar seg med mest ingi data og
	// stend attum innloggingi — daa stend hovudet med namnet i. Kva
	// side me teiknar er likegyldig; det er `CurrentPage` som styrer
	// skript-vilkori i base.html, og det er dei me prøver.
	mal, ok := tm.GetTemplate("pages/betaling")
	if !ok {
		t.Fatal("malen pages/betaling vart ikkje lasta")
	}

	// Alle verdi `CurrentPage` kann ha i base.html. Skiftar nokon paa
	// deim, skal denne lista fylgja etter — men fylgja ho ikkje, er det
	// verste som hender at me prøver ei side for lite, ikkje at me
	// slepp ein daud knapp ut.
	sidor := []string{
		"hjem", "timeplan", "klippekort", "medlemskap",
		"betaling", "admin", "arket", "profil", "",
	}

	for _, side := range sidor {
		var ut bytes.Buffer
		err := mal.ExecuteTemplate(&ut, "base", map[string]interface{}{
			"Lang": "nn", "Title": "t", "CSRFToken": "x",
			"CurrentPage": side, "UserName": "Solfrid",
		})
		if err != nil {
			t.Errorf("%q: teikning: %v", side, err)
			continue
		}
		html := ut.String()

		knapp := strings.Contains(html, `id="namn-knapp"`)
		skript := strings.Contains(html, "namnemeny.js")

		if knapp && !skript {
			t.Errorf("%q: namneknappen stend i hovudet, men namnemeny.js vert ikkje send. "+
				"Knappen er daud paa den sida.", side)
		}
		// Den andre vegen er ikkje ein feil — eit skript utan knapp gjer
		// ingen skade — men det er sløsing, og det tyder som oftast at
		// noko anna er gale.
		if skript && !knapp {
			t.Errorf("%q: namnemeny.js vert send, men knappen stend ikkje der", side)
		}
	}
}
