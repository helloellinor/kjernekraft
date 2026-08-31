package database

import (
	"database/sql"
	"time"

	"kjernekraft/models"
)

// eventKolonnar er kolonnane kvar spurning som gjev ein *heil*
// models.Event skal henta, i den rekkjefylgja skannTime les deim.
//
// Lista stod skrivi ut på nytt i kvar spurning, og ho dreiv: nokre tok
// tolv kolonnar, ein tok atten, og GetEventByID slo ikkje upp rommet i
// det heile. Same timen kom då attende med ulike tal alt etter kven som
// spurde — `Capacity` var rommet sitt tal frå den eine og 0 frå den
// andre — og ingen kunde sjå det på typen, av di baae gav ein
// models.Event. Nett den feilen kosta oss «timen er full» på kvar
// påmelding ein gong fyrr; sjå SignupUserForEvent.
//
// Kapasiteten er rekna her og ikkje i Go: timen sitt eige tal når han
// hev eitt, elles rommet sitt. `EigenPlassar` ber det rå talet attmed,
// so administrasjonen kann skilja «ikkje sett» frå «sett like høgt som
// rommet».
//
// Merk NULLIF og ikkje COALESCE på den siste: kolonnen ber 0 *og* NULL
// for «ingi eigi kapasitet», og baae skal koma fram i Go som `nil`. Ein
// COALESCE her hadde gjort fråveret til talet 0, og då er det ikkje
// lenger råd å sjå skilnad på «ikkje sett» og «sett til noko».
//
// Spurningi lyt nytta aliasi `e` for events og `r` for rooms, og ho lyt
// hava med eventFrå-koplingi — utan rommet er halve lista NULL.
const eventKolonnar = `e.id, e.title, COALESCE(e.description, ''), e.start_time, e.end_time,
	       COALESCE(e.location, ''), COALESCE(e.organizer, ''), COALESCE(e.class_type, ''),
	       COALESCE(e.teacher_name, ''), COALESCE(NULLIF(e.capacity, 0), r.capacity, 0),
	       e.current_enrolment, COALESCE(e.color, ''),
	       COALESCE(r.id, 0), COALESCE(r.name, e.location, ''), COALESCE(r.capacity, 0),
	       COALESCE(e.serie_id, 0), NULLIF(e.capacity, 0), COALESCE(e.gruppe_id, 0)`

// eventFrå er koplingi eventKolonnar reknar med.
const eventFrå = ` FROM events e LEFT JOIN rooms r ON r.id = e.room_id `

// radskannar er det *sql.Row og *sql.Rows hev sams, so ein spurning som
// gjev éi rad og ein som gjev mange kann lesast av den same koden.
type radskannar interface {
	Scan(dest ...any) error
}

// skannTime les ei rad henta med eventKolonnar inn i ein models.Event.
func skannTime(rad radskannar, e *models.Event) error {
	return rad.Scan(
		&e.ID, &e.Title, &e.Description, &e.StartTime, &e.EndTime,
		&e.Location, &e.Organizer, &e.ClassType,
		&e.TeacherName, &e.Capacity,
		&e.CurrentEnrolment, &e.Color,
		&e.RoomID, &e.RoomName, &e.RoomCapacity,
		&e.SerieID, &e.EigenPlassar, &e.GruppeID,
	)
}

// timane les alle radene ei eventKolonnar-spurning gav.
func timane(rows *sql.Rows) ([]models.Event, error) {
	defer rows.Close()
	var ut []models.Event
	for rows.Next() {
		var e models.Event
		if err := skannTime(rows, &e); err != nil {
			return nil, err
		}
		ut = append(ut, e)
	}
	return ut, rows.Err()
}

// ---- Dei smale spurningane ----
//
// Ikkje kvar spurning som rører timetabellen gjev ein *time*. Nokre spør
// um éin ting — er rommet uppteke, er nokon påmeld yver taket, kom du på
// timen — og svaret er ikkje ein models.Event.
//
// Dei gav ein like fullt: ein models.Event med fire felt fylte og resten
// på nullverdet sitt. Typen lova ein heil time, verdien var det ikkje, og
// ingen kunde sjå skilnaden. Nett den lovnaden er det som gjer at 0
// plassar ser ut som eit svar (sjå eventKolonnar). Ein eigen type med
// berre dei felti spurningi faktisk hentar kann ikkje lyga på den måten.

// Romkollisjon er timen som alt stend i rommet. Berre det ein treng for
// å segja frå um kva som er i vegen.
type Romkollisjon struct {
	Tittel string
	Lærar  string
	Start  time.Time
	Slutt  time.Time
}

// Overfylt er den fyrste timen i serien som hev fleire påmelde enn det
// nye taket. Ein treng dagen og talet for å segja kvifor det ikkje gjeng.
type Overfylt struct {
	Start    time.Time
	Paamelde int
}

// SerieTime er ein time i ein serie, slik omleggjaren treng honom: kva
// han heiter i basen og når han tek til. Resten av timen skal ikkje
// hentast for å flytta honom.
type SerieTime struct {
	ID    int64
	Start time.Time
}
