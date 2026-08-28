package database

import (
	"testing"
	"time"
)

// Ei frysing stoggar klokka. Eit aarskort som stend frose i tredive
// dagar skal ha tredive dagar meir att naar det vert sett i gang att —
// elles kostar frysingi medlemen den tidi ho varde, og daa er tolv
// maanader ikkje tolv maanader.
func TestFrysingSkuverUtlaupet(t *testing.T) {
	db := prøvebase(t)
	id := lagBrukar(t, db, "Ida")

	res, err := db.Conn.Exec(`INSERT INTO memberships (name, price, commitment_months, description, features, active)
	                          VALUES ('Aarskort', 69000, 12, '', '', TRUE)`)
	if err != nil {
		t.Fatalf("medlemskapet: %v", err)
	}
	medlemskapID, _ := res.LastInsertId()

	fornying := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	binding := time.Date(2027, 3, 11, 0, 0, 0, 0, time.UTC)
	frosenSidan := time.Now().AddDate(0, 0, -30)
	_, err = db.Conn.Exec(`INSERT INTO user_memberships
	    (user_id, membership_id, status, start_date, renewal_date, binding_end, frozen_at)
	    VALUES (?, ?, 'paused', ?, ?, ?, ?)`,
		id, medlemskapID, time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC), fornying, binding, frosenSidan)
	if err != nil {
		t.Fatalf("medlemskapet aat brukaren: %v", err)
	}

	if err := db.UnfreezeMembership(id); err != nil {
		t.Fatalf("UnfreezeMembership: %v", err)
	}

	m, err := db.GetUserMembership(id)
	if err != nil || m == nil {
		t.Fatalf("las ikkje medlemskapet att: %v", err)
	}
	if m.Status != "active" {
		t.Errorf("stoda skulde vera active, var %q", m.Status)
	}
	if m.FrozenAt != nil {
		t.Errorf("klokka skulde vera nullstelt, var %v", *m.FrozenAt)
	}

	// Tredive dagar, med ein monns slingring for tidi prøva sjølv tek.
	sjekk := func(namn string, fekk, fyre time.Time) {
		t.Helper()
		skuv := fekk.Sub(fyre)
		if skuv < 29*24*time.Hour || skuv > 31*24*time.Hour {
			t.Errorf("%s vart skuva %v, venta kring 30 dagar", namn, skuv)
		}
	}
	sjekk("fornyingi", m.RenewalDate, fornying)
	if m.BindingEnd == nil {
		t.Fatal("bindingi vart burte")
	}
	sjekk("bindingi", *m.BindingEnd, binding)
}

// Eit medlemskap som ikkje er frose skal ikkje skuvast. Vegen ut or
// 'freeze_requested' gjeng gjenom den vanlege stoda-endringi, og
// UnfreezeMembership skal ikkje gjera skade um han vert kalla der.
func TestUtanFrysingVertIngenTingSkuva(t *testing.T) {
	db := prøvebase(t)
	id := lagBrukar(t, db, "Gry")

	res, err := db.Conn.Exec(`INSERT INTO memberships (name, price, commitment_months, description, features, active)
	                          VALUES ('Maanadskort', 49000, 0, '', '', TRUE)`)
	if err != nil {
		t.Fatalf("medlemskapet: %v", err)
	}
	medlemskapID, _ := res.LastInsertId()

	fornying := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	_, err = db.Conn.Exec(`INSERT INTO user_memberships
	    (user_id, membership_id, status, start_date, renewal_date)
	    VALUES (?, ?, 'freeze_requested', ?, ?)`,
		id, medlemskapID, time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC), fornying)
	if err != nil {
		t.Fatalf("medlemskapet aat brukaren: %v", err)
	}

	if err := db.UnfreezeMembership(id); err != nil {
		t.Fatalf("UnfreezeMembership: %v", err)
	}
	m, err := db.GetUserMembership(id)
	if err != nil || m == nil {
		t.Fatalf("las ikkje medlemskapet att: %v", err)
	}
	if m.Status != "active" {
		t.Errorf("stoda skulde vera active, var %q", m.Status)
	}
	if !m.RenewalDate.Equal(fornying) {
		t.Errorf("fornyingi vart flutt til %v, skulde stade i %v", m.RenewalDate, fornying)
	}
}
