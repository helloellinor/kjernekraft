package handlers

import (
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"
)

// Merket er eit ur, og eit ur er teikna og ikkje sett saman av kassar.
//
// Difor er heile merket éin SVG med eit fast viewBox: skal han verta
// større eller mindre, er det éin skala som endrar seg, og strekar,
// skrift, rom og hyrne fylgjer med i det same forholdet. Det er dette
// som gjer at han kjennest som den same tingen som kjem nærare, og ikkje
// som ei ny teikning ved kvar storleik.
//
// Reknestykki stend her og ikkje i malen. Ein mal skal syna kva som er
// i figuren; kvar kvart punkt ligg, er ein annan slags kunnskap.
const (
	kroppB, kroppH = 46.0, 46.0
	kroppR         = 3.5 // rektangel med so vidt broti hyrna
	faneB, faneH   = 28.0, 10.4
	faneR          = 2.5
	faneLoft       = 9.0
	plettR         = 8.6
	merkeKant      = 1.0
	lappB          = 23.0 // pletten som lapp, naar timen er full

	kroppX = merkeKant
	kroppY = faneLoft + merkeKant
	faneX  = kroppX + (kroppB-faneB)/2
	faneY  = merkeKant

	plettX = kroppX + kroppB // midten aat pletten ligg i hyrna
	plettY = kroppY + kroppH

	// Maalaren heng 9,6 einingar ut til høgre og ingen ting ut til
	// vinstre. Er kassen teikna kring det, stend ikkje *uret* midt i
	// kassen, og sju merke i sju spaltor kjem ikkje under kvarandre same
	// kva ein gjer i CSS. Difor byrjar viewBox til vinstre for kroppen
	// med det same overhenget som ligg til høgre: kassen er symmetrisk
	// kring kroppen, og spaltone stend paa lina av seg sjølve.
	merkeOverheng = plettR + merkeKant       // 9,6
	MerkeVenstre  = kroppX - merkeOverheng   // −8,6
	MerkeBreidd   = kroppB + 2*merkeOverheng // 65,2
	MerkeHogd     = kroppY + kroppH + plettR + merkeKant

	skiveX = kroppX + kroppB/2
	skiveY = kroppY + kroppH/2
	spor   = 15.4 // sporet indeksane ligg paa
	kakeR  = 11.6

	ruteB, ruteH = 23.0, 9.4
	ruteX        = skiveX - ruteB/2
	ruteY        = skiveY + 4.2
	ruteR        = 1.6

	visarTime   = 7.2
	visarMinutt = 10.4
)

// Strek er ein indeks paa skiva.
type Strek struct {
	X1, Y1, X2, Y2 float64
	Kvart          bool
}

// Skrift er eit tal paa skiva.
type Skrift struct {
	X, Y  float64
	Tekst string
}

// Kake er eit kakestykke, med klassa som seier kva det tyder.
type Kake struct {
	Klasse string
	D      string
}

// Merke er alt ein mal treng for aa teikna eit timemerke.
type Merke struct {
	// Heile kassen, ferdig skriven: «−8.6 0 65.2 65.6».
	ViewBox string
	// Filtri bur inne i kvar figur og ikkje i eit sams sett paa sida.
	// Grunnen er tema: `flood-color: var(--kant)` vert løyst der
	// *filteret* stend, ikkje der det vert nytta. Eit sams sett hadde
	// difor gjeve alle merki fargane til rot-temaet, og ein ljos bolk
	// paa ei myrk sida — som verkstaden set upp — hadde fenge feil
	// lina. Ident gjer id-ane eintydige.
	Ident        string
	Breidd, Hogd float64
	Silhuett     template.HTMLAttr // umrisset: kant, ring og flate deler han
	DagX, DagY   float64
	Indeksar     []Strek
	Tal          []Skrift
	Kaker        []Kake
	SkiveX       float64
	SkiveY       float64
	VisarTime    float64
	VisarMinutt  float64
	TimeVinkel   float64
	MinuttVinkel float64
	SluttVinkel  float64
	Rute         template.HTMLAttr
	RuteTekstY   float64
	Full         bool
	Maalar       template.HTMLAttr // kakestykket i pletten
	LappX, LappY float64
}

// rutePath er eit rektangel med broti hyrne.
func rutePath(x, y, b, h, r float64) string {
	if r < 0.4 {
		r = 0.4
	}
	return fmt.Sprintf("M%.2f,%.2f H%.2f A%.2f,%.2f 0 0 1 %.2f,%.2f V%.2f "+
		"A%.2f,%.2f 0 0 1 %.2f,%.2f H%.2f A%.2f,%.2f 0 0 1 %.2f,%.2f V%.2f "+
		"A%.2f,%.2f 0 0 1 %.2f,%.2f Z",
		x+r, y, x+b-r, r, r, x+b, y+r, y+h-r,
		r, r, x+b-r, y+h, x+r, r, r, x, y+h-r, y+r,
		r, r, x+r, y)
}

func sirkelPath(cx, cy, r float64) string {
	return fmt.Sprintf("M%.2f,%.2f a%.2f,%.2f 0 1,0 %.2f,0 a%.2f,%.2f 0 1,0 %.2f,0 Z",
		cx-r, cy, r, r, 2*r, r, r, -2*r)
}

// silhuett er dei tri formene som éin veg. Umrisset og fordjupingi vert
// laga av honom med filter, ikkje ved aa teikna formene ein gong til
// litt større: utvidar ein kvar form for seg, vandrar dei innhole hyrno
// innyver i staden for utyver, og daa slepp lina upp nettupp i skøyten.
func silhuett(full bool) string {
	delar := []string{
		rutePath(faneX+merkeKant, faneY+merkeKant, faneB-2*merkeKant,
			faneH+7-2*merkeKant, faneR-merkeKant),
		rutePath(kroppX+merkeKant, kroppY+merkeKant, kroppB-2*merkeKant,
			kroppH-2*merkeKant, kroppR-merkeKant),
	}
	if full {
		delar = append(delar, rutePath(plettX+plettR-lappB+merkeKant, plettY-plettR+merkeKant,
			lappB-2*merkeKant, 2*plettR-2*merkeKant, plettR-merkeKant))
	} else {
		delar = append(delar, sirkelPath(plettX, plettY, plettR-merkeKant))
	}
	return strings.Join(delar, " ")
}

// paaSporet gjev punktet der ein vinkel møter sporet. Sporet er ein
// superellipse og ikkje ein sirkel — det er nettupp difor skiva ser
// firkanta ut, og det er ogso det som gjer avstanden fraa indeks til
// kant den same heile vegen rundt.
func paaSporet(grader, vidd float64) (float64, float64) {
	t := grader * math.Pi / 180
	dx, dy := math.Sin(t), -math.Cos(t)
	n := math.Pow(math.Pow(math.Abs(dx), 6)+math.Pow(math.Abs(dy), 6), 1.0/6.0)
	return skiveX + dx*vidd/n, skiveY + dy*vidd/n
}

func punkt(grader, r, cx, cy float64) (float64, float64) {
	t := grader * math.Pi / 180
	return cx + math.Sin(t)*r, cy - math.Cos(t)*r
}

// kakePath er eit kakestykke fraa ein vinkel til ein annan.
func kakePath(fraa, til, r, cx, cy float64) string {
	sveip := math.Mod(til-fraa, 360)
	if sveip < 0 {
		sveip += 360
	}
	if sveip < 0.4 {
		return ""
	}
	if sveip > 359.4 {
		return sirkelPath(cx, cy, r)
	}
	x1, y1 := punkt(fraa, r, cx, cy)
	x2, y2 := punkt(til, r, cx, cy)
	stor := 0
	if sveip > 180 {
		stor = 1
	}
	return fmt.Sprintf("M%.2f,%.2f L%.2f,%.2f A%.2f,%.2f 0 %d,1 %.2f,%.2f Z",
		cx, cy, x1, y1, r, r, stor, x2, y2)
}

// urvinkel er klokkeslettet paa timevisaren si skala: ei runda er tolv
// timar. Kakestykki ligg paa den same skalaen, so eit stykke er kor stor
// bit av dagen timen tek. Minuttskalaen hadde vore meir synleg — ein
// time paa 45 minutt hadde vorte trikvart av skiva — men ein time som
// varer lenger enn ein time hadde gjenge heile vegen rundt og ikkje
// kunna segja meir.
func urvinkel(t time.Time) float64 {
	return float64(t.Hour()%12)*30 + float64(t.Minute())*0.5
}

// NyttMerke reknar ut heile figuren for éin time.
func NyttMerke(ident string, start, slutt time.Time, teke, plassar int, naa time.Time, erIDag bool) Merke {
	full := plassar > 0 && teke >= plassar
	att := plassar - teke
	if att < 0 {
		att = 0
	}
	delen := 0.0
	if plassar > 0 {
		delen = float64(att) / float64(plassar)
	}

	m := Merke{
		Ident:   ident,
		ViewBox: fmt.Sprintf("%.2f 0 %.2f %.2f", MerkeVenstre, MerkeBreidd, MerkeHogd),
		Breidd:  MerkeBreidd, Hogd: MerkeHogd,
		Silhuett:     template.HTMLAttr(silhuett(full)),
		DagX:         faneX + faneB/2,
		DagY:         faneY + 8.4,
		SkiveX:       skiveX,
		SkiveY:       skiveY,
		VisarTime:    skiveY - visarTime,
		VisarMinutt:  skiveY - visarMinutt,
		TimeVinkel:   urvinkel(start),
		MinuttVinkel: float64(start.Minute()) * 6,
		SluttVinkel:  urvinkel(slutt),
		Rute:         template.HTMLAttr(rutePath(ruteX, ruteY, ruteB, ruteH, ruteR)),
		RuteTekstY:   ruteY + 6.9,
		Full:         full,
		LappX:        plettX + plettR - lappB/2,
		LappY:        plettY,
	}

	// Indeksane. Sekstalet ligg under vindauga og skal ikkje teiknast.
	for h := 0; h < 12; h++ {
		if h == 6 {
			continue
		}
		g := float64(h) * 30
		ux, uy := paaSporet(g, spor)
		kvart := h%3 == 0
		if kvart {
			tx, ty := paaSporet(g, spor-1.0)
			m.Tal = append(m.Tal, Skrift{X: tx, Y: ty + 1.8,
				Tekst: map[int]string{0: "12", 3: "3", 9: "9"}[h]})
			continue
		}
		lengd := 1.7
		ix, iy := paaSporet(g, spor-lengd)
		m.Indeksar = append(m.Indeksar, Strek{X1: ux, Y1: uy, X2: ix, Y2: iy, Kvart: false})
	}

	// Kakestykki: det bleike er ventetidi, det farga er timen sjølv.
	// Ventetidi gjeld berre timar i dag som ikkje hev byrja — elles er
	// «fram til» ikkje ein vinkel, det er dagar.
	if erIDag && naa.Before(start) {
		if d := kakePath(urvinkel(naa), urvinkel(start), kakeR, skiveX, skiveY); d != "" {
			m.Kaker = append(m.Kaker, Kake{Klasse: "kake-fyre", D: d})
		}
	}
	if d := kakePath(urvinkel(start), urvinkel(slutt), kakeR, skiveX, skiveY); d != "" {
		m.Kaker = append(m.Kaker, Kake{Klasse: "kake-timen", D: d})
	}

	if !full {
		m.Maalar = template.HTMLAttr(kakePath(0, 359.9*delen, plettR-merkeKant-1.1, plettX, plettY))
	}
	return m
}

// DaudSilhuett er ruta utan time: same maal som ei levande, so rada
// stend paa lina, men berre eit hint av ei lina.
func DaudSilhuett() template.HTMLAttr {
	return template.HTMLAttr(rutePath(kroppX, kroppY, kroppB, kroppH, kroppR))
}
