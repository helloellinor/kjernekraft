package database

import (
	"database/sql"

	"kjernekraft/models"
)

// Den private timen.
//
// Personlig Trening er ein time med éin deltakar. Han fanst som
// klippekort-kategori og som noko folk kjøpte, men det var ingen maate
// aa *setja honom upp* paa: kvar time i timeplanen var synleg for alle,
// so ei PT-økt laut anten liggja ute so heile huset saag henne, eller
// ikkje liggja der i det heile. Difor stod «Personlig Trening» med null
// timar medan Reformer hadde nitten.
//
// `private_user_id` er heile skiljet: NULL er ein vanleg time, sett er
// ein time som berre den eine ser og berre den eine kann melda seg paa.
//
// Skiljet vert handheva tvo stader, og det er med vilje:
//
//   - *synlegheit* i spurningane som byggjer timeplanen. Ser ein ikkje
//     timen, kann ein ikkje klikka paa honom.
//   - *tilgang* i SignupUserForEvent. Synlegheit er ikkje tryggleik —
//     ein som gissar eit id kann POSta seg paa ein time han aldri saag,
//     og kiosken melder folk paa utanum timeplanen med. Sjekken lyt
//     liggja der paameldingi faktisk hender.

// MigrerPrivatTime legg til kolonna um ho ikkje er der.
func MigrerPrivatTime(db *sql.DB) error {
	var finst bool
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='private_user_id'").
		Scan(&finst); err != nil {
		return err
	}
	if finst {
		return nil
	}
	if _, err := db.Exec("ALTER TABLE events ADD COLUMN private_user_id INTEGER"); err != nil {
		return err
	}
	// Timeplanen spør etter dette for kvar veke ein teiknar.
	_, err := db.Exec(
		"CREATE INDEX IF NOT EXISTS idx_events_private ON events(private_user_id)")
	return err
}

// synlegFor er vilkaaret som slepp gjenom vanlege timar og dei private
// som høyrer denne sjaaaren til. Han tek eitt argument: sjaaaren sitt id.
//
// Han stend som ein konstant og ikkje som tekst skriven for haand kvar
// gong, av di ein timeplan-spurning som gløymer honom lek ei PT-økt ut
// til heile huset — og det er ikkje noko ein ser ved aa lesa lista.
// Sidan gruppone kom, dekkjer han tvo ting: timen som høyrer éin
// person til, og timen som høyrer ei gruppe til. Baae er «kven fær sjaa
// dette», og dei stend saman av same grunnen som yver — ei spurning som
// hugsar det eine og gløymer det andre lek like fullt.
//
// Han tek sjaaaren sitt id *tvo* gonger. Det er ikkje pent, men SQL kann
// ikkje binda den same posisjonen tvo stader, og det er den trygge feilen
// aa gjera: gløymer nokon det andre argumentet, fell spurningi med
// «argument count mismatch» med ein gong. Hadde vilkaaret i staden vore
// skrive for haand paa kvar stad, hadde ein gløymd stad ikkje sagt noko —
// han hadde berre synt reformer-timane til heile huset.
const synlegFor = `(e.private_user_id IS NULL OR e.private_user_id = ?)
	AND (e.gruppe_id IS NULL OR EXISTS (
		SELECT 1 FROM gruppemedlem gm
		WHERE gm.gruppe_id = e.gruppe_id AND gm.user_id = ?))`

// EigarAvPrivatTime gjev brukaren timen er sett av til, og false naar
// timen er open for alle.
func (db *Database) EigarAvPrivatTime(eventID int64) (int64, bool, error) {
	var eigar sql.NullInt64
	err := db.Conn.QueryRow(
		`SELECT private_user_id FROM events WHERE id = ?`, eventID).Scan(&eigar)
	if err != nil {
		return 0, false, err
	}
	if !eigar.Valid {
		return 0, false, nil
	}
	return eigar.Int64, true, nil
}

// SettPrivatTime gjer ein time privat for ein brukar, eller open att
// naar userID er 0.
func (db *Database) SettPrivatTime(eventID, userID int64) error {
	if userID == 0 {
		_, err := db.Conn.Exec(
			`UPDATE events SET private_user_id = NULL WHERE id = ?`, eventID)
		return err
	}
	_, err := db.Conn.Exec(
		`UPDATE events SET private_user_id = ? WHERE id = ?`, userID, eventID)
	return err
}

// EventsSynlegeFor gjev alle timane `sjaaarID` hev lov aa sjaa.
//
// GetAllEvents heiter framleis det han gjer og gjev alt — administrasjonen
// skal sjaa PT-øktene med. Denne er den ein rettar mot brukaren.
func (db *Database) EventsSynlegeFor(sjaaarID int64) ([]models.Event, error) {
	rows, err := db.Conn.Query(`
		SELECT e.id, e.title, COALESCE(e.description, ''), e.start_time, e.end_time,
		       COALESCE(e.location, ''), COALESCE(e.organizer, ''),
		       COALESCE(e.class_type, ''), COALESCE(e.teacher_name, ''),
		       COALESCE(NULLIF(e.capacity, 0), r.capacity, 0), e.current_enrolment,
		       COALESCE(e.color, ''), COALESCE(r.name, e.location, '')
		FROM events e LEFT JOIN rooms r ON r.id = e.room_id
		WHERE `+synlegFor+`
		ORDER BY e.start_time ASC`, sjaaarID, sjaaarID)
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
		ut = append(ut, e)
	}
	return ut, rows.Err()
}
