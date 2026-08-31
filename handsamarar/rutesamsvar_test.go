package handsamarar

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Kvar adressa flata bed um, skal tenaren kunna svara på.
//
// Dette er prøva på ein feil som gjekk rett gjenom eit grønt bygg. Ei
// stor umdøyping bytte ordet «reglar» i *strengen* `/api/admin/reglar`
// — handsamarnamnet stod att, av di det ikkje er noko ordskilje inni
// `AdminReglar` — so koden bygde, prøvone gjekk, og ruta var
// flutt. Prissida bad om den gamle adressa og fekk 404, og bolken med
// medlemskapsreglar vart staaande tom.
//
// Ein streng som er ein avtale millom tvo filer vert ikkje prøvd av
// kompilatoren. Han lyt prøvast her.
func TestKvarAdressaFlataBedUmHevEiRute(t *testing.T) {
	gamal, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(gamal)

	srv, err := os.ReadFile("server.go")
	if err != nil {
		t.Fatal(err)
	}

	// Rutene tenaren set upp. `Route` er eit prefiks; hine er heile stiar.
	var stiar, prefiks []string
	for _, m := range regexp.MustCompile(
		`r\.(?:Get|Post|Put|Delete|Patch|Handle|HandleFunc)\("(/api/[^"]*)"`).
		FindAllStringSubmatch(string(srv), -1) {
		stiar = append(stiar, m[1])
	}
	for _, m := range regexp.MustCompile(`r\.Route\("(/api/[^"]*)"`).
		FindAllStringSubmatch(string(srv), -1) {
		prefiks = append(prefiks, m[1])
	}
	if len(stiar) == 0 {
		t.Fatal("fann ingen ruter i server.go — prøva ser feil stad")
	}

	dekt := func(sti string) bool {
		for _, r := range stiar {
			if r == sti {
				return true
			}
			// `/api/admin/class/*` dekkjer alt under seg.
			if strings.HasSuffix(r, "*") && strings.HasPrefix(sti, strings.TrimSuffix(r, "*")) {
				return true
			}
		}
		for _, p := range prefiks {
			if sti == p || strings.HasPrefix(sti, strings.TrimSuffix(p, "/")+"/") {
				return true
			}
		}
		return false
	}

	// Adressone flata bed um: i malar og i skript.
	bedne := map[string][]string{}
	samla := func(rot, endelse string) {
		filepath.Walk(rot, func(sti string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(sti, endelse) {
				return nil
			}
			b, err := os.ReadFile(sti)
			if err != nil {
				return nil
			}
			for _, m := range regexp.MustCompile(`["'`+"`"+`](/api/[a-zA-Z0-9/_\-]+)`).
				FindAllStringSubmatch(string(b), -1) {
				bedne[m[1]] = append(bedne[m[1]], sti)
			}
			for _, m := range regexp.MustCompile(`hx-(?:get|post|put|delete)="(/api/[a-zA-Z0-9/_\-]+)`).
				FindAllStringSubmatch(string(b), -1) {
				bedne[m[1]] = append(bedne[m[1]], sti)
			}
			return nil
		})
	}
	samla("handsamarar/templates", ".html")
	samla("static/js", ".js")

	if len(bedne) == 0 {
		t.Fatal("fann ingi adressa i malane — prøva ser feil stad")
	}
	for sti, kvar := range bedne {
		if !dekt(sti) {
			t.Errorf("%s vert bede um i %s, men ingi rute svarar", sti, strings.Join(kvar, ", "))
		}
	}
}
