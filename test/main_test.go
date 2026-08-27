package test

import (
	"os"
	"path/filepath"
	"testing"

	"kjernekraft/database"
)

// TestMain gjev kvar prøvekøyring si eigi base i ei mellombils mappa.
//
// Prøvorne skreiv i test/kjernekraft.db fyrr, og fila laag att etterpaa.
// Difor gjekk suiten fyrste gongen og fall den andre — «e-post er
// allerede i bruk» — utan at noko i koda hadde endra seg. Ei prøva som
// gjeng ein gong er ikkje ei prøva.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "kjernekraft-prova-*")
	if err != nil {
		panic(err)
	}
	if err := os.Setenv(database.DBPathEnv, filepath.Join(dir, "prova.db")); err != nil {
		panic(err)
	}

	code := m.Run()

	os.RemoveAll(dir)
	os.Exit(code)
}
