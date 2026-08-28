package database

import (
	"database/sql"
	"time"

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

// PTKategori er klippekort-kategorien ei privat økt vert betalt med.
// Ho stend som ein konstant her av di ho bind tri ting saman: kva
// `class_type` ei PT-økt fær naar ho vert sett upp, kva pakke klippet
// vert teke av, og kva prisen heiter i lista. Skil dei lag, kostar økta
// ingen ting utan at nokon ser det.
const PTKategori = "Personlig Trening"

// BokPrivatTime melder den eine paa økta si og tek klippet med ein gong.
//
// Klippet fylgjer elles krysset i døri og ikkje paameldingi — ein time
// ein ikkje fekk skal ikkje kosta (sjaa klippbruk.go). Ei PT-økt er ikkje
// den same handelen: timen er sett av til éin person, og det er
// *avtala* som er varen. Ho vert difor klippa naar ho vert sett upp.
//
// Og ho gjeng aldri i minus. Krysset i døri skriv ei skuld naar kortet
// er tomt — den timen er alt gjeven, so han skal ikkje ut or rekneskapen
// — men her er han ikkje gjeven enno. Er det ingen klipp, vert økta sett
// upp lell og ingen skuld skrivi: det er ei avtala som skal gjerast upp
// paa ein annan maate, ikkje eit tal som skal fylgja nokon.
//
// Klippet vert ført paa paameldingi (`klipp_kort_id`), so det gjeng
// attende av seg sjølv um økta vert avlyst — og so krysset i døri ser at
// det alt er teke og ikkje tek eitt til.
func (db *Database) BokPrivatTime(eventID, userID int64, naa time.Time) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Paameldingi fyrst: klippet vert ført paa henne, so ho lyt finnast.
	// `UNIQUE(user_id, event_id)` gjer at `OR IGNORE` er heile
	// prøvingi: er ho alt paameld, stend paameldingi som ho stend, og
	// klippet under ser sjølv um det alt er teke.
	if _, err := tx.Exec(`
		INSERT OR IGNORE INTO event_signups (event_id, user_id, signup_date)
		VALUES (?, ?, ?)`, eventID, userID, veggtekst(naa)); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE events SET current_enrolment = (
			SELECT COUNT(*) FROM event_signups WHERE event_id = ?
		) WHERE id = ?`, eventID, eventID); err != nil {
		return err
	}

	// Kortet som gjeng ut fyrst vert nytta fyrst — same regelen som i
	// døri, og av same grunnen: eit kort med kort frist skal ikkje liggja
	// urørt og gaa ut medan klippi vert tekne av eit som varer lenger.
	// Er klippet alt teke for denne paameldingi, er det teke. Same
	// prøva som i døri, og av same grunnen.
	var alt sql.NullInt64
	if err := tx.QueryRow(`
		SELECT klipp_kort_id FROM event_signups
		WHERE event_id = ? AND user_id = ?`, eventID, userID).Scan(&alt); err == nil && alt.Valid {
		return tx.Commit()
	}

	var kortID int64
	err = tx.QueryRow(`
		SELECT uk.id
		FROM user_klippekort uk
		JOIN klippekort_packages kp ON kp.id = uk.package_id
		WHERE uk.user_id = ? AND kp.category = ?
		  AND uk.is_active = TRUE AND uk.remaining_klipp > 0
		  AND uk.expiry_date > ?
		ORDER BY uk.expiry_date ASC
		LIMIT 1`, userID, PTKategori, veggtekst(naa)).Scan(&kortID)
	if err == sql.ErrNoRows {
		// Ingen klipp. Økta stend; ingi skuld vert skrivi.
		return tx.Commit()
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
		UPDATE user_klippekort SET remaining_klipp = remaining_klipp - 1
		WHERE id = ? AND remaining_klipp > 0`, kortID); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE event_signups SET klipp_kort_id = ?, skulda = 0
		WHERE event_id = ? AND user_id = ?`, kortID, eventID, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// MigrerLaga legg til kolonna som seier kven som sette timen upp.
//
// `teacher_name` er fri tekst — namnet paa den som *held* timen — og det
// er ikkje det same spursmaalet. Segjer nokon fraa seg ei PT-økt, skal
// meldingi til den som sette henne upp, og eit namn er ikkje ein
// motakar: tvo kann heita det same, og ein lærar treng ikkje vera
// brukar. Difor eit id.
//
// Timar som stod der fyre kolonna kom hev NULL, og det er eit ærleg
// svar: ingen veit. Meldingi gjeng til administrasjonen aaleine daa.
func MigrerLaga(db *sql.DB) error {
	var finst bool
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('events') WHERE name='created_by'").
		Scan(&finst); err != nil {
		return err
	}
	if finst {
		return nil
	}
	_, err := db.Exec("ALTER TABLE events ADD COLUMN created_by INTEGER REFERENCES users(id)")
	return err
}

// SettLagaAv skriv kven som sette timen upp.
func (db *Database) SettLagaAv(eventID, userID int64) error {
	_, err := db.Conn.Exec(`UPDATE events SET created_by = ? WHERE id = ?`, userID, eventID)
	return err
}

// ErPrivatTime segjer um timen er sett av til den eine.
//
// Spurningi og ikkje eit felt paa modellen: `private_user_id` er eit
// vilkaar i alle spurningane som byggjer timeplanen, og ikkje noko som
// vert lese ut med timen. Aa henga det paa modellen hadde tydt aa henta
// det i tjuge spurningar for aa lesa det i ein.
func (db *Database) ErPrivatTime(eventID int64) (bool, error) {
	var eigar sql.NullInt64
	if err := db.Conn.QueryRow(
		`SELECT private_user_id FROM events WHERE id = ?`, eventID).Scan(&eigar); err != nil {
		return false, err
	}
	return eigar.Valid && eigar.Int64 > 0, nil
}

// AvlysPrivatTime er det som hender naar den eine segjer fraa seg økta
// si: klippet kjem attende, paameldingi gjeng, og meldingi vert skrivi i
// den same transaksjonen.
//
// Klippet kjem attende av di det er regelen huset alt hev — «avlyser ho,
// eller møter ho ikkje, kostar det ingen ting» (klippbruk.go). Ei PT-økt
// er klippa ved tingingi og ikkje i døri, so utan dette hadde ho vore
// den eine timen i huset som kosta noko ein ikkje fekk. Er tidi for
// knapp til at det er rimeleg, er det ei avgjerd eit menneske tek — og
// meldingi er nettupp det som legg henne framfyre eit.
//
// Alt i éin transaksjon: ei melding um ei avlysing som so vart rulla
// attende sender nokon av garde etter ein time som stend.
func (db *Database) AvlysPrivatTime(eventID, userID int64, naa time.Time) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var tittel string
	var start string
	var laga sql.NullInt64
	if err := tx.QueryRow(
		`SELECT title, start_time, created_by FROM events WHERE id = ?`, eventID).
		Scan(&tittel, &start, &laga); err != nil {
		return err
	}
	timestart, _ := time.Parse("2006-01-02 15:04:05", start)

	// Klippet fyrst, medan paameldingi som ber det framleis finst.
	if err := gjevAttKlipp(tx, eventID, userID); err != nil {
		return err
	}

	res, err := tx.Exec(
		`DELETE FROM event_signups WHERE event_id = ? AND user_id = ?`, eventID, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	if _, err := tx.Exec(`
		UPDATE events SET current_enrolment = (
			SELECT COUNT(*) FROM event_signups WHERE event_id = ?
		) WHERE id = ?`, eventID, eventID); err != nil {
		return err
	}

	if err := db.LagMelding(tx, MeldingAvlystPrivat, eventID, userID, laga,
		tittel, timestart, naa); err != nil {
		return err
	}
	return tx.Commit()
}
