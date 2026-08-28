package handlers

import (
	"sync"
	"time"
)

// ---- klokka på veggen ----
//
// Tidene i databasen er lagra som «2026-08-27 16:30:00»: klokka på
// veggen i Oslo, utan sone. Drivaren les dei som UTC, av di ein streng
// utan sone må lesast som *noko*. Verdet i Go er difor 16:30 UTC — feil
// merkelapp, rett klokke.
//
// Difor: formaterer ein det rått, står det 16:30, og det er rett. Men
// reknar ein det om — `t.In(OsloLoc)` — flyt det to timar og vert 18:30.
// Det var det heimesida gjorde: ho helsa «Sest 18:30» om ein time som
// gjekk 16:30, og ingen var påmeld noko klokka halv sju.
//
// veggklokka byggjer tidi opp att av dei tala som faktisk står i
// databasen, og set Oslo på deim. Etter det er merkelappen rett òg, og
// tidi kan både skrivast og reknast med utan å flytte seg.
//
// Alternativet var å opne databasen med `_loc=Europe/Oslo`. Det gjer
// ikkje det ein trur: drivaren les framleis strengen som UTC og *reknar
// om* etterpå, so kvar einaste klokke i huset hadde vorte to timar for
// sein. Prøvd og forkasta.
func veggklokka(t time.Time) time.Time {
	loc := oslo()
	if loc == nil {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), 0, loc)
}

var osloEinGong sync.Once

// oslo gjev sona, og hentar henne sjølv um ho ikkje er sett.
//
// OsloLoc vert sett i main(). Alt som køyrer utanfor tenaren — prøvor,
// skript, ein handsamar kalla frå ein annan samanheng — såg difor ein
// nil, og då gjorde veggklokka ingen ting og gav tidi attende med den
// gale merkelappen sin. Feilen var borte i tenaren og attende alle andre
// stader, og ingen ting sa ifrå.
func oslo() *time.Location {
	if OsloLoc == nil {
		osloEinGong.Do(func() {
			if loc, err := time.LoadLocation("Europe/Oslo"); err == nil {
				OsloLoc = loc
			}
		})
	}
	return OsloLoc
}

// fyre og etter set ei *lagra* tid opp mot eit verkeleg tidspunkt.
//
// Dei to er ikkje same slaget. Ei lagra tid er klokka på veggen med
// UTC-merkelappen drivaren gav henne; `no` er eit verkeleg augneblink.
// Sette ein dei opp mot kvarandre rått, låg dei to timar frå kvarandre
// om sommaren: klokka 18:30 stod ein time som gjekk 17:00–18:00
// framleis som «ledig plass», av di 18:00 med UTC-merkelapp er 20:00
// hjå oss. veggklokka rettar merkelappen fyrst, og då tyder
// samanlikninga det ho ser ut til å tyde.
func fyre(lagra, no time.Time) bool  { return veggklokka(lagra).Before(no) }
func etter(lagra, no time.Time) bool { return veggklokka(lagra).After(no) }
