package handsamarar

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

var (
	assetVersionOnce sync.Once
	assetVersion     string
)

// staticURL gjev ei adresse som skiftar når innhaldet under static/
// skiftar. Då kan CSS og JavaScript bufraast uforanderleg i drift utan
// at ein ny utgåve sit fast bak ein gamal nettlesarbuffer.
func staticURL(path string) string {
	assetVersionOnce.Do(func() { assetVersion = reknAssetVersjon() })
	if assetVersion == "0" {
		return "/static/" + path
	}
	return "/static/" + path + "?v=" + assetVersion
}

func reknAssetVersjon() string {
	h := sha256.New()
	err := filepath.WalkDir("static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if _, err := io.WriteString(h, path+"\x00"); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = h.Write(data)
		return err
	})
	if err != nil {
		// Å ikkje kunna rekna ut avtrykket skal ikkje stogga tenaren.
		// I vanleg drift finst static/ ved oppstart; reserveverdet held
		// berre nettlesaren på dei vanlege, korte cache-reglane.
		return "0"
	}
	return hex.EncodeToString(h.Sum(nil))[:12]
}
