package yogo

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"kjernekraft/models"
)

// ---- frå utslag til rekkjor ----
//
// Yogo gjev utslag: 53 einskilde timar for ei vike. Huset her tenkjer i
// *rekkjor* — «Vinyasa Flow med Kristina, fredag 15:15» er éin ting som
// gjeng att, ikkje tri ting som liknar — og det er rekkja
// administrasjonen endrar.
//
// Difor lyt utslagi samlast fyre dei vert lagra. Nykelen er den same
// som `GrupperTimar` i handlers-pakka nyttar når han møter timar utan
// serie-id, og det er med vilje: lagar importen ei anna gruppering enn
// den lista syner, ser administrasjonen noko anna enn det som vart
// laga.

// Serie er eit knippe utslag som høyrer til den same rekkja.
type Serie struct {
	// Nykelen som samla deim. Han er ikkje noko brukaren ser; han er
	// det importen kjenner ei rekkje att på, båe i det som kjem inn
	// og i det som alt stend i basen.
	Nykel string
	Timar []models.Event
}

// Fyrste er utslaget som kjem fyrst i tid. Han er den rekkja *ser ut
// som*: namn, lærar, rom, klokke.
func (s Serie) Fyrste() models.Event {
	if len(s.Timar) == 0 {
		return models.Event{}
	}
	return s.Timar[0]
}

// SerieNykel kjenner ei rekkje att på det som er likt kvar gong ho
// gjeng: namn, lærar, rom, vekedag, klokkeslett og lengd.
//
// Klokka vert lesi raatt (`Format`, aldri `In`). Heile huset gjer det
// slik — den lagra tidi *er* veggklokka — og ei umrekning her hadde
// gjort tvo rekkjor av éi kvar gong sumartidi skifte.
func SerieNykel(e models.Event) string {
	return fmt.Sprintf("%s|%s|%s|%d|%s|%d",
		e.Title, e.TeacherName, e.RoomName, e.StartTime.Weekday(),
		e.StartTime.Format("15:04"),
		int(e.EndTime.Sub(e.StartTime).Minutes()))
}

// GrupperISeriar samlar utslag under rekkjone sine.
//
// Rekkjefylgda er timeplanen si: fyrste utslaget avgjer, so ein les
// resultatet i den same rekkja som lista syner. Utslagi inni kvar
// rekkje stend etter dato.
func GrupperISeriar(timar []models.Event) []Serie {
	under := map[string]*Serie{}
	var rekkje []string

	for _, t := range timar {
		n := SerieNykel(t)
		s, finst := under[n]
		if !finst {
			s = &Serie{Nykel: n}
			under[n] = s
			rekkje = append(rekkje, n)
		}
		s.Timar = append(s.Timar, t)
	}

	ut := make([]Serie, 0, len(rekkje))
	for _, n := range rekkje {
		s := under[n]
		sort.Slice(s.Timar, func(i, j int) bool {
			return s.Timar[i].StartTime.Before(s.Timar[j].StartTime)
		})
		ut = append(ut, *s)
	}
	sort.SliceStable(ut, func(i, j int) bool {
		return ut[i].Fyrste().StartTime.Before(ut[j].Fyrste().StartTime)
	})
	return ut
}

// UtslagNykel kjenner eit einskilt utslag att.
//
// Det er han importen nyttar til aa sjå kva som alt stend i basen, so
// ei ny køyring ikkje legg inn det same ein gong til.
//
// Læraren og rommet er med, og det er ikkje varsemd — det er noko
// timeplanen faktisk gjer. Her stod «namnet og klokkeslettet», på den
// grunngjevingi at eit studio kann ha tvo *ulike* timar samstundes men
// ikkje tvo med det same namnet. Kjernekraft hev tvo «Pilates
// Apparatus» måndag 17:30, med kvar sin lærar. Med den gamle nykelen
// var dei éin time: den andre vart lesi som ein han alt hadde, og ho
// hadde vorte burte for godt i ein import der berre den eine stod her
// frå fyrr.
func UtslagNykel(e models.Event) string {
	return strings.Join([]string{
		e.Title, e.TeacherName, e.RoomName,
		e.StartTime.Format("2006-01-02 15:04"),
	}, "|")
}

// Vekedagsnamn er berre til utskrift i importen.
func Vekedagsnamn(d time.Weekday) string {
	return [...]string{"sun", "man", "tys", "ons", "tor", "fre", "lau"}[d]
}
