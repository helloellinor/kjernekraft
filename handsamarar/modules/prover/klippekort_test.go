// Package prover tests the modules package from outside.
//
// It used to call a private nærastUtløp that did nothing but forward
// to models.NærastUtløp — a wrapper kept alive by its own test. The
// test now asks the function that actually does the work.
package prover

import (
	"testing"

	"kjernekraft/handsamarar/modules"
	"kjernekraft/models"
)

func kort(kategori string, att, dagar int) models.KlippekortWithDetails {
	var k models.KlippekortWithDetails
	k.Category = kategori
	k.RemainingKlipp = att
	k.DaysUntilExpiry = dagar
	return k
}

// What expires first is not what looks smallest.
func TestNærastUtløpVelDetSomGjengUtFyrst(t *testing.T) {
	n := models.NærastUtløp([]models.KlippekortWithDetails{
		kort("Reformer", 20, 120),
		kort("Gruppetimar", 2, 9),
		kort("Online", 8, 40),
	})
	if n == nil {
		t.Fatal("fann ingen ting")
	}
	if n.Category != "Gruppetimar" {
		t.Errorf("fekk %q, venta Gruppetimar", n.Category)
	}
}

// Eit tomt kort som gjeng ut er ikkje ein frist — det er ei kvittering.
// Og eit kort som alt er gjenge ut kann ein ikkje rekka.
func TestNærastUtløpHopparYverTomeOgUtgjengne(t *testing.T) {
	n := models.NærastUtløp([]models.KlippekortWithDetails{
		kort("Tomt", 0, 2),       // ingen klipp att
		kort("Utgjenge", 5, -3),  // fristen er ute
		kort("Gjeldande", 4, 55), // dette er det einaste som tel
	})
	if n == nil {
		t.Fatal("fann ingen ting")
	}
	if n.Category != "Gjeldande" {
		t.Errorf("fekk %q, venta Gjeldande", n.Category)
	}
}

func TestNærastUtløpGjevIngenTingNårDetIkkjeErNoko(t *testing.T) {
	if n := models.NærastUtløp(nil); n != nil {
		t.Errorf("fekk %v for ei tom lista", n)
	}
	if n := models.NærastUtløp([]models.KlippekortWithDetails{kort("Tomt", 0, 5)}); n != nil {
		t.Error("eit tomt kort skal ikkje vera ein frist")
	}
}

// An empty typed list is not "has cards".
//
// This was a real bug: the builder took interface{}, guessed with a type
// switch, and fell through to true for any typed slice — so the page
// drew the card layout with nothing in it. The list is typed now; the
// test stays because the answer must not change.
func TestTomListaErIkkjeHevKort(t *testing.T) {
	d := modules.NewKlippekortModule([]models.KlippekortWithDetails{}, "nn")
	if d.HasKlippekort {
		t.Error("ei tom lista vart rekna som «hev kort»")
	}
	d = modules.NewKlippekortModule([]models.KlippekortWithDetails{kort("Gruppetimar", 3, 20)}, "nn")
	if !d.HasKlippekort {
		t.Error("ei lista med eitt kort i vart ikkje rekna som «hev kort»")
	}
	if d.Nærast == nil {
		t.Error("fann ikkje det næraste kortet")
	}
}
