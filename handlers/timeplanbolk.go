package handlers

import (
	"fmt"
	"sort"
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
}

// Timebolk er ein time med alle dagarne han gjeng i vika.
type Timebolk struct {
	Tittel     string
	Laerar     string
	Rom        string
	Klasse     string
	Start      string // «07:15»
	Minutt     int
	Plassar    int
	Framsyning []Framsyning
	sortering  int // minutt etter midnatt, til rekkjefylgdi
}

// dagKort er dei norske stuttformene. time.Weekday.String() gjev
// engelsk, og dette er ei norsk sida.
var dagKort = map[time.Weekday]string{
	time.Monday: "man", time.Tuesday: "tys", time.Wednesday: "ons",
	time.Thursday: "tor", time.Friday: "fre", time.Saturday: "lau",
	time.Sunday: "sun",
}

// KlemVika slær saman like timar i vika.
//
// Nykelen er alt som gjer ein time til *den same* timen: kva han heiter,
// kven som held honom, kvar han er, og kva klokkeslett han byrjar.
// Skifter noko av det, er det ein annan time.
func KlemVika(events []models.Event, iDag time.Time) []Timebolk {
	iDagDato := iDag.Format("2006-01-02")
	naa := iDag

	bolkar := map[string]*Timebolk{}
	for _, e := range events {
		start := e.StartTime.Format("15:04")
		nykel := fmt.Sprintf("%s|%s|%s|%s", e.Title, e.TeacherName, e.RoomName, start)

		b, finst := bolkar[nykel]
		if !finst {
			b = &Timebolk{
				Tittel:    e.Title,
				Laerar:    e.TeacherName,
				Rom:       e.RoomName,
				Klasse:    e.ClassType,
				Start:     start,
				Minutt:    int(e.EndTime.Sub(e.StartTime).Minutes()),
				Plassar:   e.Capacity,
				sortering: e.StartTime.Hour()*60 + e.StartTime.Minute(),
			}
			bolkar[nykel] = b
		}

		prosent := 0
		if e.Capacity > 0 {
			prosent = e.CurrentEnrolment * 100 / e.Capacity
		}
		b.Framsyning = append(b.Framsyning, Framsyning{
			Event:   e,
			Dag:     e.StartTime.Weekday(),
			Dato:    e.StartTime,
			DagNamn: dagKort[e.StartTime.Weekday()],
			ErIDag:  e.StartTime.Format("2006-01-02") == iDagDato,
			ErUmme:  e.EndTime.Before(naa),
			Full:    e.Full(),
			Prosent: prosent,
		})
	}

	ut := make([]Timebolk, 0, len(bolkar))
	for _, b := range bolkar {
		sort.Slice(b.Framsyning, func(i, j int) bool {
			return b.Framsyning[i].Dato.Before(b.Framsyning[j].Dato)
		})
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
