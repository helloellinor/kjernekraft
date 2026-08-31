package handsamarar

import (
	"time"

	"kjernekraft/handsamarar/config"
)

// Vikerekninga. Ho laag inne i timeplanhandsamaren og var difor ikkje
// noko ein kunde prøva — og ho er grunnlaget for eit reknestykke som
// gjeng den andre vegen ute i nettlesaren: vikefeltet reknar seg frå
// eit vikenummer attende til den `?week=`-offseten tenaren ventar seg.
// Tvo reknestykke som lyt vera samde, og berre det eine var skrive ned.

// VikeMåndag gjev maandagen i vika som ligg `offset` vikor fram frå
// den ein stend i. Sundagen høyrer til vika som gjeng ut og ikkje til
// den som kjem — difor steget attende.
// Måndagen byrjar ved midnatt. Klokkeslettet vart med vidare fyrr, og
// då bar «måndagen for 25 veker sidan» framleis klokka du opna sida på
// — so aktivitetskartet klypte fyrste dagen sin midt på dagen, og kva
// timar som fall utanfor kom an på når du såg etter.
func VikeMåndag(no time.Time, offset int) time.Time {
	måndag := no.AddDate(0, 0, -int(no.Weekday())+1)
	if no.Weekday() == time.Sunday {
		måndag = måndag.AddDate(0, 0, -7)
	}
	måndag = måndag.AddDate(0, 0, offset*7)
	return time.Date(måndag.Year(), måndag.Month(), måndag.Day(),
		0, 0, 0, 0, måndag.Location())
}

// VikorIAaret gjev kor mange ISO-vikor det er i aaret `t` høyrer til:
// 52 dei fleste år, 53 nokre. Vikefeltet treng talet for aa skyna at
// veke 2 sedd frå veke 51 ligg *framfyre* og ikkje langt attanfyre.
//
// 28. desember ligg alltid i den siste ISO-vika i aaret sitt. Det er den
// stuttaste maaten aa spyrja um dette på som ikkje er ei tabell.
func VikorIAaret(t time.Time) int {
	år, _ := t.ISOWeek()
	_, siste := time.Date(år, time.December, 28, 0, 0, 0, 0, t.Location()).ISOWeek()
	return siste
}

// veketalNo gjev ISO-vikenummeret for i dag, etter klokka huset held
// seg til. Vekefelti i administrasjonen tel frå det.
func veketalNo() int {
	_, v := config.GetInstance().GetCurrentTime().ISOWeek()
	return v
}
