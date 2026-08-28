// Package yogo hentar timeplanen fraa bokingsystemet studioet driv i dag.
//
// Kjernekraft bokar gjenom Yogo. Sida kundane ser er
// `kjernekraftoslo.yogo.no`, og ho er ein einsides-app som sjølv hentar
// alt fraa `api.yogo.dk` — so det finst eit JSON-svar bak lista, og det
// er det me spør um. Aa skrapa HTML-en hadde vore aa lesa ei teikning av
// det talet me kann faa direkte, og teikningi endrar seg kvar gong nokon
// rører framsida.
//
// Pakka *hentar* og gjer om; ho lagrar ingen ting. Det er tvo jobbar:
// den eine kann køyrast so tidt ein vil og gjer ingen skade, den andre
// skriv i basen og skal sjaast paa fyrst. Difor gjev ho att
// `[]models.Event` — det huset alt bruker — og let kallaren avgjera kva
// som skal verta ein serie.
//
// Ingi innlogging. Timeplanen er den same lista kven som helst ser paa
// bokingsida, og me spør berre etter henne — ingen kundar, ingen
// paameldingar, ingen personopplysningar.
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
	// avgjerd `X-Yogo-Client-ID` og kva `Origin` det kjem fraa.
	APIbase = "https://api.yogo.dk"

	// Studioet si eigi side. Han stend her av di API-en godtek spurnaden
	// paa `Origin`, og ein spurnad utan honom er ein spurnad utan studio.
	Opphav = "https://kjernekraftoslo.yogo.no"

	// Kjernekraft Oslo hjaa Yogo. Talet stend i kvart einaste svar som
	// `client_id`, so det er ikkje gjeta.
	Klientnummer = 265

	// Datoformatet API-en tek imot og gjev att.
	dagformat = "2006-01-02"
)

// Klient er ein oppslag mot Yogo. Han ber ingen tilstand utanum kva
// studio han spør for, so ein kann laga honom kvar gong eller halda paa
// honom — det er det same.
type Klient struct {
	HTTP     *http.Client
	Base     string
	Opphav   string
	Klientnr int

	// Sona timane vert lesne i. Yogo gjev dagen og klokka kvar for seg —
	// «2026-08-31» og «17:30» — og det *er* klokka paa veggen i Oslo.
	// Huset lagrar den same klokka (sjaa `veggtekst` i database-pakka),
	// so tali gjeng heilt gjenom utan aa flytta seg.
	Sone *time.Location
}

// Ny lagar ein klient for Kjernekraft med rimelege innstillingar.
//
// Tidsgrensa er ikkje pynt: dette kallet gjeng mot ein tenar me ikkje
// raar yver, og ein spurnad som aldri kjem attende er verre enn ein som
// feilar — han held den som ventar.
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

// Val er det som kann stillast paa ei henting. Nullverdet er det ein
// vil ha: berre timar som gjeng.
type Val struct {
	// MedAvlyste tek med timar som er avlyste. Skal ein *importera* ein
	// timeplan, vil ein ikkje ha deim — ein avlyst time er ei melding um
	// noko som ikkje hender, ikkje ein time. Skal ein *samanlikna* tvo
	// planar, vil ein det, av di skilnaden er poenget.
	MedAvlyste bool
}

// Timar hentar alle timane i spennet, baae dagane med.
//
// Spennet er dagar og ikkje augneblink: Yogo reknar i datoar, og ein
// time høyrer til den dagen han byrjar. Fraa og til vert difor klipte
// ned til dagen sin.
func (k *Klient) Timar(ctx context.Context, fraa, til time.Time, val Val) ([]models.Event, error) {
	if til.Before(fraa) {
		return nil, fmt.Errorf("spennet gjeng baklengs: %s til %s",
			fraa.Format(dagformat), til.Format(dagformat))
	}

	spurnad := url.Values{}
	spurnad.Set("startDate", fraa.Format(dagformat))
	spurnad.Set("endDate", til.Format(dagformat))
	// Namnet paa slaget, rommet og læraren bur i eigne tabellar hjaa
	// Yogo; utan dette fær me berre id-ane deira, og ein id er ikkje
	// noko ein kann setja i ein timeplan.
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
		// Belte og bukseseler: flagget yver seier det same til tenaren,
		// men eit filter me *ser* er eit filter me kann prøva.
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

// NesteVeker er «det neste tidsrommet»: fraa i dag og so mange heile
// vikor fram.
//
// Vika er eininga av di det er henne timeplanen gjentek seg i — ein
// serie er «yoga maandag 18:00», og han sannar seg kvar vike. Ein spør
// etter tri vikor og fær tri gjentak av kvar serie, som er nettupp det
// ein treng for aa sjaa kva som er ein serie og kva som er eit unnatak.
func (k *Klient) NesteVeker(ctx context.Context, no time.Time, veker int, val Val) ([]models.Event, error) {
	if veker < 1 {
		return nil, fmt.Errorf("talet paa vikor lyt vera minst éi, ikkje %d", veker)
	}
	fraa := no.In(k.Sone)
	return k.Timar(ctx, fraa, fraa.AddDate(0, 0, 7*veker-1), val)
}

// hent gjer sjølve kallet.
func (k *Klient) hent(ctx context.Context, veg string) ([]byte, error) {
	sp, err := http.NewRequestWithContext(ctx, http.MethodGet, k.Base+veg, nil)
	if err != nil {
		return nil, err
	}
	// Yogo finn studioet paa desse tri. Utan deim svarar API-en for
	// ingen — eller for feil studio.
	sp.Header.Set("Origin", k.Opphav)
	sp.Header.Set("X-Yogo-Request-Context", "frontend")
	sp.Header.Set("X-Yogo-Client-ID", strconv.Itoa(k.Klientnr))
	// Ei ærleg merkelapp. Den som eig tenaren skal kunna sjaa i loggen
	// kven som spør, og kven han skal ringja um det er for mykje.
	sp.Header.Set("User-Agent", "kjernekraft/1.0 (timeplanhenting; post@kjernekraftoslo.no)")
	sp.Header.Set("Accept", "application/json")

	res, err := k.HTTP.Do(sp)
	if err != nil {
		return nil, fmt.Errorf("naadde ikkje Yogo: %w", err)
	}
	defer res.Body.Close()

	// Eit tak paa lesingi. Ei vike er kring 125 kB; ti megabyte er rikeleg
	// for eit heilt aar og lite nok til at eit svar som gjeng ihop ikkje
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

// ---- det Yogo sender ----
//
// Berre felti me nyttar. Svaret ber fire gonger so mange — flagg for
// ClassPass, Bruce og Urban Sports Club, bilete, lange skildringar — og
// eit felt me ikkje les er eit felt som ikkje kann driva fraa oss.

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

// tilEvent gjer ein Yogo-time um til den timen huset kjenner.
func (k klasse) tilEvent(sone *time.Location) (models.Event, error) {
	start, err := klokkeslett(k.Dato, k.Start, sone)
	if err != nil {
		return models.Event{}, err
	}
	slutt, err := klokkeslett(k.Dato, k.Slutt, sone)
	if err != nil {
		return models.Event{}, err
	}
	// Ein time som sluttar fyre han byrjar gjeng yver midnatt. Det
	// hender ikkje i eit yogastudio, men eit svar me ikkje raar yver
	// skal ikkje kunna gjeva ein time med negativ lengd.
	if !slutt.After(start) {
		slutt = slutt.AddDate(0, 0, 1)
	}

	namn := reint(k.slagnamn())
	e := models.Event{
		Title: namn,
		// Namnet er kva timen *heiter*; slaget er kva han *er*. Yogo hev
		// berre det fyrste — «Vinyasa Flow» — so slaget vert slege upp
		// (sjaa slag.go). Eit namn tabellen ikkje kjenner gjev tom
		// streng, og daa vert vengen graa i staden for aa lyga (§1).
		ClassType:   Slag(namn),
		Description: reint(k.Emne),
		StartTime:   start,
		EndTime:     slutt,
		TeacherName: k.laerar(),
		Capacity:    k.Sete,
		// Plassane er rommet sitt her — Yogo hev inga eigi/arva-skiljing —
		// so talet er timen sitt eige.
		EigenPlassar: k.Sete,
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

// laerar gjev eitt namn.
//
// Yogo let fleire lærarar staa paa ein time; huset hev eitt felt, av di
// det er eitt namn som stend i lista og paa merket. Fyrste namnet er den
// som held timen — dei hine er assistentar — og fleire enn éin er
// sjeldan nok til at det er betre aa taka den fyrste enn aa føra eit
// felt til gjenom heile huset for eit tilfelle som mest ikkje finst.
func (k klasse) laerar() string {
	if len(k.Laerarar) == 0 {
		return ""
	}
	l := k.Laerarar[0]
	return reint(strings.TrimSpace(l.Fornamn + " " + l.Etternamn))
}

// klokkeslett byggjer eit tidspunkt av dagen og klokka Yogo gjev kvar
// for seg. Det er klokka paa veggen, og ho vert merkt med Oslo.
func klokkeslett(dag, klokke string, sone *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02 15:04", dag+" "+strings.TrimSpace(klokke), sone)
	if err != nil {
		return time.Time{}, fmt.Errorf("ugild tid «%s %s»: %w", dag, klokke, err)
	}
	return t, nil
}

// reint vaskar mellomrom or endane.
//
// Yogo-namni ber deim: «Fascia Flyt », «Hatha Yoga », «Reformer ». Eit
// namn med eit mellomrom bak er eit anna namn enn det same utan, og daa
// vert det tvo seriar av éin — og tvo ulike vengar i lista.
func reint(s string) string { return strings.TrimSpace(s) }

// stutt klipper eit feilsvar so det kann staa i ei feilmelding utan aa
// taka med seg ei heil HTML-side.
func stutt(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}

// sorterEtterTid set timane i den rekkjefylgda ein les deim i.
// Yogo gjev deim i den rekkjefylgda basen hans finn deim.
func sorterEtterTid(timar []models.Event) {
	for i := 1; i < len(timar); i++ {
		for j := i; j > 0 && timar[j].StartTime.Before(timar[j-1].StartTime); j-- {
			timar[j], timar[j-1] = timar[j-1], timar[j]
		}
	}
}
