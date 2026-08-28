package handlers

import "testing"

import "kjernekraft/models"

// Namnet på eit produkt skal koma frå fakta, og ei overstyring skal
// gjelde eitt språk — ikkje alle.
//
// Prøva les dei røynlege omsetjingsfilene, so ho brest om nokon tek bort
// ein av nøklane som namna er bygde av.
func TestNamnKjemFraaFakta(t *testing.T) {
	lastMaali(t) // syt for at umsetjingane er lesne

	prøvor := []struct {
		binding int
		student bool
		nn, en  string
	}{
		{0, false, "Månadskort", "Monthly"},
		{6, false, "Halvårskort", "Six months"},
		{12, false, "Årskort", "Yearly"},
		{3, false, "3 månaders binding", "3 months' binding"},
		{12, true, "Årskort, student og senior", "Yearly, students and seniors"},
	}

	for _, p := range prøvor {
		m := models.Membership{CommitmentMonths: p.binding, IsStudentSenior: p.student}
		if fekk := MedlemskapNamn("nn", m); fekk != p.nn {
			t.Errorf("nn, %d md, student=%v: fekk %q, venta %q", p.binding, p.student, fekk, p.nn)
		}
		if fekk := MedlemskapNamn("en", m); fekk != p.en {
			t.Errorf("en, %d md, student=%v: fekk %q, venta %q", p.binding, p.student, fekk, p.en)
		}
	}
}

func TestOverstyringGjeldEittSpraak(t *testing.T) {
	lastMaali(t)

	m := models.Membership{CommitmentMonths: 12}
	// Studioet har skrive «Hausttilbod» medan dei stod i nynorsk.
	overstyrt := map[string]string{"nn": "Hausttilbod"}

	if fekk := Namn(overstyrt, "nn", MedlemskapNamn("nn", m)); fekk != "Hausttilbod" {
		t.Errorf("nynorsk skulle vera overstyrt, fekk %q", fekk)
	}
	// Engelsk er ikkje rørt, og fell attende på det systemet skriv.
	if fekk := Namn(overstyrt, "en", MedlemskapNamn("en", m)); fekk != "Yearly" {
		t.Errorf("engelsk skulle vera generert, fekk %q", fekk)
	}
	// Eit tomt namn er inga overstyring — det er slik ein tek henne bort.
	if fekk := Namn(map[string]string{"nn": ""}, "nn", "Årskort"); fekk != "Årskort" {
		t.Errorf("tomt namn skulle falla attende, fekk %q", fekk)
	}
}

// Black heiter Black. Namnet vart fyrr rekna ut or bindingi, og med
// null maanader binding kom han ut som «Månadskort» — namnet paa eit
// produkt han ikkje er.
func TestUsynlegMedlemskapHeldNamnetSitt(t *testing.T) {
	// `t()` gjeng gjenom singletonen, som leitar etter «locales»
	// relativt til arbeidskatalogen — pakkekatalogen under prøvor.
	// lastMaali peikar honom mot rota, som i localization_test.go.
	lastMaali(t)

	svart := models.Membership{Name: "Black", CommitmentMonths: 0, Skjult: true}
	for _, p := range []struct{ maal, vil string }{
		{"nn", "Svart"}, {"nb", "Svart"}, {"en", "Black Card"},
	} {
		if got := MedlemskapNamn(p.maal, svart); got != p.vil {
			t.Errorf("%s: venta «%s», fekk «%s»", p.maal, p.vil, got)
		}
	}

	// Eit usynleg medlemskap utan nykel fell attende paa raanamnet, so
	// det syner seg med ein gong i staden for aa teikna ein nykel.
	nytt := models.Membership{Name: "Platinum", Skjult: true}
	if got := MedlemskapNamn("nn", nytt); got != "Platinum" {
		t.Errorf("venta raanamnet «Platinum», fekk «%s»", got)
	}

	// Eit vanleg medlemskap skal framleis fylgja bindingi.
	vanleg := models.Membership{Name: "Basis", CommitmentMonths: 12, Skjult: false}
	if got := MedlemskapNamn("nn", vanleg); got == "Basis" {
		t.Errorf("eit synleg medlemskap skal namngjevast or bindingi, fekk raanamnet «%s»", got)
	}
}
