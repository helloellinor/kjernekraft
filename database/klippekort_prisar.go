package database

import "kjernekraft/models"

// Skrivesida for klippekortpakkane.
//
// Lesesida har alltid vore der (GetAllKlippekortPackages); det gjekk berre
// ikkje an å endre ein pakke frå administrasjonen. Prisen per klipp er
// rekna og ikkje skriven: han er prisen delt på talet på klipp, og eit
// felt der ein kan skrive noko som ikkje stemmer med det, er eit felt som
// kan lyge.

// UpdateKlippekortPackage skriv om ein pakke.
func (db *Database) UpdateKlippekortPackage(p models.KlippekortPackage) error {
	perKlipp := 0
	if p.KlippCount > 0 {
		perKlipp = p.Price / p.KlippCount
	}
	_, err := db.Conn.Exec(`UPDATE klippekort_packages SET
		name = ?, category = ?, klipp_count = ?, price = ?, price_per_session = ?,
		valid_days = ?, is_popular = ?
		WHERE id = ?`,
		p.Name, p.Category, p.KlippCount, p.Price, perKlipp,
		p.ValidDays, p.IsPopular, p.ID)
	return err
}

// CreateKlippekortPackage lagar ein ny pakke.
func (db *Database) CreateKlippekortPackage(p models.KlippekortPackage) (int64, error) {
	perKlipp := 0
	if p.KlippCount > 0 {
		perKlipp = p.Price / p.KlippCount
	}
	res, err := db.Conn.Exec(`INSERT INTO klippekort_packages
		(name, category, klipp_count, price, price_per_session, description, valid_days, active, is_popular)
		VALUES (?, ?, ?, ?, ?, ?, ?, TRUE, ?)`,
		p.Name, p.Category, p.KlippCount, p.Price, perKlipp, p.Description, p.ValidDays, p.IsPopular)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// DeactivateKlippekortPackage tek ein pakke ut av sal. Mjuk sletting:
// nokon kan ha kjøpt han, og då skal klippa deira ikkje forsvinne.
func (db *Database) DeactivateKlippekortPackage(id int64) error {
	_, err := db.Conn.Exec(`UPDATE klippekort_packages SET active = FALSE WHERE id = ?`, id)
	return err
}
