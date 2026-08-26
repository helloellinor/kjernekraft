package handlers

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"kjernekraft/models"
)

// Ein timeplan er stort sett den same vika um att. «Vinyasa Flow med
// Kristina i salen 07:15» er *éin* time som gjeng tysdag og torsdag —
// ikkje tvo ulike ting. Sett dei under kvarandre som tvo kort, og
// lesaren lyt lesa den same linja tvo gonger for aa oppdaga at det er
// den same.
//
// Difor vert vika klemd saman: éi rad per time, og eitt merke per dag
// han gjeng. Radi svarar paa spursmaalet ein faktisk stiller — «naar
// gjeng vinyasa?» — og merki svarar paa det neste, «kor full er han den
// dagen».
//
// For eit studio med same timeplan kvar veke fell fyrti kort ned til
// kring femten radar. Det er ikkje pynt; det er den same
// opplysningi paa ein tridjedel av rullingi.

// Framsyning er ein time slik han gjeng ein einskild dag.
type Framsyning struct {
	Event   models.Event
	Dag     time.Weekday
	Dato    time.Time
	DagNamn string
	ErIDag  bool
	ErUmme  bool
	Full    bool
	Prosent int
	Kolonne int // 1 = maandag … 7 = sundag

	// Klokka som vinklar, so merket kann teikna ei urskiva. Timevisaren
	// gjeng 30 grader i timen og ein halv grad i minuttet; minuttvisaren
	// seks grader i minuttet.
	Klokke       string
	TimeVinkel   float64
	MinuttVinkel float64

	// Merkelappen for skjermlesaren, ferdig umsett. Merket vert nytta
	// tvo stader, og korkje av deim skal trenga aa naa rota for aa
	// finna maalet.
	Merkelapp string
}

// Timebolk er ein time med alle dagarne han gjeng i vika.
type Timebolk struct {
	Tittel     string
	Laerar     string
	Rom        string
	Klasse     string
	Minutt     int
	Plassar    int
	Framsyning []Framsyning
	Rutor      [7]Ruta // alle sju dagarne, tome med
	// Urskiva for rada: den fyrste framsyningi i vika.
	TimeVinkel   float64
	MinuttVinkel float64
	sortering    int // minutt etter midnatt, til rekkjefylgdi
}

// Ruta er ein dag i rada — anten ein time som gjeng, eller eit tomt hol.
//
// Dei tome er med av ein grunn: ei rad med *sju* rutor der tvo er fylte
// syner kva som ikkje gjeng, ikkje berre kva som gjeng. Det er
// spursmaalet Ida stiller — kvar er holet — og det svaret finst ikkje
// naar dei tome dagarne er tome piksel.
type Ruta struct {
	Har        bool
	Framsyning Framsyning
	Dato       time.Time
	Kolonne    int
	ErIDag     bool
	ErUmme     bool
}

// dagKort er dei norske stuttformene. time.Weekday.String() gjev
// engelsk, og dette er ei norsk sida.
var dagKort = map[time.Weekday]string{
	time.Monday: "man", time.Tuesday: "tys", time.Wednesday: "ons",
	time.Thursday: "tor", time.Friday: "fre", time.Saturday: "lau",
	time.Sunday: "sun",
}

// rutor legg framsyningarne ut paa dei sju dagarne, og fyller resten
// med tome.
func rutor(b Timebolk, maandag time.Time) [7]Ruta {
	var ut [7]Ruta
	for i := 0; i < 7; i++ {
		d := maandag.AddDate(0, 0, i)
		ut[i] = Ruta{Dato: d, Kolonne: i + 1}
	}
	for _, f := range b.Framsyning {
		i := f.Kolonne - 1
		if i >= 0 && i < 7 {
			ut[i].Har = true
			ut[i].Framsyning = f
			ut[i].ErIDag = f.ErIDag
			ut[i].ErUmme = f.ErUmme
		}
	}
	return ut
}

// kolonne gjev vekedagen som 1–7 med maandag fyrst. Go sin Weekday
// byrjar paa sundag; ei norsk vika gjer ikkje det.
func kolonne(d time.Weekday) int {
	if d == time.Sunday {
		return 7
	}
	return int(d)
}

// NyFramsyning lagar framsyningi av ein time ein einskild dag.
//
// Han stend for seg sjølv av di merket vert nytta tvo stader — i vika og
// paa heimesida — og daa skal det vera *éin* stad som avgjer kva som
// stend i det.
func NyFramsyning(lang string, e models.Event, iDagDato string, naa time.Time) Framsyning {
	prosent := 0
	if e.Capacity > 0 {
		prosent = e.CurrentEnrolment * 100 / e.Capacity
	}
	dag := dagKort[e.StartTime.Weekday()]
	return Framsyning{
		Event:        e,
		Dag:          e.StartTime.Weekday(),
		Dato:         e.StartTime,
		DagNamn:      dag,
		ErIDag:       e.StartTime.Format("2006-01-02") == iDagDato,
		ErUmme:       e.EndTime.Before(naa),
		Full:         e.Full(),
		Prosent:      prosent,
		Kolonne:      kolonne(e.StartTime.Weekday()),
		Klokke:       e.StartTime.Format("15:04"),
		TimeVinkel:   float64(e.StartTime.Hour()%12)*30 + float64(e.StartTime.Minute())*0.5,
		MinuttVinkel: float64(e.StartTime.Minute()) * 6,
		Merkelapp: fmt.Sprintf("%s %s — %d %s %d",
			dag, e.StartTime.Format("15:04"),
			e.CurrentEnrolment, t(lang, "timeplan.of"), e.Capacity),
	}
}

// Framsyningar gjer ei liste med timar om til merke.
func Framsyningar(lang string, events []models.Event, naa time.Time) []Framsyning {
	iDagDato := naa.Format("2006-01-02")
	ut := make([]Framsyning, 0, len(events))
	for _, e := range events {
		ut = append(ut, NyFramsyning(lang, e, iDagDato, naa))
	}
	return ut
}

// KlemVika slær saman like timar i vika.
//
// Nykelen er alt som gjer ein time til *den same* timen: kva han heiter,
// kven som held honom, kvar han er, og kva klokkeslett han byrjar.
// Skifter noko av det, er det ein annan time.
func KlemVika(lang string, events []models.Event, iDag time.Time, maandag time.Time) []Timebolk {
	iDagDato := iDag.Format("2006-01-02")

	bolkar := map[string]*Timebolk{}
	for _, e := range events {
		nykel := fmt.Sprintf("%s|%s|%s", e.Title, e.TeacherName, e.RoomName)

		b, finst := bolkar[nykel]
		if !finst {
			b = &Timebolk{
				Tittel:    e.Title,
				Laerar:    e.TeacherName,
				Rom:       e.RoomName,
				Klasse:    e.ClassType,
				Minutt:    int(e.EndTime.Sub(e.StartTime).Minutes()),
				Plassar:   e.Capacity,
				sortering: e.StartTime.Hour()*60 + e.StartTime.Minute(),
			}
			bolkar[nykel] = b
		}

		b.Framsyning = append(b.Framsyning, NyFramsyning(lang, e, iDagDato, iDag))
	}

	ut := make([]Timebolk, 0, len(bolkar))
	for _, b := range bolkar {
		sort.Slice(b.Framsyning, func(i, j int) bool {
			return b.Framsyning[i].Dato.Before(b.Framsyning[j].Dato)
		})
		if len(b.Framsyning) > 0 {
			b.TimeVinkel = b.Framsyning[0].TimeVinkel
			b.MinuttVinkel = b.Framsyning[0].MinuttVinkel
		}
		b.Rutor = rutor(*b, maandag)
		ut = append(ut, *b)
	}

	// Etter klokka fyrst — ein timeplan vert lesen nedyver dagen — og so
	// etter namn, so rekkjefylgdi ikkje skiftar fraa lasting til lasting.
	sort.Slice(ut, func(i, j int) bool {
		if ut[i].sortering != ut[j].sortering {
			return ut[i].sortering < ut[j].sortering
		}
		return ut[i].Tittel < ut[j].Tittel
	})
	return ut
}

// Veke er eitt val i vikeveljaren.
type Veke struct {
	Nummer int
	Offset int
	Tittel string
	ErNo   bool
}

// vekeval gjev vikone kring den ein ser paa.
//
// Ein veljar med sju vikor og ikkje eit talfelt: ein veit kva veke det
// er i dag og kva veke ein vil til, men ein reknar ikkje ut vikenummer
// i hovudet. «Denne» og «neste» er dei tvo ein spør etter i praksis;
// resten er der for den som planlegg lenger fram.
func vekeval(lang string, naaVeke, naaOffset int) []Veke {
	var ut []Veke
	for d := -1; d <= 5; d++ {
		off := naaOffset + d
		if off < 0 {
			// Studioet syner ikkje vikor som er gjengne.
			continue
		}
		v := Veke{Nummer: naaVeke + d, Offset: off, ErNo: d == 0}
		switch off {
		case 0:
			v.Tittel = t(lang, "timeplan.this_week")
		case 1:
			v.Tittel = t(lang, "timeplan.next_week")
		default:
			v.Tittel = t(lang, "timeplan.week") + " " + strconv.Itoa(v.Nummer)
		}
		ut = append(ut, v)
	}
	return ut
}
