package yogo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"kjernekraft/models"
)

// The tests do not go out on the network. The fixture is a real response
// from api.yogo.dk, cut down to four classes — an ordinary one, one with
// spaces in the name, one without a teacher and one cancelled. A test that
// calls a server we do not control is not a test; it is a notification that
// somebody else changed something.
func fixtur(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/veke.json")
	if err != nil {
		t.Fatalf("fixtur: %v", err)
	}
	return b
}

// tenar gives a Yogo that answers with the fixture, and reports what was asked.
func tenar(t *testing.T, kropp []byte) (*Klient, *string) {
	t.Helper()
	var spurd string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spurd = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		w.Write(kropp)
	}))
	t.Cleanup(ts.Close)

	k, err := Ny()
	if err != nil {
		t.Fatalf("Ny: %v", err)
	}
	k.Base = ts.URL
	return k, &spurd
}

// What comes out are classes the house can use: name, teacher, room, time.
func TestTimaneKjemUtSomHusetKjennerDeim(t *testing.T) {
	k, _ := tenar(t, fixtur(t))

	timar, err := k.Timar(context.Background(),
		dag(2026, 8, 31), dag(2026, 9, 6), Val{})
	if err != nil {
		t.Fatalf("Timar: %v", err)
	}

	// Den avlyste er ikkje med. Nullverdet i `Val` er «berre det som gjeng».
	if len(timar) != 3 {
		t.Fatalf("venta tri timar utan den avlyste, fann %d", len(timar))
	}

	// Sortert på tid: 10:00 fyre 17:30 same dagen, og dagen etter sist.
	for i := 1; i < len(timar); i++ {
		if timar[i].StartTime.Before(timar[i-1].StartTime) {
			t.Errorf("timane stend ikkje i tidsrekkjefylgd: %v fyre %v",
				timar[i-1].StartTime, timar[i].StartTime)
		}
	}

	fyrste := timar[0]
	if fyrste.Title != "Fascia Flyt" {
		t.Errorf("tittel = %q; mellomromet bak skulde vore vaska burt", fyrste.Title)
	}
	// Namnet er kva han heiter; slaget er kva han er. «Fascia Flyt» er
	// fascia, og det ordet stend ikkje i namnet — det er slege upp.
	if fyrste.ClassType != SlagFascia {
		t.Errorf("slaget = %q, venta %q", fyrste.ClassType, SlagFascia)
	}
	if fyrste.TeacherName != "Gry Hamre Larsen" && fyrste.TeacherName == "" {
		t.Errorf("læraren mangla heilt: %q", fyrste.TeacherName)
	}
	if fyrste.RoomName == "" || fyrste.Location != fyrste.RoomName {
		t.Errorf("rommet = %q / %q", fyrste.RoomName, fyrste.Location)
	}
}

// Klokka er klokka på veggen i Oslo. Yogo gjev «2026-08-31» og «17:30»
// kvar for seg, og dei tali skal koma uendra ut — flyt dei ein time, er
// heile timeplanen ein time gale halve aaret.
func TestKlokkaErKlokkaPåVeggen(t *testing.T) {
	k, _ := tenar(t, fixtur(t))

	timar, err := k.Timar(context.Background(), dag(2026, 8, 31), dag(2026, 9, 6), Val{})
	if err != nil {
		t.Fatalf("Timar: %v", err)
	}

	var funne bool
	for _, e := range timar {
		if e.StartTime.Format("2006-01-02 15:04") == "2026-08-31 17:30" {
			funne = true
			if e.EndTime.Format("15:04") != "18:20" {
				t.Errorf("slutt = %s, venta 18:20", e.EndTime.Format("15:04"))
			}
			if namn := e.StartTime.Location().String(); namn != "Europe/Oslo" {
				t.Errorf("sona = %s, venta Europe/Oslo", namn)
			}
			if e.LengdMin() != 50 {
				t.Errorf("lengd = %d min, venta 50", e.LengdMin())
			}
		}
	}
	if !funne {
		t.Error("fann ikkje timen 31.8. 17:30 — klokka hev flutt seg")
	}
}

// Ein time utan lærar er ikkje ein feil. Han er ein time me ikkje veit
// kven held, og det skal han faa segja.
func TestTimeUtanLærarGjengGjenom(t *testing.T) {
	k, _ := tenar(t, fixtur(t))

	timar, err := k.Timar(context.Background(), dag(2026, 8, 31), dag(2026, 9, 6), Val{})
	if err != nil {
		t.Fatalf("Timar: %v", err)
	}
	var utan int
	for _, e := range timar {
		if e.TeacherName == "" {
			utan++
		}
	}
	if utan != 1 {
		t.Errorf("venta éin time utan lærar, fann %d", utan)
	}
}

// Cancelled classes come along when you ask for them.
func TestAvlysteKjemMedNårEinBedUmDeim(t *testing.T) {
	k, spurd := tenar(t, fixtur(t))

	timar, err := k.Timar(context.Background(), dag(2026, 8, 31), dag(2026, 9, 6),
		Val{MedAvlyste: true})
	if err != nil {
		t.Fatalf("Timar: %v", err)
	}
	if len(timar) != 4 {
		t.Errorf("venta fire timar med den avlyste, fann %d", len(timar))
	}
	if got := *spurd; contains(got, "excludeCancelledClasses") {
		t.Errorf("bad tenaren um aa sila burt avlyste likevel: %s", got)
	}
}

// The request carries what the API needs to answer with names rather than
// ids. Without populate[] the response is a list of numbers.
func TestSpurnadenBerDatoarOgPopulate(t *testing.T) {
	k, spurd := tenar(t, fixtur(t))

	if _, err := k.Timar(context.Background(), dag(2026, 8, 31), dag(2026, 9, 6), Val{}); err != nil {
		t.Fatalf("Timar: %v", err)
	}
	for _, vil := range []string{
		"startDate=2026-08-31", "endDate=2026-09-06",
		"class_type", "room", "teachers", "excludeCancelledClasses=true",
	} {
		if !contains(*spurd, vil) {
			t.Errorf("spurnaden mangla %q: %s", vil, *spurd)
		}
	}
}

// "The next stretch" is whole weeks from today: three weeks is twenty-one
// days, not twenty-two. An off-by-one here gives a fourth repetition of each
// series, and then you are looking for an exception that does not exist.
func TestNesteVekerSpennerHeileVikor(t *testing.T) {
	k, spurd := tenar(t, fixtur(t))

	no := time.Date(2026, 8, 31, 9, 0, 0, 0, k.Sone)
	if _, err := k.NesteVeker(context.Background(), no, 3, Val{}); err != nil {
		t.Fatalf("NesteVeker: %v", err)
	}
	if !contains(*spurd, "startDate=2026-08-31") || !contains(*spurd, "endDate=2026-09-20") {
		t.Errorf("tri vikor fraa 31.8. skal enda 20.9., ikkje: %s", *spurd)
	}

	if _, err := k.NesteVeker(context.Background(), no, 0, Val{}); err == nil {
		t.Error("null vikor er ikkje eit tidsrom, og skulde vore ein feil")
	}
}

// Eit spenn som gjeng baklengs er ein skrivefeil hjaa kallaren, ikkje
// eit tomt svar. Det skal segjast fyre me plagar tenaren.
func TestBaklengsSpennErEinFeil(t *testing.T) {
	k, spurd := tenar(t, fixtur(t))

	if _, err := k.Timar(context.Background(), dag(2026, 9, 6), dag(2026, 8, 31), Val{}); err == nil {
		t.Error("venta ein feil paa eit baklengs spenn")
	}
	if *spurd != "" {
		t.Errorf("spurde tenaren likevel: %s", *spurd)
	}
}

// Ein tenar som svarar noko anna enn 200 skal segja det, ikkje gjeva ei
// tom liste. Ei tom liste er eit gyldig svar — «ingen timar den vika» —
// og ein feil som ser ut som eit gyldig svar er den verste sorten.
func TestFeilFråTenarenVertEinFeil(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}))
	defer ts.Close()

	k, err := Ny()
	if err != nil {
		t.Fatalf("Ny: %v", err)
	}
	k.Base = ts.URL

	timar, err := k.Timar(context.Background(), dag(2026, 8, 31), dag(2026, 9, 6), Val{})
	if err == nil {
		t.Fatalf("venta ein feil, fekk %d timar", len(timar))
	}
	if !contains(err.Error(), "403") {
		t.Errorf("feilen seier ikkje kva tenaren svara: %v", err)
	}
}

// Fixturen er eit ekte svar. Ryk skjemaet hjaa Yogo — eit felt som
// skiftar namn — er det denne som fyrst seier ifrå.
func TestFixturenLiknarEitEkteSvar(t *testing.T) {
	var kropp struct {
		Klassar []map[string]any `json:"classes"`
	}
	if err := json.Unmarshal(fixtur(t), &kropp); err != nil {
		t.Fatalf("fixturen er ikkje gild JSON: %v", err)
	}
	if len(kropp.Klassar) == 0 {
		t.Fatal("fixturen er tom")
	}
	for _, felt := range []string{"date", "start_time", "end_time", "seats", "cancelled", "class_type", "room", "teachers"} {
		if _, ok := kropp.Klassar[0][felt]; !ok {
			t.Errorf("fixturen manglar feltet %q — skjemaet hjaa Yogo kann ha skift", felt)
		}
	}
}

func dag(år int, m time.Month, d int) time.Time {
	return time.Date(år, m, d, 0, 0, 0, 0, time.UTC)
}

func contains(s, del string) bool {
	return len(del) == 0 || (len(s) >= len(del) && indexOf(s, del) >= 0)
}

func indexOf(s, del string) int {
	for i := 0; i+len(del) <= len(s); i++ {
		if s[i:i+len(del)] == del {
			return i
		}
	}
	return -1
}

var _ = models.Event{}
