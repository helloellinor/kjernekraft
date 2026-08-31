package handsamarar

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"kjernekraft/database"
	"kjernekraft/handsamarar/config"
	"kjernekraft/models"
)

// Saving a series is *one* action.
//
// It was ten: one call per field, sent as you left the field. That did
// three things wrong. The surface saved without anyone asking, so "I
// changed my mind" did not exist. A coherent change — move the series to
// another room *and* another day — went as two calls, and the first could
// be accepted while the second was rejected, leaving the series half
// moved. And the house does not work that way anywhere else: the profile
// and the prices mark the field, the dock at the bottom carries the
// action, and the irreversible thing is *saving* (ARKET §9, .dokk).
//
// Now it is one call with the whole series in it. Everything is checked
// first and written after: if the room is taken on even one of the coming
// occurrences, nothing is written.

// serieskjema er det skjemaet sender.
type serieskjema struct {
	SerieID int64
	Tittel  string
	Lærar   string
	Slag    string
	// The teacher shown in the field when the page was drawn. It is not
	// saved; it is there only to answer "did anyone touch this field" — see
	// the write below.
	LærarFyrr string
	RomID     int64
	Vekedag   int
	Klokke    time.Time
	Minutt    int
	Plassar   int
	GruppeID  int64
	Skildring string
	Veker     int
	Avlys     map[int64]bool
	Vikar     map[int64]string
	Flytt     map[int64]time.Time
}

func lesSerieskjema(r *http.Request) (*serieskjema, error) {
	// The body can arrive in two dresses, and ParseForm only knows one.
	//
	// It reads an x-www-form-urlencoded body; if the body is multipart it
	// reads only the query in the URL — and *sets r.Form* anyway. After that
	// FormValue does nothing, because it only parses when r.Form is nil. Every
	// field came back empty, serie_id included, and the save answered 400
	// before touching anything.
	//
	// It did not show up earlier because the CSRF guard *also* reads FormValue
	// and so parsed the body first — but only when the token does not arrive
	// in a header. csrf.js always puts it in the header for fetch, so the
	// guard never touched the body.
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			return nil, fmt.Errorf("invalid form")
		}
	} else if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("invalid form")
	}
	s := &serieskjema{
		Avlys: map[int64]bool{},
		Vikar: map[int64]string{},
		Flytt: map[int64]time.Time{},
	}

	var err error
	if s.SerieID, err = strconv.ParseInt(r.FormValue("serie_id"), 10, 64); err != nil || s.SerieID == 0 {
		return nil, fmt.Errorf("invalid serie_id")
	}
	if s.Tittel = strings.TrimSpace(r.FormValue("tittel")); s.Tittel == "" {
		return nil, fmt.Errorf("tittel is required")
	}
	if s.Lærar = strings.TrimSpace(r.FormValue("laerar")); s.Lærar == "" {
		return nil, fmt.Errorf("laerar is required")
	}
	s.LærarFyrr = strings.TrimSpace(r.FormValue("laerar_fyrr"))
	// The kind can be empty: not every class is one of the kinds the house
	// knows, and a wing without colour is truer than a random one (§1).
	s.Slag = strings.TrimSpace(r.FormValue("slag"))
	if s.RomID, err = strconv.ParseInt(r.FormValue("room_id"), 10, 64); err != nil {
		return nil, fmt.Errorf("invalid room_id")
	}
	if s.Vekedag, err = strconv.Atoi(r.FormValue("vekedag")); err != nil || s.Vekedag < 0 || s.Vekedag > 6 {
		return nil, fmt.Errorf("invalid vekedag")
	}
	if s.Klokke, err = time.Parse("15:04", r.FormValue("klokke")); err != nil {
		return nil, fmt.Errorf("invalid klokke")
	}
	if s.Minutt, err = strconv.Atoi(r.FormValue("minutt")); err != nil || s.Minutt < 15 || s.Minutt > 240 {
		return nil, fmt.Errorf("invalid minutt")
	}
	// Tomt tyder «rommet rår». Det er ikkje det same som null plassar.
	if v := strings.TrimSpace(r.FormValue("plassar")); v != "" {
		if s.Plassar, err = strconv.Atoi(v); err != nil || s.Plassar < 0 || s.Plassar > 500 {
			return nil, fmt.Errorf("invalid plassar")
		}
	}
	// Null er «open for alle». Ei ukjend gruppa er ein time ingen ser.
	if v := strings.TrimSpace(r.FormValue("gruppe_id")); v != "" {
		if s.GruppeID, err = strconv.ParseInt(v, 10, 64); err != nil || s.GruppeID < 0 {
			return nil, fmt.Errorf("invalid gruppe_id")
		}
	}
	s.Skildring = r.FormValue("skildring")
	if v := strings.TrimSpace(r.FormValue("veker")); v != "" {
		if s.Veker, err = strconv.Atoi(v); err != nil || s.Veker < 0 || s.Veker > 52 {
			return nil, fmt.Errorf("invalid veker")
		}
	}

	for _, v := range r.Form["avlys"] {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			s.Avlys[id] = true
		}
	}
	for namn, verd := range r.Form {
		if id, ok := etterStreken(namn, "vikar-"); ok && len(verd) > 0 {
			if n := strings.TrimSpace(verd[0]); n != "" {
				s.Vikar[id] = n
			}
		}
		if id, ok := etterStreken(namn, "flytt-"); ok && len(verd) > 0 {
			// «2026-09-10 18:00» — dagen og klokka den eine dagen fekk.
			if når, err := time.Parse("2006-01-02 15:04", strings.TrimSpace(verd[0])); err == nil {
				s.Flytt[id] = når
			}
		}
	}
	return s, nil
}

func etterStreken(namn, prefiks string) (int64, bool) {
	if !strings.HasPrefix(namn, prefiks) {
		return 0, false
	}
	id, err := strconv.ParseInt(namn[len(prefiks):], 10, 64)
	return id, err == nil
}

// LagraSerie takes the whole series and writes it in one go.
func (a *App) LagraSerie(w http.ResponseWriter, r *http.Request) {
	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	lang := GetLanguageFromRequest(r)
	s, err := lesSerieskjema(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	no := config.GetInstance().GetCurrentTime()
	timar, err := a.DB.GetFutureEventsBySerie(s.SerieID, no)
	if err != nil {
		http.Error(w, "Could not fetch classes", http.StatusInternalServerError)
		return
	}

	romnamn, err := a.romNamn(s.RomID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ---- work out what each day should become ----
	//
	// The series sets day, time and length on all of them; a single move
	// overrides for the one day it applies to. Only once the whole picture is
	// computed can it be checked — a check per field could not see that the
	// day moving out of a slot was the same day something else was moving
	// into.
	flyttingar := make([]database.Flytting, 0, len(timar))
	for _, e := range timar {
		if s.Avlys[e.ID] {
			continue
		}
		start := serietid(e.Start, s.Vekedag, s.Klokke)
		if når, ok := s.Flytt[e.ID]; ok {
			start = time.Date(når.Year(), når.Month(), når.Day(),
				når.Hour(), når.Minute(), 0, 0, e.Start.Location())
		}
		flyttingar = append(flyttingar, database.Flytting{
			EventID: e.ID,
			Start:   start,
			Slutt:   start.Add(time.Duration(s.Minutt) * time.Minute),
		})
	}

	// ---- prøva ----
	if melding, ok := a.serieKrasjarIRom(r, s.SerieID, s.RomID, flyttingar); !ok {
		http.Error(w, melding, http.StatusConflict)
		return
	}
	if s.Plassar > 0 {
		full, err := a.DB.PaameldeYver(s.SerieID, s.Plassar, no)
		if err == nil && full != nil {
			http.Error(w, fmt.Sprintf("%s %s: %d %s",
				t(lang, "admin.seats_taken"),
				veggklokka(full.Start).Format("2.1."),
				full.Paamelde, t(lang, "admin.enrolled")), http.StatusConflict)
			return
		}
	}

	// ---- skrivinga ----
	if len(s.Avlys) > 0 {
		ider := make([]int64, 0, len(s.Avlys))
		for id := range s.Avlys {
			ider = append(ider, id)
		}
		sort.Slice(ider, func(i, j int) bool { return ider[i] < ider[j] })
		if err := a.DB.AvlysFleire(ider); err != nil {
			http.Error(w, "Could not cancel classes", http.StatusInternalServerError)
			return
		}
	}

	if err := a.DB.UpdateSerieTitle(s.SerieID, s.Tittel, no); err != nil {
		http.Error(w, "Could not save title", http.StatusInternalServerError)
		return
	}
	// The teacher is written only when someone has touched the field.
	//
	// UpdateSerieTeacher writes to *every* coming class in the run, which is
	// right when changing the series teacher. But the field held a name
	// whatever you came to do, so every save — a new title, another room, two
	// more places — wrote that name over all the classes and took with it any
	// substitute set for a particular week. A substitute is exactly the one
	// class that must *not* follow the series.
	//
	// lærar_fyrr is what stood in the field when the page was drawn. If the
	// two are equal, nobody asked for anything here.
	if s.Lærar != s.LærarFyrr {
		if err := a.DB.UpdateSerieTeacher(s.SerieID, s.Lærar, no); err != nil {
			http.Error(w, "Could not save teacher", http.StatusInternalServerError)
			return
		}
	}
	if err := a.DB.UpdateSerieClassType(s.SerieID, s.Slag, no); err != nil {
		http.Error(w, "Could not save class type", http.StatusInternalServerError)
		return
	}
	if s.RomID > 0 {
		if err := a.DB.UpdateSerieRoom(s.SerieID, s.RomID, romnamn, no); err != nil {
			http.Error(w, "Could not save room", http.StatusInternalServerError)
			return
		}
	}
	if err := a.DB.UpdateSerieCapacity(s.SerieID, s.Plassar, no); err != nil {
		http.Error(w, "Could not save capacity", http.StatusInternalServerError)
		return
	}
	// Gruppa gjeld heile serien: er reformer-timen open for dei
	// upplærde, er han det kvar veke.
	if s.GruppeID > 0 {
		finst, err := a.DB.GruppeFinst(s.GruppeID)
		if err != nil || !finst {
			http.Error(w, "Ukjend gruppe", http.StatusBadRequest)
			return
		}
	}
	if err := a.DB.SetSerieGruppe(s.SerieID, s.GruppeID, no); err != nil {
		http.Error(w, "Could not save group", http.StatusInternalServerError)
		return
	}
	if err := a.DB.UpdateSerieDescription(s.SerieID, s.Skildring, no); err != nil {
		http.Error(w, "Could not save description", http.StatusInternalServerError)
		return
	}
	if err := a.DB.FlyttFleire(flyttingar); err != nil {
		http.Error(w, "Could not save times", http.StatusInternalServerError)
		return
	}

	// Vikarane etter serien: serien set læraren på alle, og so tek
	// den eine dagen sin eigen att.
	for id, namn := range s.Vikar {
		if s.Avlys[id] {
			continue
		}
		if err := a.DB.UpdateEventTeacher(id, namn); err != nil {
			http.Error(w, "Could not save substitute", http.StatusInternalServerError)
			return
		}
	}

	if s.Veker > 0 {
		if melding, ok := a.forlengSerie(r, s.SerieID, s.Veker, no); !ok {
			http.Error(w, melding, http.StatusConflict)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// serietid set dagen og klokka serien segjer, og held datoen sin veke.
func serietid(gamal time.Time, vekedag int, klokke time.Time) time.Time {
	steg := vekedagssteg(gamal.Weekday(), vekedag)
	dag := gamal.AddDate(0, 0, steg)
	return time.Date(dag.Year(), dag.Month(), dag.Day(),
		klokke.Hour(), klokke.Minute(), 0, 0, gamal.Location())
}

// romNamn slær upp namnet på rommet. Null er «ikkje noko rom», som er
// lovleg for ein time som er lagd inn utanum romi.
func (a *App) romNamn(romID int64) (string, error) {
	if romID == 0 {
		return "", nil
	}
	rom, err := a.DB.GetRooms()
	if err != nil {
		return "", fmt.Errorf("could not fetch rooms")
	}
	for _, v := range rom {
		if int64(v.ID) == romID {
			return v.Name, nil
		}
	}
	return "", fmt.Errorf("unknown room_id")
}

// serieKrasjarIRom prøver dei nye tidene mot rommet serien skal stå i.
func (a *App) serieKrasjarIRom(r *http.Request, serieID, romID int64, flyttingar []database.Flytting) (string, bool) {
	if romID == 0 {
		return "", true
	}
	for _, f := range flyttingar {
		kollisjon, err := a.DB.RoomConflictUtanSerie(romID, serieID, f.Start, f.Slutt)
		if err != nil || kollisjon == nil {
			continue
		}
		return fmt.Sprintf("%s %s–%s: %s",
			t(GetLanguageFromRequest(r), "admin.room_busy"),
			veggklokka(kollisjon.Start).Format("2.1. 15:04"),
			veggklokka(kollisjon.Slutt).Format("15:04"),
			kollisjon.Tittel), false
	}
	return "", true
}

// forlengSerie legg fleire veker etter den siste komande timen.
func (a *App) forlengSerie(r *http.Request, serieID int64, veker int, no time.Time) (string, bool) {
	siste, err := a.DB.SisteITimeserie(serieID, no)
	if err != nil || siste == nil {
		return t(GetLanguageFromRequest(r), "admin.serie_spent"), false
	}

	var nye []models.Event
	for i := 1; i <= veker; i++ {
		start := siste.StartTime.AddDate(0, 0, 7*i)
		slutt := siste.EndTime.AddDate(0, 0, 7*i)
		if siste.RoomID > 0 {
			kollisjon, err := a.DB.RoomConflictUtanSerie(int64(siste.RoomID), serieID, start, slutt)
			if err == nil && kollisjon != nil {
				return fmt.Sprintf("%s %s–%s: %s",
					t(GetLanguageFromRequest(r), "admin.room_busy"),
					veggklokka(kollisjon.Start).Format("2.1. 15:04"),
					veggklokka(kollisjon.Slutt).Format("15:04"),
					kollisjon.Tittel), false
			}
		}
		ny := *siste
		ny.StartTime = start
		ny.EndTime = slutt
		ny.CurrentEnrolment = 0
		nye = append(nye, ny)
	}
	if _, err := a.DB.UtvidSerie(serieID, nye); err != nil {
		return "could not extend", false
	}
	return "", true
}
