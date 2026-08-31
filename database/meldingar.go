package database

import (
	"database/sql"
	"time"
)

// Notices.
//
// Something happens that a person needs to know about and that no query in
// the house would find on its own: someone gives up a PT session set aside
// for them. The class stands empty, the teacher has reserved the time, and
// without a notice nobody knows until someone turns up.
//
// It is stored, not sent. Deliberately, for two reasons:
//
//   - A notice that is only sent is gone if the sending fails. This one
//     stands in the database until someone marks it handled, and the
//     Meldingar tab is where it is read.
//   - `sendt` is the socket for an email sender that does not exist yet.
//     When it arrives, its job is a query for rows where sendt is NULL —
//     and nothing else in the house has to move.
//
// `slag` is what happened. It is there from day one with a single value,
// because a sender that has to know what it is sending is a sender that
// has to change for every new thing that can happen.
const (
	// MeldingAvlystPrivat: the one person gave up their session.
	MeldingAvlystPrivat = "avlyst-privattime"
)

// utfoerar is the small part of *sql.DB and *sql.Tx a notice needs, so it
// can be written in the same transaction as what it reports. A notice about
// something that then rolled back is worse than no notice. Its sibling
// `rader` does the same for queries (klippbruk.go).
type utfoerar interface {
	Exec(string, ...any) (sql.Result, error)
}

// Melding is a row in the tab. The title and the time are written down
// *here* rather than looked up when it is read: the class may have been
// cancelled, moved or deleted meanwhile, and the notice should still say
// what happened — not what stands now.
type Melding struct {
	ID        int64
	Slag      string
	EventID   int64
	Tittel    string
	TimeStart time.Time
	// The one who gave up the class, and the teacher who set it up.
	FråNamn  string
	FråEpost string
	TilNamn  string
	Laga     time.Time
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

// LagMelding writes the notice. `til` is the teacher who set the class up;
// null when the class predates that column and nobody knows who it was.
// The admins see it either way — the tab is all they see.
func (db *Database) LagMelding(q utfoerar, slag string, eventID, frå int64, til sql.NullInt64,
	tittel string, timestart, nå time.Time) error {
	_, err := q.Exec(`
		INSERT INTO meldingar (slag, event_id, fraa_user_id, til_user_id, tittel, timestart, laga)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		slag, eventID, frå, til, tittel, veggtekst(timestart), veggtekst(nå))
	return err
}

// VentandeMeldingar are the unhandled ones, newest first: what just
// happened is what you can still do something about.
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
			&m.FråNamn, &m.FråEpost, &m.TilNamn); err != nil {
			return nil, err
		}
		m.TimeStart, _ = time.Parse("2006-01-02 15:04:05", start)
		m.Laga, _ = time.Parse("2006-01-02 15:04:05", laga)
		ut = append(ut, m)
	}
	return ut, rows.Err()
}

// HandsamaMelding takes the notice out of the tab. It is not deleted: what
// happened, happened, and an empty tab should mean "nothing is waiting",
// not "nothing has happened".
func (db *Database) HandsamaMelding(id int64, nå time.Time) error {
	_, err := db.Conn.Exec(
		`UPDATE meldingar SET handsama = ? WHERE id = ? AND handsama IS NULL`,
		veggtekst(nå), id)
	return err
}
