package handsamarar

import (
	"bytes"
	"strings"
	"testing"
)

// The header stands on every page, and the scripts driving it have to be
// sent as widely as it is.
//
// The name menu was not. It lived in faner.js, because it is a picker and
// tabs are pickers — but faner.js was sent only to the pages that *had* a
// tab row: klippekort, membership, payment, admin and the workshop.
// The home page and the schedule have none. There the button stood in the
// header with aria-expanded="false" and nothing listening to it: you
// pressed your own name and nothing happened. It was reported as "the menu
// takes forever to appear" — which is how a dead button behaves when you do
// not know it is dead. You wait, then press again.
//
// The test is written as a *consequence* and not as a list: if the button
// is there, the script should be there. A list of which pages need it would
// have been the same mistake again — somebody adds a page, forgets the
// list, and the test stands green.
func TestNamnemenyenHevSkriptetSittPåKvarSidaHanStendPå(t *testing.T) {
	tm := lastMalane(t)

	// pages/betaling because it renders with almost no data and stands behind
	// 	// the login — so the header stands with the name in it. Which page we
	// 	// render is irrelevant; CurrentPage drives the script conditions in
	// 	// base.html, and those are what we are testing.
	mal, ok := tm.GetTemplate("pages/betaling")
	if !ok {
		t.Fatal("malen pages/betaling vart ikkje lasta")
	}

	// Every value CurrentPage can have in base.html. If somebody changes them
	// 	// this list should follow — but if it does not, the worst that happens
	// 	// is that we test one page too few, not that we let a dead button
	// 	// out.
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
