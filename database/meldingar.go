package database

import (
	"database/sql"
	"time"
)

// Meldingane.
//
// Noko hender som eit menneske lyt vita um, og som ingi spurning i huset
// ville ha funne av seg sjølv: nokon segjer fraa seg ei PT-økt som er
// sett av til honom. Timen stend att tom, læraren hev sett av tidi, og
// utan ei melding er det ingen som veit det fyre nokon møter upp.
//
// Ho er lagra og ikkje send. Det er med vilje, og det er tvo ting:
//
//   - Ei melding som berre vert send er ei melding som er burte um
//     sendingi feilar. Denne stend i basen til nokon hev sagt at ho er
//     handsama, og fana «Meldingar» i administrasjonen er der ho vert
//     lesi.
//   - `sendt` er søkket for ein e-postsendar som ikkje finst enno. Naar
//     han kjem, er arbeidet hans ei spurning etter rader der `sendt` er
//     NULL — og ingen ting anna i huset treng røra seg. Kolonna er
//     tom til det hender.
//
// `slag` er kva som hende. Han er der fraa fyrste dagen med éin verdi,
// av di ein sendar som lyt vita kva han sender er ein sendar som lyt
// endrast for kvar ny ting som kann henda.
const (
	// MeldingAvlystPrivat: den eine hev sagt fraa seg økta si.
	MeldingAvlystPrivat = "avlyst-privattime"
)

// utfoerar er det vesle av *sql.DB og *sql.Tx meldingi treng, so ho kann
// skrivast i den same transaksjonen som det ho fortel um. Ei melding um
// noko som so vart rulla attende er verre enn ingi melding. Syskenet
// `rader` gjer det same for spurningar (klippbruk.go).
type utfoerar interface {
	Exec(string, ...any) (sql.Result, error)
}

// Melding er ei rad i fana. Tittelen og tidi er skrivne av *her* og
// ikkje slegne upp naar ho vert lesi: timen kann vera avlyst, flutt
// eller sletta i millomtidi, og meldingi skal framleis fortelja kva som
// hende — ikkje kva som stend no.
type Melding struct {
	ID        int64
	Slag      string
	EventID   int64
	Tittel    string
	TimeStart time.Time
	// Den som sa fraa seg timen, og læraren som sette honom upp.
	FraaNamn  string
	FraaEpost string
	TilNamn   string
	Laga      time.Time
}

func MigrerMeldingar(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS meldingar (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slag TEXT NOT NULL,
		event_id INTEGER NOT NULL,
		fraa_user_id INTEGER NOT NULL REFERENCES users(id),
		til_user_id INTEGER REFERENCES users(id),
		tittel TEXT NOT NULL DEFAULT '',
		timestart DATETIME NOT NULL,
		laga DATETIME NOT NULL,
		handsama DATETIME,
		sendt DATETIME
	)`)
	return err
}

// LagMelding skriv meldingi. `til` er læraren som sette timen upp; null
// naar timen er eldre enn den kolonna og ingen veit kven det var.
// Administrasjonen ser henne uansett — fana er alt det dei ser.
func (db *Database) LagMelding(q utfoerar, slag string, eventID, fraa int64, til sql.NullInt64,
	tittel string, timestart, naa time.Time) error {
	_, err := q.Exec(`
		INSERT INTO meldingar (slag, event_id, fraa_user_id, til_user_id, tittel, timestart, laga)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		slag, eventID, fraa, til, tittel, veggtekst(timestart), veggtekst(naa))
	return err
}

// VentandeMeldingar er dei som ikkje er handsama. Rekkjefylgdi er den
// nyaste fyrst: det som nett hende er det ein kann gjera noko med.
func (db *Database) VentandeMeldingar() ([]Melding, error) {
	rows, err := db.Conn.Query(`
		SELECT m.id, m.slag, m.event_id, m.tittel, m.timestart, m.laga,
		       COALESCE(f.name, ''), COALESCE(f.email, ''), COALESCE(tl.name, '')
		FROM meldingar m
		JOIN users f ON f.id = m.fraa_user_id
		LEFT JOIN users tl ON tl.id = m.til_user_id
		WHERE m.handsama IS NULL
		ORDER BY m.laga DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []Melding
	for rows.Next() {
		var m Melding
		var start, laga string
		if err := rows.Scan(&m.ID, &m.Slag, &m.EventID, &m.Tittel, &start, &laga,
			&m.FraaNamn, &m.FraaEpost, &m.TilNamn); err != nil {
			return nil, err
		}
		m.TimeStart, _ = time.Parse("2006-01-02 15:04:05", start)
		m.Laga, _ = time.Parse("2006-01-02 15:04:05", laga)
		ut = append(ut, m)
	}
	return ut, rows.Err()
}

// HandsamaMelding tek meldingi ut or fana. Ho vert ikkje sletta: det som
// hende, hende, og ei tom fana skal tyda «ingen ting ventar» og ikkje
// «ingen ting hev hendt».
func (db *Database) HandsamaMelding(id int64, naa time.Time) error {
	_, err := db.Conn.Exec(
		`UPDATE meldingar SET handsama = ? WHERE id = ? AND handsama IS NULL`,
		veggtekst(naa), id)
	return err
}
