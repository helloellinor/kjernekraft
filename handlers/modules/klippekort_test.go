package modules

import (
	"kjernekraft/models"
	"testing"
)

func kort(kategori string, att, dagar int) models.KlippekortWithDetails {
	var k models.KlippekortWithDetails
	k.Category = kategori
	k.RemainingKlipp = att
	k.DaysUntilExpiry = dagar
	return k
}

// Kva som gjeng ut fyrst er ikkje det same som kva som ser minst ut.
func TestNaermastUtlopVelDetSomGjengUtFyrst(t *testing.T) {
	korta := []models.KlippekortWithDetails{
		kort("Reformer", 20, 120),
		kort("Gruppetimar", 2, 9),
		kort("Online", 8, 40),
	}
	n := naermastUtlop(korta)
	if n == nil {
		t.Fatal("fann ingen ting")
	}
	if n.Category != "Gruppetimar" {
		t.Errorf("fekk %q, venta Gruppetimar", n.Category)
	}
}

// Eit tomt kort som gjeng ut er ikkje ein frist — det er ei kvittering.
// Og eit kort som alt er gjenge ut kann ein ikkje rekka.
func TestNaermastUtlopHopparYverTomeOgUtgjengne(t *testing.T) {
	n := naermastUtlop([]models.KlippekortWithDetails{
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

func TestNaermastUtlopGjevIngenTingNaarDetIkkjeErNoko(t *testing.T) {
	if n := naermastUtlop(nil); n != nil {
		t.Errorf("fekk %v for ei tom lista", n)
	}
	if n := naermastUtlop([]models.KlippekortWithDetails{kort("Tomt", 0, 5)}); n != nil {
		t.Error("eit tomt kort skal ikkje vera ein frist")
	}
}

// Ei tom, typa lista er ikkje «hev kort». Sjekken i NyKlippekortModule
// fall til `default: true` for kvar typa lista, tom eller ei — so sida
// synte kortoppsettet med null kort i.
func TestTomListaErIkkjeHevKort(t *testing.T) {
	d, err := NewKlippekortModule([]models.KlippekortWithDetails{}, "nn")
	if err != nil {
		t.Fatal(err)
	}
	if d.HasKlippekort {
		t.Error("ei tom lista vart rekna som «hev kort»")
	}
	d, _ = NewKlippekortModule([]models.KlippekortWithDetails{kort("Gruppetimar", 3, 20)}, "nn")
	if !d.HasKlippekort {
		t.Error("ei lista med eitt kort i vart ikkje rekna som «hev kort»")
	}
	if d.Naermast == nil {
		t.Error("fann ikkje det næraste kortet")
	}
}
