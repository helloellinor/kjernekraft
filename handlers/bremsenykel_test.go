package handlers

import (
	"net/http/httptest"
	"testing"
)

// Innloggingsbremsa lyt vita kven ho bremsar.
//
// Ho talde per `RemoteAddr`. I drift kjem alt gjenom ein tunnel som
// koplar seg til 127.0.0.1, so kvar einaste brukar hadde den same
// nykelen: ti bomma passord fraa kven som helst stengde *heile* huset
// ute i femtan minutt. Det er ikkje ei tryggingsbremsa, det er ein
// av-knapp kven som helst kann trykkja paa.
//
// Men hovudlina kann ikkje trupast fraa kven som helst heller — daa
// skriv ein motstandar berre ei ny adressa for kvart forsøk og vert
// aldri bremsa. Regelen er difor: tru paa henne naar ho kjem gjenom
// vaar eigen mellomtenar (loopback), og aldri elles.
func TestBremsenykelenSkilKundarBakSameTunnelen(t *testing.T) {
	soknad := func(fjern string, hovudlinor map[string]string) string {
		r := httptest.NewRequest("POST", "/innlogging", nil)
		r.RemoteAddr = fjern
		for k, v := range hovudlinor {
			r.Header.Set(k, v)
		}
		return clientKey(r)
	}

	// Bak tunnelen: tvo kundar med kvar sin nykel, ikkje éin felles.
	ein := soknad("127.0.0.1:54321", map[string]string{"CF-Connecting-IP": "203.0.113.7"})
	tvo := soknad("127.0.0.1:54322", map[string]string{"CF-Connecting-IP": "198.51.100.9"})
	if ein == tvo {
		t.Errorf("tvo kundar bak den same tunnelen fekk den same nykelen (%q) — "+
			"daa stengjer den eine den hine ute", ein)
	}
	if ein != "203.0.113.7" {
		t.Errorf("nykelen bak tunnelen er %q, venta klientadressa 203.0.113.7", ein)
	}

	// X-Forwarded-For gjeld ogso, og det er den *fyrste* som er klienten.
	xff := soknad("127.0.0.1:54323", map[string]string{"X-Forwarded-For": "203.0.113.7, 10.0.0.1"})
	if xff != "203.0.113.7" {
		t.Errorf("X-Forwarded-For gav %q, venta 203.0.113.7", xff)
	}

	// Utanfraa er hovudlina ei paastand fraa ein framand, og ho skal
	// ikkje truast: elles skriv han ei ny adressa for kvart forsøk og
	// vert aldri bremsa.
	framand := soknad("203.0.113.99:40000", map[string]string{
		"CF-Connecting-IP": "1.1.1.1",
		"X-Forwarded-For":  "2.2.2.2",
	})
	if framand != "203.0.113.99" {
		t.Errorf("nykelen for ein framand er %q, venta den verkelege adressa "+
			"203.0.113.99 — ei diktia hovudlina slepp honom forbi bremsa", framand)
	}

	// Kjem det ingi hovudlina gjenom tunnelen, fell me attende paa det
	// me hev. Det er den gamle aatferdi, og ho er betre enn ingi bremsa
	// — e-postnykelen stend framleis.
	utan := soknad("127.0.0.1:54324", nil)
	if utan != "127.0.0.1" {
		t.Errorf("utan hovudlina vart nykelen %q, venta 127.0.0.1", utan)
	}
}
