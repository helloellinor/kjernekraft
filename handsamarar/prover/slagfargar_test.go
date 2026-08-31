package prover

import (
	"os"
	"regexp"
	"testing"

	"kjernekraft/handsamarar"
)

// class_type is free text, so the kind must be washed before it becomes
// a CSS hook — and then looked up. "Hot Yoga" used to yield
// "slag-hotyoga": a class nobody drew, absent from every source file so
// the dead-class check could not find it, and a wing wearing a colour
// the house does not have. Now it falls away and the wing goes grey.
func TestSlagklasseVaskarKlassetypen(t *testing.T) {
	fasit := map[string]string{
		"yoga":         "slag-yoga",
		"Yoga":         "slag-yoga",
		"  Reformer ":  "slag-reformer",
		"Hot Yoga":     "",
		"pilates-matt": "",
		"":             "",
		"   ":          "",
		"123":          "",
		// Ingen veg ut or klasseattributtet.
		`" onclick="x`: "",
	}
	for inn, ut := range fasit {
		if fekk := handsamarar.SlagKlasse(inn); fekk != ut {
			t.Errorf("SlagKlasse(%q) = %q, venta %q", inn, fekk, ut)
		}
	}
}

// The kinds lived in four places and disagreed. Now `Slagi` in slag.go
// is the list, and this keeps the stylesheet in step with it.
//
// Both directions: a kind without a colour looks like one nobody got
// round to drawing, and a colour without a kind is a rule no class can
// reach.
func TestSlagfargarKjennerDeiSameSlagi(t *testing.T) {
	palett, err := os.ReadFile(del("00-palett.css"))
	if err != nil {
		t.Fatal(err)
	}

	// Berre reglar som byrjar på lina. Fila *nemner* `.slag-yoga` i ein
	// kommentar, og ein kommentar er ikkje ein regel.
	iArket := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^\.slag-(\w+)`).FindAllStringSubmatch(string(palett), -1) {
		iArket[m[1]] = true
	}

	for _, slag := range handsamarar.Slagi {
		if !iArket[slag] {
			t.Errorf("`%s` stend i Slagi, men fær ingen farge i 00-palett.css", slag)
		}
		delete(iArket, slag)
	}
	for slag := range iArket {
		t.Errorf("`.slag-%s` stend i 00-palett.css, men ikkje i Slagi — ingen time kann naa honom", slag)
	}
}

// The map must live in one place. It has moved twice — out of
// 90-aktiviteten.css when the home page started reading it too, then
// into 00-palett.css next to the values it points at.
func TestKartetStendBerreEinStad(t *testing.T) {
	for _, fil := range []string{"90-aktiviteten.css", "00-token.css"} {
		b, err := os.ReadFile(del(fil))
		if err != nil {
			t.Fatal(err)
		}
		// Serie og ikkje understreng: filene *nemner* `--slagfarge` og
		// `.slag-*` i kommentarar — dei les kartet og fortel um det. Det
		// som ikkje skal stå att er ein `.slag-*`-serie som set fargen.
		if m := regexp.MustCompile(`(?m)^\.slag-\w+`).FindString(string(b)); m != "" {
			t.Errorf("kartet stend i %s òg (%s) — daa er det tvo", fil, m)
		}
	}
}
