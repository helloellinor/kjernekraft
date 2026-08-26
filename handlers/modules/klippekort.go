package modules

// KlippekortModuleData represents the data needed for the klippekort module
type KlippekortModuleData struct {
	HasKlippekort bool
	Klippekort    interface{} // This will be []models.KlippekortWithDetails in practice
	Lang          string
}

// NewKlippekortModule creates a new klippekort module with the given data
func NewKlippekortModule(klippekort interface{}, lang string) (*KlippekortModuleData, error) {
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
	}, nil
}

// GetTemplateName returns the template name for this module
func (k *KlippekortModuleData) GetTemplateName() string {
	return "klippekort_module"
}
