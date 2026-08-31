package handsamarar

import (
	"fmt"
	"strings"
	"time"
)

// Activity, drawn. Two pictures of the same half year: a heat board over
// the weeks, and a bar row over the months.
//
// The arithmetic lives here, not in the template — same rule as the
// class mark.
//
// The colour is *one* colour in four steps, not four colours. A count is
// a quantity, not an identity, and a quantity reads from how much ink is
// there. The steps live in the stylesheet as color-mix against the
// sheet, so they follow the theme.
const (
	// The board. 52 weeks divide into 13 × 4: each row is a quarter. A
	// single row of 52 was too thin to read in a side column; a board
	// gives each hole four times the room.
	brettKolonnar = 13

	bjelkeRom  = 2.0
	bjelkeR    = 2.0  // runda dataende, forankra i grunnlina — same verdet som --rund
	bjelkeHogd = 44.0 // var 64; bjelkeraden tok meir høgd enn ho sa noko med
)

// HeatCell is one *week* on the board.
//
// It was a day once, and half a year came to 182 cells. That answers
// "which days", which is not what you ask when you open the home page —
// you want to know whether you are keeping it up. A year in 52 dots
// answers that, and fits in a side column.
type HeatCell struct {
	X, Y   float64
	R      float64
	Level  int
	Slag   string // klassetypen som raadde den vika, som rååverd
	Klasse string // klassa han fær; tom når slaget ikkje er teikna
	Label  string
}

// MonthLabel is a board row's label — the month that quarter starts
// in.
type MonthLabel struct {
	Name string
}

// BarBit is one kind of training inside a month bar.
type BarBit struct {
	D      string // runda dataende på den øvste, flat fot på dei hine
	Klasse string // fargen han fær; tom når slaget ikkje er teikna
	Label  string
}

// Bar is one month in the bar row.
//
// Split by the kinds the month was made of, because the row sits right
// under the board and the board carries the same colours. Two pictures
// of the same year must say the same thing about it: a turquoise March
// on the board and a turquoise March in the row are the same March.
type Bar struct {
	X, Y, Width, Height float64
	Bitar               []BarBit
	GlimD               string // heile bjelken si form, til ljosglimet
	ValueX, ValueY      float64
	Name                string
	Value               int
	Label               string
}

// Activity is everything a template needs to draw both pictures.
type Activity struct {
	// The board: holes per row, and how many rows.
	Width, Height float64
	Cells         []HeatCell
	Months        []MonthLabel

	BarViewBox string
	// Baseline for the month name. It was a literal in the template,
	// computed for a bar height that later shrank, and the names ended up
	// on top of the rule under the figure.
	BarNameY float64
	// The monthly average, already phrased. The bar row shows each month;
	// the average is the one number that says what they are together.
	BarSnitt  string
	BarWidth  float64
	BarHeight float64
	Bars      []Bar

	Total   int
	Weeks   int
	Summary string
}

// level turns a week's class count into one of three brightness steps.
//
// A week you turned up at all is lit, but not fully: it has to be
// distinguishable from a week you were really going. Three is full
// strength — that is the week you came for. Above three it does not glow
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

// ActivityStart gives the first day both pictures need numbers for.
//
// The two do not ask about the same span: the board goes back week by week
// from this Monday, while the bars go back month by month from the first
// of this month. The handler fetched only what the board needed, so the
// oldest bar was missing the days before that Monday — up to eight of
// them. The bar then looked lower than the month was, and by how much
// depended on the date.
func ActivityStart(nå time.Time, vekor int) time.Time {
	kart := VikeMåndag(nå, 0).AddDate(0, 0, -7*(vekor-1))
	bjelkar := time.Date(nå.Year(), nå.Month(), 1, 0, 0, 0, 0, nå.Location()).
		AddDate(0, -(aktivitetMaanadar - 1), 0)
	if bjelkar.Before(kart) {
		return bjelkar
	}
	return kart
}

// NewActivity byggjer båe bileti for dei siste `vekor` vikone fram til
// og med vika `naa` ligg i.
func NewActivity(lang string, perDag map[string]int, perType map[string]map[string]int, nå time.Time, vekor int) Activity {
	// Kartet byrjar på ein måndag, elles stend ikkje vekedagane på
	// same rad heile vegen bortyver.
	sisteMåndag := VikeMåndag(nå, 0)
	fyrsteMåndag := sisteMåndag.AddDate(0, 0, -7*(vekor-1))

	a := Activity{Weeks: vekor}

	for v := 0; v < vekor; v++ {
		måndag := fyrsteMåndag.AddDate(0, 0, 7*v)

		// Ei merkelapp per rad: maanaden kvartalet byrjar i.
		if v%brettKolonnar == 0 {
			a.Months = append(a.Months, MonthLabel{Name: monthAbbrev(lang, måndag.Month())})
		}

		// Summen for vika, og kva slag trening som raadde henne.
		vikesum := 0
		perSlag := map[string]int{}
		komen := false
		for d := 0; d < 7; d++ {
			dag := måndag.AddDate(0, 0, d)
			if dag.After(nå) {
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

		// The dominant kind. On a tie the alphabetically first wins, so the board
		// does not change colour between two drawings of the same numbers.
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

		// A week that has not happened yet is an empty hole like any other: the
		// board is drilled, and the pegs arrive as they come.
		a.Cells = append(a.Cells, HeatCell{
			Level:  level(vikesum),
			Slag:   beste,
			Klasse: SlagKlasse(beste),
			Label:  weekLabel(lang, måndag, vikesum),
		})
	}

	a.Width = brettKolonnar
	a.Height = float64((vekor + brettKolonnar - 1) / brettKolonnar)

	a.buildBars(lang, perDag, perType, nå)
	a.Summary = summaryText(lang, a.Total, vekor, a.BarSnitt)
	return a
}

// buildBars sums the same days per month. The bars all carry the same
// colour: the length is the number, and colouring them by value would say
// the same thing again on a channel that then cannot
func (a *Activity) buildBars(lang string, perDag map[string]int, perType map[string]map[string]int, nå time.Time) {
	const tal = aktivitetMaanadar

	type mnd struct {
		namn string
		sum  int
		slag map[string]int
	}
	maanadar := make([]mnd, tal)

	fyrste := time.Date(nå.Year(), nå.Month(), 1, 0, 0, 0, 0, nå.Location()).AddDate(0, -(tal - 1), 0)
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
			h = bjelkeR // ein bjelke med noko i skal aldri sjå tom ut
		}
		x := float64(i) * (breidd + bjelkeRom)
		b := Bar{
			X: x, Y: bjelkeHogd - h, Width: breidd, Height: h,
			// The highlight lies over the *whole* bar and not over each piece: the
			// 			// month is the thing you take hold of, the pieces are only what it
			// 			// is made of. A highlight per piece would make every colour its own
			// 			// dome.
			GlimD:  barPath(x, bjelkeHogd-h, breidd, h, bjelkeR),
			ValueX: x + breidd/2, ValueY: bjelkeHogd - h - 5,
			Name: m.namn, Value: m.sum,
			Label: fmt.Sprintf("%s: %d", m.namn, m.sum),
		}

		// The pieces stand in the same order every month, so the eye can compare
		// 		// them across: reordered by size, you would have to read each bar
		// 		// separately.
		if m.sum > 0 {
			botn := bjelkeHogd
			for j, slag := range Slagi {
				n := m.slag[slag]
				if n == 0 {
					continue
				}
				bh := float64(n) / float64(m.sum) * h
				topp := botn - bh

				// Only the topmost piece has a rounded cap — those below are in the
				// 				// middle of the bar and must not look like separate bars.
				var d string
				if erOvst(m.slag, j) {
					d = barPath(x, topp, breidd, bh, bjelkeR)
				} else {
					d = fmt.Sprintf("M%.2f,%.2f H%.2f V%.2f H%.2f Z",
						x, topp, x+breidd, botn, x)
				}
				b.Bitar = append(b.Bitar, BarBit{
					D:      d,
					Klasse: SlagKlasse(slag),
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

// monthAbbrev gjev det stutte maanadsnamnet på det maalet sida stend i.
func monthAbbrev(lang string, m time.Month) string {
	return t(lang, fmt.Sprintf("maanad.kort_%d", int(m)))
}

// erOvst says whether the piece at position j is the last with anything in
// it — the only one that should have a rounded cap.
//
// Slagi gives both the order and the names; see slag.go. It stood here as
// slagRekkja, and the same four names stood in three other places.
func erOvst(slag map[string]int, j int) bool {
	for k := j + 1; k < len(Slagi); k++ {
		if slag[Slagi[k]] > 0 {
			return false
		}
	}
	return true
}

// weekLabel is the text you get hovering over a dot. The week, not the
// day: "the week from 3.2.2026 — 2 classes".
func weekLabel(lang string, måndag time.Time, tal int) string {
	dato := måndag.Format("2.1.2006")
	if tal == 0 {
		return fmt.Sprintf("%s %s — %s", t(lang, "aktivitet.veka_fraa"), dato, t(lang, "aktivitet.ingen"))
	}
	if tal == 1 {
		return fmt.Sprintf("%s %s — 1 %s", t(lang, "aktivitet.veka_fraa"), dato, t(lang, "aktivitet.time"))
	}
	return fmt.Sprintf("%s %s — %d %s", t(lang, "aktivitet.veka_fraa"), dato, tal, t(lang, "aktivitet.timar"))
}

// summaryText is the numbers said in words. Colour steps below 3:1 against
// the ground need a readable alternative; this is it, and it stands there
// for everyone and not only for whoever holds the pointer still.
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

// snittText says the average with a comma and one decimal, as numbers are
// written in Norwegian. %.1f gives a full stop, and a full stop in a number
// reads as a thousands separator here.
//
// It is a *tail* on the summary and not a sentence of its own. Below the bar
// row it lay half a picture away from what it describes, and the section got
// a line of text in the middle of it that broke the rhythm. At the top the
// two answer the same question — how much — and then they are one
// sentence.
func snittText(lang string, snitt float64) string {
	tal := strings.Replace(fmt.Sprintf("%.1f", snitt), ".", ",", 1)
	return fmt.Sprintf(t(lang, "aktivitet.snitt"), tal)
}

// barPath draws a bar with a rounded cap and a flat foot. Rounded at both
// ends it would float; its foot is the baseline, and a baseline is not
// rounded.
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
