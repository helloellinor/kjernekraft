package database

import (
	"database/sql"
	"log"
)

// Løyvi heitte løyve.
//
// «Løyvet» tydde tvo ting so snart gruppone kom: kva du fær *gjera* med
// studioet (admin, lærar) og kven du *høyrer til* (Reformer). Det fyrste
// er eit løyve. Det andre er ei gruppe, og han hev si eigi tabell no.
//
// Namnet er ikkje pynt: so lenge dei heitte det same, var «gjev tilgang
// til reformer» og «gjer til administrator» den same handlingi i koden,
// og det er den slags likskap som ein dag vert ein feilklikk.

func harTabell(db *sql.DB, namn string) (bool, error) {
	var n int
	err := db.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?", namn).Scan(&n)
	return n > 0, err
}

func harKolonne(db *sql.DB, tabell, kolonne string) (bool, error) {
	rows, err := db.Query("SELECT name FROM pragma_table_info(?)", tabell)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return false, err
		}
		if n == kolonne {
			return true, nil
		}
	}
	return false, rows.Err()
}

// MigrerLøyve døyper om tabellane og kolonna.
//
// Kvar prøve ser etter *båe* namni: ein base som alt er døypt om skal
// ikkje faa noko gjort med seg, og ein som ikkje er det skal ikkje faa
// tvo tabellar.
func MigrerLøyve(db *sql.DB) error {
	for _, p := range []struct{ frå, til string }{
		{"roles", "loyve"},
		{"user_roles", "brukarloyve"},
	} {
		gamal, err := harTabell(db, p.frå)
		if err != nil {
			return err
		}
		ny, err := harTabell(db, p.til)
		if err != nil {
			return err
		}
		if gamal && !ny {
			if _, err := db.Exec("ALTER TABLE " + p.frå + " RENAME TO " + p.til); err != nil {
				return err
			}
			log.Printf("Døypte om %s til %s.", p.frå, p.til)
		}
	}

	finst, err := harTabell(db, "brukarloyve")
	if err != nil || !finst {
		return err
	}
	gamal, err := harKolonne(db, "brukarloyve", "role_id")
	if err != nil {
		return err
	}
	ny, err := harKolonne(db, "brukarloyve", "loyve_id")
	if err != nil {
		return err
	}
	if gamal && !ny {
		if _, err := db.Exec("ALTER TABLE brukarloyve RENAME COLUMN role_id TO loyve_id"); err != nil {
			return err
		}
		log.Println("Døypte om role_id til loyve_id paa brukarloyve.")
	}
	return nil
}
