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
	// Rommet regelen gjeng i. Romveljaren i regelsetningi lyt vita kva
	// han stend paa no, og namnet aaleine held ikkje — tvo rom kann ha
	// same namnet, og veljaren sender id-en.
	RomID int
	// Plassar er timen si eigi kapasitet (0 = ingi eigi), RomPlassar er
	// rommet sitt tal. Feltet syner det eine som verdi og det andre som
	// framlegg, so ein ser um talet er valt eller arva.
	Plassar    int
	RomPlassar int
	Timar      []models.Event
}

// Sluttar er den siste komande timen i regelen. «Serien gjeng aatte
// gonger» segjer ikkje naar han er ute; datoen gjer det.
func (r Timeregel) Sluttar() time.Time {
	if len(r.Timar) == 0 {
		return time.Time{}
	}
	return r.Timar[len(r.Timar)-1].StartTime
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
		if !etter(t.EndTime, no) {
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
		g.RomID = fyrste.RoomID
		g.Plassar = fyrste.EigenPlassar
		g.RomPlassar = fyrste.RoomCapacity
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

// Siktdag er ein vekedag sikti kann velja. Talet er `time.Weekday`, so
// rada og veljaren samanliknar det same.
type Siktdag struct {
	Tal   int
	Nykel string
}

// Siktval er vali sikti yver regellista tilbyd.
//
// Dei vert rekna av reglane som *finst*, og ikkje av alle lærarane og
// alle romi huset hev. Ein veljar med tjuge lærarar der tri held time er
// ikkje eit sikt, han er ein katalog: kvart val du kann gjera skal gjeva
// deg noko å sjå på. Er det berre eitt val å velja millom, siktar det
// ingen ting, og malen let veljaren vera.
type Siktval struct {
	Dagar    []Siktdag
	Laerarar []string
}

// SiktvalFor plukkar dei ulike dagane, lærarane og romi ut or reglane.
// Dagane stend i timeplanen si rekkjefylgd — maandag fyrst — og hine
// alfabetisk, av di ein leitar etter eit namn ein alt veit.
func SiktvalFor(reglar []Timeregel) Siktval {
	var val Siktval

	dagsett := map[time.Weekday]bool{}
	for _, r := range reglar {
		dagsett[r.Vekedag] = true
	}
	// Maandag fyrst, som i timeplanen og som i GrupperTimar.
	for i := 0; i < 7; i++ {
		d := time.Weekday((i + 1) % 7)
		if dagsett[d] {
			val.Dagar = append(val.Dagar, Siktdag{Tal: int(d), Nykel: Timeregel{Vekedag: d}.VekedagNykel()})
		}
	}

	val.Laerarar = ulike(reglar, func(r Timeregel) string { return r.Laerar })
	return val
}

// ulike gjev dei ulike ikkje-tome verdi, sorterte.
func ulike(reglar []Timeregel, av func(Timeregel) string) []string {
	sett := map[string]bool{}
	for _, r := range reglar {
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
