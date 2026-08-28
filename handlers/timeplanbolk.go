package handlers

import (
	"fmt"
	"log"
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

// Session er ein time slik han gjeng ein einskild dag.
type Session struct {
	Event   models.Event
	Day     time.Weekday
	Date    time.Time
	DayName string
	IsToday bool
	IsPast  bool
	Full    bool
	Percent int
	Column  int // 1 = maandag … 7 = sundag

	// Klokka som vinklar, so merket kann teikna ei urskiva. Timevisaren
	// gjeng 30 grader i timen og ein halv grad i minuttet; minuttvisaren
	// seks grader i minuttet.
	Clock       string
	HourAngle   float64
	MinuteAngle float64

	// Lengdi paa timen, i minutt. Ho høyrer til namnet — ho er den same
	// kva dag timen gjeng — og ho stend attmed honom i lista, som i
	// timeplanen.
	Minute int

	// Dagen som tvo bokstavar. Han stend øvst i merket, av di det er
	// det ein spør um fyrst.
	DayAbbrev string

	// Sjølve figuren, ferdig rekna. Sjaa merkeform.go.
	Form Mark

	// Merkelappen for skjermlesaren, ferdig umsett. Merket vert nytta
	// tvo stader, og korkje av deim skal trenga aa naa rota for aa
	// finna maalet.
	Label string
}

// ClassRow er ein time med alle dagarne han gjeng i vika.
type ClassRow struct {
	Title    string
	Teacher  string
	Room     string
	Class    string
	Minute   int
	Capacity int
	Sessions []Session
	Days     [7]DayCell // alle sju dagarne, tome med
	// Urskiva for rada: den fyrste framsyningi i vika.
	HourAngle   float64
	MinuteAngle float64
	sortKey     int // minutt etter midnatt, til rekkjefylgdi

	// AllPast segjer at kvar einaste gong denne rada gjeng i vika, er
	// gjengi. Rada er då ikkje eit tilbod lenger — ho er ei kvittering,
	// og ho skal ikkje stå i vegen for den fyrste timen ein *kan* gå på.
	AllPast bool
}

// PastRowCount tel radene som er heilt gjengne. Malen treng talet til
// knappen som hentar dei fram att.
func PastRowCount(bolkar []ClassRow) int {
	n := 0
	for _, b := range bolkar {
		if b.AllPast {
			n++
		}
	}
	return n
}

// DayCell er ein dag i rada — anten ein time som gjeng, eller eit tomt hol.
//
// Dei tome er med av ein grunn: ei rad med *sju* rutor der tvo er fylte
// syner kva som ikkje gjeng, ikkje berre kva som gjeng. Det er
// spursmaalet Ida stiller — kvar er holet — og det svaret finst ikkje
// naar dei tome dagarne er tome piksel.
type DayCell struct {
	Has     bool
	Session Session
	Date    time.Time
	Column  int
	IsToday bool
	IsPast  bool
}

// dagKort er dei norske stuttformene. time.Weekday.String() gjev
// engelsk, og dette er ei norsk sida.
var dagKort = map[time.Weekday]string{
	time.Monday: "man", time.Tuesday: "tys", time.Wednesday: "ons",
	time.Thursday: "tor", time.Friday: "fre", time.Saturday: "lau",
	time.Sunday: "sun",
}

// Heile dagen paa skiltet.
//
// Her stod tvo bokstavar — «MÅ», «TY» — av di fana var 44 av 56 einingar
// og tri ikkje fekk plass. Men «TY» er ikkje eit ord; det er ein kode ein
// lyt læra, og eit merke som lyt lærast er eit merke som ikkje segjer
// noko. Fana spenner yver toppen no (52 av 56, sjaa merkeform.go), og
// daa stend dagen heil.
var dagKort2 = map[time.Weekday]string{
	time.Monday: "MÅNDAG", time.Tuesday: "TYSDAG", time.Wednesday: "ONSDAG",
	time.Thursday: "TORSDAG", time.Friday: "FREDAG", time.Saturday: "LAURDAG",
	time.Sunday: "SUNDAG",
}

// dayCells legg framsyningarne ut paa dei sju dagarne, og fyller resten
// med tome.
func dayCells(b ClassRow, maandag time.Time) [7]DayCell {
	var ut [7]DayCell
	for i := 0; i < 7; i++ {
		d := maandag.AddDate(0, 0, i)
		ut[i] = DayCell{Date: d, Column: i + 1}
	}
	for _, f := range b.Sessions {
		i := f.Column - 1
		if i >= 0 && i < 7 {
			ut[i].Has = true
			ut[i].Session = f
			ut[i].IsToday = f.IsToday
			ut[i].IsPast = f.IsPast
		}
	}
	return ut
}

// columnOf gjev vekedagen som 1–7 med maandag fyrst. Go sin Weekday
// byrjar paa sundag; ei norsk vika gjer ikkje det.
func columnOf(d time.Weekday) int {
	if d == time.Sunday {
		return 7
	}
	return int(d)
}

// NewSession lagar framsyningi av ein time ein einskild dag.
//
// Han stend for seg sjølv av di merket vert nytta tvo stader — i vika og
// paa heimesida — og daa skal det vera *éin* stad som avgjer kva som
// stend i det.
func NewSession(lang string, e models.Event, iDagDato string, naa time.Time) Session {
	prosent := 0
	if e.Capacity > 0 {
		prosent = e.CurrentEnrolment * 100 / e.Capacity
	}
	dag := dagKort[e.StartTime.Weekday()]
	return Session{
		Event:       e,
		Day:         e.StartTime.Weekday(),
		Date:        e.StartTime,
		DayName:     dag,
		IsToday:     e.StartTime.Format("2006-01-02") == iDagDato,
		IsPast:      fyre(e.EndTime, naa),
		Full:        e.Full(),
		Percent:     prosent,
		Column:      columnOf(e.StartTime.Weekday()),
		Clock:       e.StartTime.Format("15:04"),
		DayAbbrev:   dagKort2[e.StartTime.Weekday()],
		Minute:      int(e.EndTime.Sub(e.StartTime).Minutes()),
		HourAngle:   float64(e.StartTime.Hour()%12)*30 + float64(e.StartTime.Minute())*0.5,
		MinuteAngle: float64(e.StartTime.Minute()) * 6,
		Form: NewMark(fmt.Sprintf("%d-%s", e.ID, e.StartTime.Format("0102")),
			e.StartTime, e.EndTime, e.CurrentEnrolment, e.Capacity),
		Label: fmt.Sprintf("%s %s — %d %s %d",
			dag, e.StartTime.Format("15:04"),
			e.CurrentEnrolment, t(lang, "timeplan.of"), e.Capacity),
	}
}

// BuildSessions gjer ei liste med timar om til merke.
func BuildSessions(lang string, events []models.Event, naa time.Time) []Session {
	iDagDato := naa.Format("2006-01-02")
	ut := make([]Session, 0, len(events))
	for _, e := range events {
		ut = append(ut, NewSession(lang, e, iDagDato, naa))
	}
	return ut
}

// BuildWeekRows slær saman like timar i vika.
//
// Nykelen er alt som gjer ein time til *den same* timen: kva han heiter,
// kven som held honom, kvar han er, og kva klokkeslett han byrjar.
// Skifter noko av det, er det ein annan time.
func BuildWeekRows(lang string, events []models.Event, iDag time.Time, maandag time.Time) []ClassRow {
	iDagDato := iDag.Format("2006-01-02")

	bolkar := map[string]*ClassRow{}
	for _, e := range events {
		// Ein time som er gjengen er ikkje eit tilbod, og vika er ei
		// lista ein skannar etter det ein *kann* gaa paa. Han vart dimd
		// fyrr — halv styrke, men framleis ein plass i rada og framleis
		// noko auga lyt sortera burt. No vert han ikkje teikna.
		//
		// Dette skjer her og ikkje i malen: er han ikkje eit tilbod,
		// skal han korkje telja med i rada, i dagruta eller i tale.
		f := NewSession(lang, e, iDagDato, iDag)
		if f.IsPast {
			continue
		}

		nykel := fmt.Sprintf("%s|%s|%s", e.Title, e.TeacherName, e.RoomName)

		b, finst := bolkar[nykel]
		if !finst {
			b = &ClassRow{
				Title:    e.Title,
				Teacher:  e.TeacherName,
				Room:     e.RoomName,
				Class:    e.ClassType,
				Minute:   int(e.EndTime.Sub(e.StartTime).Minutes()),
				Capacity: e.Capacity,
				sortKey:  e.StartTime.Hour()*60 + e.StartTime.Minute(),
			}
			bolkar[nykel] = b
		}

		b.Sessions = append(b.Sessions, f)
	}

	ut := make([]ClassRow, 0, len(bolkar))
	for _, b := range bolkar {
		sort.Slice(b.Sessions, func(i, j int) bool {
			return b.Sessions[i].Date.Before(b.Sessions[j].Date)
		})
		if len(b.Sessions) > 0 {
			b.HourAngle = b.Sessions[0].HourAngle
			b.MinuteAngle = b.Sessions[0].MinuteAngle
			// Heilt gjengi berre naar *alle* gongene er det. Ei rad med
			// måndag attum seg og fredag framfyre er framleis eit
			// tilbod, og ho skal stå.
			b.AllPast = true
			for _, f := range b.Sessions {
				if !f.IsPast {
					b.AllPast = false
					break
				}
			}
		}
		b.Days = dayCells(*b, maandag)
		ut = append(ut, *b)
	}

	// Etter klokka fyrst — ein timeplan vert lesen nedyver dagen — og so
	// etter namn, so rekkjefylgdi ikkje skiftar fraa lasting til lasting.
	sort.Slice(ut, func(i, j int) bool {
		if ut[i].sortKey != ut[j].sortKey {
			return ut[i].sortKey < ut[j].sortKey
		}
		return ut[i].Title < ut[j].Title
	})
	return ut
}

// Week er eitt val i vikeveljaren.
type Week struct {
	Number    int
	Offset    int
	Title     string
	IsCurrent bool
}

// weekOptions gjev vikone kring den ein ser paa.
//
// Ein veljar med sju vikor og ikkje eit talfelt: ein veit kva veke det
// er i dag og kva veke ein vil til, men ein reknar ikkje ut vikenummer
// i hovudet. «Denne» og «neste» er dei tvo ein spør etter i praksis;
// resten er der for den som planlegg lenger fram.
func weekOptions(lang string, naaVeke, naaOffset int) []Week {
	var ut []Week
	for d := -1; d <= 5; d++ {
		off := naaOffset + d
		if off < 0 {
			// Studioet syner ikkje vikor som er gjengne.
			continue
		}
		v := Week{Number: naaVeke + d, Offset: off, IsCurrent: d == 0}
		switch off {
		case 0:
			v.Title = t(lang, "timeplan.this_week")
		case 1:
			v.Title = t(lang, "timeplan.next_week")
		default:
			v.Title = t(lang, "timeplan.week") + " " + strconv.Itoa(v.Number)
		}
		ut = append(ut, v)
	}
	return ut
}

// ---- listone paa heimesida ----
//
// Baade sida og brotstykki som hentar seg att teiknar dei same tvo
// listone. Dei stod i sidehandsamaren fyrr; her stend dei ein gong, so
// eit brotstykke aldri kann syna noko anna enn sida gjorde.

// EnrolledSessions er timane du hev meldt deg paa og som ikkje er
// gjengne. Dei er paamelde per definisjon, so merket veit det og dokka
// kann tilby avmelding med det same.
func EnrolledSessions(userID int64, lang string, naa time.Time) ([]Session, error) {
	komande, err := DB.GetUserUpcomingSignups(userID)
	if err != nil {
		return nil, err
	}
	for i := range komande {
		komande[i].IsUserSignedUp = true
	}
	return BuildSessions(lang, komande, naa), nil
}

// AvailableSessions er timane med ledig plass i dag og i morgon.
//
// Dei ber om du er paameld fraa fyrr. Utan det sa merket «meld deg paa»
// um ein time du alt stod paa, og dokka hadde bode deg det same ein
// gong til.
func AvailableSessions(userID int64, lang string, naa time.Time) ([]Session, error) {
	ledige, err := DB.LedigeTimar(userID)
	if err != nil {
		return nil, err
	}

	komande := make([]models.Event, 0, len(ledige))
	for _, e := range ledige {
		if etter(e.StartTime, naa) {
			komande = append(komande, e)
		}
	}
	if len(komande) == 0 {
		return nil, nil
	}

	ider := make([]int64, len(komande))
	for i, e := range komande {
		ider[i] = int64(e.ID)
	}
	// Timane du alt stend paa høyrer ikkje heime her. Dei stend i
	// «Paameld» rett yver, og ei lista som er eit *tilbod* skal ikkje
	// tilby deg det du hev teke. Det gjer paameldingi synleg med: timen
	// flyt fraa den eine lista til den andre.
	if paamelde, err := DB.GetUserSignupsForEvents(userID, ider); err != nil {
		log.Printf("paameldingar for %d: %v", userID, err)
	} else {
		att := komande[:0]
		for _, e := range komande {
			if !paamelde[int64(e.ID)] {
				att = append(att, e)
			}
		}
		komande = att
	}

	return BuildSessions(lang, komande, naa), nil
}
