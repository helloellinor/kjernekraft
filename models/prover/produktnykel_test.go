// Package prover tests models from outside.
package prover

import (
	"encoding/json"
	"testing"

	"kjernekraft/models"
)

// .ID must exist, and must mean the product.
//
// For a while it did neither: both embedded structs had a field called
// ID, so Go promoted neither and m.ID was a compile error. That this
// test *builds* is half the claim — give the row back the name ID and it
// fails here rather than out on a page.
func TestIDTyderProduktet(t *testing.T) {
	var m models.MembershipWithDetails
	m.Membership.ID = 7
	m.UserMembership.RadID = 42
	if m.ID != 7 {
		t.Errorf("m.ID er %d, venta produktet sin 7", m.ID)
	}

	var k models.KlippekortWithDetails
	k.KlippekortPackage.ID = 3
	k.UserKlippekort.RadID = 99
	if k.ID != 3 {
		t.Errorf("k.ID er %d, venta pakka sin 3", k.ID)
	}
}

// json says nothing when two embedded fields collide; it just drops the
// field. While both were called id, a membership marshalled without any
// id at all.
//
// No route sends these as JSON today. That is why the test is here: the
// first one that does should not be what discovers it.
func TestJSONBerBåeProduktetOgRaden(t *testing.T) {
	var m models.MembershipWithDetails
	m.Membership.ID = 7
	m.UserMembership.RadID = 42

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var ut map[string]any
	if err := json.Unmarshal(b, &ut); err != nil {
		t.Fatal(err)
	}
	for nykel, vil := range map[string]float64{"id": 7, "rad_id": 42} {
		fekk, finst := ut[nykel]
		if !finst {
			t.Errorf("«%s» stend ikkje i JSON-en i det heile", nykel)
			continue
		}
		if fekk != vil {
			t.Errorf("«%s» er %v, venta %v", nykel, fekk, vil)
		}
	}
}
