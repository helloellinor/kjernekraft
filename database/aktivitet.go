package database

import "time"

// AktivitetPerDag gjev talet paa timar brukaren stod paameld paa, per
// dag, for timar som alt hev vore.
//
// Merk kva dette *er*: paamelde timar som hev gjenge, ikkje frammøte.
// Frammøtelista finst ikkje enno (BLUEPRINT §6), so systemet veit at du
// hadde ein plass — ikkje at du kom. Naar eitt-trykks-lista kjem, er det
// denne spurningi som skal byta `event_signups` mot frammøtet, og
// ingen ting anna i heile kjeda.
//
// Datoen er dagen timen *gjekk*, ikkje dagen du melde deg paa. Eit
// varmekart yver naar du bestilte er eit kart yver noko anna.
func (db *Database) AktivitetPerDag(userID int64, fraa, til time.Time) (map[string]int, error) {
	rows, err := db.Conn.Query(`
		SELECT date(e.start_time) AS dag, COUNT(*)
		FROM event_signups es
		JOIN events e ON e.id = es.event_id
		WHERE es.user_id = ?
		  AND e.start_time >= ? AND e.start_time < ?
		GROUP BY dag`, userID, veggtekst(fraa), veggtekst(til))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ut := make(map[string]int)
	for rows.Next() {
		var dag string
		var tal int
		if err := rows.Scan(&dag, &tal); err != nil {
			return nil, err
		}
		ut[dag] = tal
	}
	return ut, rows.Err()
}
