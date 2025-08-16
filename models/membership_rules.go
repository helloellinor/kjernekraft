package models

// MembershipRules represents the configurable rules for membership management
type MembershipRules struct {
	ID                        int    `json:"id"`
	AllowUpgrades            bool   `json:"allow_upgrades"`
	CombineBindingPeriods    bool   `json:"combine_binding_periods"`
	AllowDowngrades          bool   `json:"allow_downgrades"`
	AllowChangeDuringBinding bool   `json:"allow_change_during_binding"`
	DefaultMembershipID      *int   `json:"default_membership_id"`
	UpdatedAt                string `json:"updated_at"`
}