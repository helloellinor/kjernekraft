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
	// Serien timen er laga av. Ein time *er* ikkje ein ting for seg —
	// han er eitt utslag av ein serie («yoga med Leon, måndag 18:00»),
	// og det er serien administrasjonen endrar. Timane ber serie-id-en
	// so dei kann finna kvarandre att.
	SerieID int64 `json:"rule_id"`
	// Timen si eigi kapasitet. `nil` tyder at han ikkje set noko sjølv,
	// og då gjeld rommet sitt tal.
	//
	// `Capacity` yver er det utrekna talet og kann ikkje svara på um det
	// *er* eit yverstyrt tal: eit rom med tolv plassar og ei yverstyring
	// på tolv gjev det same svaret. Administrasjonen lyt kunna skilja dei
	// tvo, so feltet kann syna rommet sitt tal som eit framlegg og timen
	// sitt som ein verdi.
	//
	// Han var eit `int` der 0 tydde «ingi eigi». Det er den same feilen
	// som gjer at ein spurning utan romkopling gjev 0 plassar og ser rett
	// ut: talet 0 kann ikkje segja frå om det er eit svar eller eit
	// fråver. Ein peikar kann.
	EigenPlassar *int `json:"eigen_plassar"`
	// Rommet. Kapasiteten kjem herifrå når timen ikkje set si eigi.
	// Gruppa timen er open for. Null er open for alle.
	GruppeID     int    `json:"gruppe_id"`
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

// LengdMin er lengdi på timen i minutt. JS-en i administrasjonen
// treng henne for aa rekna ny slutt når starten vert flutt.
func (e Event) LengdMin() int { return int(e.EndTime.Sub(e.StartTime).Minutes()) }

// Room er eit rom i studioet. Kapasiteten er ein eigenskap ved rommet og
// ikkje noko ein skriv inn per time.
type Room struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
}
