package database

import (
	"database/sql"
	"time"
)

// Klippet vert brukt naar nokon *var der*, ikkje naar dei melde seg paa.
//
// Til no vart eit klipp aldri brukt i det heile: `remaining_klipp` vart
// berre skrive av kjøpet og av testdata-stokkinga. Ein kunde gaa paa
// reformer heile hausten og framleis hava ti klipp att paa kortet.
//
// Kvifor krysset og ikkje paameldingi: huset skil alt paa dei tvo — sjaa
// NettFrammott, «ei paamelding aaleine segjer at ho hadde tenkt seg
// dit». Eit klipp er betaling for ein time ein *fekk*, so det fylgjer
// krysset. Avlyser ho, eller møter ho ikkje, kostar det ingen ting; tek
// nokon krysset bort att, kjem klippet attende (sjaa FjernFrammote).
//
// Kva ein time kostar er ikkje ei liste i koden. Gruppetimane i Salen
// ligg i medlemskapet og kostar ingen klipp; Reformer og Personlig
// Trening vert selde som klippekort. Bandet millom dei tvo er namnet:
// timen sin `class_type` mot pakka si `category`. Difor finn me
// kategorien ved eit uppslag og ikkje ved ein `switch` — legg nokon inn
// ein ny klippekort-kategori, verkar han med ein gong, og ingen treng
// hugsa at det ogso stend ei liste i Go.

// MigrerKlippbruk legg til kolonnorne krysset treng.
//
//	klipp_kort_id  kortet klippet vart teke fraa. NULL = ikkje klippa.
//	               Det er *kva kort* som gjer at eit angra kryss kann
//	               leggja klippet attende paa rett stad — utan dette
//	               laut ein gissa, og eit kort som gjekk ut i
//	               millomtidi hadde fenge klippet.
//	skulda         1 naar nokon vart kryssa inn utan klipp aa taka av.
//	               Timen er gjeven og skal ikkje gaa tapt or rekneskapen
//	               berre av di kortet var tomt i døri.
func MigrerKlippbruk(db *sql.DB) error {
	for _, k := range []struct{ namn, def string }{
		{"klipp_kort_id", "ALTER TABLE event_signups ADD COLUMN klipp_kort_id INTEGER"},
		{"skulda", "ALTER TABLE event_signups ADD COLUMN skulda INTEGER NOT NULL DEFAULT 0"},
	} {
		var finst bool
		if err := db.QueryRow(
			"SELECT COUNT(*) FROM pragma_table_info('event_signups') WHERE name=?", k.namn).
			Scan(&finst); err != nil {
			return err
		}
		if finst {
			continue
		}
		if _, err := db.Exec(k.def); err != nil {
			return err
		}
	}
	return nil
}

// klippKategori gjev klippekort-kategorien ein timetype skal betalast
// med, og false naar timen ikkje kostar klipp.
//
// Samanlikninga er slaakk med vilje: administrasjonen skriv `class_type`
// fritt, so «Reformer», «reformer» og « Reformer » er den same timen.
func klippKategori(q rader, timetype string) (string, bool) {
	if timetype == "" {
		return "", false
	}
	var kat string
	err := q.QueryRow(`
		SELECT category FROM klippekort_packages
		WHERE LOWER(TRIM(category)) = LOWER(TRIM(?))
		LIMIT 1`, timetype).Scan(&kat)
	if err != nil {
		return "", false
	}
	return kat, true
}

// rader er det vesle av *sql.DB og *sql.Tx me nyttar her, so den same
// koden kann gaa baade i og utanum ein transaksjon.
type rader interface {
	QueryRow(string, ...any) *sql.Row
}

// brukKlipp tek eit klipp for timen, inne i transaksjonen krysset gjeng i.
//
// Kortet som gjeng ut fyrst vert nytta fyrst. Elles kunde eit kort med
// kort frist liggja urørt og gaa ut medan klippi vart tekne av eit kort
// som varer til neste sumar.
func brukKlipp(tx *sql.Tx, eventID, userID int64, naa time.Time) error {
	var timetype string
	if err := tx.QueryRow(
		`SELECT COALESCE(class_type, '') FROM events WHERE id = ?`, eventID).
		Scan(&timetype); err != nil {
		return err
	}

	kat, kostar := klippKategori(tx, timetype)
	if !kostar {
		// Gruppetime: medlemskapet dekkjer honom.
		return nil
	}

	var kortID int64
	err := tx.QueryRow(`
		SELECT uk.id
		FROM user_klippekort uk
		JOIN klippekort_packages kp ON kp.id = uk.package_id
		WHERE uk.user_id = ? AND kp.category = ?
		  AND uk.is_active = TRUE AND uk.remaining_klipp > 0
		  AND uk.expiry_date > ?
		ORDER BY uk.expiry_date ASC
		LIMIT 1`, userID, kat, veggtekst(naa)).Scan(&kortID)

	if err == sql.ErrNoRows {
		// Ingen klipp aa taka av. Ho stend i døri og timen byrjar; me
		// skriv skuldi og lyfter henne fram for administrasjonen i
		// staden for aa stogga henne her.
		_, err := tx.Exec(`
			UPDATE event_signups SET skulda = 1
			WHERE event_id = ? AND user_id = ?`, eventID, userID)
		return err
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`
		UPDATE user_klippekort SET remaining_klipp = remaining_klipp - 1
		WHERE id = ? AND remaining_klipp > 0`, kortID); err != nil {
		return err
	}
	_, err = tx.Exec(`
		UPDATE event_signups SET klipp_kort_id = ?, skulda = 0
		WHERE event_id = ? AND user_id = ?`, kortID, eventID, userID)
	return err
}

// gjevAttKlipp legg klippet attende naar eit kryss vert teke bort.
func gjevAttKlipp(tx *sql.Tx, eventID, userID int64) error {
	var kortID sql.NullInt64
	if err := tx.QueryRow(`
		SELECT klipp_kort_id FROM event_signups
		WHERE event_id = ? AND user_id = ?`, eventID, userID).Scan(&kortID); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	if kortID.Valid {
		// Taket paa total_klipp: eit kort skal ikkje kunna gjeva fleire
		// klipp attende enn det hadde. Kann berre henda um nokon hev
		// rota i basen for haand, men eit kort med 11 av 10 er verre
		// enn eit tapt klipp.
		if _, err := tx.Exec(`
			UPDATE user_klippekort
			SET remaining_klipp = MIN(remaining_klipp + 1, total_klipp)
			WHERE id = ?`, kortID.Int64); err != nil {
			return err
		}
	}

	_, err := tx.Exec(`
		UPDATE event_signups SET klipp_kort_id = NULL, skulda = 0
		WHERE event_id = ? AND user_id = ?`, eventID, userID)
	return err
}
