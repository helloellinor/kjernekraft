// testbrukar makes one test user with something worth looking at.
//
//	go run ./scripts/testbrukar
//	go run ./scripts/testbrukar -slett     # removes her again
//
// She is not an admin. That is the point: what you show on a screen should
// show what a student sees, and an admin sees something else. She gets
// `user` and nothing more, so /admin answers 403 for her.
//
// The script is *not* random. A random user gives a heat board of even
// noise, and even noise looks like a fault — you read nothing from it. A
// pattern you recognise reads at once, so Solfrid has a story:
//
//	March   she starts, cautiously — one class a week
//	April   it sticks; two a week, and a yoga on top
//	May     three a week, and she finds fascia
//	June    the peak. Four a week, and her first reformer
//	July    away for two weeks, then carefully back
//	August  three a week again, right up to today
//
// That gives a board with a rise, a gap and a return — and a bar row where
// the kinds shift across the year.
//
// The story ends in the current week, not one before. A board empty at the
// end reads as somebody who has stopped, whatever the card beside it says.
//
// Reformer only in the spring months: the schedule carries no reformer in
// July and August, and a user cannot sign up for a class that does not
// run. The story follows the schedule, not the other way round.
//
// The classes are real classes from the database. A signup for a class
// that does not exist is a row in a table, not a class in a life.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"kjernekraft/database"

	"golang.org/x/crypto/bcrypt"
)

const (
	namn    = "Solfrid Bakketun"
	epost   = "solfrid@test.local"
	loyord  = "password123"
	telefon = "90000042"

	// Ida exists already and is not created here; she only needs signups so
	// her section is not empty.
	idaEpost = "ida@kj.no"
)

// vike is one week in the story: how many classes, and what kinds.
//
// The numbers are uneven on purpose. A week of three and a week of two
// side by side is how a person trains; three every week for twenty weeks
// is how a table looks.
type vike struct {
	tal  int
	slag []string
}

// The story, week by week, from the first whole week in March.
//
// The kinds stand in the order they are picked in. If `tal` exceeds the
// list it wraps, so "pilates, yoga" with four classes is two of each.
var solfridSoga = []vike{
	// mars — ho byrjar, og ho er varsam
	{1, []string{"pilates"}},
	{1, []string{"pilates"}},
	{2, []string{"pilates"}},
	{1, []string{"pilates"}},
	// april — det sit
	{2, []string{"pilates"}},
	{2, []string{"pilates", "yoga"}},
	{2, []string{"pilates", "yoga"}},
	{3, []string{"pilates", "yoga"}},
	// mai — tri i vika, og fascia kjem inn
	{3, []string{"pilates", "yoga", "fascia"}},
	{2, []string{"pilates", "fascia"}},
	{3, []string{"pilates", "yoga", "fascia"}},
	{3, []string{"pilates", "fascia", "yoga"}},
	// juni — toppen, og fyrste reformeren
	{4, []string{"pilates", "yoga", "fascia", "reformer"}},
	{3, []string{"pilates", "reformer", "yoga"}},
	{4, []string{"pilates", "yoga", "pilates", "fascia"}},
	{4, []string{"pilates", "reformer", "yoga", "pilates"}},
	// juli — burte, og so varsamt attende
	{0, nil},
	{0, nil},
	{1, []string{"yoga"}},
	{2, []string{"pilates", "yoga"}},
	// august — attende i full gjenge, og ho held det gaaende fram til i dag
	{3, []string{"pilates", "yoga", "fascia"}},
	{3, []string{"pilates", "yoga", "pilates"}},
	{4, []string{"pilates", "yoga", "fascia", "pilates"}},
	{3, []string{"pilates", "yoga", "pilates"}},
	{3, []string{"pilates", "fascia", "yoga"}},
	{2, []string{"pilates", "yoga"}},
}

// Ida's story is a different shape, and that is the whole reason it stands
// here rather than us giving her the same one.
//
// She runs the house. She is here anyway, so she trains *steadily* and not
// in waves: twice a week, year in and year out, with a gap when things are
// busiest and not when she travels. No rise, no peak — and that is exactly
// what stops the two boards looking like the same board twice. Put two
// identical sequences side by side and you read them as a bad copy, not as
// two people.
//
// Her kinds lean toward yoga and fascia. She teaches pilates herself, and
// what you do all day is rarely what you sign up for in the evening.
var idaSoga = []vike{
	// mars
	{2, []string{"yoga", "fascia"}},
	{2, []string{"yoga", "pilates"}},
	{1, []string{"yoga"}},
	{2, []string{"fascia", "yoga"}},
	// april
	{2, []string{"yoga", "fascia"}},
	{2, []string{"yoga", "yoga"}},
	{1, []string{"fascia"}},
	{2, []string{"yoga", "pilates"}},
	// mai — ei glipe: det er her aaret er tyngst i eit studio
	{1, []string{"yoga"}},
	{0, nil},
	{2, []string{"yoga", "fascia"}},
	{2, []string{"yoga", "pilates"}},
	// juni
	{2, []string{"yoga", "fascia"}},
	{3, []string{"yoga", "reformer", "fascia"}},
	{2, []string{"yoga", "pilates"}},
	{2, []string{"fascia", "yoga"}},
	// juli
	{1, []string{"yoga"}},
	{2, []string{"yoga", "pilates"}},
	{2, []string{"yoga", "fascia"}},
	{1, []string{"yoga"}},
	// august
	{2, []string{"yoga", "fascia"}},
	{2, []string{"yoga", "pilates"}},
	{2, []string{"fascia", "yoga"}},
	{2, []string{"yoga", "pilates"}},
	{1, []string{"yoga"}},
	{2, []string{"yoga", "fascia"}},
}

func main() {
	slett := flag.Bool("slett", false, "tak Solfrid burt att")
	ida := flag.Bool("ida", false, "meld Ida paa timar ogso — ho finst fraa fyrr")
	settEpost := flag.String("epost", "", "kven som skal faa nytt loyord")
	settLoyord := flag.String("loyord", "", "det nye loyordet; krev -epost")
	flag.Parse()

	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// Nytt loyord på ein brukar som finst. Det er ikkje seiing av
	// prøvedata, men det høyrer til den same jobben: ein prøvekonto ein
	// ikkje kjem inn på er ikkje ein prøvekonto.
	if *settLoyord != "" {
		if *settEpost == "" {
			log.Fatal("-loyord krev -epost")
		}
		if err := settNyttLoyord(db, *settEpost, *settLoyord); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s hev fenge nytt loyord.\n", *settEpost)
		return
	}

	if *slett {
		if err := taBurt(db); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Solfrid er burte.")
		return
	}

	if *ida {
		id, err := finnBrukar(db, idaEpost)
		if err != nil {
			log.Fatal(err)
		}
		tal, err := meldPå(db, id, idaSoga, 43)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Ida (id %d)\n", id)
		fmt.Printf("  %d timar yver %d vikor\n", tal, len(idaSoga))
		fmt.Println("  jamn rekkje, ikkje ei bylgja — so dei tvo brettene ikkje les seg like")
		return
	}

	id, err := lagBrukar(db)
	if err != nil {
		log.Fatal(err)
	}

	tal, err := meldPå(db, id, solfridSoga, 17)
	if err != nil {
		log.Fatal(err)
	}
	if err := gjevMedlemskap(db, id); err != nil {
		log.Fatal(err)
	}
	if err := gjevKlippekort(db, id); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s (id %d)\n", namn, id)
	fmt.Printf("  %s / %s\n", epost, loyord)
	fmt.Printf("  %d timar yver %d vikor\n", tal, len(solfridSoga))
	fmt.Println("  medlemskap: aktivt   klippekort: tvo, det eine mest brukt upp")
	fmt.Println("  ikkje administrator — /admin svarar 403")
}

// finnBrukar fetches a user that should already exist.
func finnBrukar(conn *sql.DB, epost string) (int64, error) {
	var id int64
	err := conn.QueryRow(`SELECT id FROM users WHERE email = ?`, epost).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("fann ingen brukar med eposten %s", epost)
	}
	return id, err
}

// settNyttLoyord writes a new bcrypt hash on a user.
//
// It goes through bcrypt and not straight into the database with SQL: six
// of the test users in this database carry the string "x" in the password
// field because somebody once wrote the value in directly, and so they
// cannot log in at all. bcrypt.CompareHashAndPassword accepts nothing that
// is not a hash, and it does not say so out loud — it simply refuses.
func settNyttLoyord(conn *sql.DB, epost, nytt string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(nytt), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	res, err := conn.Exec(`UPDATE users SET password = ? WHERE email = ?`, string(hash), epost)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("fann ingen brukar med eposten %s", epost)
	}
	return nil
}

// lagBrukar creates her, or finds her again if she is already there.
func lagBrukar(conn *sql.DB) (int64, error) {
	var id int64
	err := conn.QueryRow(`SELECT id FROM users WHERE email = ?`, epost).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(loyord), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	res, err := conn.Exec(`
		INSERT INTO users (name, birthdate, email, phone, address, postal_code,
		                   city, country, password, newsletter_subscription,
		                   terms_accepted, created_at)
		VALUES (?, '1991-04-17', ?, ?, 'Thorvald Meyers gate 12', '0555',
		        'Oslo', 'Norge', ?, 1, 1, ?)`,
		namn, epost, telefon, string(hash), veggtekst(byrjinga().AddDate(0, 0, -9)))
	if err != nil {
		return 0, err
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// `user` og ingen ting meir. Ho skal *ikkje* sjå administrasjonen.
	if _, err := conn.Exec(`
		INSERT OR IGNORE INTO brukarloyve (user_id, loyve_id)
		SELECT ?, id FROM loyve WHERE name = 'user'`, id); err != nil {
		return 0, err
	}
	return id, nil
}

// byrjinga er maandagen i den fyrste vika i soga: den fyrste heile vika
// i mars i det aaret timane stend i basen.
func byrjinga() time.Time {
	t := time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC)
	return t
}

// meldPå gjeng soga vika for vika og melder henne på verkelege timar.
func meldPå(conn *sql.DB, brukar int64, soga []vike, fro int64) (int, error) {
	// Fixed seed. The script should give the same user every time it runs —
	// 	// otherwise it is not a test setup, it is a new thing to get to know
	// 	// each time. Each person has their own seed, so they do not end up in
	// 	// the same classes.
	tilf := rand.New(rand.NewSource(fro))

	nå := time.Now()
	sett := 0
	for i, v := range soga {
		if v.tal == 0 {
			continue
		}
		frå := byrjinga().AddDate(0, 0, 7*i)
		til := frå.AddDate(0, 0, 7)
		if frå.After(nå) {
			break
		}

		brukte := map[int64]bool{}
		for n := 0; n < v.tal; n++ {
			slag := v.slag[n%len(v.slag)]
			id, err := finnTime(conn, slag, frå, til, tilf, brukte)
			if err != nil {
				return sett, err
			}
			if id == 0 {
				continue // ingen slik time den vika; hopp yver han
			}
			brukte[id] = true

			// Ho melde seg på nokre dagar fyre timen gjekk.
			påmeld := frå.AddDate(0, 0, -tilf.Intn(4)-1)
			res, err := conn.Exec(`
				INSERT OR IGNORE INTO event_signups (user_id, event_id, signup_date)
				VALUES (?, ?, ?)`, brukar, id, veggtekst(påmeld))
			if err != nil {
				return sett, err
			}
			if n, _ := res.RowsAffected(); n > 0 {
				sett++
				// The class carries the count of how many are in it.
				if _, err := conn.Exec(
					`UPDATE events SET current_enrolment = current_enrolment + 1 WHERE id = ?`,
					id); err != nil {
					return sett, err
				}
			}
		}
	}
	return sett, nil
}

// finnTime plukkar ein time av eit slag i ei vika, og helst ein ho ikkje
// alt stend i.
func finnTime(conn *sql.DB, slag string, frå, til time.Time, tilf *rand.Rand, brukte map[int64]bool) (int64, error) {
	rows, err := conn.Query(`
		SELECT id FROM events
		WHERE class_type = ? AND start_time >= ? AND start_time < ?
		  AND start_time < ?
		ORDER BY start_time`,
		slag, veggtekst(frå), veggtekst(til), veggtekst(time.Now()))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var kandidatar []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		if !brukte[id] {
			kandidatar = append(kandidatar, id)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(kandidatar) == 0 {
		return 0, nil
	}
	return kandidatar[tilf.Intn(len(kandidatar))], nil
}

// gjevMedlemskap gjev henne eit medlemskap som gjeng.
func gjevMedlemskap(conn *sql.DB, brukar int64) error {
	// The cheapest thing that is not a "trial": the test data in the database
	// 	// carries random prices, and a membership at 1 446 kr a month is the
	// 	// first thing anyone will point at.
	//
	// `NOT skjult` lyt staa der, og det er heile poenget med denne
	// spyrjingi. Black kostar null — han er ein tilgang og ikkje ein
	// avtale — so «det billegaste» *er* Black, kvar einaste gong.
	// Solfrid fekk difor det usynlege medlemskapet lærarar og utviklarar
	// ber, og ho er nettupp den som ikkje skal ha det: ho finst for at
	// ein skal kunna sjaa kva ein vanleg elev ser.
	//
	// `skjult` er det rette vilkaaret og ikkje `price > 0`: flagget tyder
	// «ikkje noko ein kann velja», og det er den eigenskapen me vil ha
	// burt. Kjem det ein gaava-medlemskap til 0 kr som *kann* veljast,
	// skal han vera med her. Sjaa database/svartmedlem.go.
	var medlemskap int64
	err := conn.QueryRow(`
		SELECT id FROM memberships
		WHERE name NOT LIKE 'Prøve%' AND NOT skjult AND active
		ORDER BY price LIMIT 1`).Scan(&medlemskap)
	if err != nil {
		return err
	}
	byrja := byrjinga().AddDate(0, 0, -7)
	_, err = conn.Exec(`
		INSERT OR IGNORE INTO user_memberships
			(user_id, membership_id, status, start_date, renewal_date, binding_end)
		VALUES (?, ?, 'active', ?, ?, ?)`,
		brukar, medlemskap, veggtekst(byrja),
		veggtekst(fyrsteIMaanad(time.Now().AddDate(0, 1, 0))),
		veggtekst(byrja.AddDate(1, 0, 0)))
	return err
}

// gjevKlippekort gives her two: one she has started on, and one nearly
// empty. A card with every clip intact shows nothing.
func gjevKlippekort(conn *sql.DB, brukar int64) error {
	kort := []struct {
		pakke   int64
		ialt    int
		att     int
		gjengUt time.Time
	}{
		{5, 10, 6, time.Now().AddDate(0, 7, 0)},  // reformer, godt i gang
		{2, 10, 1, time.Now().AddDate(0, 1, 12)}, // personleg trening, snart tom
	}
	for _, k := range kort {
		if _, err := conn.Exec(`
			INSERT OR IGNORE INTO user_klippekort
				(user_id, package_id, total_klipp, remaining_klipp, expiry_date)
			VALUES (?, ?, ?, ?, ?)`,
			brukar, k.pakke, k.ialt, k.att, veggtekst(k.gjengUt)); err != nil {
			return err
		}
	}
	return nil
}

func taBurt(conn *sql.DB) error {
	var id int64
	err := conn.QueryRow(`SELECT id FROM users WHERE email = ?`, epost).Scan(&id)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	// The classes she was in should not keep their count when she goes.
	if _, err := conn.Exec(`
		UPDATE events SET current_enrolment = MAX(current_enrolment - 1, 0)
		WHERE id IN (SELECT event_id FROM event_signups WHERE user_id = ?)`, id); err != nil {
		return err
	}
	for _, q := range []string{
		`DELETE FROM event_signups WHERE user_id = ?`,
		`DELETE FROM user_klippekort WHERE user_id = ?`,
		`DELETE FROM user_memberships WHERE user_id = ?`,
		`DELETE FROM brukarloyve WHERE user_id = ?`,
		`DELETE FROM users WHERE id = ?`,
	} {
		if _, err := conn.Exec(q, id); err != nil {
			return err
		}
	}
	return nil
}

func fyrsteIMaanad(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// veggtekst skriv ei tid slik basen ber henne: veggklokka i Oslo, utan
// sone. Sjå handlers/tid.go — drivaren les desse som UTC, so dei lyt
// skrivast ut med dei tali som faktisk skal stå der.
func veggtekst(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
