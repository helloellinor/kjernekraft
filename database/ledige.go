package database

import (
	"time"

	"kjernekraft/models"
)

// Timar med ledig plass i dag og i morgon.
//
// Heimesida spør ikkje «kva går i dag» lenger — ho spør «kvar er det
// plass». Ein time som er full, er ikkje eit tilbod; han er berre ei rad
// ein les forbi. Og i dag åleine er for kort: er klokka sju om kvelden,
// er det ingenting att av dagen, og då står bolken tom kvar kveld.
// LedigeTimar gjev det `sjaaarID` skal sjå: opne timar, og private som
// er hans eigne. Sjå privattime.go.
func (db *Database) LedigeTimar(sjaårID int64, nå time.Time) ([]models.Event, error) {
	start := time.Date(nå.Year(), nå.Month(), nå.Day(), 0, 0, 0, 0, nå.Location())
	slutt := start.AddDate(0, 0, 2)
	rows, err := db.Conn.Query(`SELECT `+eventKolonnar+eventFrå+`
		WHERE e.start_time >= ? AND e.start_time < ?
		  AND `+synlegFor+`
		ORDER BY e.start_time ASC, e.id ASC`,
		veggtekst(start), veggtekst(slutt), sjaårID, sjaårID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []models.Event
	for rows.Next() {
		var e models.Event
		if err := skannTime(rows, &e); err != nil {
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
