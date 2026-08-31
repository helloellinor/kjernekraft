package handsamarar

import (
	"fmt"
	"html"
	"net/http"
)

// svarFeil er den eine vegen eit feilsvar skal ganga ut or ein handsamar.
//
// Han svarar i brukaren sitt eige mål, og i den skapnaden mottakaren
// kann syna fram: eit stykke for htmx — feil.js byter det inn der svaret
// elles skulde stade — og rein tekst elles. Same skapnaden som
// Recoverer i berging.go sender ved panikk, so alle feil ber den eine
// regelen som finst.
//
// Det han løyser av er http.Error med ein handskriven konstant: slike
// stod her i seks stilar og tri mål, og htmx synte ingen av deim fram
// fyrr feil.js lærde seg aa byta inn 4xx (sjå docs/MASKINEN.md, B1).
// Nye handsamarar skal hit, ikkje til http.Error.
func svarFeil(w http.ResponseWriter, r *http.Request, status int, nykel string) {
	svarFeilMelding(w, r, status, t(GetLanguageFromRequest(r), nykel))
}

// svarFeilMelding er den same vegen ut, for ei melding som alt er skrivi.
//
// Han finst av di ikkje kvar grunn er ein nykel enno: reglane for
// medlemskapsbyte — «Kan ikke bytte medlemskap under bindingsperiode» og
// dei fem systeri hennar — stend som strengar i database/medlemskap og
// ikkje i `maal/`. Dei skal dit, og til dess er det betre at grunnen naar
// fram i den skapnaden sida kann syna enn at han vert bytt ut med ei
// tekst som ikkje segjer kvifor.
//
// Éin utgang, tvo inngangar. Ein tridje maate aa svara paa ein feil er
// det som gjer at det vert ni av deim (docs/MASKINEN.md, F29).
func svarFeilMelding(w http.ResponseWriter, r *http.Request, status int, melding string) {
	trygg := html.EscapeString(melding)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprintf(w, `<div class="error" role="alert">%s</div>`, trygg)
		return
	}
	http.Error(w, trygg, status)
}
