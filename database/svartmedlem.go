package database

import (
	"database/sql"
	"time"

	"kjernekraft/models"
)

// Det usynlege medlemskapet.
//
// Eit medlemskap kunde til no berre vera tvo ting: aktivt av di nokon
// hadde kjøpt det, eller ikkje der. Lærarane og utviklarane skal hava
// full tilgang utan aa kjøpa noko, og utan aa staa i prislista der folk
// vel — difor «usynleg»: han finst som medlemskap, men han er ikkje
// noko ein kann velja.
//
// Han er *avleidd av løyvet*, ikkje ei rad i user_memberships. Det er
// den viktige avgjerdi her, og grunnen er kva som skjer naar løyvet
// gjeng bort: ei lagra rad hadde vorte liggjande att og gjeve ein
// tidlegare lærar fri tilgang til nokon rydda henne for haand. Eit løyve
// som vert teki bort tek medlemskapet med seg i same augneblinken, av
// di det aldri var noko anna enn løyvet.
//
// Det tyder ogso at han ikkje kann frysast, seiast upp eller endrast:
// det finst ingen ting aa endra. Malen skal gøyma dei knappane, og
// `Tildelt` er flagget ho ser paa.

// SvartMedlemskap er namnet i basen.
const SvartMedlemskap = "Black"

// MigrerSvartMedlemskap legg til skjult-flagget og set upp Black.
//
//	skjult = TRUE   medlemskapet er ikkje noko ein kann velja. Det held
//	                honom ute or prislista og or veljarane, men han er
//	                framleis eit fullgodt medlemskap for den som hev han.
//	                Flagget er «skjult» og ikkje «synleg» so standarden —
//	                det ingen hev sagt noko om — er synleg.
func MigrerSvartMedlemskap(db *sql.DB) error {
	var finst bool
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('memberships') WHERE name='skjult'").
		Scan(&finst); err != nil {
		return err
	}
	if !finst {
		if _, err := db.Exec(
			"ALTER TABLE memberships ADD COLUMN skjult BOOLEAN NOT NULL DEFAULT FALSE"); err != nil {
			return err
		}
	}

	// Black kostar ingen ting og bind ingen. Han er ein tilgang, ikkje
	// ein avtale.
	if _, err := db.Exec(`
		INSERT INTO memberships (name, price, commitment_months, is_student_senior,
		                         is_special_offer, description, features, active, skjult)
		SELECT ?, 0, 0, FALSE, FALSE,
		       'Full tilgang. Fylgjer løyvet, og kann ikkje kjøpast.',
		       '["Ubegrensa gruppetimar","Ubegrensa reformer","Fylgjer løyvet"]',
		       TRUE, TRUE
		WHERE NOT EXISTS (SELECT 1 FROM memberships WHERE name = ?)`,
		SvartMedlemskap, SvartMedlemskap); err != nil {
		return err
	}

	// Fanst Black alt — han gjorde det i basar som hadde flagget fyre
	// det skifte namn — vart han ikkje sett inn paa nytt, og daa hadde
	// han stade i prislista med standardverdien til den nye kolonna.
	// Han skal vera skjult, uansett kva veg han kom inn.
	_, err := db.Exec(
		`UPDATE memberships SET skjult = TRUE WHERE name = ?`, SvartMedlemskap)
	return err
}

// HarFriMedlemskap segjer um brukaren fær Black.
//
// Tvo kjeldor, og dei er ulike med vilje:
//
//   - lærarløyvet, som ein administrator gjev ut gjenom flata
//   - utviklarlista, som stend i ei fil paa tenaren og ikkje kann
//     skrivast gjenom nettet i det heile (sjaa utviklar.go)
//
// Eit løyve ein administrator kann gjeva seg sjølv, og ei han ikkje kann.
func (db *Database) HarFriMedlemskap(userID int64) (bool, error) {
	var tal int
	if err := db.Conn.QueryRow(`
		SELECT COUNT(*) FROM brukarloyve ur
		JOIN loyve r ON r.id = ur.loyve_id
		WHERE ur.user_id = ? AND r.name = ?`, userID, LoyveLaerar).Scan(&tal); err != nil {
		return false, err
	}
	if tal > 0 {
		return true, nil
	}
	return db.ErUtviklarID(userID)
}

// svartMedlemskapFor byggjer det avleidde medlemskapet.
//
// Datoane er ikkje dikta upp for aa sjaa ut som eit kjøp: han byrja den
// dagen brukaren vart oppretta, og han vert ikkje fornya — han varer so
// lenge løyvet varer. RenewalDate eit aar fram er berre so sidor som
// reknar «dagar til fornying» ikkje viser noko rart.
func (db *Database) svartMedlemskapFor(userID int64) (*models.MembershipWithDetails, error) {
	var m models.Membership
	// `skjult` lyt vera med. Utan henne kom Black attende med flagget
	// usett, og MedlemskapNamn — som nyttar nett det flagget til aa sjaa
	// at eit medlemskap ikkje er eit bindingsprodukt — rekna namnet hans
	// ut or bindingi i staden. Med null maanader binding vart «Black»
	// til «Månadskort» paa skjermen, medan basen heile tidi sa Black.
	err := db.Conn.QueryRow(`
		SELECT id, name, price, commitment_months, is_student_senior,
		       is_special_offer, description, features, active, skjult
		FROM memberships WHERE name = ?`, SvartMedlemskap).
		Scan(&m.ID, &m.Name, &m.Price, &m.CommitmentMonths, &m.IsStudentSenior,
			&m.IsSpecialOffer, &m.Description, &m.Features, &m.Active, &m.Skjult)
	if err != nil {
		return nil, err
	}

	// Same grunnen som i MedlemSidan: kolonna beint, ikkje eit uttrykk.
	start, err := db.MedlemSidan(userID)
	if err != nil || start.IsZero() {
		start = time.Now()
	}

	ut := &models.MembershipWithDetails{Membership: m}
	ut.UserMembership = models.UserMembership{
		UserID:       int(userID),
		MembershipID: m.ID,
		Status:       "active",
		StartDate:    start,
		RenewalDate:  time.Now().AddDate(1, 0, 0),
		LastBilled:   start,
		CreatedAt:    start,
	}
	// Ingen ting aa seia upp eller frysa: det finst ingi rad aa endra.
	ut.CanCancel = false
	ut.CanPause = false
	ut.Tildelt = true
	return ut, nil
}

// SynkFriMedlemskap stoggar faktureringi naar nokon vert forfremja.
//
// Det held ikkje aa syna Black paa skjermen. Ligg det ein aktiv avtale i
// user_memberships, er det han som vert fakturert — kortet er berre
// biletet. Ein lærar som fekk løyvet i gaar og eit trekk i dag hev
// betalt for noko han fær gratis.
//
// Difor: løyvet kjem, avtalen stend av. Han vert *avslutta*, ikkje
// sletta, so det stend att i basen kva han var og naar han gjekk av —
// ein rekneskap treng aa kunna svara for eit trekk som alt er gjort.
//
// Det er ein einvegs port med vilje. Gjeng løyvet bort att, kjem ikkje
// den gamle avtalen attende av seg sjølv: han vart avslutta, og det som
// er avslutta lyt kjøpast paa nytt. Aa vekkja ein avtale til live og
// taka til aa krevja pengar for honom att, utan at nokon bad um det, er
// verre enn aa lata folk velja sjølve.
func (db *Database) SynkFriMedlemskap(userID int64) (bool, error) {
	fri, err := db.HarFriMedlemskap(userID)
	if err != nil || !fri {
		return false, err
	}

	res, err := db.Conn.Exec(`
		UPDATE user_memberships
		SET status = 'cancelled', end_date = CURRENT_TIMESTAMP
		WHERE user_id = ?
		  AND status IN ('active', 'paused', 'freeze_requested')`, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// MedlemSidan gjev dagen kontoen vart oppretta.
//
// «Medlem sidan» er ikkje det same som «denne avtalen tok til». Ein som
// kom i 2019 og bytte medlemskap i fjor hev vore medlem sidan 2019;
// tek ein datoen fraa user_memberships, vert han nullstilt kvar gong
// nokon byter plan, og kortet ljug um noko folk er stolte av.
//
// Det er ogso den einaste kjelda som *finst* for eit tildelt medlemskap:
// Black hev ingi rad i user_memberships i det heile.
func (db *Database) MedlemSidan(userID int64) (time.Time, error) {
	// Utan COALESCE i spurningi, med vilje. SQLite lagrar tidi som tekst,
	// og drivaren gjer henne um til time.Time *berre* naar kolonna er
	// utsagd DATETIME. Eit COALESCE er eit uttrykk og ikkje ei kolonna
	// lenger, so verdet kom attende som streng og skanningi fall med
	// «unsupported Scan». Fallet tek me i Go i staden.
	var naar sql.NullTime
	if err := db.Conn.QueryRow(
		`SELECT created_at FROM users WHERE id = ?`, userID).Scan(&naar); err != nil {
		return time.Time{}, err
	}
	if !naar.Valid {
		return time.Time{}, nil
	}
	return naar.Time, nil
}
