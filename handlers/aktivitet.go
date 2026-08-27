package handlers

import (
	"fmt"
	"time"
)

// Aktiviteten, teikna. To bilete av det same halve aaret: eit varmekart
// yver kvar einskild dag, og ei bjelkerad yver kvar maanad.
//
// Reknestykki stend her og ikkje i malen — same regelen som for
// timemerket. Ein mal skal syna kva som er i figuren; kvar kvart punkt
// ligg, er ein annan slags kunnskap.
//
// Fargen er *ein* farge i fire trinn, ikkje fire fargar. Talet er ei
// mengd og ikkje ein identitet, og ei mengd les ein av kor mykje blekk
// som ligg der. Trinni stend i stilarket som color-mix mot arket, so dei
// fylgjer temaet: i det ljose gjeng dei fraa ljost til metta, i det
// myrke fraa nesten-grunnen upp til full farge.
const (
	ruteSide = 11.0
	ruteRom  = 2.0 // mellomrommet i flatefargen, som i mark-spec-en
	dagR     = 2.0

	bjelkeRom  = 2.0
	bjelkeR    = 2.0 // runda dataende, forankra i grunnlina — same verdet som --rund
	bjelkeHogd = 64.0
)

// Rute er ein dag i varmekartet.
type Rute struct {
	X, Y      float64
	Niva      int
	Merkelapp string
}

// Maanadsmerke er ei maanadsoverskrift yver kartet.
type Maanadsmerke struct {
	X    float64
	Namn string
}

// Bjelke er ein maanad i bjelkeraden.
type Bjelke struct {
	X, Y, Breidd, Hogd float64
	D                  string // runda dataende, flat fot
	TalX, TalY         float64
	Namn               string
	Tal                int
	Merkelapp          string
}

// Aktivitet er alt ein mal treng for aa teikna dei tvo bileti.
type Aktivitet struct {
	ViewBox      string
	Breidd, Hogd float64
	Rutar        []Rute
	Maanadar     []Maanadsmerke

	BjelkeViewBox string
	BjelkeBreidd  float64
	BjelkeHogd    float64
	Bjelkar       []Bjelke

	Totalt    int
	Vekor     int
	Samandrag string
}

// niva gjer eit tal um til eit av fire trinn. Fire er nok: fleire trinn
// enn auga kann skilja er ikkje meir informasjon, det er berre fleire
// fargar.
func niva(tal int) int {
	switch {
	case tal <= 0:
		return 0
	case tal == 1:
		return 1
	case tal == 2:
		return 2
	case tal == 3:
		return 3
	default:
		return 4
	}
}

// aktivitetMaanadar er kor mange maanadsbjelkar raden syner.
const aktivitetMaanadar = 6

// AktivitetFraa gjev den fyrste dagen begge bileti treng tal for.
//
// Dei tvo spyrja ikkje um det same tidsrommet: kartet gjeng vikevis
// attende fraa maandagen i denne vika, medan bjelkane gjeng maanadsvis
// attende fraa den fyrste i denne maanaden. Handsamaren henta berre det
// kartet trong, og daa mangla den eldste bjelken dei dagane som laag
// fyre den maandagen — upp til aatte av deim. Bjelken saag daa lægre ut
// enn maanaden var, og kor mykje kom an paa kva dag i maanaden det var.
func AktivitetFraa(naa time.Time, vekor int) time.Time {
	kart := VikeMaandag(naa, 0).AddDate(0, 0, -7*(vekor-1))
	bjelkar := time.Date(naa.Year(), naa.Month(), 1, 0, 0, 0, 0, naa.Location()).
		AddDate(0, -(aktivitetMaanadar - 1), 0)
	if bjelkar.Before(kart) {
		return bjelkar
	}
	return kart
}

// NyAktivitet byggjer baae bileti for dei siste `vekor` vikone fram til
// og med vika `naa` ligg i.
func NyAktivitet(lang string, perDag map[string]int, naa time.Time, vekor int) Aktivitet {
	// Kartet byrjar paa ein maandag, elles stend ikkje vekedagane paa
	// same rad heile vegen bortyver.
	sisteMaandag := VikeMaandag(naa, 0)
	fyrsteMaandag := sisteMaandag.AddDate(0, 0, -7*(vekor-1))

	a := Aktivitet{Vekor: vekor}

	var sisteMaanad time.Month
	for v := 0; v < vekor; v++ {
		maandag := fyrsteMaandag.AddDate(0, 0, 7*v)
		x := float64(v) * (ruteSide + ruteRom)

		// Maanadsmerket stend yver den vika maanaden byrjar i.
		if maandag.Month() != sisteMaanad {
			// Berre naar det er plass til ordet; elles klumpar dei seg.
			if len(a.Maanadar) == 0 || x-a.Maanadar[len(a.Maanadar)-1].X > 26 {
				a.Maanadar = append(a.Maanadar, Maanadsmerke{X: x, Namn: maanadskort(lang, maandag.Month())})
			}
			sisteMaanad = maandag.Month()
		}

		for d := 0; d < 7; d++ {
			dag := maandag.AddDate(0, 0, d)
			if dag.After(naa) {
				continue // dagar som ikkje hev vore enno
			}
			tal := perDag[dag.Format("2006-01-02")]
			a.Totalt += tal
			a.Rutar = append(a.Rutar, Rute{
				X:         x,
				Y:         float64(d) * (ruteSide + ruteRom),
				Niva:      niva(tal),
				Merkelapp: merkelapp(lang, dag, tal),
			})
		}
	}

	a.Breidd = float64(vekor)*(ruteSide+ruteRom) - ruteRom
	a.Hogd = 7*(ruteSide+ruteRom) - ruteRom
	a.ViewBox = fmt.Sprintf("0 0 %.2f %.2f", a.Breidd, a.Hogd)

	a.byggBjelkar(lang, perDag, naa)
	a.Samandrag = samandrag(lang, a.Totalt, vekor)
	return a
}

// byggBjelkar summerer dei same dagane per maanad. Bjelkane ber alle
// same fargen: lengdi er talet, og aa fargeleggja dei etter verdien
// hadde sagt det same ein gong til med ein kanal som daa ikkje kann
// segja noko anna.
func (a *Aktivitet) byggBjelkar(lang string, perDag map[string]int, naa time.Time) {
	const tal = aktivitetMaanadar

	type mnd struct {
		namn string
		sum  int
	}
	maanadar := make([]mnd, tal)

	fyrste := time.Date(naa.Year(), naa.Month(), 1, 0, 0, 0, 0, naa.Location()).AddDate(0, -(tal - 1), 0)
	for i := 0; i < tal; i++ {
		m := fyrste.AddDate(0, i, 0)
		maanadar[i].namn = maanadskort(lang, m.Month())
		slutt := m.AddDate(0, 1, 0)
		for d := m; d.Before(slutt); d = d.AddDate(0, 0, 1) {
			maanadar[i].sum += perDag[d.Format("2006-01-02")]
		}
	}

	toppen := 1
	for _, m := range maanadar {
		if m.sum > toppen {
			toppen = m.sum
		}
	}

	breidd := 26.0
	for i, m := range maanadar {
		h := float64(m.sum) / float64(toppen) * bjelkeHogd
		if m.sum > 0 && h < bjelkeR {
			h = bjelkeR // ein bjelke med noko i skal aldri sjaa tom ut
		}
		x := float64(i) * (breidd + bjelkeRom)
		a.Bjelkar = append(a.Bjelkar, Bjelke{
			X: x, Y: bjelkeHogd - h, Breidd: breidd, Hogd: h,
			D:    bjelkePath(x, bjelkeHogd-h, breidd, h, bjelkeR),
			TalX: x + breidd/2, TalY: bjelkeHogd - h - 5,
			Namn: m.namn, Tal: m.sum,
			Merkelapp: fmt.Sprintf("%s: %d", m.namn, m.sum),
		})
	}
	a.BjelkeBreidd = float64(tal)*(breidd+bjelkeRom) - bjelkeRom
	a.BjelkeHogd = bjelkeHogd + 16 // rom til maanadsnamnet under
	a.BjelkeViewBox = fmt.Sprintf("0 -14 %.2f %.2f", a.BjelkeBreidd, a.BjelkeHogd+14)
}

// maanadskort gjev det stutte maanadsnamnet paa det maalet sida stend i.
func maanadskort(lang string, m time.Month) string {
	return t(lang, fmt.Sprintf("maanad.kort_%d", int(m)))
}

// merkelapp er det ein fær naar ein held peikaren yver ein dag. Det er
// ogso det ein skjermlesar les — difor stend talet i ord og ikkje som
// ein farge.
func merkelapp(lang string, dag time.Time, tal int) string {
	dato := dag.Format("2.1.2006")
	if tal == 0 {
		return fmt.Sprintf("%s — %s", dato, t(lang, "aktivitet.ingen"))
	}
	if tal == 1 {
		return fmt.Sprintf("%s — 1 %s", dato, t(lang, "aktivitet.time"))
	}
	return fmt.Sprintf("%s — %d %s", dato, tal, t(lang, "aktivitet.timar"))
}

// samandrag er tali sagde i ord. Fargesteg under 3:1 mot grunnen skal
// ha ei avlasting som kann lesast; dette er henne, og ho stend der for
// alle og ikkje berre for den som held peikaren i ro.
func samandrag(lang string, totalt, vekor int) string {
	ord := t(lang, "aktivitet.timar")
	if totalt == 1 {
		ord = t(lang, "aktivitet.time")
	}
	return fmt.Sprintf(t(lang, "aktivitet.samandrag"), totalt, ord, vekor)
}

// bjelkePath teiknar ein bjelke med runda dataende og flat fot. Runding
// i baae endar hadde late honom flyta; foten hans er grunnlina, og ei
// grunnlina er ikkje runda.
func bjelkePath(x, y, b, h, r float64) string {
	if h <= 0 {
		return ""
	}
	if r > h {
		r = h
	}
	if r > b/2 {
		r = b / 2
	}
	return fmt.Sprintf("M%.2f,%.2f V%.2f A%.2f,%.2f 0 0 1 %.2f,%.2f H%.2f A%.2f,%.2f 0 0 1 %.2f,%.2f V%.2f Z",
		x, y+h, y+r, r, r, x+r, y, x+b-r, r, r, x+b, y+r, y+h)
}
