package handsamarar

import (
	"fmt"
	"kjernekraft/handsamarar/config"
	"strings"
	"time"
	"unicode"

	"kjernekraft/models"
)

// First name. "See you Thursday morning, Anna Larsen!" is not a
// greeting, it is a summons.
func fyrenamnet(namn string) string {
	if i := strings.IndexByte(namn, ' '); i > 0 {
		return namn[:i]
	}
	return namn
}

// HelsingTittel is the page title: "See you tomorrow morning, Anna".
//
// Deliberately vague about when — the exact time belongs in the briefing
// below, and a title carrying a clock reads as a summons. An empty
// `naar` means nothing is coming up, so the title cannot promise a
// reunion and falls back to a plain greeting.
func HelsingTittel(lang, namn, når string, nå time.Time, nettFerdig bool) string {
	// Just out of a class: what she is wondering is not when the next
	// one is — she is standing in the changing room. This comes first
	// because it is the freshest thing that happened.
	if nettFerdig {
		return fmt.Sprintf(t(lang, "greeting.takk_for_no"), fyrenamnet(namn))
	}
	if når == "" {
		// Nothing ahead, so the title cannot promise a reunion. The
		// clock still knows whether it is morning or night, and a
		// greeting that knows the time of day is the one you get at the
		// door.
		return fmt.Sprintf(t(lang, "greeting.hi_fmt"), doegnhelsing(lang, nå), fyrenamnet(namn))
	}
	return fmt.Sprintf(t(lang, "greeting.title_see_you"), når, fyrenamnet(namn))
}

// doegnhelsing gives the greeting for the hour: morning, day, evening,
// night. The boundaries are where a person changes words, not where the
// clock does.
func doegnhelsing(lang string, nå time.Time) string {
	switch h := nå.In(OsloLoc).Hour(); {
	case h < 5:
		return t(lang, "greeting.god_natt")
	case h < 10:
		return t(lang, "greeting.god_morgon")
	case h < 17:
		return t(lang, "greeting.god_dag")
	case h < 20:
		return t(lang, "greeting.god_eftan")
	case h < 23:
		return t(lang, "greeting.god_kveld")
	default:
		return t(lang, "greeting.god_natt")
	}
}

// HelsingNår gives the vague when the title carries: "tomorrow
// morning", "Tuesday evening", "2 September". Empty when nothing is
// left.
func HelsingNår(lang string, neste *models.Event, nå time.Time) string {
	if neste == nil {
		return ""
	}

	// The stored time is the clock on the wall; rebuild it, do not
	// convert it. See tid.go — a conversion here made the greeting run
	// two hours late.
	start := veggklokka(neste.StartTime)
	nå = nå.In(OsloLoc)

	iDag := nå.Format("2006-01-02")
	iMorgon := nå.AddDate(0, 0, 1).Format("2006-01-02")
	dagen := start.Format("2006-01-02")

	switch {
	// The last quarter hour. The question is no longer when but where,
	// and every class is on the second floor.
	case start.Sub(nå) <= 15*time.Minute:
		return t(lang, "greeting.andre_hogd")
	case dagen == iDag:
		// Today needs the hour, not the day — but the hour alone read as
		// "See you 18:00, Anna", which is not a sentence, so it carries
		// the word that makes it one.
		return fmt.Sprintf(t(lang, "greeting.today_at"), start.Format("15:04"))
	case dagen == iMorgon:
		// "tomorrow morning" needs its own phrasing in Norwegian; the
		// other parts of the day follow "tomorrow" plainly.
		if start.Hour() < 10 {
			return t(lang, "greeting.tomorrow_early")
		}
		return t(lang, "greeting.tomorrow") + " " + tidbolk(lang, start)
	case start.Sub(nå) < 7*24*time.Hour:
		return vekedag(lang, start) + " " + tidbolk(lang, start)
	default:
		// Further out than a week: the date matters more than the hour.
		return fmt.Sprintf(t(lang, "greeting.date_fmt"), start.Day(), maanad(lang, start))
	}
}

// NestePresis gives the exact when the briefing carries: "today,
// 18:00". The title is vague because it is a greeting; this is precise
// because it is an answer. It always carries the clock, even when the
// title already said "tomorrow".
func NestePresis(lang string, neste *models.Event, nå time.Time) string {
	if neste == nil {
		return ""
	}
	start := veggklokka(neste.StartTime)
	nå = nå.In(OsloLoc)

	iDag := nå.Format("2006-01-02")
	iMorgon := nå.AddDate(0, 0, 1).Format("2006-01-02")
	dagen := start.Format("2006-01-02")

	var dag string
	switch {
	case dagen == iDag:
		dag = t(lang, "greeting.today")
	case dagen == iMorgon:
		dag = t(lang, "greeting.tomorrow")
	case start.Sub(nå) < 7*24*time.Hour:
		dag = vekedag(lang, start)
	default:
		dag = fmt.Sprintf(t(lang, "greeting.date_fmt"), start.Day(), maanad(lang, start))
	}
	return dag + ", " + start.Format("15:04")
}

// Briefing is the sentence under the title, split into the parts the
// template joins. The logic lives here rather than in the template
// because this is where it can be tested — it is the assembly that can
// read oddly, not the words.
//
// Three links, any of which can fall away:
//
//	"Your next class is Tuesday, 18:00."      whenever something is left
//	"You are signed up for 3 classes"         only when more than one
//	" and have 30 clips to use."              only when clips exist
//
// Drop the middle one and the last cannot open with "and" — the
// conjunction would dangle. Hence two forms of the clip link, and hence
// Go choosing the key: the template cannot see what precedes it.
type Briefing struct {
	Når        string // «tysdag, 18:00»; tom når ingen time stend att
	Med        string // læraren; tom når timen ikkje hev nokon
	Stad       string // rommet; tom når det ikkje er sett
	VekeTal    int
	VekeNykel  string // tom = lekken vert ikkje sagd
	KlippTal   int
	KlippNykel string // tom = lekken vert ikkje sagd
	// The "no classes, but clips" sentence. Whole on its own, and it
	// replaces both the empty link and the clip link.
	TomNykel string
}

// NyBriefing picks the links and their forms.
//
// iVeka counts classes this week, the next one included. One class is
// the one already named in the sentence before, so it is not repeated;
// only from two is it news.
func NyBriefing(lang string, neste *models.Event, nå time.Time, iVeka, klipp int) Briefing {
	b := Briefing{Når: NestePresis(lang, neste, nå)}

	// Who and where. "Your next class is tomorrow, 07:45" says when but
	// not what you are going to: a class is a teacher in a room. Both are
	// optional; empty, the link falls away and the sentence still
	// stands.
	if neste != nil {
		b.Med = neste.TeacherName
		b.Stad = neste.RoomName
		if b.Stad == "" {
			b.Stad = neste.Location
		}
	}

	// With no class ahead there is no week to count in. The rule must
	// live here, not in the template: left to the template, Go still
	// picked the "and have …" form and the sentence became "You have no
	// classes ahead. and have 30 clips to use." The conjunction pointed
	// at something never drawn.
	if b.Når == "" {
		iVeka = 0
	}

	if iVeka >= 2 {
		b.VekeTal = iVeka
		b.VekeNykel = "greeting.week_count_many"
	}

	// No classes but clips on the book is not two sentences side by
	// side. It is one sentence with a "but" in it, and the "but" is the
	// whole difference: the first half is a lack, the second is a way
	// out of it.
	if b.Når == "" && klipp > 0 {
		b.KlippTal = klipp
		if klipp == 1 {
			b.TomNykel = "greeting.ingen_timar_men_eitt"
		} else {
			b.TomNykel = "greeting.ingen_timar_men_klipp"
		}
		return b
	}

	if klipp > 0 {
		b.KlippTal = klipp
		eitt := klipp == 1
		switch {
		case b.VekeNykel != "" && eitt:
			b.KlippNykel = "greeting.klipp_join_one"
		case b.VekeNykel != "":
			b.KlippNykel = "greeting.klipp_join_many"
		case eitt:
			b.KlippNykel = "greeting.klipp_solo_one"
		default:
			b.KlippNykel = "greeting.klipp_solo_many"
		}
	}
	return b
}

// vekedag and maanad give the name in the language the user chose.
// They were two fixed Norwegian tables once, so the English page said
// "tysdag".
//
// Case is data, not code: English capitalises weekdays mid-sentence,
// Norwegian does not.
func vekedag(lang string, t0 time.Time) string {
	nykel := [...]string{
		"timeplan.sunday", "timeplan.monday", "timeplan.tuesday",
		"timeplan.wednesday", "timeplan.thursday", "timeplan.friday",
		"timeplan.saturday",
	}[t0.Weekday()]
	return iSetning(lang, t(lang, nykel))
}

func maanad(lang string, t0 time.Time) string {
	nykel := [...]string{
		"month.jan", "month.feb", "month.mar", "month.apr",
		"month.may", "month.jun", "month.jul", "month.aug",
		"month.sep", "month.oct", "month.nov", "month.dec",
	}[int(t0.Month())-1]
	return t(lang, nykel)
}

// iSetning puts a word in the case the language uses mid-sentence.
func iSetning(lang, ord string) string {
	if ord == "" || t(lang, "greeting.day_case") == "title" {
		return ord
	}
	r := []rune(ord)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// tidbolk gives morning, forenoon, afternoon or evening. You say "see
// you Thursday morning", not "Thursday 07:15".
func tidbolk(lang string, t0 time.Time) string {
	switch h := t0.Hour(); {
	case h < 10:
		return t(lang, "greeting.morning")
	case h < 12:
		return t(lang, "greeting.forenoon")
	case h < 17:
		return t(lang, "greeting.afternoon")
	default:
		return t(lang, "greeting.evening")
	}
}

// Kvalifisert says whether the user gets to see the student and senior
// rates. Senior follows from the birth date the system already has;
// nobody should tick a box to say they turned 67. A student card is
// something you show at the desk.
func Kvalifisert(u *models.User) bool {
	if u == nil {
		return false
	}
	return u.KvalifisertFor(config.GetInstance().GetCurrentTime())
}

// Heimehovudet computes the two parts at the top of the home page: the
// greeting and the briefing.
//
// It exists once because they are drawn twice — when the page loads, and
// again when you sign up for a class and the block refetches itself. Two
// copies could disagree, and that would not show until someone signed up.
//
// It fetches nothing: signups, clips and the check-in are already in
// hand at the call site.
func Heimehovudet(lang, namn string, paamelde []Session, klippAtt int, nettFerdig bool, nå time.Time) (string, Briefing) {
	// The first class signed up for. It carries the greeting.
	var neste *models.Event
	if len(paamelde) > 0 {
		neste = &paamelde[0].Event
	}

	// How many classes fall in *this* week. The list holds everything
	// coming, so it must be cut at the week boundary: "signed up for 4
	// classes this week" is untrue if three of them are next month.
	//
	// Wall clock, not conversion — same reason as in the greeting.
	måndag := VikeMåndag(nå, 0)
	nesteMåndag := måndag.AddDate(0, 0, 7)
	iVeka := 0
	for _, s := range paamelde {
		d := veggklokka(s.Event.StartTime)
		if !d.Before(måndag) && d.Before(nesteMåndag) {
			iVeka++
		}
	}

	når := HelsingNår(lang, neste, nå)
	return HelsingTittel(lang, namn, når, nå, nettFerdig),
		NyBriefing(lang, neste, nå, iVeka, klippAtt)
}
