// Package prover holds the tests that look at the house from outside.
//
// They reach only exported API. That is worth something on its own: a
// test that gets by without digging into the private parts is a sign
// the seam it tests is a real boundary.
package prover

import (
	"os"
	"path/filepath"
	"testing"
)

// rota, absolute. Tests read files from here rather than through a path
// relative to the working directory: `go test` caches a result while the
// files a test opened are unchanged, and it resolves the name the test
// asked for. Asking for "static/css/deler/…" after a chdir found no such
// file under the package directory, and the stale answer came back
// marked "(cached)" while the stylesheet had been rewritten.
var rota string

func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		panic("kom ikkje til rota: " + err.Error())
	}
	stig, err := os.Getwd()
	if err != nil {
		panic("fann ikkje rota: " + err.Error())
	}
	rota = stig
	os.Exit(m.Run())
}

// del gives the absolute path to one part of the stylesheet.
func del(namn string) string { return filepath.Join(rota, "static", "css", "deler", namn) }
