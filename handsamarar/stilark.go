package handsamarar

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// The stylesheet is *one* file for the browser and many files on disk.
//
// The style book says there is one stylesheet, and that is still true where
// it counts: one <link>, one fetch, no @import chain loading in rounds, and
// no build step — Go assembles them when they are asked for.
//
// They lie apart not out of tidiness. It was one file of nearly four
// thousand lines, and a failed replacement across the token block took the
// whole house with it, dark theme included. One file per section means a
// mistake reaches as far as the section and no further.
//
// The order is the filename, which is why every name starts with a number.
// CSS is order-dependent: 00-token has to come before anything that reads
// the tokens, and 45-smale-glas after what it overrides. Adding a section
// means giving it a number that puts it where it belongs — not one that
// happens to be free.
//
// The directory is a var and not a const because the tests run with
// handlers/ as the working directory and have to point one step up.
var stilarkMappe = "static/css/deler"

var (
	stilarkLås   sync.RWMutex
	stilarkBytes []byte
	// Avtrykket av delmappa slik ho såg ut sist arket vart sett saman.
	stilarkAvtrykkSist string
)

// stilarkAvtrykk makes a string of name, size and modification time for
// the parts, so you can see whether anything changed without reading the
// contents.
func stilarkAvtrykk() string {
	oppf, err := os.ReadDir(stilarkMappe)
	if err != nil {
		return ""
	}
	var b strings.Builder
	for _, o := range oppf {
		if o.IsDir() || !strings.HasSuffix(o.Name(), ".css") {
			continue
		}
		info, err := o.Info()
		if err != nil {
			return ""
		}
		fmt.Fprintf(&b, "%s|%d|%d\n", o.Name(), info.Size(), info.ModTime().UnixNano())
	}
	return b.String()
}

// stilarkEndra says whether anything has changed since we last assembled.
func stilarkEndra() bool {
	nå := stilarkAvtrykk()
	stilarkLås.RLock()
	same := nå == stilarkAvtrykkSist && nå != ""
	stilarkLås.RUnlock()
	return !same
}

// utanKommentar tek `/* ... */` ut or arket.
//
// Kommentarane er sjølve grunnen til at desse filone er verdt å arbeida i,
// og dei skal liggja urørde på disken. Men dei er 70 % av det lesaren lastar
// ned — 216 kB prosa til kvar einaste som opnar timeplanen — og dei segjer
// ingen ting til nettlesaren. Difor: heile på disken, rulla saman på vegen ut.
//
// Handsamar ikkje strengar med `/*` i seg, av di CSS ikkje hev nokon. `url()`
// kann bera det i teorien; ingen av delane våre gjer det, og
// TestArketMistarIkkjeReglarNaarKommentaraneGjeng vaktar det.
func utanKommentar(css []byte) []byte {
	ut := make([]byte, 0, len(css))
	for i := 0; i < len(css); {
		if css[i] == '/' && i+1 < len(css) && css[i+1] == '*' {
			slutt := bytes.Index(css[i+2:], []byte("*/"))
			if slutt < 0 {
				break // uavslutta kommentar: resten er kommentar
			}
			i += 2 + slutt + 2
			continue
		}
		ut = append(ut, css[i])
		i++
	}
	// Tomme liner som stod att etter kommentarane.
	return bytes.ReplaceAll(bytes.TrimSpace(ut), []byte("\n\n\n"), []byte("\n"))
}

// byggStilark reads all the parts and puts them together.
func byggStilark() ([]byte, error) {
	oppf, err := os.ReadDir(stilarkMappe)
	if err != nil {
		return nil, err
	}
	var namn []string
	for _, o := range oppf {
		if !o.IsDir() && strings.HasSuffix(o.Name(), ".css") {
			namn = append(namn, o.Name())
		}
	}
	sort.Strings(namn)

	var b strings.Builder
	for i, n := range namn {
		data, err := os.ReadFile(filepath.Join(stilarkMappe, n))
		if err != nil {
			return nil, err
		}
		if i > 0 {
			b.WriteString("\n")
		}
		b.Write(data)
	}
	return utanKommentar([]byte(b.String())), nil
}

// Stilark serves the assembled sheet.
//
// In development it is reassembled per request, like the templates: you
// should see your change by loading the page, not by restarting the server.
// In production it is assembled once.
func (a *App) Stilark(w http.ResponseWriter, r *http.Request) {
	stilarkLås.RLock()
	bufra := stilarkBytes
	stilarkLås.RUnlock()

	// In development the sheet was reassembled per request: all 31 parts
	// 	// read, every time, even when nothing had changed. Same move as for the
	// 	// templates — stat first, read only when something changed. Your changes
	// 	// still show on the next reload.
	if bufra != nil && IsDevelopment() && !stilarkEndra() {
		// the sheet we have is correct
	} else if bufra == nil || IsDevelopment() {
		ny, err := byggStilark()
		if err != nil {
			log.Printf("stilarket: %v", err)
			http.Error(w, "Fann ikkje stilarket", http.StatusInternalServerError)
			return
		}
		avtrykk := stilarkAvtrykk()
		stilarkLås.Lock()
		stilarkBytes = ny
		stilarkAvtrykkSist = avtrykk
		stilarkLås.Unlock()
		bufra = ny
	}

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	if IsDevelopment() {
		w.Header().Set("Cache-Control", "no-store")
	} else if r.URL.Query().Get("v") != "" {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=300")
	}
	w.Write(bufra)
}
