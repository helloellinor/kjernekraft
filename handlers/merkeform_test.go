package handlers

import (
	"math"
	"regexp"
	"strconv"
	"testing"
	"time"
)

// Merket er teikna av tal som stend kvar for seg, og daa kann eitt av
// deim flytta seg utan at dei hine fylgjer med. Det hende: skiva vart
// flutt, og klyppet som heldt kakestykket inne stod att med dei gamle
// tali i ein annan fil — so stykket vart klypt mot der skiva *hadde*
// vore. Prøvone her held delane saman.

// innanfor seier um ein kasse ligg heilt inne i ein annan.
func innanfor(x, y, b, h, ytreX, ytreY, ytreB, ytreH float64) bool {
	return x >= ytreX && y >= ytreY && x+b <= ytreX+ytreB && y+h <= ytreY+ytreH
}

func TestAltITimemerketLiggInnanforKassen(t *testing.T) {
	kasse := func(namn string, x, y, b, h float64) {
		t.Helper()
		if !innanfor(x, y, b, h, MerkeVenstre, 0, MerkeBreidd, MerkeHogd) {
			t.Errorf("%s ligg utanfor kassen: %.1f,%.1f %.1f×%.1f mot 0,0 %.1f×%.1f",
				namn, x, y, b, h, MerkeBreidd, MerkeHogd)
		}
	}
	kasse("kroppen", kroppX, kroppY, kroppB, kroppH)
	kasse("fana", faneX, faneY, faneB, faneH)
	kasse("vindauga", ruteX, ruteY, ruteB, ruteH)
	kasse("plassmerket", plettX-plettR, plettY-plettR, 2*plettR, 2*plettR)
	kasse("skiva", skiveX-spor, skiveY-spor, 2*spor, 2*spor)
}

// Vindauga nede til vinstre og plassmerket nede til høgre skal begge
// liggja inne i kroppen, og dei skal ikkje ta i kvarandre.
func TestBotnstripaStendInneIKroppenOgKolliderarIkkje(t *testing.T) {
	if !innanfor(ruteX, ruteY, ruteB, ruteH, kroppX, kroppY, kroppB, kroppH) {
		t.Error("vindauga stikk utanfor kroppen")
	}
	if !innanfor(plettX-plettR, plettY-plettR, 2*plettR, 2*plettR, kroppX, kroppY, kroppB, kroppH) {
		t.Error("plassmerket stikk utanfor kroppen")
	}
	if ruteX+ruteB > plettX-plettR {
		t.Errorf("vindauga sluttar paa %.1f og plassmerket byrjar paa %.1f — dei ligg uppaa kvarandre",
			ruteX+ruteB, plettX-plettR)
	}
}

// Skiva skal ikkje gaa ned i botnstripa. Ho gjorde det fyrr:
// klokkeslettet var teikna tvert yver henne, og daa las korkje uret
// eller talet seg.
func TestSkivaNaarIkkjeNedIBotnstripa(t *testing.T) {
	botn := math.Min(ruteY, plettY-plettR)
	if skiveY+spor > botn {
		t.Errorf("skiva naar til %.1f, botnstripa byrjar paa %.1f", skiveY+spor, botn)
	}
	if skiveY-spor < kroppY {
		t.Errorf("skiva byrjar paa %.1f, kroppen paa %.1f", skiveY-spor, kroppY)
	}
}

// Kakestykket er timen lagd ut paa døgeret, og det er grunnen til at
// merket er eit ur. Det skal teiknast, og det skal halda seg innanfor
// skiva av seg sjølv — utan eit klypp som kann skilja lag fraa henne.
func TestKakestykketVertTeiknaOgHeldSegInnanforSkiva(t *testing.T) {
	if kakeR > spor {
		t.Fatalf("kakestykket (%.1f) er større enn sporet indeksane ligg paa (%.1f)", kakeR, spor)
	}

	start := time.Date(2026, 8, 27, 18, 0, 0, 0, time.UTC)
	slutt := start.Add(75 * time.Minute)
	m := NyttMerke("prove", start, slutt, 4, 18)

	var timen string
	for _, k := range m.Kaker {
		if k.Klasse == "kake-timen" {
			timen = k.D
		}
	}
	if timen == "" {
		t.Fatal("timen sjølv fekk ikkje noko kakestykke")
	}

	// Berre koordinatane i stykket skal prøvast — ikkje radius-og-flagg-
	// lista i bogen, som stend i den same strengen og ser ut som eit
	// punkt om ein deler han for grovt.
	punkt := regexp.MustCompile(`[ML](-?[\d.]+),(-?[\d.]+)|A[\d.]+,[\d.]+ 0 \d,\d (-?[\d.]+),(-?[\d.]+)`)
	treff := punkt.FindAllStringSubmatch(timen, -1)
	if len(treff) < 3 {
		t.Fatalf("stykket hadde %d punkt, venta minst tri: %s", len(treff), timen)
	}
	for _, tr := range treff {
		xs, ys := tr[1], tr[2]
		if xs == "" {
			xs, ys = tr[3], tr[4]
		}
		x, _ := strconv.ParseFloat(xs, 64)
		y, _ := strconv.ParseFloat(ys, 64)
		if d := math.Hypot(x-skiveX, y-skiveY); d > kakeR+0.05 {
			t.Errorf("punktet %.2f,%.2f ligg %.2f fraa skivemidten, kakeR er %.2f", x, y, d, kakeR)
		}
	}
}

// Kakestykket skal veksa ut or minuttvisaren og ikkje or timevisaren.
//
// Det låg på timevisaren fyrr — den korte, tjukke, den eine som ikkje
// rører seg medan timen gjeng. Det er minuttvisaren som gjeng medan du
// ligg på matta. Prøva held stykket der, og held sluttvisaren i den
// andre kanten av det, so dei tri les seg som éin ting.
func TestKakestykketHengIMinuttvisaren(t *testing.T) {
	fyrstePunkt := regexp.MustCompile(`^M[-\d.]+,[-\d.]+ L([-\d.]+),([-\d.]+)`)

	for _, p := range []struct {
		namn         string
		start, slutt string
		grader       float64
	}{
		{"ein time", "2026-08-27 07:15", "2026-08-27 08:15", 30},
		{"trikvart", "2026-08-27 12:00", "2026-08-27 12:45", 22.5},
		{"fem og sytti minutt", "2026-08-27 18:30", "2026-08-27 19:45", 37.5},
		{"yver eit heilt timeskifte", "2026-08-27 11:50", "2026-08-27 12:50", 30},
	} {
		start, _ := time.Parse("2006-01-02 15:04", p.start)
		slutt, _ := time.Parse("2006-01-02 15:04", p.slutt)
		m := NyttMerke("prova", start, slutt, 4, 18)

		if got := varigheit(start, slutt); math.Abs(got-p.grader) > 0.01 {
			t.Errorf("%s: stykket vart %.1f grader, venta %.1f", p.namn, got, p.grader)
		}

		// Sluttvisaren stend der stykket sluttar.
		venta := minuttvinkel(start) + p.grader
		if math.Abs(m.SluttVinkel-venta) > 0.01 {
			t.Errorf("%s: sluttvisaren stend paa %.1f, venta %.1f", p.namn, m.SluttVinkel, venta)
		}

		// Og stykket sjølv byrjar i minuttvisaren: fyrste linestykket i
		// pathen gjeng fraa midten ut til startvinkelen.
		var timen string
		for _, k := range m.Kaker {
			if k.Klasse == "kake-timen" {
				timen = string(k.D)
			}
		}
		if timen == "" {
			t.Fatalf("%s: fann ikkje kakestykket", p.namn)
		}
		hit := fyrstePunkt.FindStringSubmatch(timen)
		if hit == nil {
			t.Fatalf("%s: skjøna ikkje pathen %q", p.namn, timen)
		}
		x, _ := strconv.ParseFloat(hit[1], 64)
		y, _ := strconv.ParseFloat(hit[2], 64)
		vx, vy := punkt(minuttvinkel(start), kakeR, skiveX, skiveY)
		if math.Abs(x-vx) > 0.05 || math.Abs(y-vy) > 0.05 {
			t.Errorf("%s: stykket byrjar i %.2f,%.2f, men minuttvisaren stend i %.2f,%.2f",
				p.namn, x, y, vx, vy)
		}
	}
}
