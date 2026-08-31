package handsamarar

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"net/http"
	"runtime/debug"
)

// Recoverer turns a panic into a response with meaning.
//
// chi's Recoverer wrote 500 without a body, and only when nothing had been
// written: the handlers stream templates straight into the ResponseWriter,
// so a panic mid-render came after 200 had already been sent — and then
// WriteHeader(500) is a no-op, the user got half a page, and the log said
// all was well. And htmx never showed the empty 500 at all: the button
// simply did nothing, and the error existed only in the network tab.
//
// So this one holds the whole response back until the handler is done. If
// all goes well it is sent as it was. If a panic comes, the half-written
// response is discarded and the user gets a complete 500 with a message in
// their own language: a fragment for htmx — which swaps it in where the
// response should have gone, see static/js/feil.js — and a small page
// otherwise.
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
				// net/http's own way of aborting a response. It should go up to the
				// 				// server, not become a 500.
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
			// No template here: the template system may be the very thing that fell.
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

// bufra holds the response in memory until the handler has said its piece.
// That is what makes the 500 above always possible: before the first byte
// has reached the wire, nothing is promised.
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
