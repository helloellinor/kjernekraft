package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Skjemaet lyt lesast same kva drakt kroppen kjem i.
//
// Dette er prøva paa ein feil som gøymde seg godt. `ParseForm` les berre
// `x-www-form-urlencoded`; er kroppen `multipart`, les ho spurjingi i
// adressa og set `r.Form` likevel — og etter det gjer `FormValue` ingen
// ting, av di ho berre parsar naar `r.Form` er nil. Kvart felt kom
// attende tomt og lagringi svara 400 fyre ho hadde teke i noko.
//
// Han synte seg ikkje i noka anna prøve av di CSRF-vakti *ogso* les
// `FormValue`, og dermed parsa kroppen fyrst — men berre naar
// kjennemerket ikkje kjem i ei hovudlinja. `csrf.js` legg det alltid i
// hovudlinja for fetch, so i nettlesaren rørte vakti aldri kroppen.
func TestSkjemaetLesestIBaaeDraktene(t *testing.T) {
	verdi := map[string]string{
		"rule_id": "10", "tittel": "Yoga", "laerar": "Leon",
		"room_id": "1", "vekedag": "3", "klokke": "18:00",
		"minutt": "60", "plassar": "12", "skildring": "noko",
	}

	for _, p := range []struct {
		namn string
		lag  func() *http.Request
	}{
		{"urlkoda", func() *http.Request {
			f := url.Values{}
			for k, v := range verdi {
				f.Set(k, v)
			}
			f.Add("avlys", "2")
			f.Set("vikar-3", "Kristina")
			r := httptest.NewRequest("POST", "/api/admin/rule/lagra", strings.NewReader(f.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			return r
		}},
		{"multipart", func() *http.Request {
			var b bytes.Buffer
			w := multipart.NewWriter(&b)
			for k, v := range verdi {
				w.WriteField(k, v)
			}
			w.WriteField("avlys", "2")
			w.WriteField("vikar-3", "Kristina")
			w.Close()
			r := httptest.NewRequest("POST", "/api/admin/rule/lagra", &b)
			r.Header.Set("Content-Type", w.FormDataContentType())
			return r
		}},
	} {
		t.Run(p.namn, func(t *testing.T) {
			s, err := lesRegelskjema(p.lag())
			if err != nil {
				t.Fatalf("skjemaet lét seg ikkje lesa: %v", err)
			}
			if s.RegelID != 10 || s.Tittel != "Yoga" || s.Laerar != "Leon" {
				t.Errorf("regelen kom feil att: %+v", s)
			}
			if s.RomID != 1 || s.Vekedag != 3 || s.Minutt != 60 || s.Plassar != 12 {
				t.Errorf("tali kom feil att: rom=%d dag=%d min=%d plassar=%d",
					s.RomID, s.Vekedag, s.Minutt, s.Plassar)
			}
			if s.Klokke.Format("15:04") != "18:00" {
				t.Errorf("klokka kom feil att: %s", s.Klokke.Format("15:04"))
			}
			if !s.Avlys[2] {
				t.Error("den avlyste dagen kom ikkje med")
			}
			if s.Vikar[3] != "Kristina" {
				t.Errorf("vikaren kom ikkje med: %v", s.Vikar)
			}
		})
	}
}

// Eit tomt plasstal tyder «rommet raar», ikkje null plassar.
func TestTomtPlasstalErIkkjeNull(t *testing.T) {
	f := url.Values{"rule_id": {"10"}, "tittel": {"Yoga"}, "laerar": {"Leon"},
		"room_id": {"1"}, "vekedag": {"3"}, "klokke": {"18:00"}, "minutt": {"60"},
		"plassar": {""}}
	r := httptest.NewRequest("POST", "/x", strings.NewReader(f.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	s, err := lesRegelskjema(r)
	if err != nil {
		t.Fatal(err)
	}
	if s.Plassar != 0 {
		t.Errorf("tomt plasstal vart %d", s.Plassar)
	}
}
