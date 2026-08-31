package modules

import "kjernekraft/models"

// KlippekortModuleData is what the klippekort block needs to render.
// Its template is "klippekort_module".
type KlippekortModuleData struct {
	HasKlippekort bool
	Klippekort    []models.KlippekortWithDetails
	Lang          string

	// Nærast is the card that expires first among those with clips
	// left — the one question a list of cards does not answer by itself,
	// since a card with twenty clips and four days left looks richer
	// than one with two clips and half a year.
	//
	// No template reads it today; the line that showed it (.klippfyrst)
	// was taken out. Kept because the question it answers has not gone
	// away, and it is tested.
	Nærast *models.KlippekortWithDetails
}

// NewKlippekortModule builds the block. An empty list means no cards.
func NewKlippekortModule(kort []models.KlippekortWithDetails, lang string) *KlippekortModuleData {
	return &KlippekortModuleData{
		HasKlippekort: len(kort) > 0,
		Klippekort:    kort,
		Lang:          lang,
		Nærast:        models.NærastUtløp(kort),
	}
}
