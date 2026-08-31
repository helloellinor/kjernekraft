package handsamarar

import (
	"regexp"
	"strings"
	"testing"
)

// Kommentarane er 70 % av det arket veg. Dei skal liggja urørde på
// disken og aldri gå ut på netet — men *berre* dei. Fell ein regel med
// dei, er sida i stykke, og det syner seg fyrst i nettlesaren.
func TestArketMistarIkkjeReglarNaarKommentaraneGjeng(t *testing.T) {
	inn := []byte(`/* fyrste kommentar */
.kort { color: red; }
/* ein
   som gjeng yver
   fleire liner */
.merke { --a: 1; }
.hank { color: blue; } /* på slutten av lina */
`)
	ut := string(utanKommentar(inn))

	for _, regel := range []string{".kort", "color: red", ".merke", "--a: 1", ".hank", "color: blue"} {
		if !strings.Contains(ut, regel) {
			t.Errorf("regelen %q fall ut med kommentarane:\n%s", regel, ut)
		}
	}
	if strings.Contains(ut, "kommentar") || strings.Contains(ut, "fleire liner") {
		t.Errorf("kommentar stod att:\n%s", ut)
	}
}

// Ein uavslutta kommentar skal ikkje eta resten av arket stillteiande i
// tillegg til å vera ein feil — men han er ein feil i kjelda, so det
// einaste kravet er at me ikkje krasjar.
func TestUavslutaKommentarKrasjarIkkje(t *testing.T) {
	ut := utanKommentar([]byte(".a { color: red; }\n/* aldri lukka"))
	if !strings.Contains(string(ut), ".a") {
		t.Errorf("regelen fyre den uavslutta kommentaren fall ut: %q", ut)
	}
}

// Det verkelege arket: like mange klammeparentesar fyre og etter, so
// ingen regel er broten på vegen.
func TestHeileArketHevLikeManageKlammerEtterStriping(t *testing.T) {
	gamal := stilarkMappe
	stilarkMappe = "../static/css/deler"
	defer func() { stilarkMappe = gamal }()

	heile, err := byggStilark()
	if err != nil {
		t.Fatalf("bygde ikkje arket: %v", err)
	}
	tel := func(s []byte, r byte) int {
		n := 0
		for _, b := range s {
			if b == r {
				n++
			}
		}
		return n
	}
	if a, b := tel(heile, '{'), tel(heile, '}'); a != b {
		t.Errorf("ubalanserte klammer etter striping: %d { mot %d }", a, b)
	}
	if regexp.MustCompile(`/\*`).Match(heile) {
		t.Errorf("det stend kommentarar att i det ferdige arket")
	}
	t.Logf("arket etter striping: %d B", len(heile))
}
