package handsamarar

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// proveFanerekkje byggjer ei rekkje for ei prøva. `vald` er fana ho skal
// standa paa, og `nyklar` er alle fanone i rekkja.
//
// Ho gjeng gjenom fanerekkje() og ikkje utanum: ei prøva som byggjer
// structen for hand prøver ikkje det sida gjer.
func proveFanerekkje(id, vald string, nyklar ...string) Fanerekkje {
	faner := make([]Tab, len(nyklar))
	for i, k := range nyklar {
		faner[i] = Tab{Key: k, Name: k}
	}
	r := httptest.NewRequest(http.MethodGet, "/?fane="+vald, nil)
	return fanerekkje(r, id, "fane", id, faner)
}

// Ein nykel som ikkje finst er ikkje ein feil. Det er ei gamal lenkja,
// eller ei fana som er burte for denne brukaren — byt-fana finst ikkje
// for eit tildelt medlemskap — og daa lyt fyrste fana svara. Gjorde ho
// ikkje det, teikna sida ingen bolk i det heile.
func TestUkjendFaneFellAttendePaaFyrste(t *testing.T) {
	faner := []Tab{{Key: "korta"}, {Key: "kjop-klipp"}}

	for _, adresse := range []string{
		"/elev/klippekort",
		"/elev/klippekort?fane=",
		"/elev/klippekort?fane=finst-ikkje",
	} {
		r := httptest.NewRequest(http.MethodGet, adresse, nil)
		if got := fanerekkje(r, "a", "fane", "l", faner).Vald; got != "korta" {
			t.Errorf("%s: vald = %q, venta «korta»", adresse, got)
		}
	}

	r := httptest.NewRequest(http.MethodGet, "/elev/klippekort?fane=kjop-klipp", nil)
	if got := fanerekkje(r, "a", "fane", "l", faner).Vald; got != "kjop-klipp" {
		t.Errorf("vald = %q, venta «kjop-klipp»", got)
	}
}

// Lenkjone tek vare paa resten av adressa.
//
// «Fyll paa» fører til `?fane=kjop-klipp&fill=Reformer`, og trykkjer ein
// seg derifraa til korta sine skal `fill` fylgja med — han er ikkje
// fanen sin, han er sida sin. Og andre vegen: ei ytre rekkje nullstiller
// dei indre, so `?fane=folk` ikkje slæpar med seg kva prisfana ein stod
// paa i ei fana ein hev gjenge ut or.
func TestLenkjaTekVarePaaResten(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/elev/klippekort?fane=kjop-klipp&fill=Reformer", nil)
	rekkja := fanerekkje(r, "a", "fane", "l", []Tab{{Key: "korta"}, {Key: "kjop-klipp"}})
	if h := rekkja.Tabs[0].Href; h != "/elev/klippekort?fane=korta&fill=Reformer" {
		t.Errorf("lenkja miste noko: %q", h)
	}

	r = httptest.NewRequest(http.MethodGet, "/admin?fane=prisar&prisfane=reglar", nil)
	ytre := fanerekkje(r, "a", "fane", "l", []Tab{{Key: "prisar"}, {Key: "folk"}}, "prisfane")
	if h := ytre.Tabs[1].Href; h != "/admin?fane=folk" {
		t.Errorf("den ytre rekkja slæpte den indre med seg: %q", h)
	}
}
