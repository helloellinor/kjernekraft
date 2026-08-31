package handsamarar

import (
	"testing"
	"time"
)

// Vikefeltet reknar «kva offset skal eg be um for aa koma til veke N?»
// som `naaOffset + (N - naa)`. Det stemmer berre so lenge vikenummeret
// veks med eitt for kvar vike tenaren legg til. Prøva held det.
func TestVikenummeretVeksMedEittPerOffset(t *testing.T) {
	// Ein torsdag midt i aaret, og ein som ligg tett på aarsskiftet.
	for _, start := range []time.Time{
		time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 12, 17, 12, 0, 0, 0, time.UTC),
	} {
		_, fyrre := VikeMåndag(start, 0).ISOWeek()
		vekerIÅr := VikorIAaret(VikeMåndag(start, 0))

		for offset := 1; offset <= 60; offset++ {
			_, denne := VikeMåndag(start, offset).ISOWeek()
			venta := fyrre + 1
			if venta > vekerIÅr {
				venta = 1
				vekerIÅr = VikorIAaret(VikeMåndag(start, offset))
			}
			if denne != venta {
				t.Fatalf("fraa %s: offset %d gav veke %d, venta %d",
					start.Format("2006-01-02"), offset, denne, venta)
			}
			fyrre = denne
		}
	}
}

// Sundagen høyrer til vika som gjeng ut. Ein som opnar timeplanen
// sundag kveld skal sjå vika han er ferdig med, ikkje den som byrjar
// dagen etter — timane han eventuelt hev att ligg i den han stend i.
func TestSundagenHoyrerTilVikaSomGjengUt(t *testing.T) {
	sundag := time.Date(2026, 8, 30, 20, 0, 0, 0, time.UTC)
	if sundag.Weekday() != time.Sunday {
		t.Fatalf("prøvedagen er %v og ikkje sundag", sundag.Weekday())
	}
	m := VikeMåndag(sundag, 0)
	if m.Weekday() != time.Monday {
		t.Fatalf("fekk %v, venta ein maandag", m.Weekday())
	}
	if d := sundag.Sub(m).Hours() / 24; d < 6 || d > 7 {
		t.Errorf("maandagen ligg %.1f døger fyre sundagen; han skal liggja seks", d)
	}
}

// Kvar måndag VikeMåndag gjev, skal faktisk vera ein måndag — ogso
// når ein kryssar sommartid.
func TestVikeMåndagGjevAlltidEinMåndag(t *testing.T) {
	oslo, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Skipf("inga tidssonetabell: %v", err)
	}
	start := time.Date(2026, 3, 1, 12, 0, 0, 0, oslo)
	for offset := 0; offset <= 80; offset++ {
		if d := VikeMåndag(start, offset).Weekday(); d != time.Monday {
			t.Fatalf("offset %d gav %v", offset, d)
		}
	}
}

// 2026 hev 53 vikor; 2027 hev 52. Talet er det vikefeltet treng for aa
// skyna at eit lægre vikenummer nær aarsskiftet ligg framfyre.
func TestVikorIAaret(t *testing.T) {
	for _, p := range []struct {
		dato  time.Time
		venta int
	}{
		{time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), 53},
		{time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC), 52},
		{time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC), 53},
	} {
		if fekk := VikorIAaret(p.dato); fekk != p.venta {
			t.Errorf("%d: fekk %d vikor, venta %d", p.dato.Year(), fekk, p.venta)
		}
	}
}
