package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"kjernekraft/handlers/config"
	"kjernekraft/models"
)

// Innsjekkskjermen ved disken.
//
// Han hev ingi personleg innlogging. Det er eit val og ikkje ei
// forsøming: skjermen stend framme i studioet, og skulde nokon lyt logga
// seg paa fyre dei kryssa av, hadde ingen kryssa av. Skjermen sjølv hev
// ein firesifra kode (sjaa InnsjekkLaasHandler) som vert skriven ein
// gong per apparat — han skil iPaden paa disken fraa resten av verdi,
// ikkje folk fraa kvarandre.
//
// Personvernet ligg elles i *vindauget*. Sida syner berre timar som
// byrjar snart eller gjeng no, og berre so lenge dei gjer det. Ho kann
// ikkje svara paa «kven trenar her» — berre paa «kven er her no», og det
// er det same som ein ser med augo naar ein stend i rommet.
const (
	innsjekkFyre  = 30 * time.Minute // kor lenge fyre timen lista kjem fram
	innsjekkEtter = 20 * time.Minute // kor lenge etter han sluttar ho stend
)

const (
	// Søket i kiosken er ope, som resten av sida. Desse tvo er verjet
	// mot at nokon blar i medlemslista fraa vestibylen: du lyt vita
	// kven du leitar etter (minst tri teikn), og du fær aldri meir enn
	// ei handfull svar.
	innsjekkSokMinst  = 3
	innsjekkSokGrense = 6
)

// InnsjekkTime er ein time paa skjermen, med lista si.
type InnsjekkTime struct {
	Event    models.Event
	Klokke   string
	Paamelde []Innsjekkrad
	Talde    int
	Ledige   int
}

// Innsjekkrad er ein person paa lista.
type Innsjekkrad struct {
	UserID   int64
	Namn     string
	Frammott bool
	Klokke   string // naar han vart kryssa av
}

// Koden for skjermen. Det stend éin iPad paa disken, og koden er ikkje
// eit vern mot aatak — han er terskelen som gjer at ein gjest med
// telefonen sin ikkje kann bla i namnesøket heimanfraa. Difor er han
// fire siffer i kjeldekoden og ikkje eit løyndomsapparat.
const innsjekkKode = "6767"

const innsjekkKake = "innsjekklaas"

// Kaka ber ikkje koden sjølv, berre eit avtrykk av honom — so ein som
// les kaka paa iPaden ikkje les koden med det same.
var innsjekkKakeverd = fmt.Sprintf("%x", sha256.Sum256([]byte("innsjekk:"+innsjekkKode)))

// innsjekkLaastOpp segjer um denne skjermen hev skrive koden.
func innsjekkLaastOpp(r *http.Request) bool {
	k, err := r.Cookie(innsjekkKake)
	return err == nil && k.Value == innsjekkKakeverd
}

// InnsjekkLaasHandler er døri til kiosken: fire siffer, ein gong per
// iPad. Kaka stend eit aar, so skjermen paa disken spør ikkje att.
func InnsjekkLaasHandler(w http.ResponseWriter, r *http.Request) {
	lang := GetLanguageFromRequest(r)

	feil := false
	if r.Method == http.MethodPost {
		if r.FormValue("kode") == innsjekkKode {
			http.SetCookie(w, &http.Cookie{
				Name:     innsjekkKake,
				Value:    innsjekkKakeverd,
				Path:     "/",
				MaxAge:   365 * 24 * 60 * 60,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			http.Redirect(w, r, "/innsjekk", http.StatusSeeOther)
			return
		}
		feil = true
	}

	renderPage(w, r, "pages/innsjekklaas", map[string]any{
		"Title":       t(lang, "innsjekk.title"),
		"CurrentPage": "innsjekk",
		"Lang":        lang,
		"Feil":        feil,
		"CSRFToken":   CSRFToken(r),
	})
}

// InnsjekkHandler teiknar skjermen.
func InnsjekkHandler(w http.ResponseWriter, r *http.Request) {
	if !innsjekkLaastOpp(r) {
		http.Redirect(w, r, "/innsjekk/laas", http.StatusSeeOther)
		return
	}
	lang := GetLanguageFromRequest(r)
	naa := config.GetInstance().GetCurrentTime()

	timar, err := DB.TimarIVindauget(naa, innsjekkFyre, innsjekkEtter)
	if err != nil {
		log.Printf("innsjekk: %v", err)
		http.Error(w, "Kunne ikkje henta timane", http.StatusInternalServerError)
		return
	}

	eventIDs := make([]int64, len(timar))
	for i, e := range timar {
		eventIDs[i] = int64(e.ID)
	}
	paamelde, err := DB.PaameldeTilTimar(eventIDs)
	if err != nil {
		log.Printf("innsjekk, paamelde: %v", err)
		http.Error(w, "Kunne ikkje henta dei påmelde", http.StatusInternalServerError)
		return
	}

	var ut []InnsjekkTime
	for _, e := range timar {
		folk := paamelde[int64(e.ID)]
		t := InnsjekkTime{Event: e, Klokke: veggklokka(e.StartTime).Format("15:04"),
			Ledige: e.Ledige()}
		for _, p := range folk {
			rad := Innsjekkrad{UserID: p.UserID, Namn: p.Namn}
			if p.Frammott != nil {
				rad.Frammott = true
				rad.Klokke = veggklokka(*p.Frammott).Format("15:04")
				t.Talde++
			}
			t.Paamelde = append(t.Paamelde, rad)
		}
		ut = append(ut, t)
	}

	renderPage(w, r, "pages/innsjekk", map[string]any{
		"Title":       t(lang, "innsjekk.title"),
		"CurrentPage": "innsjekk",
		"Lang":        lang,
		"Timar":       ut,
		// Kiosken hev ingi innlogging, men han fær ei økt og eit
		// kjennemerke som kvar annan soknad. Difor treng CSRF-vakta
		// ikkje eit hol for denne ruta.
		"CSRFToken": CSRFToken(r),
	})
}

// InnsjekkMerkHandler kryssar av éin person.
func InnsjekkMerkHandler(w http.ResponseWriter, r *http.Request) {
	if !innsjekkLaastOpp(r) {
		http.Error(w, "Skjermen er laast", http.StatusForbidden)
		return
	}
	eventID, err1 := strconv.ParseInt(r.FormValue("event"), 10, 64)
	userID, err2 := strconv.ParseInt(r.FormValue("user"), 10, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "Ugyldig", http.StatusBadRequest)
		return
	}

	naa := config.GetInstance().GetCurrentTime()

	// Vindauget gjeld her med, og ikkje berre paa sida. Utan denne
	// sjekken kunde kven som helst POSta seg inn paa kva time som helst
	// naar som helst — sida er open, so ruta er det ogso.
	if !innsjekkOpen(w, naa, eventID) {
		return
	}

	if err := DB.MerkFrammote(eventID, userID, naa); err != nil {
		log.Printf("merk frammøte: %v", err)
		http.Error(w, "Feil", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"klokke": naa.Format("15:04"),
	})
}

// InnsjekkAngreHandler tek eit kryss bort att: feil person vart kryssa
// av. Ruta gjer ikkje meir enn det — paameldingi stend, og kva ho elles
// skulde kosta eller telja er ikkje kiosken si sak.
func InnsjekkAngreHandler(w http.ResponseWriter, r *http.Request) {
	if !innsjekkLaastOpp(r) {
		http.Error(w, "Skjermen er laast", http.StatusForbidden)
		return
	}
	eventID, err1 := strconv.ParseInt(r.FormValue("event"), 10, 64)
	userID, err2 := strconv.ParseInt(r.FormValue("user"), 10, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "Ugyldig", http.StatusBadRequest)
		return
	}
	naa := config.GetInstance().GetCurrentTime()
	if !innsjekkOpen(w, naa, eventID) {
		return
	}

	if err := DB.FjernFrammote(eventID, userID); err != nil {
		log.Printf("innsjekk, angra kryss: %v", err)
		http.Error(w, "Feil", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

// innsjekkOpen segjer um timen er i vindauget no, og skriv svaret sjølv
// naar han ikkje er det. Kvar rute paa kiosken lyt spyrja fyrst — sida
// er open, so vindauget er heile vaktholdet.
func innsjekkOpen(w http.ResponseWriter, naa time.Time, eventID int64) bool {
	timar, err := DB.TimarIVindauget(naa, innsjekkFyre, innsjekkEtter)
	if err != nil {
		http.Error(w, "Feil", http.StatusInternalServerError)
		return false
	}
	for _, e := range timar {
		if int64(e.ID) == eventID {
			return true
		}
	}
	http.Error(w, "Timen er ikkje open for innsjekk", http.StatusForbidden)
	return false
}

// InnsjekkSokHandler leitar etter eit namn, til drop-in: nokon stend i
// vestibylen utan aa ha tinga, og det er plass att.
//
// Ruta er open som resten av kiosken, og eit namnesøk er det næraste
// sida kjem aa svara paa «kven trenar her» — det ho elles aldri skal.
// Verjet er tri ting: søket finst berre medan ein time er i vindauget,
// spurningi lyt ha minst tri teikn, og svaret er aldri meir enn seks
// namn og ingenting anna enn namn.
func InnsjekkSokHandler(w http.ResponseWriter, r *http.Request) {
	if !innsjekkLaastOpp(r) {
		http.Error(w, "Skjermen er laast", http.StatusForbidden)
		return
	}
	eventID, err := strconv.ParseInt(r.FormValue("event"), 10, 64)
	if err != nil {
		http.Error(w, "Ugyldig", http.StatusBadRequest)
		return
	}
	naa := config.GetInstance().GetCurrentTime()
	if !innsjekkOpen(w, naa, eventID) {
		return
	}

	q := strings.TrimSpace(r.FormValue("q"))
	if utf8.RuneCountInString(q) < innsjekkSokMinst {
		http.Error(w, "For stutt", http.StatusBadRequest)
		return
	}

	folk, err := DB.SokTilInnsjekk(q, eventID, innsjekkSokGrense)
	if err != nil {
		log.Printf("innsjekk, søk: %v", err)
		http.Error(w, "Feil", http.StatusInternalServerError)
		return
	}

	type treff struct {
		ID   int64  `json:"id"`
		Namn string `json:"namn"`
	}
	ut := make([]treff, 0, len(folk))
	for _, p := range folk {
		ut = append(ut, treff{ID: p.UserID, Namn: p.Namn})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"folk": ut})
}

// InnsjekkDropinHandler melder ein paa og kryssar av, i eitt: personen
// stend jo her. Paameldingi gjeng gjenom same vegen som all onnor
// paamelding, so kapasiteten vert halden i same transaksjonen.
func InnsjekkDropinHandler(w http.ResponseWriter, r *http.Request) {
	if !innsjekkLaastOpp(r) {
		http.Error(w, "Skjermen er laast", http.StatusForbidden)
		return
	}
	eventID, err1 := strconv.ParseInt(r.FormValue("event"), 10, 64)
	userID, err2 := strconv.ParseInt(r.FormValue("user"), 10, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "Ugyldig", http.StatusBadRequest)
		return
	}
	naa := config.GetInstance().GetCurrentTime()
	if !innsjekkOpen(w, naa, eventID) {
		return
	}

	if err := DB.SignupUserForEvent(userID, eventID); err != nil {
		// Alt som hindrar paameldingi — fullt, alt paameld, ingi
		// kapasitet — er «det vart ikkje plass» sedd fraa skjermen.
		// Skiljet ligg i loggen, ikkje i vestibylen.
		log.Printf("innsjekk, drop-in paa %d: %v", eventID, err)
		http.Error(w, "Vart ikkje plass", http.StatusConflict)
		return
	}
	if err := DB.MerkFrammote(eventID, userID, naa); err != nil {
		// Paameldingi stend; det er berre krysset som vanta. Rada kjem
		// paa skjermen ved neste teikning, og personen kann kryssa der.
		log.Printf("innsjekk, kryss etter drop-in: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"klokke": naa.Format("15:04"),
	})
}
