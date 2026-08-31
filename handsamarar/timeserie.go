package handsamarar

import (
	"fmt"
	"sort"
	"time"

	"kjernekraft/models"
)

// Timeserie is the class itself: "yoga with Leon, Monday 18:00". The
// classes are its occurrences — they carry its serie id in the events
// table, and it is the series the admin edits. A one-off is a series with
// one occurrence; the only thing you do with a single occurrence is cancel
// it.
type Timeserie struct {
	SerieID int64
	Tittel  string
	Lærar   string
	// Every name appearing on the classes in the run: the teacher first, then
	// the substitutes. Lærar is the one the series *has*; this is who
	// actually holds the classes, and it is longer than one name for the week
	// somebody stepped in.
	Laerarar []string
	Rom      string
	// Kva slag trening rekkja er — yoga, pilates, reformer. Han ber
	// vengefargen, og han er den same kvar gong rekkja gjeng.
	Slag       string
	Vekedag    time.Weekday
	Klokke     string // «18:00»
	Lengd      int    // minutt
	Beskriving string
	// The room the series runs in. The room select in the series sentence has
	// to know what it currently stands on, and the name alone will not do —
	// two rooms can share a name, and the select sends the id.
	RomID int
	// Gruppa serien er open for. Null er open for alle.
	GruppeID int
	// Plassar er serien si eigi kapasitet; `nil` tyder at han ikkje set
	// noko sjølv. RomPlassar er rommet sitt tal. Skjemaet syner det eine
	// som verdi og det andre som framlegg, so ein kann sjå um talet er
	// valt eller ervt.
	Plassar    *int
	RomPlassar int
	Timar      []models.Event
}

// Sluttar is the last coming class in the series. "The series runs eight
// times" does not say when it is over; the date does.
func (r Timeserie) Sluttar() time.Time {
	if len(r.Timar) == 0 {
		return time.Time{}
	}
	return r.Timar[len(r.Timar)-1].StartTime
}

// VekedagNykel gives the translation key for the weekday, so the template
// can say "Måndag" in the language the user chose (§11).
func (r Timeserie) VekedagNykel() string {
	return [...]string{
		"timeplan.sunday", "timeplan.monday", "timeplan.tuesday",
		"timeplan.wednesday", "timeplan.thursday", "timeplan.friday",
		"timeplan.saturday",
	}[r.Vekedag]
}

// VekedagFleirtalNykel gives the weekday in the plural. A run happening
// more than once happens on "maandagar", not on "maandag" — the list on the
// left is the runs and not their occurrences, so it has to speak of the day
// as it is: repeated.
func (r Timeserie) VekedagFleirtalNykel() string {
	return [...]string{
		"timeplan.sundays", "timeplan.mondays", "timeplan.tuesdays",
		"timeplan.wednesdays", "timeplan.thursdays", "timeplan.fridays",
		"timeplan.saturdays",
	}[r.Vekedag]
}

// GrupperTimar samlar timane under seriane sine. Serie-id-en på timen
// avgjer; ein time utan (frå fyre flyttingi, eller lagd inn utanum)
// vert kjend att på det gamle samantreffet av like felt. Berre timar
// som ikkje er avslutta vert med; rekkjefylgdi er timeplanen si —
// måndag fyrst, so klokkeslett — og timane i kvar serie stend etter
// dato.
func GrupperTimar(timar []models.Event, no time.Time) []Timeserie {
	grupper := map[string]*Timeserie{}
	var rekkje []string

	// Klokka vert lesi som ho stend. Heile huset formaterer lagra tider
	// raatt — `StartTime.Format("15:04")`, aldri `.In()` — so den lagra
	// tidi *er* veggklokka. Ei umrekning her hadde synt «19:45» for ein
	// time heile timeplanen kallar «17:45».
	for _, t := range timar {
		if !etter(t.EndTime, no) {
			continue
		}
		st := t.StartTime
		nykel := fmt.Sprintf("serie:%d", t.SerieID)
		if t.SerieID == 0 {
			nykel = fmt.Sprintf("%s|%s|%s|%d|%s|%d",
				t.Title, t.TeacherName, t.RoomName, st.Weekday(),
				st.Format("15:04"), int(t.EndTime.Sub(t.StartTime).Minutes()))
		}

		g, finst := grupper[nykel]
		if !finst {
			g = &Timeserie{SerieID: t.SerieID}
			grupper[nykel] = g
			rekkje = append(rekkje, nykel)
		}
		g.Timar = append(g.Timar, t)
	}

	seriar := make([]Timeserie, 0, len(rekkje))
	for _, nykel := range rekkje {
		g := grupper[nykel]
		sort.Slice(g.Timar, func(i, j int) bool {
			return g.Timar[i].StartTime.Before(g.Timar[j].StartTime)
		})
		// Serien ser ut som den næraste komande timen sin. Er eldre
		// utslag annleis (ein vikar i fjor), er det historia — serien
		// er det han er no.
		fyrste := g.Timar[0]
		st := fyrste.StartTime
		g.Tittel = fyrste.Title
		// The teacher is the exception to that rule, and it was a real loss of
		// 		// data.
		// 		//
		// 		// It stood as the nearest class's teacher, and the field in the series
		// 		// sentence carries it — the same field UpdateSerieTeacher writes to
		// 		// *every* class in the run. Give the nearest class a substitute, and
		// 		// the substitute was the series' teacher on the next load, and the
		// 		// next save — whatever it was for, the title, the room, the places —
		// 		// wrote the substitute's name across the whole run.
		// 		//
		// 		// A substitute is by definition the exception, so the name recurring
		// 		// most often is the teacher.
		g.Lærar, g.Laerarar = laerarane(g.Timar)
		g.Rom = fyrste.RoomName
		g.Slag = fyrste.ClassType
		g.Vekedag = st.Weekday()
		g.Klokke = st.Format("15:04")
		g.Lengd = int(fyrste.EndTime.Sub(fyrste.StartTime).Minutes())
		g.Beskriving = fyrste.Description
		g.RomID = fyrste.RoomID
		g.GruppeID = fyrste.GruppeID
		g.Plassar = fyrste.EigenPlassar
		g.RomPlassar = fyrste.RoomCapacity
		seriar = append(seriar, *g)
	}

	// Måndag fyrst, som i timeplanen — ikkje sundag, som time.Weekday
	// tel frå.
	dag := func(w time.Weekday) int { return (int(w) + 6) % 7 }
	sort.SliceStable(seriar, func(i, j int) bool {
		if dag(seriar[i].Vekedag) != dag(seriar[j].Vekedag) {
			return dag(seriar[i].Vekedag) < dag(seriar[j].Vekedag)
		}
		if seriar[i].Klokke != seriar[j].Klokke {
			return seriar[i].Klokke < seriar[j].Klokke
		}
		return seriar[i].Tittel < seriar[j].Tittel
	})
	return seriar
}

// Siktdag is a weekday the filter can pick. The number is a time.Weekday,
// so the row and the select compare the same thing.
type Siktdag struct {
	Tal   int
	Nykel string
}

// Siktval are the options the filter over the series list offers.
//
// They are computed from the series that *exist*, not from every teacher
// and every room the house has. A select with twenty teachers where three
// hold classes is not a filter, it is a catalogue: every choice you can
// make should give you something to look at. If there is only one option it
// filters nothing, and the template leaves the select out.
type Siktval struct {
	Dagar    []Siktdag
	Laerarar []string
}

// laerarane gjev læraren rekkja hev, og heile namnelista attaat.
//
// Ingi spyrjing: timane ligg alt her, so det som skal til er aa telja
// dei. Flest timar vinn; stend det likt, vinn den som kjem fyrst, og
// rekkja er sortert på dato — so det er den næraste timen.
//
// Namnelista byrjar med læraren og held fram med vikarane i den
// rekkjefylgda dei kjem: «Leon og Vikar» les seg som «Leon, og so ein
// vikar ei vika».
func laerarane(timar []models.Event) (string, []string) {
	tal := map[string]int{}
	var rekkje []string
	for _, t := range timar {
		namn := t.TeacherName
		if namn == "" {
			continue
		}
		if tal[namn] == 0 {
			rekkje = append(rekkje, namn)
		}
		tal[namn]++
	}
	if len(rekkje) == 0 {
		return "", nil
	}

	hovud := rekkje[0]
	for _, namn := range rekkje[1:] {
		if tal[namn] > tal[hovud] {
			hovud = namn
		}
	}

	ut := make([]string, 0, len(rekkje))
	ut = append(ut, hovud)
	for _, namn := range rekkje {
		if namn != hovud {
			ut = append(ut, namn)
		}
	}
	return hovud, ut
}

// SiktvalFor plukkar dei ulike dagane, lærarane og romi ut or seriane.
// Dagane stend i timeplanen si rekkjefylgd — måndag fyrst — og hine
// alfabetisk, av di ein leitar etter eit namn ein alt veit.
func SiktvalFor(seriar []Timeserie) Siktval {
	var val Siktval

	dagsett := map[time.Weekday]bool{}
	for _, r := range seriar {
		dagsett[r.Vekedag] = true
	}
	// Måndag fyrst, som i timeplanen og som i GrupperTimar.
	for i := 0; i < 7; i++ {
		d := time.Weekday((i + 1) % 7)
		if dagsett[d] {
			val.Dagar = append(val.Dagar, Siktdag{Tal: int(d), Nykel: Timeserie{Vekedag: d}.VekedagNykel()})
		}
	}

	val.Laerarar = ulike(seriar, func(r Timeserie) string { return r.Lærar })
	return val
}

// ulike gjev dei ulike ikkje-tome verdi, sorterte.
func ulike(seriar []Timeserie, av func(Timeserie) string) []string {
	sett := map[string]bool{}
	for _, r := range seriar {
		if v := av(r); v != "" {
			sett[v] = true
		}
	}
	ut := make([]string, 0, len(sett))
	for v := range sett {
		ut = append(ut, v)
	}
	sort.Strings(ut)
	return ut
}
