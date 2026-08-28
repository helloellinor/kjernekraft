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
		if !innanfor(x, y, b, h, MarkLeft, 0, MarkWidth, MarkHeight) {
			t.Errorf("%s ligg utanfor kassen: %.1f,%.1f %.1f×%.1f mot 0,0 %.1f×%.1f",
				namn, x, y, b, h, MarkWidth, MarkHeight)
		}
	}
	kasse("kroppen", kroppX, kroppY, kroppB, kroppH)
	kasse("fana", faneX, faneY, faneB, faneH)
	kasse("vindauga", ruteX, ruteY, ruteB, ruteH)
	kasse("skiva", skiveX-sporB, skiveY-sporOpp, 2*sporB, sporOpp+sporNed)
}

// Vindauga nede til vinstre skal liggja inne i kroppen.
//
// Plassmerket laag nede til høgre og vart prøvd her med. Det er ikkje
// teikna i SVG-en lenger — det er eit HTML-element uppaa figuren — so
// det som kann prøvast her er staden det vert sentrert paa.
func TestBotnstripaStendInneIKroppenOgKolliderarIkkje(t *testing.T) {
	if !innanfor(ruteX, ruteY, ruteB, ruteH, kroppX, kroppY, kroppB, kroppH) {
		t.Error("vindauga stikk utanfor kroppen")
	}
	// Hyrna merket vert sentrert paa *er* hyrna nede til høgre paa
	// kroppen. Gjeng dei fraa kvarandre, flyt bobla ut i lufta.
	if vil := (kroppX + kroppB) / MarkWidth * 100; hyrneX != vil {
		t.Errorf("hyrneX er %.3f%%, kroppen sluttar paa %.3f%%", hyrneX, vil)
	}
	if vil := (kroppY + kroppH) / MarkHeight * 100; hyrneY != vil {
		t.Errorf("hyrneY er %.3f%%, kroppen sluttar paa %.3f%%", hyrneY, vil)
	}
}

// Skiva gjeng under vindauga, men ikkje ut or kroppen.
//
// Ho laut halda seg yver botnstripa fyrr, av di klokkeslettet var
// teikna tvert yver henne og daa las korkje uret eller talet seg. Det
// er ikkje sant lenger: vindauga er ei ugjenomsynleg flate som vert
// teikna *etter* skiva, so han dekkjer det han dekkjer. Difor kann
// indeksane naa dei nedre hyrni, som paa eit firkanta ur.
func TestSkivaHeldSegInneIKroppen(t *testing.T) {
	if skiveY+sporNed > kroppY+kroppH {
		t.Errorf("skiva naar til %.1f, kroppen sluttar paa %.1f", skiveY+sporNed, kroppY+kroppH)
	}
	if skiveY-sporOpp < kroppY {
		t.Errorf("skiva byrjar paa %.1f, kroppen paa %.1f", skiveY-sporOpp, kroppY)
	}
}

func TestSporetErEinFirkantOgIkkjeEinSirkel(t *testing.T) {
	avstand := func(g float64) float64 {
		x, y := onDialTrack(g, 0)
		return math.Hypot(x-skiveX, y-skiveY)
	}
	// Rett upp og ned møter han toppen og botnen; til sidone, sidone.
	for _, p := range []struct {
		g   float64
		vil float64
	}{{0, sporOpp}, {180, sporNed}, {90, sporB}, {270, sporB}} {
		if d := avstand(p.g); math.Abs(d-p.vil) > 0.01 {
			t.Errorf("%.0f grader naar %.2f, venta %.2f", p.g, d, p.vil)
		}
	}
	// Og i hyrni naar han lenger enn baae — det er nettupp det som gjer
	// ein firkant til ein firkant. Ein sirkel naar like langt overalt.
	for _, g := range []float64{45, 135, 225, 315} {
		d := avstand(g)
		if d <= sporB || d <= sporOpp {
			t.Errorf("%.0f grader naar %.2f — ikkje lenger enn sidone (%.1f/%.1f), so det er ein sirkel",
				g, d, sporB, sporOpp)
		}
	}
	// Klokka tvo og ti skal liggja i hyrni, ikkje paa ein sirkel: dei
	// naar lenger enn tolv gjer.
	for _, g := range []float64{60, 120, 240, 300} {
		if d := avstand(g); d <= sporOpp {
			t.Errorf("%.0f grader naar %.2f, ikkje meir enn toppen (%.1f)", g, d, sporOpp)
		}
	}
	// Ingen ting av skiva gjeng ut or kroppen eller ned i vindauga.
	for g := 0.0; g < 360; g += 3 {
		x, y := onDialTrack(g, 0)
		if x < kroppX || x > kroppX+kroppB || y < kroppY || y > kroppY+kroppH {
			t.Errorf("%.0f grader gjeng ut or kroppen: %.1f,%.1f", g, x, y)
		}
	}
}

func TestVisaraneErRettVegen(t *testing.T) {
	start, _ := time.Parse("2006-01-02 15:04", "2026-08-27 06:00")
	slutt, _ := time.Parse("2006-01-02 15:04", "2026-08-27 07:00")
	m := NewMark("prova", start, slutt, 4, 18)

	timeLengd := skiveY - m.HourHand
	minuttLengd := skiveY - m.MinuteHand
	if timeLengd >= minuttLengd {
		t.Errorf("teikna timevisar er %.1f og minuttvisar %.1f — timevisaren skal vera kortast",
			timeLengd, minuttLengd)
	}
	// Sluttvisaren *er* ein timevisar og skal vera like lang.
	if m.HourHand != m.HourHand {
		t.Error("sluttvisaren nyttar ikkje timevisaren si lengd")
	}
	// Klokka seks: timevisaren peikar ned, minuttvisaren rett upp.
	if m.HourAngle != 180 {
		t.Errorf("timevisaren stend paa %.1f grader klokka seks, venta 180", m.HourAngle)
	}
	if m.MinuteAngle != 0 {
		t.Errorf("minuttvisaren stend paa %.1f grader klokka seks, venta 0", m.MinuteAngle)
	}
}

func TestTimevisarenErKortareEnnMinuttvisaren(t *testing.T) {
	if visarTime >= visarMinutt {
		t.Errorf("timevisaren er %.1f og minuttvisaren %.1f — timevisaren skal vera den korte",
			visarTime, visarMinutt)
	}
}

// Ingen visar skal stikka ut or kroppen. Skiva ligg nærare yverkanten
// enn dei hine kantane, so det er dit ein visar naar fyrst — og det er
// nettupp der ein visar som peikar paa tolv gjeng.
//
// Timevisaren gjorde det: 23,9 lang fraa ei skiva som ligg 22 fraa
// yverkanten, so han stakk ut kvar heile time.
func TestVisaraneStikkIkkjeUtOrKroppen(t *testing.T) {
	// Halve stroken tel med: linjeenden er runda.
	const halvStrok = 1.15
	rom := skiveY - kroppY - halvStrok
	for _, v := range []struct {
		namn  string
		lengd float64
	}{
		{"timevisaren", visarTime},
		{"minuttvisaren", visarMinutt},
	} {
		if v.lengd > rom {
			t.Errorf("%s er %.1f lang, men det er berre %.1f upp til kanten aat kroppen",
				v.namn, v.lengd, rom)
		}
	}
}

func TestKakestykketVertTeiknaOgHeldSegInnanforSkiva(t *testing.T) {
	// Stykket skal aldri naa lenger ut enn visarane det stend for. Naar
	// det gjer det, ser det ut som eit felt uret ligg *i* i staden for
	// noko uret syner: peikaren sluttar, og flata held fram.
	if kakeR > visarTime {
		t.Errorf("kakestykket (%.1f) er lenger enn timevisaren (%.1f)", kakeR, visarTime)
	}
	if kakeR > visarMinutt {
		t.Errorf("kakestykket (%.1f) er lenger enn minuttvisaren (%.1f)", kakeR, visarMinutt)
	}
	if kakeR > spor {
		t.Fatalf("kakestykket (%.1f) er større enn sporet indeksane ligg paa (%.1f)", kakeR, spor)
	}

	start := time.Date(2026, 8, 27, 18, 0, 0, 0, time.UTC)
	slutt := start.Add(75 * time.Minute)
	m := NewMark("prove", start, slutt, 4, 18)

	var timen string
	for _, k := range m.Slices {
		if k.Class == "kake-timen" {
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
func TestKakestykketHengITimevisaren(t *testing.T) {
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
		m := NewMark("prova", start, slutt, 4, 18)

		if got := durationFraction(start, slutt); math.Abs(got-p.grader) > 0.01 {
			t.Errorf("%s: stykket vart %.1f grader, venta %.1f", p.namn, got, p.grader)
		}

		// Sluttvisaren stend der stykket sluttar — paa timeskalaen, som
		// er den same skalaen stykket sjølv maaler seg paa. Han hekk i
		// minuttvisaren fyrr, og daa var dei tvo tali i figuren rekna
		// paa kvar sin maate.
		venta := hourAngle(start) + p.grader
		if math.Abs(m.EndAngle-venta) > 0.01 {
			t.Errorf("%s: sluttvisaren stend paa %.1f, venta %.1f", p.namn, m.EndAngle, venta)
		}

		// Og stykket sjølv byrjar i timevisaren: fyrste linestykket i
		// pathen gjeng fraa midten ut til startvinkelen.
		var timen string
		for _, k := range m.Slices {
			if k.Class == "kake-timen" {
				timen = string(k.D)
			}
		}
		if timen == "" {
			t.Fatalf("%s: fann ikkje kakestykket", p.namn)
		}
		// Sluttvisaren skal peika der timevisaren hadde stade daa timen
		// var ute. Det er heile poenget med aa flytta honom hit.
		if vil := hourAngle(slutt); math.Abs(math.Mod(m.EndAngle, 360)-math.Mod(vil, 360)) > 0.01 {
			t.Errorf("%s: sluttvisaren stend paa %.1f, timevisaren ved slutt paa %.1f",
				p.namn, math.Mod(m.EndAngle, 360), math.Mod(vil, 360))
		}

		hit := fyrstePunkt.FindStringSubmatch(timen)
		if hit == nil {
			t.Fatalf("%s: skjøna ikkje pathen %q", p.namn, timen)
		}
		x, _ := strconv.ParseFloat(hit[1], 64)
		y, _ := strconv.ParseFloat(hit[2], 64)
		vx, vy := point(hourAngle(start), kakeR, skiveX, skiveY)
		if math.Abs(x-vx) > 0.05 || math.Abs(y-vy) > 0.05 {
			t.Errorf("%s: stykket byrjar i %.2f,%.2f, men timevisaren stend i %.2f,%.2f",
				p.namn, x, y, vx, vy)
		}
	}
}
