package modules

// AdminStatsModuleData represents the data needed for the admin stats module
type AdminStatsModuleData struct {
	TotalUsers            int
	TotalEvents           int
	PendingFreezeRequests int
	Lang                  string
}

// NewAdminStatsModule creates a new admin stats module with the given data
func NewAdminStatsModule(totalUsers, totalEvents, pendingFreezeRequests int, lang string) (*AdminStatsModuleData, error) {
	return &AdminStatsModuleData{
		TotalUsers:            totalUsers,
		TotalEvents:           totalEvents,
		PendingFreezeRequests: pendingFreezeRequests,
		Lang:                  lang,
	}, nil
}

// GetTemplateName returns the template name for this module
func (a *AdminStatsModuleData) GetTemplateName() string {
	return "admin_stats_module"
}
