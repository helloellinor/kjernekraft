package database

import "database/sql"

// Namn på produkt.
//
// Eit medlemskap heiter ikkje noko i basen lenger — det *er* nokre fakta,
// og namnet vert skrive av systemet i det språket den som les har valt.
// «12 månaders binding» er ikkje eit ord nokon skreiv; det er noko basen
// alt visste.
//
// Men studioet skal kunne overstyre. Eit haustkampanjekort heiter
// «Hausttilbod» og ikkje «Årskort», og det er ikkje noko systemet kan
// rekne seg fram til. Ei overstyring gjeld eitt språk — det språket
// administrasjonen stod i då ho vart skriven — og resten held fram med
// det genererte namnet.
//
// Tabellen er felles for medlemskap og klippekort. Slaget skil dei.

// MigrerProduktnamn lagar tabellen, og flyttar namna som alt står i
// `memberships` og `klippekort_packages` inn i han som bokmålsnamn.
//
// Dei er skrivne på bokmål — «12-måneder», «Ingen binding», «5 Klipp -
// Gruppetimer» — so det er det dei er. Nynorsk og engelsk får genererte
// namn frå fakta, og bokmål held fram med å sjå det studioet skreiv.
func (db *Database) MigrerProduktnamn() error {
	if _, err := db.Conn.Exec(`CREATE TABLE IF NOT EXISTS produktnamn (
		slag TEXT NOT NULL,
		produkt_id INTEGER NOT NULL,
		lang TEXT NOT NULL,
		namn TEXT NOT NULL,
		PRIMARY KEY (slag, produkt_id, lang)
	)`); err != nil {
		return err
	}

	var alt int
	if err := db.Conn.QueryRow(`SELECT COUNT(*) FROM produktnamn`).Scan(&alt); err != nil {
		return err
	}
	if alt > 0 {
		return nil // flytta ein gong er nok
	}

	for _, k := range []struct{ slag, tabell string }{
		{"medlemskap", "memberships"},
		{"klippekort", "klippekort_packages"},
	} {
		// Les ferdig fyrst, skriv etterpå. Å skrive medan markøren står
		// open på den same sambandet, er noko SQLite ikkje gjer — og av
		// di migreringa køyrer ved oppstart med log.Fatal attom seg, var
		// resultatet ein tenar som bygde fint og so la seg ned med det
		// same. Han sa ✓ i loggen og var likevel ikkje der.
		type namnrad struct {
			id   int
			namn string
		}
		var funne []namnrad

		rows, err := db.Conn.Query(`SELECT id, name FROM ` + k.tabell)
		if err != nil {
			return err
		}
		for rows.Next() {
			var r namnrad
			if err := rows.Scan(&r.id, &r.namn); err != nil {
				rows.Close()
				return err
			}
			if r.namn != "" {
				funne = append(funne, r)
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()

		for _, r := range funne {
			if _, err := db.Conn.Exec(
				`INSERT OR IGNORE INTO produktnamn (slag, produkt_id, lang, namn) VALUES (?, ?, 'nb', ?)`,
				k.slag, r.id, r.namn); err != nil {
				return err
			}
		}
	}
	return nil
}

// Produktnamn gjev overstyringane for eit slag, som id → språk → namn.
func (db *Database) Produktnamn(slag string) (map[int]map[string]string, error) {
	rows, err := db.Conn.Query(
		`SELECT produkt_id, lang, namn FROM produktnamn WHERE slag = ?`, slag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ut := map[int]map[string]string{}
	for rows.Next() {
		var id int
		var lang, namn string
		if err := rows.Scan(&id, &lang, &namn); err != nil {
			return nil, err
		}
		if ut[id] == nil {
			ut[id] = map[string]string{}
		}
		ut[id][lang] = namn
	}
	return ut, nil
}

// SetProduktnamn skriv ei overstyring. Eit tomt namn tek henne bort, so
// produktet fell attende til det genererte namnet — det er slik ein
// «angrar» eit namn utan ein eigen knapp for det.
func (db *Database) SetProduktnamn(slag string, id int, lang, namn string) error {
	if namn == "" {
		_, err := db.Conn.Exec(
			`DELETE FROM produktnamn WHERE slag = ? AND produkt_id = ? AND lang = ?`,
			slag, id, lang)
		return err
	}
	_, err := db.Conn.Exec(`INSERT INTO produktnamn (slag, produkt_id, lang, namn)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(slag, produkt_id, lang) DO UPDATE SET namn = excluded.namn`,
		slag, id, lang, namn)
	return err
}

// SlettProduktnamn tek bort alle namn for eit produkt.
func (db *Database) SlettProduktnamn(slag string, id int) error {
	_, err := db.Conn.Exec(`DELETE FROM produktnamn WHERE slag = ? AND produkt_id = ?`, slag, id)
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}
