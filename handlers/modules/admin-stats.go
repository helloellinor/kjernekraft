package modules

import (
	"io/ioutil"
	"path/filepath"
)

// AdminStatsModuleData represents the data needed for the admin stats module
type AdminStatsModuleData struct {
	TotalUsers             int
	TotalEvents            int
	PendingFreezeRequests  int
	Lang                   string
	AdminStatsCSS          string
}

// NewAdminStatsModule creates a new admin stats module with the given data
func NewAdminStatsModule(totalUsers, totalEvents, pendingFreezeRequests int, lang string) (*AdminStatsModuleData, error) {
	// Load CSS content
	cssPath := filepath.Join("handlers", "templates", "modules", "admin-stats.css")
	cssContent, err := ioutil.ReadFile(cssPath)
	if err != nil {
		cssContent = []byte("/* CSS loading failed */")
	}

	return &AdminStatsModuleData{
		TotalUsers:            totalUsers,
		TotalEvents:           totalEvents,
		PendingFreezeRequests: pendingFreezeRequests,
		Lang:                  lang,
		AdminStatsCSS:         string(cssContent),
	}, nil
}

// GetTemplateName returns the template name for this module
func (a *AdminStatsModuleData) GetTemplateName() string {
	return "admin_stats_module"
}