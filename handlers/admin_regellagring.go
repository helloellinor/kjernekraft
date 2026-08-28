package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"kjernekraft/database"
	"kjernekraft/handlers/config"
	"kjernekraft/models"
)

// Lagringi av ein regel er *éi* handling.
//
// Ho var ti: eit kall per felt, sendt i det du forlét feltet. Det gjorde
// tri ting gale. Flata lagra utan at nokon bad om det, so «eg ombestemte
// meg» fanst ikkje. Ei endring som heng saman — flytt regelen til eit
// anna rom *og* ein annan dag — gjekk som tvo kall, og det fyrste kunde
// verta teke imot og det andre avvist, so regelen stod att halvvegs
// flutt. Og huset gjer det ikkje slik nokon annan stad: profilen og
// prisane merkjer feltet, dokka nedst ber handlingi, og det ugjenkallege
// er *å lagra* (ARKET §9, `.dokk`).
//
// No er det eitt kall med heile regelen i seg. Alt vert prøvt fyrst og
// skrive etterpaa: er rommet uppteke ein einaste av dei komande gongene,
// vert ingen ting skrive.

// regelskjema er det skjemaet sender.
type regelskjema struct {
	RegelID   int64
	Tittel    string
	Laerar    string
	RomID     int64
	Vekedag   int
	Klokke    time.Time
	Minutt    int
	Plassar   int
	Skildring string
	Veker     int
	Avlys     map[int64]bool
	Vikar     map[int64]string
	Flytt     map[int64]time.Time
}

func lesRegelskjema(r *http.Request) (*regelskjema, error) {
	// Kroppen kann koma i tvo drakter, og `ParseForm` kann berre den eine.
	//
	// Ho les eit `x-www-form-urlencoded`-lik; er kroppen `multipart`,
	// les ho berre spurjingi i adressa — og *set `r.Form`* likevel. Etter
	// det gjer `FormValue` ingen ting, av di ho berre parsar naar
	// `r.Form` er nil. Kvart felt kom attende tomt, `rule_id` med, og
	// lagringi svara 400 fyre ho hadde teke i noko.
	//
	// Det synte seg ikkje fyrr av di CSRF-vakti *ogso* les `FormValue` og
	// dermed parsa kroppen fyrst — men berre naar kjennemerket ikkje kjem
	// i ei hovudlinja. `csrf.js` legg det alltid i hovudlinja for fetch,
	// so vakti rørte aldri kroppen, og handsamaren stod att aaleine med
	// ein kropp han ikkje kunde lesa.
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			return nil, fmt.Errorf("invalid form")
		}
	} else if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("invalid form")
	}
	s := &regelskjema{
		Avlys: map[int64]bool{},
		Vikar: map[int64]string{},
		Flytt: map[int64]time.Time{},
	}

	var err error
	if s.RegelID, err = strconv.ParseInt(r.FormValue("rule_id"), 10, 64); err != nil || s.RegelID == 0 {
		return nil, fmt.Errorf("invalid rule_id")
	}
	if s.Tittel = strings.TrimSpace(r.FormValue("tittel")); s.Tittel == "" {
		return nil, fmt.Errorf("tittel is required")
	}
	if s.Laerar = strings.TrimSpace(r.FormValue("laerar")); s.Laerar == "" {
		return nil, fmt.Errorf("laerar is required")
	}
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
	// Tomt tyder «rommet raar». Det er ikkje det same som null plassar.
	if v := strings.TrimSpace(r.FormValue("plassar")); v != "" {
		if s.Plassar, err = strconv.Atoi(v); err != nil || s.Plassar < 0 || s.Plassar > 500 {
			return nil, fmt.Errorf("invalid plassar")
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
			if naar, err := time.Parse("2006-01-02 15:04", strings.TrimSpace(verd[0])); err == nil {
				s.Flytt[id] = naar
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

// SaveRuleHandler tek imot heile regelen og skriv honom i eitt.
func SaveRuleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.

	lang := GetLanguageFromRequest(r)
	s, err := lesRegelskjema(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	no := config.GetInstance().GetCurrentTime()
	timar, err := AdminDB.GetFutureEventsByRule(s.RegelID, no)
	if err != nil {
		http.Error(w, "Could not fetch classes", http.StatusInternalServerError)
		return
	}

	romnamn, err := romNamn(s.RomID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ---- rekna ut kva kvar dag skal verta ----
	//
	// Regelen set dagen, klokka og lengdi paa alle; ei einskild flytting
	// yverstyrer for den eine dagen ho gjeld. Fyrst naar heile biletet
	// er rekna ut, kann det provast — ein prøve per felt kunde ikkje
	// sjaa at dagen som flytta seg burt fri sporet var den same dagen
	// som noko anna skulde flytta seg inn i.
	flyttingar := make([]database.Flytting, 0, len(timar))
	for _, e := range timar {
		if s.Avlys[int64(e.ID)] {
			continue
		}
		start := regeltid(e.StartTime, s.Vekedag, s.Klokke)
		if naar, ok := s.Flytt[int64(e.ID)]; ok {
			start = time.Date(naar.Year(), naar.Month(), naar.Day(),
				naar.Hour(), naar.Minute(), 0, 0, e.StartTime.Location())
		}
		flyttingar = append(flyttingar, database.Flytting{
			EventID: int64(e.ID),
			Start:   start,
			Slutt:   start.Add(time.Duration(s.Minutt) * time.Minute),
		})
	}

	// ---- prøva ----
	if melding, ok := regelKrasjarIRom(r, s.RegelID, s.RomID, flyttingar); !ok {
		http.Error(w, melding, http.StatusConflict)
		return
	}
	if s.Plassar > 0 {
		full, err := AdminDB.PaameldeYver(s.RegelID, s.Plassar, no)
		if err == nil && full != nil {
			http.Error(w, fmt.Sprintf("%s %s: %d %s",
				t(lang, "admin.seats_taken"),
				veggklokka(full.StartTime).Format("2.1."),
				full.CurrentEnrolment, t(lang, "admin.enrolled")), http.StatusConflict)
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
		if err := AdminDB.AvlysFleire(ider); err != nil {
			http.Error(w, "Could not cancel classes", http.StatusInternalServerError)
			return
		}
	}

	if err := AdminDB.UpdateRuleTitle(s.RegelID, s.Tittel, no); err != nil {
		http.Error(w, "Could not save title", http.StatusInternalServerError)
		return
	}
	if err := AdminDB.UpdateRuleTeacher(s.RegelID, s.Laerar, no); err != nil {
		http.Error(w, "Could not save teacher", http.StatusInternalServerError)
		return
	}
	if s.RomID > 0 {
		if err := AdminDB.UpdateRuleRoom(s.RegelID, s.RomID, romnamn, no); err != nil {
			http.Error(w, "Could not save room", http.StatusInternalServerError)
			return
		}
	}
	if err := AdminDB.UpdateRuleCapacity(s.RegelID, s.Plassar, no); err != nil {
		http.Error(w, "Could not save capacity", http.StatusInternalServerError)
		return
	}
	if err := AdminDB.UpdateRuleDescription(s.RegelID, s.Skildring, no); err != nil {
		http.Error(w, "Could not save description", http.StatusInternalServerError)
		return
	}
	if err := AdminDB.FlyttFleire(flyttingar); err != nil {
		http.Error(w, "Could not save times", http.StatusInternalServerError)
		return
	}

	// Vikarane etter regelen: regelen set læraren paa alle, og so tek
	// den eine dagen sin eigen att.
	for id, namn := range s.Vikar {
		if s.Avlys[id] {
			continue
		}
		if err := AdminDB.UpdateEventTeacher(id, namn); err != nil {
			http.Error(w, "Could not save substitute", http.StatusInternalServerError)
			return
		}
	}

	if s.Veker > 0 {
		if melding, ok := forlengRegel(r, s.RegelID, s.Veker, no); !ok {
			http.Error(w, melding, http.StatusConflict)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// regeltid set dagen og klokka regelen segjer, og held datoen sin veke.
func regeltid(gamal time.Time, vekedag int, klokke time.Time) time.Time {
	steg := vekedagssteg(gamal.Weekday(), vekedag)
	dag := gamal.AddDate(0, 0, steg)
	return time.Date(dag.Year(), dag.Month(), dag.Day(),
		klokke.Hour(), klokke.Minute(), 0, 0, gamal.Location())
}

// romNamn slær upp namnet paa rommet. Null er «ikkje noko rom», som er
// lovleg for ein time som er lagd inn utanum romi.
func romNamn(romID int64) (string, error) {
	if romID == 0 {
		return "", nil
	}
	rom, err := AdminDB.GetRooms()
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

// regelKrasjarIRom prøver dei nye tidene mot rommet regelen skal staa i.
func regelKrasjarIRom(r *http.Request, ruleID, romID int64, flyttingar []database.Flytting) (string, bool) {
	if romID == 0 {
		return "", true
	}
	for _, f := range flyttingar {
		kollisjon, err := AdminDB.RoomConflictUtanRegel(romID, ruleID, f.Start, f.Slutt)
		if err != nil || kollisjon == nil {
			continue
		}
		return fmt.Sprintf("%s %s–%s: %s",
			t(GetLanguageFromRequest(r), "admin.room_busy"),
			veggklokka(kollisjon.StartTime).Format("2.1. 15:04"),
			veggklokka(kollisjon.EndTime).Format("15:04"),
			kollisjon.Title), false
	}
	return "", true
}

// forlengRegel legg fleire veker etter den siste komande timen.
func forlengRegel(r *http.Request, ruleID int64, veker int, no time.Time) (string, bool) {
	siste, err := AdminDB.SisteITimeregel(ruleID, no)
	if err != nil || siste == nil {
		return t(GetLanguageFromRequest(r), "admin.rule_spent"), false
	}

	var nye []models.Event
	for i := 1; i <= veker; i++ {
		start := siste.StartTime.AddDate(0, 0, 7*i)
		slutt := siste.EndTime.AddDate(0, 0, 7*i)
		if siste.RoomID > 0 {
			kollisjon, err := AdminDB.RoomConflictUtanRegel(int64(siste.RoomID), ruleID, start, slutt)
			if err == nil && kollisjon != nil {
				return fmt.Sprintf("%s %s–%s: %s",
					t(GetLanguageFromRequest(r), "admin.room_busy"),
					veggklokka(kollisjon.StartTime).Format("2.1. 15:04"),
					veggklokka(kollisjon.EndTime).Format("15:04"),
					kollisjon.Title), false
			}
		}
		ny := *siste
		ny.StartTime = start
		ny.EndTime = slutt
		ny.CurrentEnrolment = 0
		nye = append(nye, ny)
	}
	if _, err := AdminDB.UtvidRegel(ruleID, nye); err != nil {
		return "could not extend", false
	}
	return "", true
}
