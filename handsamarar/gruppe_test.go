package handsamarar

import (
	"os"
	"strings"
	"testing"

	"kjernekraft/database"
)

func teiknFolk(t *testing.T, folk []database.Person, grupper []database.Gruppe) string {
	t.Helper()
	gamal, _ := os.Getwd()
	if err := os.Chdir(".."); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(gamal)

	tm := GetTemplateManager()
	tm.ReloadTemplates()
	mal, ok := tm.GetTemplate("pages/admin")
	if !ok {
		t.Fatal("malsettet lét seg ikkje lasta")
	}
	var ut strings.Builder
	if err := mal.ExecuteTemplate(&ut, "admin_users_table", map[string]interface{}{
		"Lang": "nn", "Folk": folk, "Grupper": grupper, "CSRFToken": "x",
	}); err != nil {
		t.Fatalf("malen feila: %v", err)
	}
	return ut.String()
}

// Gruppone stend på personen, og dei syner kven han er med i.
func TestGruppemerkaSynerKvaPersonenErMedI(t *testing.T) {
	grupper := []database.Gruppe{{ID: 1, Namn: "Reformer"}, {ID: 2, Namn: "Kurs"}}
	folk := []database.Person{{
		ID: 7, Namn: "Ola", Epost: "ola@do.me",
		GruppeSett: map[int64]bool{1: true},
	}}
	html := teiknFolk(t, folk, grupper)

	if !strings.Contains(html, "Reformer") || !strings.Contains(html, "Kurs") {
		t.Fatal("gruppone stend ikkje paa personen")
	}
	// Berre merka på personen — `data-gruppe` stend ogso i lista yver
	// gruppone, og den fyrste treffen der er ei anna rad.
	rad := strings.Index(html, `class="grupper"`)
	if rad < 0 {
		t.Fatal("gruppone stend ikkje paa personen")
	}
	merka := html[rad:]

	// Den han er med i er på, den andre er av — båe stend der, av di
	// ei gruppe ein ikkje er med i er ikkje noko ein *ser* at manglar.
	i := strings.Index(merka, `data-gruppe="1"`)
	j := strings.Index(merka, `data-gruppe="2"`)
	if i < 0 || j < 0 {
		t.Fatal("gruppemerka stend ikkje der")
	}
	html = merka
	if !strings.Contains(html[i:i+120], `aria-pressed="true"`) {
		t.Error("gruppa han er med i stend ikkje som paa")
	}
	if !strings.Contains(html[j:j+120], `aria-pressed="false"`) {
		t.Error("gruppa han ikkje er med i stend som paa")
	}
}

// Ei gruppe skal ikkje sjå ut som eit løyve.
//
// «Gjev tilgang til reformer» og «gjer til administrator» er tvo ulike
// slag handling, og dei skal ikkje bera det same merket. §5: store
// bokstavar og 68 % breidd er maalet til det som namngjev ein kategori —
// eit løyve er det, eit gruppenamn er ikkje.
func TestGruppaSerIkkjeUtSomEitLøyve(t *testing.T) {
	html := teiknFolk(t,
		[]database.Person{{ID: 7, Namn: "Ola", GruppeSett: map[int64]bool{}}},
		[]database.Gruppe{{ID: 1, Namn: "Reformer"}})

	i := strings.Index(html, `class="gruppemerke"`)
	if i < 0 {
		t.Fatal("gruppemerket stend ikkje der")
	}
	// Merket ber korkje `.held` eller `.løyvet` — dei tvo er løyveforma.
	byrjing := strings.LastIndex(html[:i], "<button")
	knapp := html[byrjing : i+40]
	for _, feil := range []string{"held", "løyvet"} {
		if strings.Contains(knapp, feil) {
			t.Errorf("gruppemerket ber `%s` — det er forma til eit løyve", feil)
		}
	}
	// Og løyvi ber henne framleis.
	if !strings.Contains(html, `class="held loyvemerke"`) {
		t.Error("løyvi hev misst løyveforma si")
	}
}

// Gruppone lét seg laga og sletta.
func TestGrupponeLetSegLagaOgSletta(t *testing.T) {
	html := teiknFolk(t,
		[]database.Person{{ID: 7, Namn: "Ola", GruppeSett: map[int64]bool{}}},
		[]database.Gruppe{{ID: 1, Namn: "Reformer", Medlem: 3}})

	if !strings.Contains(html, `id="ny-gruppe"`) || !strings.Contains(html, `id="lag-gruppe"`) {
		t.Error("det gjeng ikkje an aa laga ei gruppe")
	}
	if !strings.Contains(html, "gruppeslett") {
		t.Error("det gjeng ikkje an aa sletta ei gruppe")
	}
	// Slettingi bed um trykk nummer tvo (§7).
	i := strings.Index(html, "gruppeslett")
	byrjing := strings.LastIndex(html[:i], "<button")
	if !strings.Contains(html[byrjing:i+200], "data-stadfest") {
		t.Error("slettingi bed ikkje um trykk nummer tvo")
	}
	// Medlemstalet stend, so ein ser kva ein tek burt.
	if !strings.Contains(html, "3 "+t2("nn", "admin.group_members")) {
		t.Error("medlemstalet stend ikkje paa gruppa")
	}
}
