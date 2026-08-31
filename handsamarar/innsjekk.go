package handsamarar

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

	"kjernekraft/handsamarar/config"
	"kjernekraft/models"
)

// The check-in screen at the desk.
//
// It has no personal login. That is a choice and not an oversight: the
// screen stands out in the studio, and if people had to log in before
// ticking themselves off, nobody would tick themselves off. The screen
// itself has a four-digit code (see InnsjekkLås) entered once per
// device — it separates the iPad at the desk from the rest of the world,
// not people from each other.
//
// The privacy is in the *window*. The page shows only classes starting
// soon or running now, and only while they are. It cannot answer "who
// trains here" — only "who is here now", which is what you see with your
// eyes standing in the room.
const (
	innsjekkFyre  = 30 * time.Minute // kor lenge fyre timen lista kjem fram
	innsjekkEtter = 20 * time.Minute // kor lenge etter han sluttar ho stend
)

const (
	// Search in the kiosk is open, like the rest of the page. These two are
	// the guard against somebody browsing the member list from the lobby: you
	// have to know who you are looking for (at least three characters), and
	// you never get more than a handful of answers.
	innsjekkSokMinst  = 3
	innsjekkSokGrense = 6
)

// InnsjekkTime er ein time på skjermen, med lista si.
type InnsjekkTime struct {
	Event    models.Event
	Klokke   string
	Paamelde []Innsjekkrad
	Talde    int
	Ledige   int
}

// Innsjekkrad er ein person på lista.
type Innsjekkrad struct {
	UserID   int64
	Namn     string
	Frammott bool
	Klokke   string // når han vart kryssa av
}

// The code for the screen. There is one iPad at the desk, and the code is
// not protection against attack — it is the threshold that stops a guest
// with their phone browsing the name search from home. Hence four digits
// in the source and not a secrets apparatus.
const innsjekkKode = "6767"

const innsjekkKake = "innsjekklaas"

// The cookie does not carry the code itself, only a digest of it — so
// reading the cookie on the iPad does not read the code.
var innsjekkKakeverd = fmt.Sprintf("%x", sha256.Sum256([]byte("innsjekk:"+innsjekkKode)))

// innsjekkLaastOpp says whether this screen has entered the code.
func innsjekkLaastOpp(r *http.Request) bool {
	k, err := r.Cookie(innsjekkKake)
	return err == nil && k.Value == innsjekkKakeverd
}

// InnsjekkLås er døri til kiosken: fire siffer, ein gong per
// iPad. Kaka stend eit år, so skjermen på disken spør ikkje att.
func (a *App) InnsjekkLås(w http.ResponseWriter, r *http.Request) {
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

	renderPage(w, r, "pages/innsjekklaas", sidedata(r, SidaInnsjekk, t(lang, "innsjekk.title"), map[string]any{
		"Feil": feil,
	}))
}

// Innsjekk teiknar skjermen.
func (a *App) Innsjekk(w http.ResponseWriter, r *http.Request) {
	if !innsjekkLaastOpp(r) {
		http.Redirect(w, r, "/innsjekk/laas", http.StatusSeeOther)
		return
	}
	lang := GetLanguageFromRequest(r)
	nå := config.GetInstance().GetCurrentTime()

	timar, err := a.DB.TimarIVindauget(nå, innsjekkFyre, innsjekkEtter)
	if err != nil {
		log.Printf("innsjekk: %v", err)
		http.Error(w, "Kunne ikkje henta timane", http.StatusInternalServerError)
		return
	}

	eventIDs := make([]int64, len(timar))
	for i, e := range timar {
		eventIDs[i] = int64(e.ID)
	}
	paamelde, err := a.DB.PaameldeTilTimar(eventIDs)
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

	// The kiosk has no login, but it gets a session and a token like any
	// other request. So the CSRF guard needs no hole for this route.
	renderPage(w, r, "pages/innsjekk", sidedata(r, SidaInnsjekk, t(lang, "innsjekk.title"), map[string]any{
		"Timar": ut,
	}))
}

// innsjekkKrav is the preamble the three writing kiosk handlers share:
// is the screen unlocked, and which class and person.
//
// The lock is the only thing separating the kiosk from an open page. A
// lock you have to remember three times is one you can forget the
// fourth; now a handler missing it shows.
func innsjekkKrav(w http.ResponseWriter, r *http.Request) (int64, int64, bool) {
	if !innsjekkLaastOpp(r) {
		http.Error(w, "Skjermen er laast", http.StatusForbidden)
		return 0, 0, false
	}
	eventID, ok := skjemaTal(w, r, "event")
	if !ok {
		return 0, 0, false
	}
	userID, ok := skjemaTal(w, r, "user")
	if !ok {
		return 0, 0, false
	}
	return eventID, userID, true
}

// InnsjekkKryss kryssar av éin person.
func (a *App) InnsjekkKryss(w http.ResponseWriter, r *http.Request) {
	eventID, userID, ok := innsjekkKrav(w, r)
	if !ok {
		return
	}

	nå := config.GetInstance().GetCurrentTime()

	// The window applies here too, not only on the page. Without this check
	// anyone could POST themselves into any class at any time — the page is
	// open, so the route is too.
	if !a.innsjekkOpen(w, nå, eventID) {
		return
	}

	if err := a.DB.MerkFrammote(eventID, userID, nå); err != nil {
		log.Printf("merk frammøte: %v", err)
		http.Error(w, "Feil", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"klokke": nå.Format("15:04"),
	})
}

// InnsjekkAngre tek eit kryss bort att: feil person vart kryssa
// av. Ruta gjer ikkje meir enn det — paameldingi stend, og kva ho elles
// skulde kosta eller telja er ikkje kiosken si sak.
func (a *App) InnsjekkAngre(w http.ResponseWriter, r *http.Request) {
	eventID, userID, ok := innsjekkKrav(w, r)
	if !ok {
		return
	}
	nå := config.GetInstance().GetCurrentTime()
	if !a.innsjekkOpen(w, nå, eventID) {
		return
	}

	if err := a.DB.FjernFrammote(eventID, userID); err != nil {
		log.Printf("innsjekk, angra kryss: %v", err)
		http.Error(w, "Feil", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

// innsjekkOpen says whether the class is in the window now, and writes the
// answer itself when it is not. Every kiosk route has to ask first — the
// page is open, so the window is the whole guard.
func (a *App) innsjekkOpen(w http.ResponseWriter, nå time.Time, eventID int64) bool {
	timar, err := a.DB.TimarIVindauget(nå, innsjekkFyre, innsjekkEtter)
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

// InnsjekkSok leitar etter eit namn, til drop-in: nokon stend i
// vestibylen utan aa ha tinga, og det er plass att.
//
// Ruta er open som resten av kiosken, og eit namnesøk er det næraste
// sida kjem aa svara på «kven trenar her» — det ho elles aldri skal.
// Verjet er tri ting: søket finst berre medan ein time er i vindauget,
// spurningi lyt ha minst tri teikn, og svaret er aldri meir enn seks
// namn og ingenting anna enn namn.
func (a *App) InnsjekkSok(w http.ResponseWriter, r *http.Request) {
	if !innsjekkLaastOpp(r) {
		http.Error(w, "Skjermen er laast", http.StatusForbidden)
		return
	}
	eventID, err := strconv.ParseInt(r.FormValue("event"), 10, 64)
	if err != nil {
		http.Error(w, "Ugyldig", http.StatusBadRequest)
		return
	}
	nå := config.GetInstance().GetCurrentTime()
	if !a.innsjekkOpen(w, nå, eventID) {
		return
	}

	q := strings.TrimSpace(r.FormValue("q"))
	if utf8.RuneCountInString(q) < innsjekkSokMinst {
		http.Error(w, "For stutt", http.StatusBadRequest)
		return
	}

	folk, err := a.DB.SokTilInnsjekk(q, eventID, innsjekkSokGrense)
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

// InnsjekkDropin melder ein på og kryssar av, i eitt: personen
// stend jo her. Paameldingi gjeng gjenom same vegen som all onnor
// paamelding, so kapasiteten vert halden i same transaksjonen.
func (a *App) InnsjekkDropin(w http.ResponseWriter, r *http.Request) {
	eventID, userID, ok := innsjekkKrav(w, r)
	if !ok {
		return
	}
	nå := config.GetInstance().GetCurrentTime()
	if !a.innsjekkOpen(w, nå, eventID) {
		return
	}

	if err := a.DB.SignupUserForEvent(userID, eventID); err != nil {
		// Alt som hindrar paameldingi — fullt, alt påmeld, ingi
		// kapasitet — er «det vart ikkje plass» sedd frå skjermen.
		// Skiljet ligg i loggen, ikkje i vestibylen.
		log.Printf("innsjekk, drop-in paa %d: %v", eventID, err)
		http.Error(w, "Vart ikkje plass", http.StatusConflict)
		return
	}
	if err := a.DB.MerkFrammote(eventID, userID, nå); err != nil {
		// Paameldingi stend; det er berre krysset som vanta. Rada kjem
		// på skjermen ved neste teikning, og personen kann kryssa der.
		log.Printf("innsjekk, kryss etter drop-in: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"klokke": nå.Format("15:04"),
	})
}
