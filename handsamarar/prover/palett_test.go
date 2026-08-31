package prover

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"kjernekraft/handsamarar"
)

// arket gives the assembled stylesheet, fetched the way a browser
// fetches it. The test used to glue the parts together itself, which
// meant it tested its own ordering rather than the house's.
func arket(t *testing.T) string {
	t.Helper()
	// Handsamaren les delane med ein stig som gjeng ut frå arbeidsmappa,
	// og `go test` fær då ikkje vita kva filer svaret kviler på. So me
	// opnar deim sjølve, med heile stigen, fyre me spør: då stend dei i
	// testloggen, og eit skift i kva del som helst gjer svaret ugildt.
	delar, err := filepath.Glob(del("*.css"))
	if err != nil || len(delar) == 0 {
		t.Fatalf("fann ingen delar av stilarket: %v", err)
	}
	for _, d := range delar {
		if _, err := os.ReadFile(d); err != nil {
			t.Fatal(err)
		}
	}

	w := httptest.NewRecorder()
	// Stilarket les frå disken og rører ikkje basen, so huset treng ingi
	// tilkopling for å svara her.
	handsamarar.NyApp(nil).Stilark(w, httptest.NewRequest(http.MethodGet, "/static/css/kjernekraft.css", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("stilarket svara %d", w.Code)
	}
	return w.Body.String()
}

// Values and meaning live apart, and must keep doing so.
//
// ARKET §1 has always said each flat colour is written once as --l-/--m-
// and the four blocks only give it meaning. That was not true of the
// files: both lived in 00-token.css. The values are in 00-palett.css now
// — the one file you rewrite when the house changes colour.
//
// This looks for declarations, not mentions: 00-token.css reads --l-ark
// in every block and should. What it must not do is write them.
func TestVerdiOgTydingBurKvarForSeg(t *testing.T) {
	// Namnet med eit kolon etter er ei utsegn; `var(--l-ark)` hev ein
	// parentes der og vert ikkje teken. Ho var ankra til lineskiftet
	// fyrst, og då slapp `:root { --l-nytt: #fff; }` — same lina som
	// veljaren — beint gjenom.
	utsegn := regexp.MustCompile(`(--[lm]-[a-z-]+)\s*:`)
	// Prosa som nemner eit token er ikkje ei utsegn.
	kommentar := regexp.MustCompile(`(?s)/\*.*?\*/`)
	les := func(fil string) string {
		b, err := os.ReadFile(del(fil))
		if err != nil {
			t.Fatal(err)
		}
		return kommentar.ReplaceAllString(string(b), "")
	}

	for _, m := range utsegn.FindAllStringSubmatch(les("00-token.css"), -1) {
		t.Errorf("%s vert skriven i 00-token.css — flate verd bur i 00-palett.css", m[1])
	}

	skrivne := map[string]bool{}
	for _, m := range utsegn.FindAllStringSubmatch(les("00-palett.css"), -1) {
		skrivne[m[1]] = true
	}
	if len(skrivne) == 0 {
		t.Fatal("00-palett.css skriv ingen flate verd i det heile")
	}

	// Kvart flatt verd nokon les lyt vera skrive. Ein `var(--l-noko)` som
	// ingen hev sett gjev ingen farge og ingi åtvaring — han vert berre
	// usynleg, og det er den verste maaten aa mangla på.
	for _, m := range regexp.MustCompile(`var\((--[lm]-[a-z-]+)\)`).FindAllStringSubmatch(arket(t), -1) {
		if !skrivne[m[1]] {
			t.Errorf("%s vert lesen i arket, men stend ikkje i 00-palett.css", m[1])
		}
	}
}
