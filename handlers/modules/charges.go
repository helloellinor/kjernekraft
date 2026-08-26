package modules

// ChargesModuleData represents the data needed for the charges module
type ChargesModuleData struct {
	HasCharges bool
	Charges    interface{} // This will be []models.Charge in practice
	Lang       string
}

// NewChargesModule creates a new charges module with the given data
func NewChargesModule(charges interface{}, lang string) (*ChargesModuleData, error) {
	hasCharges := false
	if charges != nil {
		// Check if charges slice has items (this assumes it's a slice)
		switch v := charges.(type) {
		case []interface{}:
			hasCharges = len(v) > 0
		default:
			hasCharges = true // Assume true if not a slice
		}
	}

	return &ChargesModuleData{
		HasCharges: hasCharges,
		Charges:    charges,
		Lang:       lang,
	}, nil
}

// GetTemplateName returns the template name for this module
func (c *ChargesModuleData) GetTemplateName() string {
	return "charges_module"
}
