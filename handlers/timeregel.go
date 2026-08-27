package handlers

import (
	"fmt"
	"sort"
	"time"

	"kjernekraft/models"
)

// Timeregel er sjølve klassen: «yoga med Leon, maandag 18:00». Timane
// er utslagi hans — dei ber regel-id-en sin i events-tabellen, og det
// er regelen administrasjonen endrar. Ein einskildtime er ein regel
// med eitt utslag; det einaste ein gjer med eit einskilt utslag er aa
// avlysa det.
type Timeregel struct {
	RegelID    int64
	Tittel     string
	Laerar     string
	Rom        string
	Vekedag    time.Weekday
	Klokke     string // «18:00»
	Lengd      int    // minutt
	Beskriving string
	Timar      []models.Event
}

// VekedagNykel gjev umsetjingsnykelen for vekedagen, so malen kann
// segja «Måndag» på det maalet brukaren hev valt (§11).
func (r Timeregel) VekedagNykel() string {
	return [...]string{
		"timeplan.sunday", "timeplan.monday", "timeplan.tuesday",
		"timeplan.wednesday", "timeplan.thursday", "timeplan.friday",
		"timeplan.saturday",
	}[r.Vekedag]
}

// GrupperTimar samlar timane under reglane sine. Regel-id-en paa timen
// avgjer; ein time utan (fraa fyre flyttingi, eller lagd inn utanum)
// vert kjend att paa det gamle samantreffet av like felt. Berre timar
// som ikkje er avslutta vert med; rekkjefylgdi er timeplanen si —
// maandag fyrst, so klokkeslett — og timane i kvar regel stend etter
// dato.
func GrupperTimar(timar []models.Event, no time.Time) []Timeregel {
	grupper := map[string]*Timeregel{}
	var rekkje []string

	// Klokka vert lesi som ho stend. Heile huset formaterer lagra tider
	// raatt — `StartTime.Format("15:04")`, aldri `.In()` — so den lagra
	// tidi *er* veggklokka. Ei umrekning her hadde synt «19:45» for ein
	// time heile timeplanen kallar «17:45».
	for _, t := range timar {
		if !t.EndTime.After(no) {
			continue
		}
		st := t.StartTime
		nykel := fmt.Sprintf("regel:%d", t.RuleID)
		if t.RuleID == 0 {
			nykel = fmt.Sprintf("%s|%s|%s|%d|%s|%d",
				t.Title, t.TeacherName, t.RoomName, st.Weekday(),
				st.Format("15:04"), int(t.EndTime.Sub(t.StartTime).Minutes()))
		}

		g, finst := grupper[nykel]
		if !finst {
			g = &Timeregel{RegelID: t.RuleID}
			grupper[nykel] = g
			rekkje = append(rekkje, nykel)
		}
		g.Timar = append(g.Timar, t)
	}

	reglar := make([]Timeregel, 0, len(rekkje))
	for _, nykel := range rekkje {
		g := grupper[nykel]
		sort.Slice(g.Timar, func(i, j int) bool {
			return g.Timar[i].StartTime.Before(g.Timar[j].StartTime)
		})
		// Regelen ser ut som den næraste komande timen sin. Er eldre
		// utslag annleis (ein vikar i fjor), er det historia — regelen
		// er det han er no.
		fyrste := g.Timar[0]
		st := fyrste.StartTime
		g.Tittel = fyrste.Title
		g.Laerar = fyrste.TeacherName
		g.Rom = fyrste.RoomName
		g.Vekedag = st.Weekday()
		g.Klokke = st.Format("15:04")
		g.Lengd = int(fyrste.EndTime.Sub(fyrste.StartTime).Minutes())
		g.Beskriving = fyrste.Description
		reglar = append(reglar, *g)
	}

	// Maandag fyrst, som i timeplanen — ikkje sundag, som time.Weekday
	// tel fraa.
	dag := func(w time.Weekday) int { return (int(w) + 6) % 7 }
	sort.SliceStable(reglar, func(i, j int) bool {
		if dag(reglar[i].Vekedag) != dag(reglar[j].Vekedag) {
			return dag(reglar[i].Vekedag) < dag(reglar[j].Vekedag)
		}
		if reglar[i].Klokke != reglar[j].Klokke {
			return reglar[i].Klokke < reglar[j].Klokke
		}
		return reglar[i].Tittel < reglar[j].Tittel
	})
	return reglar
}
