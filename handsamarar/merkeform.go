package handsamarar

import (
	"fmt"
	"html/template"
	"math"
	"time"
)

// The mark is a clock, and a clock is drawn — not assembled from boxes.
//
// So it is one SVG with a fixed viewBox: resizing changes a single
// scale, and strokes, type, spacing and corners follow in the same
// proportion. That is what makes it feel like the same object coming
// closer rather than a new drawing at every size.
//
// The arithmetic lives here, not in the template. A template shows what
// is in the figure; where each point lies is another kind of knowledge.
const (
	// A clock in a tall case. The face is in three: the day name across
	// the top as a tab, the dial in the middle, and a bottom strip with
	// the time at the left and the seat mark at the right.
	//
	// The aspect is not a choice — it is one step on the e-ladder: the
	// case height is the width times e^(1/4). So kroppH is computed, not
	// picked. To resize the mark you change the width; the height follows
	// the step, and the strip, the seat mark and the dial move with it.
	stegFjerdedel = 1.2840254 // e^(1/4)

	kroppB = 56.0
	kroppH = MarkWidth*stegFjerdedel - kroppY - merkeKant
	kroppR = 4.0 // rektangel med so vidt broti hyrna
	// The tab spans the top so the day name stands whole. At 44 of 56 it
	// had to be cut to two letters, and "TY" is not a word, it is a code
	// you have to learn. At 52 of 56 "TYSDAG" fits, and two units of
	// shoulder each side still show the tab sitting behind the case.
	faneB, faneH = 52.0, 13.0
	faneR        = 3.0
	faneLoft     = 11.0
	merkeKant    = 1.0

	kroppX = merkeKant
	kroppY = faneLoft + merkeKant
	faneX  = kroppX + (kroppB-faneB)/2
	faneY  = merkeKant

	// The bottom strip. Everything in it sits inside the case; nothing
	// hangs outside, so the body fills 96.6 % of the box rather than
	// 70.6 %. At 5 the window bottom clears the rebate by 2 — a visible
	// floor between screen and edge.
	botnRom = 5.0

	// The window is centred in the strip. The seat mark now sits on the
	// corner, on top of the edge, so the strip belongs to the window
	// alone.
	ruteB, ruteH = 32.0, 14.5
	ruteX        = kroppX + (kroppB-ruteB)/2
	ruteY        = kroppY + kroppH - ruteH - botnRom
	ruteR        = 1.8

	// The seat mark is no longer drawn in the SVG. It was a shape that
	// *resembled* a button and never quite matched, because the button
	// carries border, radius, depth and a one-pixel press that SVG has to
	// imitate separately. It is an HTML element on top of the figure now,
	// taking the button's own style (.btn, …, .plettmerke).
	//
	// What stays here is the position: the bottom-right corner, in
	// percent of the box, so the mark can be centred on it.
	hyrneX = (kroppX + kroppB) / MarkWidth * 100
	hyrneY = (kroppY + kroppH) / MarkHeight * 100

	MarkLeft   = 0.0
	MarkWidth  = kroppB + 2*merkeKant
	MarkHeight = kroppY + kroppH + merkeKant

	// The margin from the case edge in to the dial face — the rebate left
	// around the sunken plate. The index track starts 4 in from the case,
	// so the plate's lip does not meet the indices.
	skiveMarg = 3.0

	skiveX = kroppX + kroppB/2
	// The dial sits at the centre of the case, in width and in height.
	// Angles distributed around a point that is not the centre cannot come
	// out even: twelve crowded the top while three and nine stood above
	// the midline and the whole bottom was empty.
	//
	// The window covers the lowest part of the dial. It is an opaque
	// layer on top, so what falls under it stays under it.
	skiveY = kroppY + kroppH/2

	// Skiva set høgdi på heile merket.
	//
	// Ho er ikkje kakestykket — ho er *sporet* indeksane ligg på, og
	// det er breidare. Kassen lyt vera høg nok til å halda henne heil
	// med botnstripa under, so kroppH kan ikkje kortast utan at skiva
	// follows. merkeform_test holds this together: a dial that does not
	// fit is a dial drawn down into the clock text.
	//
	// The track the indices sit on — a rounded square filling the face,
	// not a small square inside a big one. Symmetric, because the dial
	// sits at the centre of the case.
	//
	// And 3.5 *inside the dial face*, not out to its edge: on the seam
	// between rebate and face, "12" drowned in the shadow from the top
	// edge and "9" and "3" stood half on each surface. Print on a watch
	// sits on the dial, with margin to the edge.
	sporB   = 21.5 // halvbreidd
	sporOpp = 23.5 // upp frå midten
	sporNed = 23.5 // ned frå midten
	spor    = sporOpp
	// The slice reaches exactly as far as the hour hand it stands for —
	// not a notch short, and never longer.
	kakeR = visarTime

	// The hands, as *lengths* from the dial centre. NewMark turns them
	// into y endpoints with skiveY - lengd.
	//
	// Read once as endpoints, they were "corrected" the wrong way round:
	// the hour hand became the long one, which is a clock that lies about
	// the time — and the tests computed the same expression the same
	// wrong way and passed it. Hence the unit written here, and hence
	// TestVisaraneErRettVegen checking the finished Mark fields rather
	// than the constants.
	visarTime   = 12.0 // timevisaren, den korte
	visarMinutt = 17.5 // minuttvisaren, den lange
)

// Tick is one index on the dial.
type Tick struct {
	X1, Y1, X2, Y2 float64
	Quarter        bool
}

// Numeral is one number on the dial.
type Numeral struct {
	X, Y float64
	Text string
}

// Indeksane på skiva er dei same på kvart merke, so dei vert rekna ein
// gong og teikna ein gong. Dei stod i kvart einskilt merke fyrr: åtte
// strekar og tri tal, gonga med talet på timar i vika — 418 element som
// alle var det same.
//
// Sekstalet ligg under vindauga og skal ikkje teiknast.
var skiveIndeks, skiveTal = byggIndeks()

func byggIndeks() ([]Tick, []Numeral) {
	var t []Tick
	var n []Numeral
	for h := 0; h < 12; h++ {
		if h == 6 {
			continue
		}
		g := float64(h) * 30
		ux, uy := onDialTrack(g, 0)
		if h%3 == 0 {
			tx, ty := onDialTrack(g, 1.4)
			n = append(n, Numeral{X: tx, Y: ty + 2.4,
				Text: map[int]string{0: "12", 3: "3", 9: "9"}[h]})
			continue
		}
		ix, iy := onDialTrack(g, 2.6)
		t = append(t, Tick{X1: ux, Y1: uy, X2: ix, Y2: iy, Quarter: false})
	}
	return t, n
}

// Kroppen og dagfana er dei same på kvart merke — dei stend i
// `merke_defs` og vert henta med <use>.
func Silhuett() template.HTMLAttr { return template.HTMLAttr(silhouette()) }
func Dagfane() template.HTMLAttr  { return template.HTMLAttr(dayTabPath()) }

// Glaset sitt lys ligg på same staden i kvart merke — tali under er
// konstantar. Difor kann gradienten teiknast ein gong.
func GlasMidt() (float64, float64, float64) {
	return kroppX + 0.34*kroppB, kroppY + 0.22*kroppH, 0.62 * kroppB
}

// SkiveIndeks og SkiveTal gjev indeksane til malen, som teiknar dei ein
// gong i `merke_defs`.
func SkiveIndeks() []Tick { return skiveIndeks }
func SkiveTal() []Numeral { return skiveTal }

// Slice is a pie slice, with the class that says what it means.
type Slice struct {
	Class string
	D     string
}

// Mark is everything a template needs to draw a class mark.
type Mark struct {
	// Heile kassen, ferdig skriven: «−8.6 0 65.2 65.6».
	ViewBox string
	// Filtri bur inne i kvar figur og ikkje i eit sams sett på sida.
	// Grunnen er tema: `flood-color: var(--kant)` vert løyst der
	// *filteret* stend, ikkje der det vert nytta. Eit sams sett hadde
	// difor gjeve alle merki fargane til rot-temaet, og ein ljos bolk
	// på ei myrk sida — som verkstaden set upp — hadde fenge feil
	// lina. Ident gjer id-ane eintydige.
	Ident         string
	Width, Height float64
	Silhouette    template.HTMLAttr // umrisset: kant, ring og flate deler han
	DayX, DayY    float64
	// The glass. The sun is at 34 % 22 % of the body, like everything round
	// in the house (--sol). The numbers are reckoned in the mark's *own*
	// space, not as a percentage of the bounding box: a box taller than it is
	// wide stretches an objectBoundingBox gradient into an oval, and the
	// reflection becomes a smear rather than a point of light.
	GlasX, GlasY float64
	GlasR        float64
	Ticks        []Tick
	Numerals     []Numeral
	Slices       []Slice
	DialX        float64
	DialY        float64
	HourHand     float64
	MinuteHand   float64
	HourAngle    float64
	MinuteAngle  float64
	EndAngle     float64
	DayTab       template.HTMLAttr // dagfana, teikna for seg attum kroppen
	// Skiveflata: den senka plata som fyller andlitet, med jamn marg
	// mot kanten av kroppen. Vindauga er stansa ned i *henne* — kasse,
	// senka skiva, og eit djupare datovindauga, som på eit ur.
	Skive     template.HTMLAttr
	Box       template.HTMLAttr
	BoxTextX  float64
	BoxTextY  float64
	Full      bool
	Remaining int // plassar att
	Capacity  int // av so mange
	// Hyrna merket skal sentrerast på, i prosent av kassen.
	HyrneX float64
	HyrneY float64
}

// boxPath er eit rektangel med broti hyrne.
func boxPath(x, y, b, h, r float64) string {
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

func circlePath(cx, cy, r float64) string {
	return fmt.Sprintf("M%.2f,%.2f a%.2f,%.2f 0 1,0 %.2f,0 a%.2f,%.2f 0 1,0 %.2f,0 Z",
		cx-r, cy, r, r, 2*r, r, r, -2*r)
}

// silhouette is the three shapes as one path. The outline and the recess
// are made from it with filters, not by drawing the shapes again slightly
// larger: expand each shape separately and the concave corners travel
// inward instead of outward, so the line breaks open at the joins.
func silhouette() string {
	return boxPath(kroppX+merkeKant, kroppY+merkeKant, kroppB-2*merkeKant,
		kroppH-2*merkeKant, kroppR-merkeKant)
}

// dayTabPath is the day tab. It used to be drawn together with the body as
// one shape, and then there was no tab — there was one outline with a bump
// on it. A tab is a *separate* thing lying behind its sheet: it goes down
// under the top edge of the body, the body covers its lower edge, and what
// sticks up is the tab.
func dayTabPath() string {
	return boxPath(faneX+merkeKant, faneY+merkeKant, faneB-2*merkeKant,
		faneH+6-2*merkeKant, faneR-merkeKant)
}

// onDialTrack gives a point on the track, or `inn` units inside it.
//
// The track is a rounded square and not a circle: the piece is a square
// clock, so the indices should follow the case. A circle leaves the
// corners empty.
//
// It was a superellipse (|dx|⁶+|dy|⁶). That is *nearly* square — diagonally
// it reaches 1.26 of what it reaches straight out, where a square reaches
// 1.41 — and turning the exponent moved the indices by a couple of per
// cent.
//
// `inn` is a distance *along the ray*, not a smaller track. Scaling the
// track instead would make the indices different lengths around the dial,
// because the track is oblong.
func onDialTrack(grader, inn float64) (float64, float64) {
	t := grader * math.Pi / 180
	dx, dy := math.Sin(t), -math.Cos(t)

	// How far the ray travels before it meets one of the sides.
	lengd := math.Inf(1)
	if dx != 0 {
		lengd = math.Min(lengd, sporB/math.Abs(dx))
	}
	if dy != 0 {
		hogd := sporOpp
		if dy > 0 {
			hogd = sporNed
		}
		lengd = math.Min(lengd, hogd/math.Abs(dy))
	}

	// Er treffet inne i eit hyrne, er det buen han møter og ikkje sida.
	hogd := sporOpp
	if dy > 0 {
		hogd = sporNed
	}
	r := math.Min(sporB, hogd) * (kroppR / (kroppB / 2))
	naerB, naerH := sporB-r, hogd-r
	if math.Abs(dx*lengd) > naerB && math.Abs(dy*lengd) > naerH {
		sx := math.Copysign(naerB, dx)
		sy := math.Copysign(naerH, dy)
		b := dx*sx + dy*sy
		c := sx*sx + sy*sy - r*r
		if d := b*b - c; d >= 0 {
			lengd = b + math.Sqrt(d)
		}
	}
	lengd -= inn
	return skiveX + dx*lengd, skiveY + dy*lengd
}

func point(grader, r, cx, cy float64) (float64, float64) {
	t := grader * math.Pi / 180
	return cx + math.Sin(t)*r, cy - math.Cos(t)*r
}

// slicePath er eit kakestykke frå ein vinkel til ein annan.
func slicePath(frå, til, r, cx, cy float64) string {
	sveip := math.Mod(til-frå, 360)
	if sveip < 0 {
		sveip += 360
	}
	if sveip < 0.4 {
		return ""
	}
	if sveip > 359.4 {
		return circlePath(cx, cy, r)
	}
	x1, y1 := point(frå, r, cx, cy)
	x2, y2 := point(til, r, cx, cy)
	stor := 0
	if sveip > 180 {
		stor = 1
	}
	return fmt.Sprintf("M%.2f,%.2f L%.2f,%.2f A%.2f,%.2f 0 %d,1 %.2f,%.2f Z",
		cx, cy, x1, y1, r, r, stor, x2, y2)
}

// hourAngle er klokkeslettet på timevisaren si skala: ei runda er tolv
// timar. Kakestykki ligg på den same skalaen, so eit stykke er kor stor
// bit av dagen timen tek. Minuttskalaen hadde vore meir synleg — ein
// time på 45 minutt hadde vorte trikvart av skiva — men ein time som
// varer lenger enn ein time hadde gjenge heile vegen rundt og ikkje
// kunna segja meir.
func hourAngle(t time.Time) float64 {
	return float64(t.Hour()%12)*30 + float64(t.Minute())*0.5
}

// minuteAngle er kvar minuttvisaren stend: ei runda er ein time.
func minuteAngle(t time.Time) float64 {
	return float64(t.Minute()) * 6
}

// durationFraction er kor stort stykke timen tek av skiva, på timevisaren si
// skala: ei runda er tolv timar, so ein time er tretti grader.
//
// Minuttskalaen hadde vore meir synleg — ein time på 45 minutt hadde
// vorte trikvart av skiva — men ein time som varer lenger enn ein time
// hadde gjenge heile vegen rundt og ikkje kunna segja meir. Difor stend
// stykket i minuttvisaren, men maaler seg på timeskalaen.
func durationFraction(start, slutt time.Time) float64 {
	return slutt.Sub(start).Hours() * 30
}

// NewMark reknar ut heile figuren for éin time.
func NewMark(ident string, start, slutt time.Time, teke, plassar int) Mark {
	full := plassar > 0 && teke >= plassar
	att := plassar - teke
	if att < 0 {
		att = 0
	}
	m := Mark{
		Ident:   ident,
		ViewBox: fmt.Sprintf("%.2f 0 %.2f %.2f", MarkLeft, MarkWidth, MarkHeight),
		Width:   MarkWidth, Height: MarkHeight,
		Silhouette:  template.HTMLAttr(silhouette()),
		DayTab:      template.HTMLAttr(dayTabPath()),
		GlasX:       kroppX + 0.34*kroppB,
		GlasY:       kroppY + 0.22*kroppH,
		GlasR:       0.62 * kroppB,
		DayX:        faneX + faneB/2,
		DayY:        faneY + 9.4,
		DialX:       skiveX,
		DialY:       skiveY,
		HourHand:    skiveY - visarTime,
		MinuteHand:  skiveY - visarMinutt,
		HourAngle:   hourAngle(start),
		MinuteAngle: minuteAngle(start),
		// The end hand hangs off the *hour* hand, not the minute hand.
		//
		// It hung off the minute hand while the slice measured itself on the hour
		// scale — two scales in one figure. A sixty-minute class then swept thirty
		// degrees out from a pointer that itself moves six degrees per minute, and
		// the picture answered neither "what time is it" nor "how long is the
		// class".
		//
		// On the hour scale both are true at once: the hour hand stands where the
		// class begins, the end hand where it is over, and the slice between them
		// *is* the class.
		EndAngle: hourAngle(start) + durationFraction(start, slutt),
		Skive:    template.HTMLAttr(boxPath(kroppX+skiveMarg, kroppY+skiveMarg, kroppB-2*skiveMarg, kroppH-2*skiveMarg, kroppR-1)),
		Box:      template.HTMLAttr(boxPath(ruteX, ruteY, ruteB, ruteH, ruteR)),
		BoxTextX: ruteX + ruteB/2,
		BoxTextY: ruteY + ruteH/2,
		Full:     full,
		HyrneX:   hyrneX,
		HyrneY:   hyrneY,
	}

	// Indeksane er dei same på kvart einaste merke — dei stend i
	// `merke_defs` og vert henta med <use>. Sjå SkiveIndeks.

	// The slice: the class itself, and nothing more.
	//
	// There was a pale slice here too — the time until the class began. It was
	// unreadable: it looked like a night-and-day diagram and nobody knew what
	// it did. The dial answers "when does it run and for how long"; "how long
	// until" is a different question, and it stands in plain words in the
	// dock.
	//
	// The slice is shorter than the hands it stands for, so the pointers
	// always project beyond the surface rather than the other way round.
	frå := hourAngle(start)
	if d := slicePath(frå, frå+durationFraction(start, slutt), kakeR, skiveX, skiveY); d != "" {
		m.Slices = append(m.Slices, Slice{Class: "kake-timen", D: d})
	}

	m.Remaining, m.Capacity = att, plassar
	return m
}

// DeadSilhouette er ruta utan time: same mål som ei levande, so rada
// stend på lina, men berre eit hint av ei lina.
func DeadSilhouette() template.HTMLAttr {
	return template.HTMLAttr(boxPath(kroppX, kroppY, kroppB, kroppH, kroppR))
}
