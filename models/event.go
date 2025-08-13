package models

import "time"

// Event represents a generic event with common fields.
type Event struct {
	ID               int                 `json:"id"`
	Title            string              `json:"title"`
	Description      string              `json:"description"`
	RoleRequirements map[string]struct{} `json:"role_requirements"`
	StartTime        time.Time           `json:"start_time"`
	EndTime          time.Time           `json:"end_time"`
	Location         string              `json:"location"`
	Organizer        string              `json:"organizer"`
	Attendees        []string            `json:"attendees"`
}
