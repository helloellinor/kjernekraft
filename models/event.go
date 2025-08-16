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
	// Class-specific fields
	ClassType        string              `json:"class_type"`        // e.g. "yoga", "pilates"
	TeacherName      string              `json:"teacher_name"`      // Name of the teacher
	Capacity         int                 `json:"capacity"`          // Maximum number of attendees
	CurrentEnrolment int                 `json:"current_enrolment"` // Current number of enrolled
	Color            string              `json:"color"`             // Color for the class type
	// User-specific fields (populated for specific users)
	IsUserSignedUp   bool                `json:"is_user_signed_up"` // Whether the current user is signed up for this event
}
