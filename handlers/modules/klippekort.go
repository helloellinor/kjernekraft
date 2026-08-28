package modules

import "kjernekraft/models"

// KlippekortModuleData represents the data needed for the klippekort module
type KlippekortModuleData struct {
	HasKlippekort bool
	Klippekort    interface{} // This will be []models.KlippekortWithDetails in practice
	Lang          string

	// Naermast er det kortet som gjeng ut fyrst av dei som framleis hev
	// klipp att. Det er det einaste spursmaalet ei liste yver kort ikkje
	// svarar paa av seg sjølv: eit kort med tjue klipp att og fire dagar
	// til utløp ser rikare ut enn eit med tvo klipp og eit halvt aar.
	//
	// Kort som er tome eller alt utgjengne tel ikkje med. Eit tomt kort
	// som gjeng ut er ikkje ein frist, det er ei kvittering.
	//
	// Ingen mal les honom i dag: lina som synte honom
	// (`.klippfyrst`) er teki ut. Rekninga stend att med vilje —
	// ho er prøvd, og spursmaalet ho svarar paa er ikkje burte.
	Naermast *models.KlippekortWithDetails
}

// naermastUtlop er regelen i models, naadd herifraa. Prøvone i denne
// pakka held framleis paa honom.
func naermastUtlop(kort []models.KlippekortWithDetails) *models.KlippekortWithDetails {
	return models.NaermastUtlop(kort)
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

	d := &KlippekortModuleData{
		HasKlippekort: hasKlippekort,
		Klippekort:    klippekort,
		Lang:          lang,
	}
	if kort, ok := klippekort.([]models.KlippekortWithDetails); ok {
		// Ei tom, typa lista er *ikkje* «hev kort». Sjekken over fell
		// til `default: true` for kvar typa lista, tom eller ei.
		d.HasKlippekort = len(kort) > 0
		d.Naermast = naermastUtlop(kort)
	}
	return d, nil
}

// GetTemplateName returns the template name for this module
func (k *KlippekortModuleData) GetTemplateName() string {
	return "klippekort_module"
}
