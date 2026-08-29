package handlers

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"net/http"
	"runtime/debug"
)

// Recoverer gjer ein panikk um til eit svar med meining.
//
// chi sin Recoverer skreiv 500 utan kropp, og det berre naar ingen ting
// var skrive: handsamarane strøymer malarne beint i ResponseWriter, so
// ein panikk midt i teikningi kom etter at 200 alt var sendt — daa er
// WriteHeader(500) eit inkje, brukaren fekk ei halv sida, og loggen sa
// at alt gjekk vel. Og den tome 500-en synte htmx aldri fram: knappen
// gjorde berre ingen ting, og feilen fanst einast i nettverksfana.
//
// Difor held denne heile svaret attende til handsamaren er ferdug.
// Gjeng alt vel, vert det sendt som det var. Kjem ein panikk, vert det
// halvskrivne kasta, og brukaren fær ein heil 500 med ei melding i sitt
// eige maal: eit stykke for htmx — som byter det inn der svaret skulde
// stade, sjaa static/js/feil.js — og ei lita sida elles.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := &bufra{w: w}
		defer func() {
			p := recover()
			if p == nil {
				b.send()
				return
			}
			if p == http.ErrAbortHandler {
				// net/http sin eigen maate aa stogga eit svar paa.
				// Han skal upp til tenaren, ikkje verta 500.
				panic(p)
			}
			log.Printf("panikk i %s %s: %v\n%s", r.Method, r.URL.Path, p, debug.Stack())

			hovud := w.Header()
			hovud.Del("Content-Length")
			hovud.Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)

			lang := GetLanguageFromRequest(r)
			melding := html.EscapeString(t(lang, "feil.krasj"))
			if r.Header.Get("HX-Request") == "true" {
				// Same skapnad som skjemafeilarne elles sender, so
				// meldingi ber den regelen som alt finst.
				fmt.Fprintf(w, `<div class="error" role="alert">%s</div>`, melding)
				return
			}
			// Ingen mal her: malverket kann vera nettupp det som fall.
			fmt.Fprintf(w, `<!doctype html>
<html lang="%s"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Kjernekraft</title>
<link rel="stylesheet" href="/static/css/kjernekraft.css">
</head><body><main><p class="error" role="alert">%s</p></main></body></html>
`, lang, melding)
		}()
		next.ServeHTTP(b, r)
	})
}

// bufra held svaret i minnet til handsamaren hev sagt sitt. Det er
// dette som gjer at 500-en ovanfor alltid er mogeleg: fyrr fyrste byten
// hev naatt lina, er ingen ting lova.
type bufra struct {
	w      http.ResponseWriter
	status int
	kropp  bytes.Buffer
}

func (b *bufra) Header() http.Header { return b.w.Header() }

// WriteHeader hugsar berre det fyrste talet, som net/http sjølv gjer.
func (b *bufra) WriteHeader(status int) {
	if b.status == 0 {
		b.status = status
	}
}

func (b *bufra) Write(p []byte) (int, error) {
	return b.kropp.Write(p)
}

func (b *bufra) send() {
	if b.status != 0 {
		b.w.WriteHeader(b.status)
	}
	if b.kropp.Len() > 0 {
		b.w.Write(b.kropp.Bytes())
	}
}
