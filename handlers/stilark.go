package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Stilarket er *éi* fil for nettlesaren og mange filer paa disken.
//
// Stilboki segjer at det er eitt stilark, og det er framleis sant der
// det tel: éin `<link>`, éi henting, ingen `@import`-kjede som lastar i
// fleire rundar, og ingi byggjesteg — Go set deim saman naar dei vert
// bedne um.
//
// Grunnen til at dei ligg kvar for seg er ikkje ryddesans. Det var éi
// fil paa nær fire tusen liner, og ei mislukka utskifting yver
// token-blokka tok med seg heile huset — mørketemaet med. Ei fil per
// bolk gjer at ein feil naar so langt som bolken rekk og ikkje lenger.
//
// Rekkjefylgda er filnamnet, og difor byrjar kvart namn med eit tal.
// CSS er rekkjefylgd-avhengig: `00-token` lyt koma fyre alt som les
// tokeni, og `45-smale-glas` lyt koma etter det han skriv um. Skal du
// leggja til ein bolk, gjev honom eit tal som set honom der han skal
// standa — ikkje eit som er ledig.
// Mappa er ein `var` og ikkje ein `const` av di prøvone gjeng med
// `handlers/` som arbeidsmappa og lyt peika eit hakk opp.
var stilarkMappe = "static/css/deler"

var (
	stilarkLaas  sync.RWMutex
	stilarkBytes []byte
	// Avtrykket av delmappa slik ho saag ut sist arket vart sett saman.
	stilarkAvtrykkSist string
)

// stilarkAvtrykk lagar ein streng av namn, storleik og endringstid for
// delfilene, so ein ser um noko er endra utan aa lesa innhaldet.
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

// stilarkEndra segjer um noko hev skift sidan sist me sette arket saman.
func stilarkEndra() bool {
	naa := stilarkAvtrykk()
	stilarkLaas.RLock()
	same := naa == stilarkAvtrykkSist && naa != ""
	stilarkLaas.RUnlock()
	return !same
}

// byggStilark les alle delane og set deim saman.
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
	return []byte(b.String()), nil
}

// StilarkHandler serverer det samansette arket.
//
// I utvikling vert det sett saman paa nytt for kvar soknad, som malarne:
// ein skal sjaa endringi si ved aa lasta sida, ikkje ved aa starta
// tenaren. I drift vert det sett saman ein gong.
func StilarkHandler(w http.ResponseWriter, r *http.Request) {
	stilarkLaas.RLock()
	bufra := stilarkBytes
	stilarkLaas.RUnlock()

	// I utvikling vart arket sett saman paa nytt for kvar soknad: alle
	// 31 delfilene lesne, kvar gong, ogso naar ingen ting var endra.
	// Same grepet som for malarne — stat fyrst, les berre naar noko er
	// endra. Endringane dine syner seg framleis ved neste oppdatering.
	if bufra != nil && IsDevelopment() && !stilarkEndra() {
		// arket me hev er rett
	} else if bufra == nil || IsDevelopment() {
		ny, err := byggStilark()
		if err != nil {
			log.Printf("stilarket: %v", err)
			http.Error(w, "Fann ikkje stilarket", http.StatusInternalServerError)
			return
		}
		avtrykk := stilarkAvtrykk()
		stilarkLaas.Lock()
		stilarkBytes = ny
		stilarkAvtrykkSist = avtrykk
		stilarkLaas.Unlock()
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
