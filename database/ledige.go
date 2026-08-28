package database

import "kjernekraft/models"

// Timar med ledig plass i dag og i morgon.
//
// Heimesida spør ikkje «kva går i dag» lenger — ho spør «kvar er det
// plass». Ein time som er full, er ikkje eit tilbod; han er berre ei rad
// ein les forbi. Og i dag åleine er for kort: er klokka sju om kvelden,
// er det ingenting att av dagen, og då står bolken tom kvar kveld.
// LedigeTimar gjev det `sjaaarID` skal sjaa: opne timar, og private som
// er hans eigne. Sjaa privattime.go.
func (db *Database) LedigeTimar(sjaaarID int64) ([]models.Event, error) {
	query := `
		SELECT e.id, e.title, COALESCE(e.description, ''), e.start_time, e.end_time,
		       COALESCE(e.location, ''), COALESCE(e.organizer, ''), COALESCE(e.class_type, ''),
		       COALESCE(e.teacher_name, ''),
		       COALESCE(NULLIF(e.capacity, 0), r.capacity, 0), e.current_enrolment,
		       COALESCE(e.color, ''), COALESCE(r.name, e.location, '')
		FROM events e LEFT JOIN rooms r ON r.id = e.room_id
		WHERE DATE(e.start_time) IN (DATE('now', 'localtime'), DATE('now', 'localtime', '+1 day'))
		  AND ` + synlegFor + `
		ORDER BY e.start_time ASC
	`
	rows, err := db.Conn.Query(query, sjaaarID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.StartTime, &e.EndTime,
			&e.Location, &e.Organizer, &e.ClassType, &e.TeacherName,
			&e.Capacity, &e.CurrentEnrolment, &e.Color, &e.RoomName); err != nil {
			return nil, err
		}
		// Kapasiteten kjem frå rommet når timen ikkje set si eiga, so
		// «full» kan fyrst avgjerast her.
		if e.Full() {
			continue
		}
		ut = append(ut, e)
	}
	return ut, rows.Err()
}
