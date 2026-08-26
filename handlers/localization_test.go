package handlers

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Umsetjingarne var ein paastand i README fyrr: «full støtte for nb, nn
// og en». Ein paastand i eit dokument veit ikkje naar han vert usann.
// Desse prøvone gjer han til noko som brest naar han brest.

const maalKatalog = "../locales"

// lastMaali les umsetjingsfilone med den same koden som tenaren nyttar,
// berre med ein annan sti. loadTranslations tegjer naar ei fil ikkje let
// seg lesa — difor prøver me at kvart maal *hev* noko i seg fyrst.
func lastMaali(t *testing.T) map[string]map[string]interface{} {
	t.Helper()

	l := &Localization{
		languages: make(map[string]map[string]interface{}),
		basePath:  maalKatalog,
	}
	l.loadTranslations()

	for _, maal := range l.GetSupportedLanguages() {
		if len(l.languages[maal]) == 0 {
			t.Fatalf("%s/common.json gav ingen nyklar — fila manglar, er tom, eller er ugild JSON", maal)
		}
	}
	return l.languages
}

// flatNyklar gjer det nøsta kartet um til «admin.class_table.title».
func flatNyklar(m map[string]interface{}, fyre string, ut map[string]string) {
	for k, v := range m {
		nykel := k
		if fyre != "" {
			nykel = fyre + "." + k
		}
		switch t := v.(type) {
		case map[string]interface{}:
			flatNyklar(t, nykel, ut)
		case string:
			ut[nykel] = t
		}
	}
}

func TestAlleMaalHevDeiSameNyklane(t *testing.T) {
	maali := lastMaali(t)

	flate := make(map[string]map[string]string, len(maali))
	alle := make(map[string]bool)
	for maal, m := range maali {
		flate[maal] = make(map[string]string)
		flatNyklar(m, "", flate[maal])
		for nykel := range flate[maal] {
			alle[nykel] = true
		}
	}

	for maal, nyklar := range flate {
		var manglar []string
		for nykel := range alle {
			if _, finst := nyklar[nykel]; !finst {
				manglar = append(manglar, nykel)
			}
		}
		if len(manglar) > 0 {
			sort.Strings(manglar)
			t.Errorf("%s manglar %d nyklar som hine maali hev:\n  %s",
				maal, len(manglar), strings.Join(manglar, "\n  "))
		}
	}
}

func TestIngenUmsetjingErTom(t *testing.T) {
	for maal, m := range lastMaali(t) {
		flat := make(map[string]string)
		flatNyklar(m, "", flat)
		for nykel, verd := range flat {
			if strings.TrimSpace(verd) == "" {
				t.Errorf("%s.%s er tom", maal, nykel)
			}
		}
	}
}

// Nyklar som malane bed um, men som ingen stad er umsette, kjem ut paa
// skjermen som seg sjølve — «admin.people» stend der ordet skulde staa.
// Malen brest ikkje, sida brest ikkje; det er berre nykelen som er der.
var nykelIMal = regexp.MustCompile(`\{\{\s*t\s+\.[A-Za-z.]+\s+"([^"]+)"`)

func TestMalaneBedBerreUmNyklarSomFinst(t *testing.T) {
	maali := lastMaali(t)
	nn := make(map[string]string)
	flatNyklar(maali["nn"], "", nn)

	sakna := make(map[string][]string)

	err := filepath.Walk("templates", func(sti string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(sti) != ".html" {
			return nil
		}
		innhald, err := os.ReadFile(sti)
		if err != nil {
			return err
		}
		for _, treff := range nykelIMal.FindAllStringSubmatch(string(innhald), -1) {
			nykel := treff[1]
			if _, finst := nn[nykel]; !finst {
				sakna[nykel] = append(sakna[nykel], sti)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("kunde ikkje ganga gjenom malane: %v", err)
	}

	if len(sakna) > 0 {
		nyklar := make([]string, 0, len(sakna))
		for nykel := range sakna {
			nyklar = append(nyklar, nykel)
		}
		sort.Strings(nyklar)
		for _, nykel := range nyklar {
			t.Errorf("malen bed um «%s», som ikkje er umsett: %s",
				nykel, strings.Join(sakna[nykel], ", "))
		}
	}
}
