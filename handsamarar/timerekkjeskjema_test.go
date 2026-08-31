package handsamarar

import (
	"strings"
	"testing"
	"time"

	"kjernekraft/models"
)

func proveTime(id int, serie int64, tittel string, dag int) models.Event {
	st := time.Date(2026, 9, dag, 12, 0, 0, 0, time.UTC)
	return models.Event{ID: id, SerieID: serie, Title: tittel, TeacherName: "Ida",
		RoomName: "Salen", RoomID: 1, Capacity: 12, RoomCapacity: 12,
		StartTime: st, EndTime: st.Add(time.Hour)}
}

// Kva som stend i skjemaet, avgjort av tvo spursmål som ikkje er det
// same: finst det ein *serie* aa endra, og finst det meir enn *ein time*.
//
// Prøva finst av di det freistar aa slaa dei tvo saman — ein time utan
// serie er mest alltid éin time — og av di det gjeng gale i nettupp det
// tilfellet ein ikkje tenkjer på: `GrupperTimar` samlar serielause
// timar på samantreff av tittel, lærar, rom, vekedag, klokke og lengd,
// so tvo like drop-in-timar er éi gruppa med tvo timar. Styrde
// `$utanserie` daglista, var ho burte nettupp der ein trong henne.
func TestSkjemaetFylgjerSerienOgTalet(t *testing.T) {
	tilfelle := []struct {
		namn                              string
		timar                             []models.Event
		fanor, veke, dagliste, vikar, ein bool
	}{
		{"serie med fleire", []models.Event{proveTime(1, 5, "Yoga", 2), proveTime(2, 5, "Yoga", 9)},
			true, true, true, true, false},
		{"serie med éin att", []models.Event{proveTime(3, 6, "Pilates", 3)},
			false, false, false, false, true},
		{"utan serie, éin", []models.Event{proveTime(4, 0, "Drop-in", 4)},
			false, false, false, false, true},
		// Tilfellet som gjer at $utanserie ikkje kann styra lista:
		// tvo like serielause timar vert éi gruppa med tvo timar.
		{"utan serie, tvo like", []models.Event{proveTime(5, 0, "Drop-in", 5), proveTime(6, 0, "Drop-in", 12)},
			false, true, true, true, false},
	}

	for _, tf := range tilfelle {
		t.Run(tf.namn, func(t *testing.T) {
			h := teiknTimestyringa(t, tf.timar)
			sjekk := func(namn, naal string, venta bool) {
				if strings.Contains(h, naal) != venta {
					t.Errorf("%s: venta %v", namn, venta)
				}
			}
			sjekk("fanor", "faneark rekkjevidd", tf.fanor)
			sjekk("vekefelt", "vekesetning", tf.veke)
			sjekk("daghovud", "daghovud", tf.dagliste)
			sjekk("vikar", `class="felt val-vikar"`, tf.vikar)
			// Avlysingsknappen stend i alle tilfelle; berre ordet skifter.
			sjekk("avlysknapp", `class="btn-danger avlys-rekkja"`, true)
			sjekk("eitt-ord", "Avlys timen", tf.ein)
			sjekk("fleir-ord", "Avlys heile rekkja", !tf.ein)
			// Lista er alltid i dokumentet — merket bur på rada.
			sjekk("daglista", `class="daglista"`, true)
			if tf.ein && !strings.Contains(h, `class="daglista" hidden`) {
				t.Error("lista skulde vore gøymd, ikkje burte")
			}
		})
	}
}
