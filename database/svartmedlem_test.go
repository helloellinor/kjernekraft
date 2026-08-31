package database

import (
	"os"
	"path/filepath"
	"testing"
)

// Lærarrolla gjev Black, utan at nokon hev kjøpt noko.
func TestLærarFårSvartMedlemskap(t *testing.T) {
	db := prøvebase(t)
	id := lagBrukar(t, db, "Ida")

	fyre, err := db.GetUserMembership(id)
	if err != nil {
		t.Fatalf("fyre: %v", err)
	}
	if fyre != nil {
		t.Fatalf("utan løyvet skal ho ikkje hava medlemskap, hadde %q", fyre.Membership.Name)
	}

	if err := db.SettLøyve(id, LøyveLærar, true); err != nil {
		t.Fatalf("SettLoyve: %v", err)
	}
	etter, err := db.GetUserMembership(id)
	if err != nil {
		t.Fatalf("etter: %v", err)
	}
	if etter == nil {
		t.Fatal("læraren fekk ikkje medlemskap")
	}
	if etter.Membership.Name != SvartMedlemskap {
		t.Errorf("venta %q, fekk %q", SvartMedlemskap, etter.Membership.Name)
	}
	if !etter.Tildelt {
		t.Error("medlemskapet skal vera merkt som tildelt")
	}
	if etter.Membership.Price != 0 {
		t.Errorf("Black skal ikkje kosta noko, kosta %d", etter.Membership.Price)
	}
}

// Utviklaren fær det same, men gjenom fila og ikkje gjenom eit løyve.
func TestUtviklarFårSvartMedlemskap(t *testing.T) {
	db := prøvebase(t)
	id := lagBrukar(t, db, "Carl")

	// lagBrukar skriv namn+«@dømet.no» som e-post.
	skrivUtviklarfil(t, "# dei som byggjer huset\nCarl@dømet.no\n")

	m, err := db.GetUserMembership(id)
	if err != nil {
		t.Fatalf("GetUserMembership: %v", err)
	}
	if m == nil || m.Membership.Name != SvartMedlemskap {
		t.Fatalf("utviklaren skal hava Black, fekk %v", m)
	}
}

// skrivUtviklarfil peikar lista mot ei fil prøva eig, og nullstiller
// bufferet etterpå so prøvone ikkje ser kvarandre si fil.
func skrivUtviklarfil(t *testing.T, innhald string) {
	t.Helper()
	stig := filepath.Join(t.TempDir(), "utviklarar")
	if err := os.WriteFile(stig, []byte(innhald), 0o600); err != nil {
		t.Fatalf("skriva utviklarfila: %v", err)
	}
	t.Setenv(UtviklarfilEnv, stig)
	nullstillUtviklarbuffer()
	t.Cleanup(nullstillUtviklarbuffer)
}

// Utviklarlista skal ikkje kunna skrivast gjenom flata. LøyveFinst er
// porten administrasjonen skriv gjenom, og «developer» skal ikkje sleppa
// gjenom honom.
func TestUtviklarErIkkjeEiRollaAdminKannGjeva(t *testing.T) {
	if LøyveFinst("developer") {
		t.Error("ein administrator kunde gjeva ut utviklarløyvet")
	}
}

// Ein som ikkje stend i fila fær ingen ting.
func TestUtanforUtviklarfilaFårIngenTing(t *testing.T) {
	db := prøvebase(t)
	id := lagBrukar(t, db, "Bjørn")
	skrivUtviklarfil(t, "carl@cbcustomhom.es\n")

	m, err := db.GetUserMembership(id)
	if err != nil {
		t.Fatalf("GetUserMembership: %v", err)
	}
	if m != nil {
		t.Errorf("Bjørn stend ikkje i fila og skal ikkje hava %q", m.Membership.Name)
	}
}

// Kommentarar og tome liner er ikkje e-postar.
func TestUtviklarfilaHopparYverKommentarar(t *testing.T) {
	skrivUtviklarfil(t, "# ein kommentar\n\n  carl@cbcustomhom.es  \n")
	if ErUtviklar("# ein kommentar") {
		t.Error("ein kommentar vart lesen som ein utviklar")
	}
	if !ErUtviklar("CARL@CBCUSTOMHOM.ES") {
		t.Error("e-posten skal finnast utan umsyn til store bokstavar og mellomrom")
	}
}

// Det viktige med at medlemskapet er avleidd: tek ein løyvet bort,
// gjeng medlemskapet med i same augneblinken. Ei lagra rad hadde vorte
// liggjande att.
func TestMedlemskapetGjengNårRollaGjeng(t *testing.T) {
	db := prøvebase(t)
	id := lagBrukar(t, db, "Ida")

	if err := db.SettLøyve(id, LøyveLærar, true); err != nil {
		t.Fatalf("paa: %v", err)
	}
	if m, _ := db.GetUserMembership(id); m == nil {
		t.Fatal("oppsettet: læraren skulde hava Black")
	}

	if err := db.SettLøyve(id, LøyveLærar, false); err != nil {
		t.Fatalf("av: %v", err)
	}
	m, err := db.GetUserMembership(id)
	if err != nil {
		t.Fatalf("GetUserMembership: %v", err)
	}
	if m != nil {
		t.Errorf("utan løyvet skal medlemskapet vera burte, fekk %q", m.Membership.Name)
	}
}

// Black skal ikkje stande i lista folk vel or.
func TestSvartMedlemskapErUteAvPrislista(t *testing.T) {
	db := prøvebase(t)

	if _, err := db.Conn.Exec(`
		INSERT INTO memberships (name, price, commitment_months, is_student_senior,
		                         is_special_offer, description, features, active, skjult)
		VALUES ('12-maanader', 104000, 12, FALSE, FALSE, '', '[]', TRUE, FALSE)`); err != nil {
		t.Fatalf("oppsettet: %v", err)
	}

	val, err := db.MedlemskapFor(false)
	if err != nil {
		t.Fatalf("MedlemskapFor: %v", err)
	}
	for _, m := range val {
		if m.Name == SvartMedlemskap {
			t.Fatal("Black stod i lista ein kann velja or")
		}
	}
	if len(val) == 0 {
		t.Error("dei vanlege medlemskapi skal framleis standa der")
	}
}

// Ei ukjend løyvet skal framleis avvisast — lista er ei lista yver noko.
func TestUkjendRollaFinstIkkje(t *testing.T) {
	if LøyveFinst("svartebørs") {
		t.Error("ukjend løyvet vart godteki")
	}
	for _, r := range []string{LøyveAdmin, LøyveLærar} {
		if !LøyveFinst(r) {
			t.Errorf("%q skal vera ei kjend løyvet", r)
		}
	}
}

// Black skal koma attende med skjult-flagget sitt. Utan det rekna
// namnegjevingi honom som eit bindingsprodukt og kalla honom
// «Månadskort».
func TestSvartMedlemskapBerSkjultflagget(t *testing.T) {
	db := prøvebase(t)
	id := lagBrukar(t, db, "Ida")
	if err := db.SettLøyve(id, LøyveLærar, true); err != nil {
		t.Fatalf("SettLoyve: %v", err)
	}
	m, err := db.GetUserMembership(id)
	if err != nil || m == nil {
		t.Fatalf("GetUserMembership: %v", err)
	}
	if !m.Membership.Skjult {
		t.Error("Black kom attende utan skjult-flagget")
	}
}
