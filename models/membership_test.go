package models

import (
	"testing"
	"time"
)

// Utlaupet på kortet. Fire greiner og ei klokka, og kvar prøva her
// svarar til ei linja i kommentaren yver GjeldTil.
func TestGjeldTil(t *testing.T) {
	nå := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	fornying := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	binding := time.Date(2027, 3, 11, 0, 0, 0, 0, time.UTC)

	kort := func(stoda string, maanader int, medBinding bool) MembershipWithDetails {
		var m MembershipWithDetails
		m.Status = stoda
		m.CommitmentMonths = maanader
		m.RenewalDate = fornying
		if medBinding {
			b := binding
			m.BindingEnd = &b
		}
		return m
	}

	prov := []struct {
		namn string
		m    MembershipWithDetails
		vil  *time.Time
	}{
		{"maanadskort varer ein maanad", kort("active", 0, false), &fornying},
		{"aarskort varer tolv maanader", kort("active", 12, true), &binding},
		{"uppsagt aarskort varer ut det du hev betalt", kort("cancelled", 12, true), &fornying},
		{"uppsagt maanadskort med", kort("cancelled", 0, false), &fornying},
	}
	for _, p := range prov {
		if fekk := p.m.GjeldTil(nå); fekk == nil || !fekk.Equal(*p.vil) {
			t.Errorf("%s: fekk %v, vilde ha %v", p.namn, fekk, *p.vil)
		}
	}

	// Tildelt hev ingen dato i det heile.
	tildelt := kort("active", 0, false)
	tildelt.Tildelt = true
	if fekk := tildelt.GjeldTil(nå); fekk != nil {
		t.Errorf("tildelt skulde ikkje ha nokon dato, fekk %v", fekk)
	}

	// Klokka: eit aarskort som hev stade frose i tredive dagar hev tredive
	// dagar meir att enn det som stend i basen.
	frose := kort("paused", 12, true)
	frå := nå.AddDate(0, 0, -30)
	frose.FrozenAt = &frå
	vil := binding.AddDate(0, 0, 30)
	if fekk := frose.GjeldTil(nå); fekk == nil || !fekk.Equal(vil) {
		t.Errorf("frose i 30 dagar: fekk %v, vilde ha %v", fekk, vil)
	}

	// Ei frysing som ikkje er godkjend enno stoggar ingi klokka: stoda
	// er 'freeze_requested', og medlemskapet gjeng som vanleg til
	// studioet svarar.
	bede := kort("freeze_requested", 12, true)
	bede.FrozenAt = &frå
	if fekk := bede.GjeldTil(nå); fekk == nil || !fekk.Equal(binding) {
		t.Errorf("venteande frysing skal ikkje skuva utlaupet, fekk %v", fekk)
	}

	// Eit frose medlemskap utan klokka — det stod frose fyre kolonna
	// fanst — skal ikkje gissast på.
	utanKlokka := kort("paused", 12, true)
	if fekk := utanKlokka.GjeldTil(nå); fekk == nil || !fekk.Equal(binding) {
		t.Errorf("frose utan frozen_at: fekk %v, vilde ha %v", fekk, binding)
	}
}
