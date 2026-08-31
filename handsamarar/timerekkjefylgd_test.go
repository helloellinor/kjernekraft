package handsamarar

import (
	"testing"
	"time"

	"kjernekraft/models"
)

// timarPaaSameKlokke gjev tri «Reformer» klokka 17 med kvar sin lærar.
// Dei hev same klokka *og* same tittelen, so alt som skil dei er læraren
// — nett den uavgjorde som fall attende på rekkjefylgdi basen gav.
func timarPaaSameKlokke(måndag time.Time) []models.Event {
	start := time.Date(måndag.Year(), måndag.Month(), måndag.Day(), 17, 0, 0, 0, måndag.Location())
	var ut []models.Event
	for i, lærar := range []string{"Kristina", "Leon", "Anja"} {
		ut = append(ut, models.Event{
			ID:          100 + i,
			Title:       "Reformer",
			TeacherName: lærar,
			RoomName:    "Salen",
			ClassType:   "reformer",
			StartTime:   start,
			EndTime:     start.Add(50 * time.Minute),
			Capacity:    4,
		})
	}
	return ut
}

// Radene skal koma i den same rekkjefylgdi kvar gong, kva rekkjefylgd
// timane kom inn i.
//
// Sorteringi samanlikna klokka og so tittelen. Radene er bolka på
// «tittel|lærar|rom», so tri timar med same tittel og same klokke gav
// tri rader der samanlikningi ikkje hadde meir å gå på — og daa fall
// utfallet attende på rekkjefylgdi frå basen, som heller ikkje var
// avgjord. Ti henteringar av den same sida gav fem ulike rekkjefylgder,
// og radene bytte plass under lesaren utan at noko hadde endra seg.
func TestRadeneStendILikRekkjefylgdKvarGong(t *testing.T) {
	måndag := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	nå := måndag.Add(6 * time.Hour)

	timar := timarPaaSameKlokke(måndag)
	fasit := BuildWeekRows("nn", timar, nå, måndag)
	if len(fasit) != 3 {
		t.Fatalf("venta tri rader, fekk %d", len(fasit))
	}

	// Same timane, i kvar einaste rekkjefylgd dei kann koma i.
	for _, bytt := range [][]int{{0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}} {
		stokka := []models.Event{timar[bytt[0]], timar[bytt[1]], timar[bytt[2]]}
		fekk := BuildWeekRows("nn", stokka, nå, måndag)
		if len(fekk) != len(fasit) {
			t.Fatalf("%v: venta %d rader, fekk %d", bytt, len(fasit), len(fekk))
		}
		for i := range fasit {
			if fekk[i].Teacher != fasit[i].Teacher {
				t.Errorf("%v: rad %d vart «%s», venta «%s» — rekkjefylgdi fylgjer inndata",
					bytt, i, fekk[i].Teacher, fasit[i].Teacher)
			}
		}
	}
}

// Og ordningi skal vera den ein ventar: klokka, so tittel, so lærar.
func TestRadeneErOrdaEtterKlokkeTittelLaerar(t *testing.T) {
	måndag := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	nå := måndag.Add(6 * time.Hour)

	rader := BuildWeekRows("nn", timarPaaSameKlokke(måndag), nå, måndag)
	vil := []string{"Anja", "Kristina", "Leon"}
	for i, v := range vil {
		if i >= len(rader) {
			t.Fatalf("venta %d rader, fekk %d", len(vil), len(rader))
		}
		if rader[i].Teacher != v {
			t.Errorf("rad %d: «%s», venta «%s»", i, rader[i].Teacher, v)
		}
	}
}
