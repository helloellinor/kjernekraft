package handlers

import "time"

// Vikerekninga. Ho laag inne i timeplanhandsamaren og var difor ikkje
// noko ein kunde prøva — og ho er grunnlaget for eit reknestykke som
// gjeng den andre vegen ute i nettlesaren: vikefeltet reknar seg fraa
// eit vikenummer attende til den `?week=`-offseten tenaren ventar seg.
// Tvo reknestykke som lyt vera samde, og berre det eine var skrive ned.

// VikeMaandag gjev maandagen i vika som ligg `offset` vikor fram fraa
// den ein stend i. Sundagen høyrer til vika som gjeng ut og ikkje til
// den som kjem — difor steget attende.
func VikeMaandag(no time.Time, offset int) time.Time {
	maandag := no.AddDate(0, 0, -int(no.Weekday())+1)
	if no.Weekday() == time.Sunday {
		maandag = maandag.AddDate(0, 0, -7)
	}
	return maandag.AddDate(0, 0, offset*7)
}

// VikorIAaret gjev kor mange ISO-vikor det er i aaret `t` høyrer til:
// 52 dei fleste aar, 53 nokre. Vikefeltet treng talet for aa skyna at
// veke 2 sedd fraa veke 51 ligg *framfyre* og ikkje langt attanfyre.
//
// 28. desember ligg alltid i den siste ISO-vika i aaret sitt. Det er den
// stuttaste maaten aa spyrja um dette paa som ikkje er ei tabell.
func VikorIAaret(t time.Time) int {
	aar, _ := t.ISOWeek()
	_, siste := time.Date(aar, time.December, 28, 0, 0, 0, 0, t.Location()).ISOWeek()
	return siste
}
