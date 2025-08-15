package modules

import (
	"io/ioutil"
	"path/filepath"
)

// MembershipModuleData represents the data needed for the membership module
type MembershipModuleData struct {
	HasMembership bool
	Membership    interface{} // This will be *models.MembershipWithDetails in practice
	Lang          string
	MembershipCSS string
}

// NewMembershipModule creates a new membership module with the given data
func NewMembershipModule(membership interface{}, lang string) (*MembershipModuleData, error) {
	// Load CSS content
	cssPath := filepath.Join("handlers", "templates", "modules", "membership", "membership.css")
	cssContent, err := ioutil.ReadFile(cssPath)
	if err != nil {
		cssContent = []byte("/* CSS loading failed */")
	}

	hasMembership := membership != nil

	return &MembershipModuleData{
		HasMembership: hasMembership,
		Membership:    membership,
		Lang:          lang,
		MembershipCSS: string(cssContent),
	}, nil
}

// GetTemplateName returns the template name for this module
func (m *MembershipModuleData) GetTemplateName() string {
	return "membership_module"
}