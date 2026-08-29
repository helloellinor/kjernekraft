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
	//
	// Kassen var 72 høg mot 56 brei, og det er ikkje eit merke lenger —
	// det er ei søyle. Ei rad på heimesida vart hundre piksel høg av
	// merket åleine, og ei liste med fem timar fylte skjermen. Han er
	// 60 no: framleis høgare enn brei, av di han *er* ein kasse med tri
	// bolkar, men ikkje så høg at rada bøygjer seg etter honom.
	//
	// Vil ein stilla på dette, er kroppH talet. Alt anna er rekna ut or
	// honom — botnstripa, plassmerket og skiva flytter seg med.
	kroppB, kroppH = 56.0, 60.0
	kroppR         = 4.0 // rektangel med so vidt broti hyrna
	// Fana spenner yver toppen, so dagen stend *heil*. Ho var 44 av 56
	// og laut forkortast til tvo bokstavar — «TY» er ikkje eit ord, det
	// er ein kode ein lyt læra. Med 52 av 56 stend «TYSDAG», og tvo
	// einingar skulder paa kvar sida syner framleis at fana ligg attum
	// kroppen.
	faneB, faneH = 52.0, 13.0
	faneR        = 3.0
	faneLoft     = 11.0
	merkeKant    = 1.0

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

	// Vindauga stend midt i botnen. Det hekk i vinstre kanten fyrr —
	// `kroppX + botnRom` — av di plassmerket laag nede til høgre og
	// dei tvo delte stripa. Merket ligg i hyrna no, uppaa kanten, so
	// stripa er vindauga sin aaleine og han kann staa der han skal.
	ruteB, ruteH = 32.0, 14.5
	ruteX        = kroppX + (kroppB-ruteB)/2
	ruteY        = kroppY + kroppH - ruteH - botnRom
	ruteR        = 1.8

	// Plassmerket er ikkje teikna i SVG-en lenger.
	//
	// Det var ein figur som *likna* ein knapp: ei lapp med ein kuppel
	// male i ein gradient. Ho vart aldri heilt lik, av di knappen ber
	// line, radius, djupn og eit piksels sig som SVG-en lyt herma etter
	// kvar for seg. No er merket eit HTML-element uppaa figuren, og det
	// tek stilen aat knappen beinveges (`.btn, … , .plettmerke`).
	//
	// Det som stend att her er staden: hyrna nede til høgre, gjeven i
	// prosent av kassen, so merket kan sentrerast *paa* honom.
	hyrneX = (kroppX + kroppB) / MarkWidth * 100
	hyrneY = (kroppY + kroppH) / MarkHeight * 100

	MarkLeft   = 0.0
	MarkWidth  = kroppB + 2*merkeKant
	MarkHeight = kroppY + kroppH + merkeKant

	// Skiva stend midt i rommet yver botnstripa, og visarane stend midt
	// i skiva. Ho laag halvt nedi vindauga fyrr — klokkeslettet var
	// teikna tvert yver henne — og daa las korkje uret eller talet seg.
	//
	// Midten vert rekna ut og ikkje skriven inn. Han stod som 38 —
	// eit tal som høvde til den kroppshøgdi som var — og skiva hadde
	// vorte hangande for lågt med ein gong kassen skifte høgd.
	// Margen fraa kroppkanten inn til skiveflata — det som stend att av
	// kroppen kring den senka plata, kassen sin fals. Sporet byrjar 4
	// inne fraa kroppen, so lippa paa plata (kring 1–2 einingar) møter
	// ikkje indeksane.
	skiveMarg = 3.0

	skiveX = kroppX + kroppB/2
	// Skiva stend midt i kroppen — baade i breidd og i høgd.
	//
	// Midten hennar laag paa (kroppY+ruteY)/2 fyrr, altso ni einingar
	// yver midten av kassen. Ei vinkelfordeling kring eit punkt som
	// ikkje er midten *kann* ikkje verta jamn: tolv stod tett uppi
	// toppen medan tri og ni stod høgt yver midtlina, og heile botnen
	// stod tom under.
	//
	// Vindauga dekkjer den nedste biten av skiva no. Det ligg i sitt
	// eige lag uppaa henne og er ei ugjenomsynleg flate, so det som
	// hamnar under, hamnar under.
	skiveY = kroppY + kroppH/2

	// Skiva set høgdi paa heile merket.
	//
	// Ho er ikkje kakestykket — ho er *sporet* indeksane ligg paa, og
	// det er breidare. Kassen lyt vera høg nok til å halda henne heil
	// med botnstripa under, so kroppH kan ikkje kortast utan at skiva
	// fylgjer med. Prøvone i merkeform_test held det saman: ei skiva som
	// ikkje får plass er ei skiva som vert teikna ned i klokkeslettet.
	// Sporet indeksane ligg paa. Det er ein runda firkant som fyller
	// andlitet — ikkje ein liten firkant inni ein stor.
	//
	// Det var 17,6 i alle leider: forma var firkanta, men so mykje
	// mindre enn kassen at hyrni hennar aldri kom i nærleiken av hyrni
	// hans, og daa ser skiva rund ut likevel. Kassen gjev 21 upp og ned
	// (vindauga stengjer nedantil) og 28 til sidone; sporet tek det
	// meste av det og let ein jamn marg staa.
	// Sporet er ikkje symmetrisk um midten, av di kassen ikkje er det
	// heller: skiva svingar kring eit punkt som ligg 21 under toppen og
	// 39 yver botnen. Med same maalet upp og ned stogga dei nedre
	// hyrni midt paa kroppen, medan dei øvre naadde heilt ut — og daa
	// ser skiva ut som ho hev skore botnen av seg sjølv.
	//
	// Vindauga stend i vegen nedantil, men det ligg i sitt eige lag
	// uppaa skiva og dekkjer det det dekkjer. Indeksane gjeng under.
	// Symmetrisk, av di skiva stend midt i kassen. Det var skeivt (18
	// upp, 30 ned) so lenge midten hennar laag høgt — ei bot paa at ho
	// stod feil, ikkje ei form.
	//
	// Og 3,5 *innanfor skiveflata*, ikkje ut til kanten hennar. Sporet
	// gjekk til 24/27 daa flata kom (29.8.2026), og daa stod indeksane
	// beint paa saumen millom fals og skiva — «12» drukna i skuggen
	// fraa yverkanten, og «9» og «3» stod halvt paa kvar si flata.
	// Prenten paa eit ur stend paa skiva, med marg til kanten.
	sporB   = 21.5 // halvbreidd
	sporOpp = 23.5 // upp fraa midten
	sporNed = 23.5 // ned fraa midten
	spor    = sporOpp
	// Stykket naar nettupp so langt som timevisaren det stend for —
	// ikkje eit hakk kortare, og aldri lenger.
	kakeR = visarTime

	// Visarane, i *lengd* fraa skivemidten. `NewMark` gjer deim um til
	// y-endepunkt med `skiveY - lengd`.
	//
	// Eg las dei ein gong som endepunkt og «retta» dei difor motsett
	// veg: timevisaren vart den lange og minuttvisaren den korte, som er
	// eit ur som lyg um kva klokka er. Prøvone mine rekna det same
	// uttrykket paa den same gale maaten og sa god for det. Difor stend
	// eininga skriven her no, og difor prøver `TestVisaraneErRettVegen`
	// dei ferdige felti i `Mark` og ikkje konstantane.
	visarTime   = 12.0 // timevisaren, den korte
	visarMinutt = 17.5 // minuttvisaren, den lange
)

// Tick er ein indeks paa skiva.
type Tick struct {
	X1, Y1, X2, Y2 float64
	Quarter        bool
}

// Numeral er eit tal paa skiva.
type Numeral struct {
	X, Y float64
	Text string
}

// Slice er eit kakestykke, med klassa som seier kva det tyder.
type Slice struct {
	Class string
	D     string
}

// Mark er alt ein mal treng for aa teikna eit timemerke.
type Mark struct {
	// Heile kassen, ferdig skriven: «−8.6 0 65.2 65.6».
	ViewBox string
	// Filtri bur inne i kvar figur og ikkje i eit sams sett paa sida.
	// Grunnen er tema: `flood-color: var(--kant)` vert løyst der
	// *filteret* stend, ikkje der det vert nytta. Eit sams sett hadde
	// difor gjeve alle merki fargane til rot-temaet, og ein ljos bolk
	// paa ei myrk sida — som verkstaden set upp — hadde fenge feil
	// lina. Ident gjer id-ane eintydige.
	Ident         string
	Width, Height float64
	Silhouette    template.HTMLAttr // umrisset: kant, ring og flate deler han
	DayX, DayY    float64
	// Glaset. Sola stend paa 34 % 22 % av kroppen, som alt anna runda i
	// huset (--sol). Tali er rekna i merket sitt *eige* rom og ikkje i
	// prosent av omrisset: eit omriss som er høgare enn det er breidt
	// strekkjer ein objectBoundingBox-gradient til ein oval, og daa vert
	// speglingi ei flekk i staden for eit ljospunkt.
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
	// senka skiva, og eit djupare datovindauga, som paa eit ur.
	Skive        template.HTMLAttr
	Box          template.HTMLAttr
	BoxTextX     float64
	BoxTextY     float64
	Full         bool
	Remaining    int // plassar att
	Capacity     int // av so mange
	// Hyrna merket skal sentrerast paa, i prosent av kassen.
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

// silhouette er dei tri formene som éin veg. Umrisset og fordjupingi vert
// laga av honom med filter, ikkje ved aa teikna formene ein gong til
// litt større: utvidar ein kvar form for seg, vandrar dei innhole hyrno
// innyver i staden for utyver, og daa slepp lina upp nettupp i skøyten.
func silhouette() string {
	return boxPath(kroppX+merkeKant, kroppY+merkeKant, kroppB-2*merkeKant,
		kroppH-2*merkeKant, kroppR-merkeKant)
}

// dayTabPath er dagfana. Ho vart teikna saman med kroppen som éi form
// fyrr, og daa fanst det ingi fana — det fanst eitt umriss med ein
// kul paa. Ei fana er ein *eigen* ting som ligg attum arket sitt: ho
// gjeng ned under yverkanten aat kroppen, kroppen dekkjer nedkanten
// hennar, og det som stikk upp er fana. Same grepet som fanone i
// administrasjonen, berre teikna.
func dayTabPath() string {
	return boxPath(faneX+merkeKant, faneY+merkeKant, faneB-2*merkeKant,
		faneH+6-2*merkeKant, faneR-merkeKant)
}

// onDialTrack gjev eit punkt paa sporet, eller `inn` einingar innanfor
// det.
//
// Sporet er ein runda firkant og ikkje ein sirkel: brikka er eit
// firkanta ur, og daa skal indeksane fylgja kassen. Ein sirkel let
// hyrni staa tome.
//
// Det stod ein superellipse her fyrr (|dx|⁶+|dy|⁶). Han er *nesten*
// firkanta — paa skraa naar han 1,26 av det han naar rett ut, medan ein
// firkant naar 1,41 — og aa skru paa eksponenten flytte indeksane knapt
// eit par prosent.
//
// `inn` er eit stykke *langs straalen*, ikkje eit mindre spor. Skalerte
// me sporet i staden, vart indeksane ulike lange rundt skiva: eit hakk
// paa toppen hadde vorte kortare enn eit paa sida, av di sporet er
// oblongt.
func onDialTrack(grader, inn float64) (float64, float64) {
	t := grader * math.Pi / 180
	dx, dy := math.Sin(t), -math.Cos(t)

	// Kor langt straalen gjeng fyrr han møter ei av sidone.
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

// slicePath er eit kakestykke fraa ein vinkel til ein annan.
func slicePath(fraa, til, r, cx, cy float64) string {
	sveip := math.Mod(til-fraa, 360)
	if sveip < 0 {
		sveip += 360
	}
	if sveip < 0.4 {
		return ""
	}
	if sveip > 359.4 {
		return circlePath(cx, cy, r)
	}
	x1, y1 := point(fraa, r, cx, cy)
	x2, y2 := point(til, r, cx, cy)
	stor := 0
	if sveip > 180 {
		stor = 1
	}
	return fmt.Sprintf("M%.2f,%.2f L%.2f,%.2f A%.2f,%.2f 0 %d,1 %.2f,%.2f Z",
		cx, cy, x1, y1, r, r, stor, x2, y2)
}

// hourAngle er klokkeslettet paa timevisaren si skala: ei runda er tolv
// timar. Kakestykki ligg paa den same skalaen, so eit stykke er kor stor
// bit av dagen timen tek. Minuttskalaen hadde vore meir synleg — ein
// time paa 45 minutt hadde vorte trikvart av skiva — men ein time som
// varer lenger enn ein time hadde gjenge heile vegen rundt og ikkje
// kunna segja meir.
func hourAngle(t time.Time) float64 {
	return float64(t.Hour()%12)*30 + float64(t.Minute())*0.5
}

// minuteAngle er kvar minuttvisaren stend: ei runda er ein time.
func minuteAngle(t time.Time) float64 {
	return float64(t.Minute()) * 6
}

// durationFraction er kor stort stykke timen tek av skiva, paa timevisaren si
// skala: ei runda er tolv timar, so ein time er tretti grader.
//
// Minuttskalaen hadde vore meir synleg — ein time paa 45 minutt hadde
// vorte trikvart av skiva — men ein time som varer lenger enn ein time
// hadde gjenge heile vegen rundt og ikkje kunna segja meir. Difor stend
// stykket i minuttvisaren, men maaler seg paa timeskalaen.
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
		// Sluttvisaren heng i *timevisaren*, ikkje i minuttvisaren.
		//
		// Han hekk i minuttvisaren fyrr, medan stykket maalte seg paa
		// timeskalaen — tvo skalaer i den same figuren. Ein time paa
		// seksti minutt tok daa tretti grader ut fraa ein peikar som
		// sjølv gjeng seks grader i minuttet, og biletet svara ikkje paa
		// noko: korkje «kva klokka er» eller «kor lang timen er».
		//
		// Paa timeskalaen er baae sanne samstundes. Timevisaren stend
		// der timen byrjar, sluttvisaren der han er ute, og stykket
		// millom dei *er* timen — same maaten eit ur syner at klokka
		// hev gjenge.
		EndAngle: hourAngle(start) + durationFraction(start, slutt),
		Skive:    template.HTMLAttr(boxPath(kroppX+skiveMarg, kroppY+skiveMarg, kroppB-2*skiveMarg, kroppH-2*skiveMarg, kroppR-1)),
		Box:      template.HTMLAttr(boxPath(ruteX, ruteY, ruteB, ruteH, ruteR)),
		BoxTextX: ruteX + ruteB/2,
		BoxTextY: ruteY + ruteH/2,
		Full:     full,
		HyrneX:   hyrneX,
		HyrneY:   hyrneY,
	}

	// Indeksane. Sekstalet ligg under vindauga og skal ikkje teiknast.
	for h := 0; h < 12; h++ {
		if h == 6 {
			continue
		}
		g := float64(h) * 30
		ux, uy := onDialTrack(g, 0)
		kvart := h%3 == 0
		if kvart {
			tx, ty := onDialTrack(g, 1.4)
			m.Numerals = append(m.Numerals, Numeral{X: tx, Y: ty + 2.4,
				Text: map[int]string{0: "12", 3: "3", 9: "9"}[h]})
			continue
		}
		lengd := 2.6
		ix, iy := onDialTrack(g, lengd)
		m.Ticks = append(m.Ticks, Tick{X1: ux, Y1: uy, X2: ix, Y2: iy, Quarter: false})
	}

	// Kakestykket: timen sjølv, og ikkje noko meir.
	//
	// Det låg eit bleikt stykke her med — tidi fram til timen byrja. Det
	// var ikkje til aa lesa: det saag ut som ei nott-og-dag-teikning, og
	// ingen visste kva det gjorde. Skiva svarar paa «naar gjeng han og
	// kor lenge»; «kor lenge til» er eit anna spursmaal, og det stend i
	// klartekst i dokka.
	// Timen sjølv veks ut or *timevisaren*.
	//
	// Han hekk i minuttvisaren ei stund, medan han maalte seg paa
	// timeskalaen — tvo skalaer i den same figuren, og daa svara biletet
	// korkje paa «kva klokka er» eller «kor lang timen er». Paa
	// timeskalaen er baae sanne samstundes: timevisaren stend der timen
	// byrjar, sluttvisaren der han er ute, og stykket millom dei er
	// timen. Dei tri les seg som éin ting.
	//
	// Stykket er kortare enn visarane det stend for, so peikarane alltid
	// stikk ut or flata og ikkje omvendt.
	fraa := hourAngle(start)
	if d := slicePath(fraa, fraa+durationFraction(start, slutt), kakeR, skiveX, skiveY); d != "" {
		m.Slices = append(m.Slices, Slice{Class: "kake-timen", D: d})
	}

	m.Remaining, m.Capacity = att, plassar
	return m
}

// DeadSilhouette er ruta utan time: same maal som ei levande, so rada
// stend paa lina, men berre eit hint av ei lina.
func DeadSilhouette() template.HTMLAttr {
	return template.HTMLAttr(boxPath(kroppX, kroppY, kroppB, kroppH, kroppR))
}
