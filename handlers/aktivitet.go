package handlers

import (
	"fmt"
	"strings"
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
	// Brettet. 52 vikor gjeng upp i 13 × 4: kvar rad er eit kvartal.
	// Ei rad paa 52 var so tunn i ei sidespalta at ho ikkje kunde
	// lesast; eit brett gjev kvart hol fire gonger so mykje rom.
	brettKolonnar = 13

	bjelkeRom  = 2.0
	bjelkeR    = 2.0  // runda dataende, forankra i grunnlina — same verdet som --rund
	bjelkeHogd = 44.0 // var 64; bjelkeraden tok meir høgd enn ho sa noko med
)

// HeatCell er ei *vika* i aktivitetsraden.
//
// Han var ein dag fyrr, og eit halvt aar var 182 ruter i sju rader. Det
// svara paa «kva dagar», og det er ikkje spursmaalet ein stiller seg
// naar ein opnar heimesida — ein vil vita um ein held det gaaende. Eit
// aar med 52 prikkar paa ei rad svarar paa det, og fær rom i ei
// sidespalta attpaa.
type HeatCell struct {
	X, Y   float64
	R      float64
	Level  int
	Slag   string // klassetypen som raadde den vika, som rååverd
	Klasse string // klassa han fær; tom naar slaget ikkje er teikna
	Label  string
}

// MonthLabel er merkelappen til ei rad paa brettet — maanaden det
// kvartalet byrjar i.
type MonthLabel struct {
	Name string
}

// BarBit er eitt slag trening inni ein maanadsbjelke.
type BarBit struct {
	D      string // runda dataende paa den øvste, flat fot paa dei hine
	Klasse string // fargen han fær; tom naar slaget ikkje er teikna
	Label  string
}

// Bar er ein maanad i bjelkeraden.
//
// Han var eitt einsfarga stykke fyrr. No er han delt i dei slagi
// maanaden var sett saman av, av di raden stend rett under brettet og
// brettet ber dei same fargane: er dei tvo bileti av det same aaret,
// skal dei segja det same um det. Ein turkis mars paa brettet og ein
// turkis mars i raden er den same marsen.
type Bar struct {
	X, Y, Width, Height float64
	Bitar               []BarBit
	GlimD               string // heile bjelken si form, til ljosglimet
	ValueX, ValueY      float64
	Name                string
	Value               int
	Label               string
}

// Activity er alt ein mal treng for aa teikna dei tvo bileti.
type Activity struct {
	// Brettet: kor mange hol paa rad, og kor mange rader.
	Width, Height float64
	Cells         []HeatCell
	Months        []MonthLabel

	BarViewBox string
	// Grunnlina aat maanadsnamnet. Han stod som eit tal i malen fyrr,
	// rekna for ei bjelkehøgd som sidan vart mindre — og daa la namni
	// seg uppaa bilina under figuren.
	BarNameY float64
	// Snittet per maanad, ferdig sagt. Bjelkeraden syner kvar maanad
	// for seg; snittet er det eine talet som segjer kva dei er saman.
	BarSnitt  string
	BarWidth  float64
	BarHeight float64
	Bars      []Bar

	Total   int
	Weeks   int
	Summary string
}

// level gjer talet paa timar i ei vika um til eit av tri lysstyrketrinn.
//
// Ei vika ein var innom i det heile er tend, men ikkje heilt: han skal
// kunna skiljast fraa ei vika ein verkeleg var i gang. Tri gonger er
// full styrke — det er den vika ein kom for. Yver tri lyser han ikkje
// meir; det finst ingen sterkare enn full.
func level(tal int) int {
	switch {
	case tal <= 0:
		return 0
	case tal == 1:
		return 1
	case tal == 2:
		return 2
	default:
		return 3
	}
}

// aktivitetMaanadar er kor mange maanadsbjelkar raden syner.
const aktivitetMaanadar = 6

// ActivityStart gjev den fyrste dagen begge bileti treng tal for.
//
// Dei tvo spyrja ikkje um det same tidsrommet: kartet gjeng vikevis
// attende fraa maandagen i denne vika, medan bjelkane gjeng maanadsvis
// attende fraa den fyrste i denne maanaden. Handsamaren henta berre det
// kartet trong, og daa mangla den eldste bjelken dei dagane som laag
// fyre den maandagen — upp til aatte av deim. Bjelken saag daa lægre ut
// enn maanaden var, og kor mykje kom an paa kva dag i maanaden det var.
func ActivityStart(naa time.Time, vekor int) time.Time {
	kart := VikeMaandag(naa, 0).AddDate(0, 0, -7*(vekor-1))
	bjelkar := time.Date(naa.Year(), naa.Month(), 1, 0, 0, 0, 0, naa.Location()).
		AddDate(0, -(aktivitetMaanadar - 1), 0)
	if bjelkar.Before(kart) {
		return bjelkar
	}
	return kart
}

// NewActivity byggjer baae bileti for dei siste `vekor` vikone fram til
// og med vika `naa` ligg i.
func NewActivity(lang string, perDag map[string]int, perType map[string]map[string]int, naa time.Time, vekor int) Activity {
	// Kartet byrjar paa ein maandag, elles stend ikkje vekedagane paa
	// same rad heile vegen bortyver.
	sisteMaandag := VikeMaandag(naa, 0)
	fyrsteMaandag := sisteMaandag.AddDate(0, 0, -7*(vekor-1))

	a := Activity{Weeks: vekor}

	for v := 0; v < vekor; v++ {
		maandag := fyrsteMaandag.AddDate(0, 0, 7*v)

		// Ei merkelapp per rad: maanaden kvartalet byrjar i.
		if v%brettKolonnar == 0 {
			a.Months = append(a.Months, MonthLabel{Name: monthAbbrev(lang, maandag.Month())})
		}

		// Summen for vika, og kva slag trening som raadde henne.
		vikesum := 0
		perSlag := map[string]int{}
		komen := false
		for d := 0; d < 7; d++ {
			dag := maandag.AddDate(0, 0, d)
			if dag.After(naa) {
				continue // dagar som ikkje hev vore enno
			}
			komen = true
			nykel := dag.Format("2006-01-02")
			vikesum += perDag[nykel]
			for slag, tal := range perType[nykel] {
				perSlag[slag] += tal
			}
		}
		a.Total += vikesum

		// Raadde slaget. Ved likt vinn det alfabetisk fyrste, so brettet
		// ikkje skiftar farge mellom tvo teikningar av dei same tali.
		beste := ""
		if komen {
			for slag, tal := range perSlag {
				if slag == "" {
					continue
				}
				if tal > perSlag[beste] || (tal == perSlag[beste] && (beste == "" || slag < beste)) {
					beste = slag
				}
			}
		}

		// Ei vika som ikkje hev vore enno er eit tomt hol som alle
		// andre: brettet er bora ferdig, og pinnane kjem etter kvart.
		a.Cells = append(a.Cells, HeatCell{
			Level:  level(vikesum),
			Slag:   beste,
			Klasse: slagKlasse[beste],
			Label:  weekLabel(lang, maandag, vikesum),
		})
	}

	a.Width = brettKolonnar
	a.Height = float64((vekor + brettKolonnar - 1) / brettKolonnar)

	a.buildBars(lang, perDag, perType, naa)
	a.Summary = summaryText(lang, a.Total, vekor, a.BarSnitt)
	return a
}

// buildBars summerer dei same dagane per maanad. Bjelkane ber alle
// same fargen: lengdi er talet, og aa fargeleggja dei etter verdien
// hadde sagt det same ein gong til med ein kanal som daa ikkje kann
// segja noko anna.
func (a *Activity) buildBars(lang string, perDag map[string]int, perType map[string]map[string]int, naa time.Time) {
	const tal = aktivitetMaanadar

	type mnd struct {
		namn string
		sum  int
		slag map[string]int
	}
	maanadar := make([]mnd, tal)

	fyrste := time.Date(naa.Year(), naa.Month(), 1, 0, 0, 0, 0, naa.Location()).AddDate(0, -(tal - 1), 0)
	for i := 0; i < tal; i++ {
		m := fyrste.AddDate(0, i, 0)
		maanadar[i].namn = monthAbbrev(lang, m.Month())
		maanadar[i].slag = map[string]int{}
		slutt := m.AddDate(0, 1, 0)
		for d := m; d.Before(slutt); d = d.AddDate(0, 0, 1) {
			nykel := d.Format("2006-01-02")
			maanadar[i].sum += perDag[nykel]
			for slag, n := range perType[nykel] {
				maanadar[i].slag[slag] += n
			}
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
		b := Bar{
			X: x, Y: bjelkeHogd - h, Width: breidd, Height: h,
			// Glimet ligg yver *heile* bjelken og ikkje yver kvar bit:
			// maanaden er tingen ein tek i, bitane er berre kva han er
			// laga av. Eit glim per bit hadde gjort kvar farge til sin
			// eigen kuppel.
			GlimD:  barPath(x, bjelkeHogd-h, breidd, h, bjelkeR),
			ValueX: x + breidd/2, ValueY: bjelkeHogd - h - 5,
			Name: m.namn, Value: m.sum,
			Label: fmt.Sprintf("%s: %d", m.namn, m.sum),
		}

		// Bitane stend i den same rekkjefylgda kvar maanad, so auga kann
		// samanlikna dei paa tvers: skifte dei plass etter kor mange det
		// var, laut ein lesa kvar bjelke for seg.
		if m.sum > 0 {
			botn := bjelkeHogd
			for j, slag := range slagRekkja {
				n := m.slag[slag]
				if n == 0 {
					continue
				}
				bh := float64(n) / float64(m.sum) * h
				topp := botn - bh

				// Berre den øvste biten hev runda dataende — dei under
				// er midt inne i bjelken og skal ikkje sjaa ut som
				// eigne bjelkar.
				var d string
				if erOvst(m.slag, j) {
					d = barPath(x, topp, breidd, bh, bjelkeR)
				} else {
					d = fmt.Sprintf("M%.2f,%.2f H%.2f V%.2f H%.2f Z",
						x, topp, x+breidd, botn, x)
				}
				b.Bitar = append(b.Bitar, BarBit{
					D:      d,
					Klasse: slagKlasse[slag],
					Label:  fmt.Sprintf("%s — %s: %d", m.namn, slag, n),
				})
				botn = topp
			}
		}

		a.Bars = append(a.Bars, b)
	}
	sum := 0
	for _, m := range maanadar {
		sum += m.sum
	}
	a.BarSnitt = snittText(lang, float64(sum)/float64(tal))

	a.BarWidth = float64(tal)*(breidd+bjelkeRom) - bjelkeRom
	a.BarNameY = bjelkeHogd + 14
	a.BarHeight = bjelkeHogd + 16 // rom til maanadsnamnet under
	a.BarViewBox = fmt.Sprintf("0 -14 %.2f %.2f", a.BarWidth, a.BarHeight+14)
}

// monthAbbrev gjev det stutte maanadsnamnet paa det maalet sida stend i.
func monthAbbrev(lang string, m time.Month) string {
	return t(lang, fmt.Sprintf("maanad.kort_%d", int(m)))
}

// slagKlasse er dei klassane ein prikk kann faa av eit klasseslag.
//
// Tvo grunnar til at dette er ei liste og ikkje `"slag-" + slag`:
//
// Det eine er at `class_type` kjem raatt or basen. Sette malen klassen
// saman sjølv, kunde kva som helst som stod i den kolonna hamna som eit
// klassenamn i dokumentet — ein verdi ingen hev teikna, og ingen ser
// fyre dei skriv honom inn.
//
// Det andre er at ein klasse som berre finst samansett ikkje kann
// finnast att. `scripts/daude-klassar.sh` leitar etter namnet slik det
// stend i stilarket, og `slag-yoga` fanst ingen stad i kjelda — so
// prøva melde honom daud medan han var i bruk. No stend han her.
// slagRekkja er rekkjefylgda bitane stend i, nedanfraa og upp. Ho er
// fast: skifte dei plass etter storleik, laut ein lesa kvar bjelke for
// seg i staden for aa samanlikna deim.
var slagRekkja = []string{"fascia", "yoga", "pilates", "reformer"}

// erOvst segjer um biten paa plass `j` er den siste som hev noko i seg —
// den einaste som skal ha runda dataende.
func erOvst(slag map[string]int, j int) bool {
	for k := j + 1; k < len(slagRekkja); k++ {
		if slag[slagRekkja[k]] > 0 {
			return false
		}
	}
	return true
}

var slagKlasse = map[string]string{
	"fascia":   "slag-fascia",
	"yoga":     "slag-yoga",
	"pilates":  "slag-pilates",
	"reformer": "slag-reformer",
}

// weekLabel er teksten ein fær naar ein held peikaren yver ein prikk.
// Vika, ikkje dagen: «veka fraa 3.2.2026 — 2 timar».
func weekLabel(lang string, maandag time.Time, tal int) string {
	dato := maandag.Format("2.1.2006")
	if tal == 0 {
		return fmt.Sprintf("%s %s — %s", t(lang, "aktivitet.veka_fraa"), dato, t(lang, "aktivitet.ingen"))
	}
	if tal == 1 {
		return fmt.Sprintf("%s %s — 1 %s", t(lang, "aktivitet.veka_fraa"), dato, t(lang, "aktivitet.time"))
	}
	return fmt.Sprintf("%s %s — %d %s", t(lang, "aktivitet.veka_fraa"), dato, tal, t(lang, "aktivitet.timar"))
}

// summaryText er tali sagde i ord. Fargesteg under 3:1 mot grunnen skal
// ha ei avlasting som kann lesast; dette er henne, og ho stend der for
// alle og ikkje berre for den som held peikaren i ro.
func summaryText(lang string, totalt, vekor int, snitt string) string {
	ord := t(lang, "aktivitet.timar")
	if totalt == 1 {
		ord = t(lang, "aktivitet.time")
	}
	setning := fmt.Sprintf(t(lang, "aktivitet.samandrag"), totalt, ord, vekor)
	if snitt != "" {
		setning += snitt
	}
	return setning
}

// snittText segjer snittet med komma og éin desimal, som ein skriv tal
// paa norsk. `%.1f` gjev punktum, og eit punktum i eit tal les seg som
// tusenskilje her.
//
// Han er ein *hale* paa samandraget og ikkje ei setning for seg. Stod
// han under bjelkeraden, laag han eit halvt bilete unna det han
// skildrar, og bolken fekk ei lina med tekst midt inni seg som braut
// takti. Yverst svarar dei tvo paa det same spursmaalet — kor mykje —
// og daa er dei éi setning.
func snittText(lang string, snitt float64) string {
	tal := strings.Replace(fmt.Sprintf("%.1f", snitt), ".", ",", 1)
	return fmt.Sprintf(t(lang, "aktivitet.snitt"), tal)
}

// barPath teiknar ein bjelke med runda dataende og flat fot. Runding
// i baae endar hadde late honom flyta; foten hans er grunnlina, og ei
// grunnlina er ikkje runda.
func barPath(x, y, b, h, r float64) string {
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
