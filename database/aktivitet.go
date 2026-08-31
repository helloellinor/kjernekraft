package database

import "time"

// AktivitetPerDag gjev talet på timar brukaren stod påmeld på, per
// dag, for timar som alt hev vore.
//
// Merk kva dette *er*: paamelde timar som hev gjenge, ikkje frammøte.
// Frammøtelista finst ikkje enno (BLUEPRINT §6), so systemet veit at du
// hadde ein plass — ikkje at du kom. Når eitt-trykks-lista kjem, er det
// denne spurningi som skal byta `event_signups` mot frammøtet, og
// ingen ting anna i heile kjeda.
//
// Datoen er dagen timen *gjekk*, ikkje dagen du melde deg på. Eit
// varmekart yver når du bestilte er eit kart yver noko anna.
// AktivitetPerDagType er det same, men brote ned på klassetype.
//
// Vekeprikken tek fargen sin av kva slag trening det var, so talet
// aaleine er ikkje nok — han lyt vita *kva*. Nykelen er dagen, verdet er
// kvar type med talet sitt den dagen; vekone vert summerte i Go, av di
// vika her byrjar på måndag og `strftime` ikkje er samd um det.
//
// Ein time utan `class_type` fell under tom streng og fær grunnfargen.
func (db *Database) AktivitetPerDagType(userID int64, frå, til time.Time) (map[string]map[string]int, error) {
	rows, err := db.Conn.Query(`
		SELECT date(e.start_time) AS dag, COALESCE(e.class_type, '') AS slag, COUNT(*)
		FROM event_signups es
		JOIN events e ON e.id = es.event_id
		WHERE es.user_id = ?
		  AND e.start_time >= ? AND e.start_time < ?
		GROUP BY dag, slag`, userID, veggtekst(frå), veggtekst(til))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ut := make(map[string]map[string]int)
	for rows.Next() {
		var dag, slag string
		var tal int
		if err := rows.Scan(&dag, &slag, &tal); err != nil {
			return nil, err
		}
		if ut[dag] == nil {
			ut[dag] = make(map[string]int)
		}
		ut[dag][slag] = tal
	}
	return ut, rows.Err()
}
