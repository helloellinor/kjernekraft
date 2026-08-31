package handsamarar

import "net/http"

// Sidenykel er namnet base.html kjenner sida på.
//
// Malen avgjer kva javascript som vert lasta med å samanlikna
// `.CurrentPage` mot faste strengar — sju vilkår i base.html:77-100. Nykelen
// stod skriven som ein lauslivd streng på baae sidor: femten handsamarar
// skreiv sin, og malen skreiv sin um att. Ein skrivefeil i den eine enden
// gav ei sida heilt utan javascript, og ingen ting sa frå — sida teikna
// seg, ho berre gjorde ingen ting.
//
// Difor ein type. Go-sida fær ein konstant kompilatoren kjenner, og
// TestMalenKjennerBerreSidenyklarSomFinst held malen si lista i takt med
// denne. Sjølve verdet gjeng ut i data-kartet som ein rein streng, av di
// det er det malen samanliknar.
type Sidenykel string

const (
	SidaHeim          Sidenykel = "hjem"
	SidaTimeplan      Sidenykel = "timeplan"
	SidaMedlemskap    Sidenykel = "medlemskap"
	SidaKlippekort    Sidenykel = "klippekort"
	SidaBetaling      Sidenykel = "betaling"
	SidaAdmin         Sidenykel = "admin"
	SidaProfil        Sidenykel = "profil"
	SidaArket         Sidenykel = "arket"
	SidaInnsjekk      Sidenykel = "innsjekk"
	SidaInnlogging    Sidenykel = "innlogging"
	SidaRegistrering  Sidenykel = "registrering"
	SidaVilkaar       Sidenykel = "vilkaar"
	SidaGloymtPassord Sidenykel = "gloymt-passord"
	SidaTestdata      Sidenykel = "testdata"
)

// sidenyklar er alle nyklane som finst. Prøva samanliknar henne med det
// malen spør etter.
var sidenyklar = []Sidenykel{
	SidaHeim, SidaTimeplan, SidaMedlemskap, SidaKlippekort, SidaBetaling,
	SidaAdmin, SidaProfil, SidaArket, SidaInnsjekk, SidaInnlogging,
	SidaRegistrering, SidaVilkaar, SidaGloymtPassord, SidaTestdata,
}

// sidedata byggjer det kvar sida treng, og legg ekstra uppå.
//
// Dei fem same nyklane — sida, målet, kjennemerket, løyvet og namnet —
// vart sette saman for hand i femten handsamarar. Det er ikkje berre
// gjentaking: `CSRFToken` som fell ut gjev eit skjema som ikkje kann
// sendast, og `IsAdmin` som fell ut gøymer administrasjonslenkja for den
// som skal hava henne. Slikt skal ein ikkje kunna gløyma.
func sidedata(r *http.Request, side Sidenykel, tittel string, ekstra map[string]any) map[string]any {
	data := map[string]any{
		"CurrentPage": string(side),
		"Title":       tittel,
		"Lang":        GetLanguageFromRequest(r),
		"CSRFToken":   CSRFToken(r),
		"IsAdmin":     sessionIsAdmin(r),
		"UserName":    sessionUserName(r),
		"ExternalCSS": []string{},
	}
	for k, v := range ekstra {
		data[k] = v
	}
	return data
}
