package handlers

import (
	"fmt"
	"html/template"
	"math"
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
	// Merket er eit ur i ein høg kasse. Framsida er delt i tri: dagnamnet
	// langs toppen som ei fane, skiva midt i, og ei botnstripe med
	// klokkeslettet nede til vinstre og plassmerket nede til høgre.
	kroppB, kroppH = 56.0, 72.0
	kroppR         = 4.0 // rektangel med so vidt broti hyrna
	faneB, faneH   = 44.0, 13.0
	faneR          = 3.0
	faneLoft       = 11.0
	merkeKant      = 1.0

	kroppX = merkeKant
	kroppY = faneLoft + merkeKant
	faneX  = kroppX + (kroppB-faneB)/2
	faneY  = merkeKant

	// Botnstripa. Alt i henne ligg *inne* i kroppen no. Plassmerket laag
	// med midten sin paa hyrna fyrr og hekk ein heil radius utanfor, og
	// daa laut kassen vera so mykje breidare enn kroppen paa baae sidor —
	// dei einingane merket ikkje fekk bruka. Ingen ting heng utanfor
	// lenger, og kroppen fyller difor 96,6 % av kassen mot 70,6 % fyrr.
	botnRom = 3.5

	ruteB, ruteH = 27.0, 12.5
	ruteX        = kroppX + botnRom
	ruteY        = kroppY + kroppH - ruteH - botnRom
	ruteR        = 1.8

	plettR = 9.0
	plettX = kroppX + kroppB - botnRom - plettR
	plettY = kroppY + kroppH - botnRom - plettR
	lappB  = 24.0 // plassmerket som lapp, naar timen er full

	MerkeVenstre = 0.0
	MerkeBreidd  = kroppB + 2*merkeKant
	MerkeHogd    = kroppY + kroppH + merkeKant

	// Skiva stend midt i rommet yver botnstripa, og visarane stend midt
	// i skiva. Ho laag halvt nedi vindauga fyrr — klokkeslettet var
	// teikna tvert yver henne — og daa las korkje uret eller talet seg.
	skiveX = kroppX + kroppB/2
	skiveY = 38.0
	spor   = 21.0 // sporet indeksane ligg paa
	kakeR  = 17.5

	visarTime   = 12.0
	visarMinutt = 17.5
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
	Fane         template.HTMLAttr // dagfana, teikna for seg attum kroppen
	Plett        template.HTMLAttr // plassmerket, teikna for seg uppaa kroppen
	Rute         template.HTMLAttr
	RuteTekstX   float64
	RuteTekstY   float64
	Full         bool
	Att          int // plassar att
	Plassar      int // av so mange
	PlettX       float64
	PlettY       float64
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
func silhuett() string {
	return rutePath(kroppX+merkeKant, kroppY+merkeKant, kroppB-2*merkeKant,
		kroppH-2*merkeKant, kroppR-merkeKant)
}

// fanePath er dagfana. Ho vart teikna saman med kroppen som éi form
// fyrr, og daa fanst det ingi fana — det fanst eitt umriss med ein
// kul paa. Ei fana er ein *eigen* ting som ligg attum arket sitt: ho
// gjeng ned under yverkanten aat kroppen, kroppen dekkjer nedkanten
// hennar, og det som stikk upp er fana. Same grepet som fanone i
// administrasjonen, berre teikna.
func fanePath() string {
	return rutePath(faneX+merkeKant, faneY+merkeKant, faneB-2*merkeKant,
		faneH+6-2*merkeKant, faneR-merkeKant)
}

// plettPath er plassmerket. Ein sirkel naar det er plassar att, ei lapp
// naar timen er full og ordet treng breidd.
func plettPath(full bool) string {
	if full {
		return rutePath(plettX+plettR-lappB+merkeKant, plettY-plettR+merkeKant,
			lappB-2*merkeKant, 2*plettR-2*merkeKant, plettR-merkeKant)
	}
	return sirkelPath(plettX, plettY, plettR-merkeKant)
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

// minuttvinkel er kvar minuttvisaren stend: ei runda er ein time.
func minuttvinkel(t time.Time) float64 {
	return float64(t.Minute()) * 6
}

// varigheit er kor stort stykke timen tek av skiva, paa timevisaren si
// skala: ei runda er tolv timar, so ein time er tretti grader.
//
// Minuttskalaen hadde vore meir synleg — ein time paa 45 minutt hadde
// vorte trikvart av skiva — men ein time som varer lenger enn ein time
// hadde gjenge heile vegen rundt og ikkje kunna segja meir. Difor stend
// stykket i minuttvisaren, men maaler seg paa timeskalaen.
func varigheit(start, slutt time.Time) float64 {
	return slutt.Sub(start).Hours() * 30
}

// NyttMerke reknar ut heile figuren for éin time.
func NyttMerke(ident string, start, slutt time.Time, teke, plassar int) Merke {
	full := plassar > 0 && teke >= plassar
	att := plassar - teke
	if att < 0 {
		att = 0
	}
	m := Merke{
		Ident:   ident,
		ViewBox: fmt.Sprintf("%.2f 0 %.2f %.2f", MerkeVenstre, MerkeBreidd, MerkeHogd),
		Breidd:  MerkeBreidd, Hogd: MerkeHogd,
		Silhuett:     template.HTMLAttr(silhuett()),
		Fane:         template.HTMLAttr(fanePath()),
		Plett:        template.HTMLAttr(plettPath(full)),
		DagX:         faneX + faneB/2,
		DagY:         faneY + 9.4,
		SkiveX:       skiveX,
		SkiveY:       skiveY,
		VisarTime:    skiveY - visarTime,
		VisarMinutt:  skiveY - visarMinutt,
		TimeVinkel:   urvinkel(start),
		MinuttVinkel: minuttvinkel(start),
		SluttVinkel:  minuttvinkel(start) + varigheit(start, slutt),
		Rute:         template.HTMLAttr(rutePath(ruteX, ruteY, ruteB, ruteH, ruteR)),
		RuteTekstX:   ruteX + ruteB/2,
		RuteTekstY:   ruteY + 8.6,
		Full:         full,
		PlettX:       plettX,
		PlettY:       plettY,
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
			tx, ty := paaSporet(g, spor-1.4)
			m.Tal = append(m.Tal, Skrift{X: tx, Y: ty + 2.4,
				Tekst: map[int]string{0: "12", 3: "3", 9: "9"}[h]})
			continue
		}
		lengd := 2.6
		ix, iy := paaSporet(g, spor-lengd)
		m.Indeksar = append(m.Indeksar, Strek{X1: ux, Y1: uy, X2: ix, Y2: iy, Kvart: false})
	}

	// Kakestykket: timen sjølv, og ikkje noko meir.
	//
	// Det låg eit bleikt stykke her med — tidi fram til timen byrja. Det
	// var ikkje til aa lesa: det saag ut som ei nott-og-dag-teikning, og
	// ingen visste kva det gjorde. Skiva svarar paa «naar gjeng han og
	// kor lenge»; «kor lenge til» er eit anna spursmaal, og det stend i
	// klartekst i dokka.
	// Timen sjølv veks ut or *minuttvisaren* og ikkje or timevisaren.
	//
	// Han låg på timevisaren fyrr, og då voks stykket ut or den korte,
	// tjukke visaren — den eine som ikkje rører seg medan timen gjeng.
	// Det er minuttvisaren som gjeng medan du ligg på matta, og difor er
	// det han stykket skal henge i. Sluttvisaren stend i den andre kanten
	// av stykket, so dei tri tingi — visar, stykke, visar — les seg som
	// éin ting og ikkje som tri.
	fraa := minuttvinkel(start)
	if d := kakePath(fraa, fraa+varigheit(start, slutt), kakeR, skiveX, skiveY); d != "" {
		m.Kaker = append(m.Kaker, Kake{Klasse: "kake-timen", D: d})
	}

	m.Att, m.Plassar = att, plassar
	return m
}

// DaudSilhuett er ruta utan time: same maal som ei levande, so rada
// stend paa lina, men berre eit hint av ei lina.
func DaudSilhuett() template.HTMLAttr {
	return template.HTMLAttr(rutePath(kroppX, kroppY, kroppB, kroppH, kroppR))
}
