package modules

import (
	"io/ioutil"
	"path/filepath"
)

// KlippekortModuleData represents the data needed for the klippekort module
type KlippekortModuleData struct {
	HasKlippekort bool
	Klippekort    interface{} // This will be []models.KlippekortWithDetails in practice
	Lang          string
	KlippekortCSS string
}

// NewKlippekortModule creates a new klippekort module with the given data
func NewKlippekortModule(klippekort interface{}, lang string) (*KlippekortModuleData, error) {
	// Load CSS content
	cssPath := filepath.Join("handlers", "templates", "modules", "membership", "klippekort.css")
	cssContent, err := ioutil.ReadFile(cssPath)
	if err != nil {
		cssContent = []byte("/* CSS loading failed */")
	}

	hasKlippekort := false
	if klippekort != nil {
		// Check if klippekort slice has items
		switch v := klippekort.(type) {
		case []interface{}:
			hasKlippekort = len(v) > 0
		default:
			hasKlippekort = true // Assume true if not a slice
		}
	}

	return &KlippekortModuleData{
		HasKlippekort: hasKlippekort,
		Klippekort:    klippekort,
		Lang:          lang,
		KlippekortCSS: string(cssContent),
	}, nil
}

// GetTemplateName returns the template name for this module
func (k *KlippekortModuleData) GetTemplateName() string {
	return "klippekort_module"
}