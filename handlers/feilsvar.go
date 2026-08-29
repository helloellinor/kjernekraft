package handlers

import (
	"fmt"
	"html"
	"net/http"
)

// svarFeil er den eine vegen eit feilsvar skal ganga ut or ein handsamar.
//
// Han svarar i brukaren sitt eige maal, og i den skapnaden mottakaren
// kann syna fram: eit stykke for htmx — feil.js byter det inn der svaret
// elles skulde stade — og rein tekst elles. Same skapnaden som
// Recoverer i berging.go sender ved panikk, so alle feil ber den eine
// regelen som finst.
//
// Det han løyser av er http.Error med ein handskriven konstant: slike
// stod her i seks stilar og tri maal, og htmx synte ingen av deim fram
// fyrr feil.js lærde seg aa byta inn 4xx (sjaa docs/MASKINEN.md, B1).
// Nye handsamarar skal hit, ikkje til http.Error.
func svarFeil(w http.ResponseWriter, r *http.Request, status int, nykel string) {
	lang := GetLanguageFromRequest(r)
	melding := html.EscapeString(t(lang, nykel))
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprintf(w, `<div class="error" role="alert">%s</div>`, melding)
		return
	}
	http.Error(w, melding, status)
}
