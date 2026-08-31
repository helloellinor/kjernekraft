package handsamarar

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var sidenykelIMal = regexp.MustCompile(`\.CurrentPage\s+"([^"]+)"`)

// Malen avgjer kva javascript ei sida fær med å samanlikna .CurrentPage
// mot faste strengar. Skriv Go-sida ein nykel malen ikkje kjenner — eller
// døyper ein um utan å fylgja etter i malen — fell samanlikningi berre
// igjenom, og sida kjem upp utan javascript. Ingi feilmelding, ingi tom
// sida: alt ser rett ut, og ingen ting verkar.
//
// Difor lyt kvar streng malen spør etter svara til ein Sidenykel som
// finst.
func TestMalenKjennerBerreSidenyklarSomFinst(t *testing.T) {
	kjende := make(map[string]bool, len(sidenyklar))
	for _, n := range sidenyklar {
		kjende[string(n)] = true
	}

	ukjende := make(map[string][]string)
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
		for _, treff := range sidenykelIMal.FindAllStringSubmatch(string(innhald), -1) {
			if !kjende[treff[1]] {
				ukjende[treff[1]] = append(ukjende[treff[1]], sti)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(ukjende) > 0 {
		nyklar := make([]string, 0, len(ukjende))
		for n := range ukjende {
			nyklar = append(nyklar, n)
		}
		sort.Strings(nyklar)
		for _, n := range nyklar {
			t.Errorf("malen spør etter sidenykelen %q, men han finst ikkje i sidenyklar: %s",
				n, strings.Join(ukjende[n], ", "))
		}
	}
}
