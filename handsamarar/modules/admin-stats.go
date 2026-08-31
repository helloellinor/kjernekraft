package modules

// AdminStatsModuleData is what the admin header needs to render.
// Its template is "admin_stats_module".
type AdminStatsModuleData struct {
	TotalUsers            int
	TotalEvents           int
	PendingFreezeRequests int
	Lang                  string
}

// NewAdminStatsModule builds the block.
func NewAdminStatsModule(totalUsers, totalEvents, pendingFreezeRequests int, lang string) *AdminStatsModuleData {
	return &AdminStatsModuleData{
		TotalUsers:            totalUsers,
		TotalEvents:           totalEvents,
		PendingFreezeRequests: pendingFreezeRequests,
		Lang:                  lang,
	}
}
