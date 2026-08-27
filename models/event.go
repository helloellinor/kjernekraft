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
	ClassType        string `json:"class_type"`        // e.g. "yoga", "pilates"
	TeacherName      string `json:"teacher_name"`      // Name of the teacher
	Capacity         int    `json:"capacity"`          // Maximum number of attendees
	CurrentEnrolment int    `json:"current_enrolment"` // Current number of enrolled
	Color            string `json:"color"`             // Color for the class type
	// Regelen timen er laga av. Ein time *er* ikkje ein ting for seg —
	// han er eitt utslag av ein regel («yoga med Leon, maandag 18:00»),
	// og det er regelen administrasjonen endrar. Timane ber regel-id-en
	// so dei kann finna kvarandre att.
	RuleID int64 `json:"rule_id"`
	// Rommet. Kapasiteten kjem herifraa naar timen ikkje set si eigi.
	RoomID       int    `json:"room_id"`
	RoomName     string `json:"room_name"`
	RoomCapacity int    `json:"room_capacity"`
	// User-specific fields (populated for specific users)
	IsUserSignedUp bool `json:"is_user_signed_up"` // Whether the current user is signed up for this event
}

// Ledige segjer kor mange plassar som er att.
func (e Event) Ledige() int {
	if n := e.Capacity - e.CurrentEnrolment; n > 0 {
		return n
	}
	return 0
}

// Full segjer um timen er utseld.
func (e Event) Full() bool { return e.Capacity > 0 && e.CurrentEnrolment >= e.Capacity }

// LengdMin er lengdi paa timen i minutt. JS-en i administrasjonen
// treng henne for aa rekna ny slutt naar starten vert flutt.
func (e Event) LengdMin() int { return int(e.EndTime.Sub(e.StartTime).Minutes()) }

// Room er eit rom i studioet. Kapasiteten er ein eigenskap ved rommet og
// ikkje noko ein skriv inn per time.
type Room struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}
