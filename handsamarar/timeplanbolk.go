package handsamarar

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"kjernekraft/models"
)

// A schedule is mostly the same week again. "Vinyasa Flow with Kristina
// in the hall at 07:15" is one class running Tuesday and Thursday, not
// two different things — stack them as two cards and the reader has to
// read the same line twice to notice.
//
// So the week is folded: one row per class, one mark per day it runs.
// The row answers the question you actually ask — when is vinyasa? — and
// the marks answer the next one, how full is it that day.
//
// For a studio with the same schedule each week, forty cards fall to
// about fifteen rows.

// Session is a class as it runs on one particular day.
type Session struct {
	Event   models.Event
	Day     time.Weekday
	Date    time.Time
	DayName string
	IsToday bool
	IsPast  bool
	Full    bool
	Percent int
	Column  int // 1 = måndag … 7 = sundag

	// The time as angles, so the mark can draw a dial. The hour hand
	// moves 30° per hour and half a degree per minute; the minute hand
	// six degrees per minute.
	Clock       string
	HourAngle   float64
	MinuteAngle float64

	// Length in minutes. It belongs to the name — the same whatever day
	// the class runs — and stands beside it in the list.
	Minute int

	// The day as two letters, at the top of the mark: it is what you ask
	// first.
	DayAbbrev string
	// Day and date in full — "fredag 28. august" — for the dock. The mark
	// does not carry .Lang, so the translation has to happen where lang
	// exists, which is here.
	DatoTekst string

	// The figure itself, already computed. See merkeform.go.
	Form Mark

	// The screen-reader label, already translated. The mark is used in
	// two places and neither should have to reach the root for the
	// language.
	Label string
}

// ClassRow is one class with every day it runs that week.
type ClassRow struct {
	Title    string
	Teacher  string
	Room     string
	Class    string
	Minute   int
	Capacity int
	Sessions []Session
	Days     [7]DayCell // alle sju dagarne, tome med
	// The dial for the row: the week's first occurrence.
	HourAngle   float64
	MinuteAngle float64
	sortKey     int // minutt etter midnatt, til rekkjefylgdi

	// AllPast means every occurrence this week has been. The row is no
	// longer an offer, it is a receipt, and it must not stand in the way
	// of the first class you can still attend.
	AllPast bool
}

// PastRowCount counts the rows entirely in the past. The template needs
// the number for the button that brings them back.
func PastRowCount(bolkar []ClassRow) int {
	n := 0
	for _, b := range bolkar {
		if b.AllPast {
			n++
		}
	}
	return n
}

// DayCell is one day in the row — either a class or an empty hole.
//
// The empty ones are there on purpose: seven cells with two filled shows
// what does *not* run, not only what does. That is the question being
// asked — where is the gap — and it has no answer when the empty days
// are empty pixels.
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

// vekedagNykel and maanadNykel give the translation key for a day and a
// month. The same keys the template functions use — it is one truth, and
// it must not stand in two places under two names.
func vekedagNykel(d time.Weekday) string {
	return [...]string{
		"timeplan.sunday", "timeplan.monday", "timeplan.tuesday",
		"timeplan.wednesday", "timeplan.thursday", "timeplan.friday",
		"timeplan.saturday",
	}[d]
}

func maanadNykel(m time.Month) string {
	return [...]string{
		"", "timeplan.month_january", "timeplan.month_february",
		"timeplan.month_march", "timeplan.month_april", "timeplan.month_may",
		"timeplan.month_june", "timeplan.month_july", "timeplan.month_august",
		"timeplan.month_september", "timeplan.month_october",
		"timeplan.month_november", "timeplan.month_december",
	}[int(m)]
}

// The whole day on the sign.
//
// It was two letters — "MÅ", "TY" — because the tab was 44 of 56 units and
// three did not fit. But "TY" is not a word, it is a code you have to
// learn, and a mark that has to be learnt is a mark that says nothing. The
// tab spans the top now (52 of 56, see merkeform.go), so the day stands
// whole.
//
// It was a fixed table of Nynorsk words, which meant the clock said
// "MÅNDAG" even when the rest of the page was in Bokmål or English — and
// the mark is the thing that appears most often on a page, so it was the
// most visible place in the house to forget the translation.
//
// The capitals are applied here and not in the stylesheet: .merkedag
// carries no text-transform, and the word should be uppercase in all three
// languages.
//
// datoIKlartekst gives "fredag 28. august" in the page's language.
func datoIKlartekst(lang string, tid time.Time) string {
	return fmt.Sprintf("%s %d. %s",
		t(lang, vekedagNykel(tid.Weekday())), tid.Day(), t(lang, maanadNykel(tid.Month())))
}

func dagPåSkiltet(lang string, d time.Weekday) string {
	return strings.ToUpper(t(lang, vekedagNykel(d)))
}

// dayCells legg framsyningarne ut på dei sju dagarne, og fyller resten
// med tome.
func dayCells(b ClassRow, måndag time.Time) [7]DayCell {
	var ut [7]DayCell
	for i := 0; i < 7; i++ {
		d := måndag.AddDate(0, 0, i)
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

// columnOf gjev vekedagen som 1–7 med måndag fyrst. Go sin Weekday
// byrjar på sundag; ei norsk vika gjer ikkje det.
func columnOf(d time.Weekday) int {
	if d == time.Sunday {
		return 7
	}
	return int(d)
}

// NewSession builds the presentation of a class on one particular day.
//
// It stands on its own because the mark is used in two places — the week
// and the home page — and there should be *one* place deciding what is in
// it.
func NewSession(lang string, e models.Event, iDagDato string, nå time.Time) Session {
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
		IsPast:      fyre(e.EndTime, nå),
		Full:        e.Full(),
		Percent:     prosent,
		Column:      columnOf(e.StartTime.Weekday()),
		Clock:       e.StartTime.Format("15:04"),
		DayAbbrev:   dagPåSkiltet(lang, e.StartTime.Weekday()),
		DatoTekst:   datoIKlartekst(lang, e.StartTime),
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
func BuildSessions(lang string, events []models.Event, nå time.Time) []Session {
	iDagDato := nå.Format("2006-01-02")
	ut := make([]Session, 0, len(events))
	for _, e := range events {
		ut = append(ut, NewSession(lang, e, iDagDato, nå))
	}
	return ut
}

// BuildWeekRows folds identical classes in the week together.
//
// The key is everything that makes a class *the same* class: what it is
// called, who holds it, where it is, and what time it starts. Change any
// of that and it is a different class.
func BuildWeekRows(lang string, events []models.Event, iDag time.Time, måndag time.Time) []ClassRow {
	iDagDato := iDag.Format("2006-01-02")

	bolkar := map[string]*ClassRow{}
	for _, e := range events {
		// A class that has been is not an offer, and the week is a list you scan
		// for what you *can* attend. It used to be dimmed — half strength, but
		// still a place in the row and still something the eye has to sort away.
		// Now it is not drawn.
		//
		// This happens here and not in the template: if it is not an offer it
		// should count neither in the row, nor in the day cell, nor in the
		// number.
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
			// Entirely past only when *every* occurrence is. A row with Monday behind
			// it and Friday ahead is still an offer, and should stand.
			b.AllPast = true
			for _, f := range b.Sessions {
				if !f.IsPast {
					b.AllPast = false
					break
				}
			}
		}
		b.Days = dayCells(*b, måndag)
		ut = append(ut, *b)
	}

	// Etter klokka fyrst — ein timeplan vert lesen nedyver dagen — og so
	// etter namn, so rekkjefylgdi ikkje skiftar frå lasting til lasting.
	//
	// Namn åleine var ikkje nok. Radene er bolka på «tittel|lærar|rom»,
	// so tri «Reformer» klokka 17 med kvar sin lærar er tri rader med
	// same klokka *og* same tittelen. Då hadde samanlikningi ikkje meir
	// å gå på, og rekkjefylgdi fall attende på den basen gav oss — som
	// heller ikkje var avgjord (sjå ORDER BY i GetEventsForWeek). Ti
	// henteringar av den same sida gav fem ulike rekkjefylgder.
	//
	// Difor heile bolknykelen: tittel, lærar, rom. Det er ei *fullstendig*
	// ordning på det som skil ei rad frå ei onnor, so utfallet er det
	// same kva rekkjefylgd radene kom inn i.
	sort.Slice(ut, func(i, j int) bool {
		if ut[i].sortKey != ut[j].sortKey {
			return ut[i].sortKey < ut[j].sortKey
		}
		if ut[i].Title != ut[j].Title {
			return ut[i].Title < ut[j].Title
		}
		if ut[i].Teacher != ut[j].Teacher {
			return ut[i].Teacher < ut[j].Teacher
		}
		return ut[i].Room < ut[j].Room
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

// weekOptions gives the weeks around the one being viewed.
//
// A picker with seven weeks rather than a number field: you know what week
// it is today and what week you want, but you do not compute week numbers
// in your head. "This" and "next" are the two you actually ask for.
func weekOptions(lang string, nåVeke, nåOffset int) []Week {
	var ut []Week
	for d := -1; d <= 5; d++ {
		off := nåOffset + d
		if off < 0 {
			// Studioet syner ikkje vikor som er gjengne.
			continue
		}
		v := Week{Number: nåVeke + d, Offset: off, IsCurrent: d == 0}
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

// ---- the lists on the home page ----
//
// Both the page and the fragments that refetch themselves draw the same
// two lists. They used to live in the page handler; here they stand once,
// so a fragment can never show something different from what the page
// did.

// EnrolledSessions are the classes you have signed up for that have not
// been. They are signed-up by definition, so the mark knows it and the
// dock can offer cancellation right away.
func (a *App) EnrolledSessions(userID int64, lang string, nå time.Time) ([]Session, error) {
	komande, err := a.DB.GetUserUpcomingSignups(userID)
	if err != nil {
		return nil, err
	}
	for i := range komande {
		komande[i].IsUserSignedUp = true
	}
	return BuildSessions(lang, komande, nå), nil
}

// AvailableSessions are the classes with room today and tomorrow.
//
// They carry whether you are already signed up. Without that the mark said
// "sign up" for a class you were already on, and the dock offered it to
// you a second time.
func (a *App) AvailableSessions(userID int64, lang string, nå time.Time) ([]Session, error) {
	ledige, err := a.DB.LedigeTimar(userID, nå)
	if err != nil {
		return nil, err
	}

	komande := make([]models.Event, 0, len(ledige))
	for _, e := range ledige {
		if etter(e.StartTime, nå) {
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
	// Classes you are already on do not belong here. They stand in "Signed
	// up" just above, and a list that is an *offer* should not offer you what
	// you have taken. It also makes the signup visible: the class flows from
	// one list to the other.
	if paamelde, err := a.DB.GetUserSignupsForEvents(userID, ider); err != nil {
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

	return BuildSessions(lang, komande, nå), nil
}
