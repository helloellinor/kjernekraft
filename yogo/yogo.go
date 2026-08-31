// Package yogo fetches the schedule from the booking system the studio
// runs today.
//
// The customer-facing page is a single-page app that fetches everything
// from api.yogo.dk, so there is a JSON response behind the list and that
// is what we ask for. Scraping the HTML would be reading a drawing of a
// number we can have directly, and the drawing changes whenever someone
// touches the front end.
//
// The package fetches and converts; it stores nothing. Those are two
// jobs: one can be run as often as you like and does no harm, the other
// writes to the database and should be looked at first. So it returns
// []models.Event and leaves the caller to decide what becomes a series.
//
// No login. The schedule is the same list anyone sees on the booking
// page — no customers, no signups, no personal data.
package yogo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"kjernekraft/models"
)

const (
	// APIbasen er den same for alle studio Yogo driv; kven som spør
	// avgjerd `X-Yogo-Client-ID` og kva `Origin` det kjem frå.
	APIbase = "https://api.yogo.dk"

	// The studio's own site. It is here because the API accepts the request
	// on Origin, and a request without one is a request without a studio.
	Opphav = "https://kjernekraftoslo.yogo.no"

	// Kjernekraft Oslo hjaa Yogo. Talet stend i kvart einaste svar som
	// `client_id`, so det er ikkje gjeta.
	Klientnummer = 265

	// Datoformatet API-en tek imot og gjev att.
	dagformat = "2006-01-02"
)

// Klient is one lookup against Yogo. It carries no state beyond which
// studio it asks for, so build it each time or keep it — same thing.
type Klient struct {
	HTTP     *http.Client
	Base     string
	Opphav   string
	Klientnr int

	// The zone the classes are read in. Yogo gives the day and the time
	// separately — "2026-08-31" and "17:30" — and that *is* the clock on the
	// wall in Oslo. The house stores the same clock (see veggtekst in the
	// database package), so the numbers pass straight through.
	Sone *time.Location
}

// Ny makes a client with sensible settings.
//
// The timeout is not decoration: this call goes to a server we do not
// control, and a request that never returns is worse than one that
// fails — it holds whoever is waiting.
func Ny() (*Klient, error) {
	sone, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		return nil, fmt.Errorf("fann ikkje sona Europe/Oslo: %w", err)
	}
	return &Klient{
		HTTP:     &http.Client{Timeout: 30 * time.Second},
		Base:     APIbase,
		Opphav:   Opphav,
		Klientnr: Klientnummer,
		Sone:     sone,
	}, nil
}

// Val is what can be set on a fetch. The zero value is what you want:
// only classes that run.
type Val struct {
	// MedAvlyste includes cancelled classes. Importing a schedule, you do
	// not want them — a cancelled class is a notice about something not
	// happening, not a class. Comparing two schedules, you do, because the
	// difference is the point.
	MedAvlyste bool
}

// Timar hentar alle timane i spennet, båe dagane med.
//
// Spennet er dagar og ikkje augneblink: Yogo reknar i datoar, og ein
// time høyrer til den dagen han byrjar. Frå og til vert difor klipte
// ned til dagen sin.
func (k *Klient) Timar(ctx context.Context, frå, til time.Time, val Val) ([]models.Event, error) {
	if til.Before(frå) {
		return nil, fmt.Errorf("spennet gjeng baklengs: %s til %s",
			frå.Format(dagformat), til.Format(dagformat))
	}

	spurnad := url.Values{}
	spurnad.Set("startDate", frå.Format(dagformat))
	spurnad.Set("endDate", til.Format(dagformat))
	// The names of the kind, the room and the teacher live in separate
	// tables at Yogo; without this we get only their ids, and an id is not
	// something you can put in a schedule.
	spurnad.Add("populate[]", "class_type")
	spurnad.Add("populate[]", "room")
	spurnad.Add("populate[]", "teachers")
	if !val.MedAvlyste {
		spurnad.Set("excludeCancelledClasses", "true")
	}

	svar, err := k.hent(ctx, "/classes?"+spurnad.Encode())
	if err != nil {
		return nil, err
	}

	var kropp struct {
		Klassar []klasse `json:"classes"`
	}
	if err := json.Unmarshal(svar, &kropp); err != nil {
		return nil, fmt.Errorf("kunde ikkje lesa svaret fraa Yogo: %w", err)
	}

	ut := make([]models.Event, 0, len(kropp.Klassar))
	for _, kl := range kropp.Klassar {
		// Belt and braces: the flag above tells the server the same thing, but a
		// filter we can *see* is a filter we can test.
		if kl.Avlyst && !val.MedAvlyste {
			continue
		}
		time, err := kl.tilEvent(k.Sone)
		if err != nil {
			return nil, fmt.Errorf("time %d (%s %s): %w", kl.ID, kl.Dato, kl.Start, err)
		}
		ut = append(ut, time)
	}
	sorterEtterTid(ut)
	return ut, nil
}

// NesteVeker is "the next stretch": from today and that many whole weeks
// forward.
//
// The week is the unit because that is what a schedule repeats in — a
// series is "yoga Monday 18:00", and it comes true every week. Ask for
// three weeks and you get three repetitions of each series, which is
// exactly what you need to see what is a series and what is an
// exception.
func (k *Klient) NesteVeker(ctx context.Context, no time.Time, veker int, val Val) ([]models.Event, error) {
	if veker < 1 {
		return nil, fmt.Errorf("talet paa vikor lyt vera minst éi, ikkje %d", veker)
	}
	frå := no.In(k.Sone)
	return k.Timar(ctx, frå, frå.AddDate(0, 0, 7*veker-1), val)
}

// hent gjer sjølve kallet.
func (k *Klient) hent(ctx context.Context, veg string) ([]byte, error) {
	sp, err := http.NewRequestWithContext(ctx, http.MethodGet, k.Base+veg, nil)
	if err != nil {
		return nil, err
	}
	// Yogo finds the studio from these three. Without them the API answers
	// for nobody — or for the wrong studio.
	sp.Header.Set("Origin", k.Opphav)
	sp.Header.Set("X-Yogo-Request-Context", "frontend")
	sp.Header.Set("X-Yogo-Client-ID", strconv.Itoa(k.Klientnr))
	// An honest label. Whoever owns the server should be able to see in the
	// log who is asking, and who to call if it is too much.
	sp.Header.Set("User-Agent", "kjernekraft/1.0 (timeplanhenting; post@kjernekraftoslo.no)")
	sp.Header.Set("Accept", "application/json")

	res, err := k.HTTP.Do(sp)
	if err != nil {
		return nil, fmt.Errorf("naadde ikkje Yogo: %w", err)
	}
	defer res.Body.Close()

	// Eit tak på lesingi. Ei vike er kring 125 kB; ti megabyte er rikeleg
	// for eit heilt år og lite nok til at eit svar som gjeng ihop ikkje
	// tek minnet med seg.
	kropp, err := io.ReadAll(io.LimitReader(res.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("kunde ikkje lesa svaret: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Yogo svara %d: %s", res.StatusCode, stutt(kropp))
	}
	return kropp, nil
}

// ---- what Yogo sends ----
//
// Only the fields we use. The response carries four times as many — flags
// for ClassPass, Bruce and Urban Sports Club, images, long descriptions —
// and a field we do not read is a field that cannot drift away from us.

type klasse struct {
	ID     int    `json:"id"`
	Dato   string `json:"date"`       // «2026-08-31»
	Start  string `json:"start_time"` // «17:30»
	Slutt  string `json:"end_time"`   // «18:20»
	Emne   string `json:"subtitle"`   // «Åpent nivå»
	Sete   int    `json:"seats"`
	Avlyst bool   `json:"cancelled"`

	Slag *struct {
		Namn  string `json:"name"`
		Farge string `json:"color"`
	} `json:"class_type"`

	Rom *struct {
		ID   int    `json:"id"`
		Namn string `json:"name"`
	} `json:"room"`

	Laerarar []struct {
		Fornamn   string `json:"first_name"`
		Etternamn string `json:"last_name"`
	} `json:"teachers"`
}

// tilEvent turns a Yogo class into the class the house knows.
func (k klasse) tilEvent(sone *time.Location) (models.Event, error) {
	start, err := klokkeslett(k.Dato, k.Start, sone)
	if err != nil {
		return models.Event{}, err
	}
	slutt, err := klokkeslett(k.Dato, k.Slutt, sone)
	if err != nil {
		return models.Event{}, err
	}
	// A class that ends before it starts runs past midnight. That does not
	// happen in a yoga studio, but a response we do not control must not be
	// able to give a class a negative length.
	if !slutt.After(start) {
		slutt = slutt.AddDate(0, 0, 1)
	}

	namn := reint(k.slagnamn())
	sete := k.Sete
	e := models.Event{
		Title: namn,
		// The name is what the class is *called*; the kind is what it *is*. Yogo
		// has only the first — "Vinyasa Flow" — so the kind is looked up (see
		// slag.go). A name the table does not know gives an empty string, and
		// then the wing goes grey rather than lying (§1).
		ClassType:   Slag(namn),
		Description: reint(k.Emne),
		StartTime:   start,
		EndTime:     slutt,
		TeacherName: k.lærar(),
		Capacity:    k.Sete,
		// Places are the room's here — Yogo has no own/inherited distinction — so
		// the number is the class's own.
		EigenPlassar: &sete,
	}
	if k.Rom != nil {
		e.RoomName = reint(k.Rom.Namn)
		e.Location = e.RoomName
	}
	if k.Slag != nil {
		e.Color = k.Slag.Farge
	}
	return e, nil
}

func (k klasse) slagnamn() string {
	if k.Slag == nil {
		return ""
	}
	return k.Slag.Namn
}

// lærar gives one name.
//
// Yogo allows several teachers on a class; the house has one field,
// because it is one name that stands in the list and on the mark. The
// first is the one holding the class, the rest are assistants, and more
// than one is rare enough that taking the first beats threading a field
// through the whole house for a case that barely exists.
func (k klasse) lærar() string {
	if len(k.Laerarar) == 0 {
		return ""
	}
	l := k.Laerarar[0]
	return reint(strings.TrimSpace(l.Fornamn + " " + l.Etternamn))
}

// klokkeslett byggjer eit tidspunkt av dagen og klokka Yogo gjev kvar
// for seg. Det er klokka på veggen, og ho vert merkt med Oslo.
func klokkeslett(dag, klokke string, sone *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02 15:04", dag+" "+strings.TrimSpace(klokke), sone)
	if err != nil {
		return time.Time{}, fmt.Errorf("ugild tid «%s %s»: %w", dag, klokke, err)
	}
	return t, nil
}

// reint strips whitespace from the ends.
//
// Yogo names carry it: "Fascia Flyt ", "Hatha Yoga ", "Reformer ". A name
// with a trailing space is a different name from the same one without, and
// then one series becomes two — and two different wings in the list.
func reint(s string) string { return strings.TrimSpace(s) }

// stutt trims an error response so it can stand in a message without
// dragging a whole HTML page with it.
func stutt(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}

// sorterEtterTid puts the classes in the order you read them in. Yogo
// gives them in whatever order its database finds them.
func sorterEtterTid(timar []models.Event) {
	for i := 1; i < len(timar); i++ {
		for j := i; j > 0 && timar[j].StartTime.Before(timar[j-1].StartTime); j-- {
			timar[j], timar[j-1] = timar[j-1], timar[j]
		}
	}
}
