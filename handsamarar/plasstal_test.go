package handsamarar

import (
	"bytes"
	"html/template"
	"testing"
)

// Skjemaet i admin-class-management.html skriv plasstalet slik:
//
//	{{if $r.Plassar}}{{$r.Plassar}}{{else}}{{$r.RomPlassar}}{{end}}
//
// Med `Plassar` som peikar tyder «tom» no `nil` og ikkje 0. Malen skal
// framleis falla attende på rommet sitt tal når serien ikkje set noko,
// og skriva talet — ikkje ei adressa — når han gjer.
func TestPlasstaletFellAttendePaaRommet(t *testing.T) {
	mal := template.Must(template.New("t").Parse(
		`{{if .Plassar}}{{.Plassar}}{{else}}{{.RomPlassar}}{{end}}`))

	tolv := 12
	for _, p := range []struct {
		namn    string
		plassar *int
		rom     int
		vil     string
	}{
		{"ingi eigi kapasitet", nil, 18, "18"},
		{"eigi kapasitet sett", &tolv, 18, "12"},
	} {
		var ut bytes.Buffer
		if err := mal.Execute(&ut, struct {
			Plassar    *int
			RomPlassar int
		}{p.plassar, p.rom}); err != nil {
			t.Fatalf("%s: %v", p.namn, err)
		}
		if ut.String() != p.vil {
			t.Errorf("%s: fekk %q, venta %q", p.namn, ut.String(), p.vil)
		}
	}
}
