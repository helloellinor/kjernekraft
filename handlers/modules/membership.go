package modules

// MembershipModuleData represents the data needed for the membership module
type MembershipModuleData struct {
	HasMembership bool
	Membership    interface{} // This will be *models.MembershipWithDetails in practice
	Lang          string
}

// NewMembershipModule creates a new membership module with the given data
func NewMembershipModule(membership interface{}, lang string) (*MembershipModuleData, error) {
	hasMembership := membership != nil

	return &MembershipModuleData{
		HasMembership: hasMembership,
		Membership:    membership,
		Lang:          lang,
	}, nil
}

// GetTemplateName returns the template name for this module
func (m *MembershipModuleData) GetTemplateName() string {
	return "membership_module"
}
