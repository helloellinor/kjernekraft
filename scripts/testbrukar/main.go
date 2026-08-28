// testbrukar lagar éin prøvebrukar som det er noko aa sjaa paa.
//
//	go run ./scripts/testbrukar
//	go run ./scripts/testbrukar -slett     # tek henne burt att
//
// Han er ikkje ein administrator. Det er heile poenget: alt ein syner
// fram paa ein skjerm skal syna det ein elev ser, og ein administrator
// ser noko anna. Ho fær `user` og ingen ting meir, so `/admin` svarar
// 403 for henne.
//
// Skriptet er *ikkje* tilfeldig. Ein tilfeldig brukar gjev eit
// varmekart med jamn stoy i, og jamn stoy ser ut som ein feil — ein
// les ingen ting av det. Ei rekkje ein kjenner att les ein med ein
// gong, og difor hev Solfrid ei soga:
//
//	mars    ho byrjar, og ho er varsam — ein time i vika
//	april   det sit; tvo i vika, og ein yoga attaat
//	mai     tri i vika, og ho finn fascia
//	juni    toppen. Fire i vika, og fyrste reformeren
//	juli    burte i tvo vikor, og so varsamt attende
//	august  tri i vika att, heilt fram til i dag
//
// Det gjev eit brett med ei stigning, eit hol og ei attkoma — og ei
// bjelkerad der slagi skifter yver aaret. Baae bileti hev daa noko aa
// segja, som er det dei er der for.
//
// Soga endar i den vika me stend i, og ikkje ei vika fyre. Eit brett
// som er tomt i enden les seg som ein som hev slutta, kva so det stend
// i kortet ved sida av.
//
// Reformer stend berre i vaarmaanadene: timeplanen ber ingen reformer
// i juli og august, og ein brukar kann ikkje melda seg paa ein time som
// ikkje gjeng. Soga fylgjer timeplanen og ikkje umvendt.
//
// Timane er verkelege timar or basen. Skriptet finn dei som gjekk i den
// vika det er tale um, av det slaget det er tale um, og melder henne
// paa dei. Ei paamelding paa ein time som ikkje finst er ei line i ein
// tabell og ikkje ein time i eit liv.
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

	// Ida finst fraa fyrr og vert ikkje laga her; ho skal berre
	// meldast paa timar so bolken hennar ikkje stend tom.
	idaEpost = "ida@kj.no"
)

// vike er ei vika i soga: kor mange timar, og kva slag dei var.
//
// Tali er ikkje jamne med vilje. Ei vika med tri timar og ei med tvo
// attmed kvarandre er slik ein person trenar; tri kvar vika i tjue
// vikor er slik ein tabell ser ut.
type vike struct {
	tal  int
	slag []string
}

// soga, vika for vika, fraa den fyrste heile vika i mars.
//
// Slagi stend i den rekkjefylgda dei skal veljast i. Er `tal` større
// enn lista, gjeng ho rundt paa nytt — so «pilates, yoga» og fire timar
// er tvo pilates og tvo yoga.
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

// Ida si soga er ei onnor form, og det er heile grunnen til at ho stend
// her i staden for at me gjev henne den same.
//
// Ho driv huset. Ho er her likevel, og difor trenar ho *jamt* og ikkje i
// bylgjor: tvo i vika, aar ut og aar inn, med ei glipe naar det stend
// paa som verst og ikkje naar ho reiser burt. Ingen stigning, ingen
// topp — og det er nettupp det som gjer at dei tvo brettene ikkje ser
// ut som det same brettet tvo gonger. Set ein tvo like rekkjor attmed
// kvarandre, les ein deim som ei feilkopiering og ikkje som tvo folk.
//
// Slagi hennar heller mot yoga og fascia. Ho lærer pilates sjølv, og det
// ein gjer heile dagen er sjeldan det ein melder seg paa om kvelden.
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

	// Nytt loyord paa ein brukar som finst. Det er ikkje seiing av
	// prøvedata, men det høyrer til den same jobben: ein prøvekonto ein
	// ikkje kjem inn paa er ikkje ein prøvekonto.
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
		tal, err := meldPaa(db, id, idaSoga, 43)
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

	tal, err := meldPaa(db, id, solfridSoga, 17)
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

// finnBrukar hentar ein brukar som skal finnast fraa fyrr.
func finnBrukar(conn *sql.DB, epost string) (int64, error) {
	var id int64
	err := conn.QueryRow(`SELECT id FROM users WHERE email = ?`, epost).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("fann ingen brukar med eposten %s", epost)
	}
	return id, err
}

// settNyttLoyord skriv eit nytt bcrypt-hash paa ein brukar.
//
// Det gjeng gjenom bcrypt og ikkje rett i basen med SQL: seks av
// prøvebrukarane i denne basen ber strengen «x» i loyordfeltet av di
// nokon ein gong skreiv verdet beint inn, og dei kann difor ikkje logga
// inn i det heile. `bcrypt.CompareHashAndPassword` godtek ingen ting som
// ikkje er eit hash, og han segjer det ikkje høgt — han berre nektar.
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

// lagBrukar lagar henne, eller finn henne att um ho alt stend der.
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

	// `user` og ingen ting meir. Ho skal *ikkje* sjaa administrasjonen.
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

// meldPaa gjeng soga vika for vika og melder henne paa verkelege timar.
func meldPaa(conn *sql.DB, brukar int64, soga []vike, fro int64) (int, error) {
	// Fast frø. Skriptet skal gjeva den same brukaren kvar gong det
	// gjeng — elles er det ikkje eit prøveoppsett, det er ei ny sak aa
	// setja seg inn i kvar gong. Kvar person hev sitt eige frø, so dei
	// ikkje endar paa dei same timane.
	tilf := rand.New(rand.NewSource(fro))

	naa := time.Now()
	sett := 0
	for i, v := range soga {
		if v.tal == 0 {
			continue
		}
		fraa := byrjinga().AddDate(0, 0, 7*i)
		til := fraa.AddDate(0, 0, 7)
		if fraa.After(naa) {
			break
		}

		brukte := map[int64]bool{}
		for n := 0; n < v.tal; n++ {
			slag := v.slag[n%len(v.slag)]
			id, err := finnTime(conn, slag, fraa, til, tilf, brukte)
			if err != nil {
				return sett, err
			}
			if id == 0 {
				continue // ingen slik time den vika; hopp yver han
			}
			brukte[id] = true

			// Ho melde seg paa nokre dagar fyre timen gjekk.
			paameld := fraa.AddDate(0, 0, -tilf.Intn(4)-1)
			res, err := conn.Exec(`
				INSERT OR IGNORE INTO event_signups (user_id, event_id, signup_date)
				VALUES (?, ?, ?)`, brukar, id, veggtekst(paameld))
			if err != nil {
				return sett, err
			}
			if n, _ := res.RowsAffected(); n > 0 {
				sett++
				// Timen ber talet paa kor mange som stend i honom.
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
func finnTime(conn *sql.DB, slag string, fraa, til time.Time, tilf *rand.Rand, brukte map[int64]bool) (int64, error) {
	rows, err := conn.Query(`
		SELECT id FROM events
		WHERE class_type = ? AND start_time >= ? AND start_time < ?
		  AND start_time < ?
		ORDER BY start_time`,
		slag, veggtekst(fraa), veggtekst(til), veggtekst(time.Now()))
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
	// Det billegaste som ikkje er ein «Prøve»: prøvedataen i basen ber
	// tilfeldige prisar, og eit medlemskap til 1 446 kr i maanaden er
	// det fyrste nokon kjem til aa peika paa.
	var medlemskap int64
	err := conn.QueryRow(`
		SELECT id FROM memberships
		WHERE name NOT LIKE 'Prøve%' ORDER BY price LIMIT 1`).Scan(&medlemskap)
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

// gjevKlippekort gjev henne tvo: eitt ho hev teke hol paa, og eitt som
// snart er tomt. Eit kort med alle klippi i behald syner ingen ting.
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
	// Timane ho stod i skal ikkje halda paa talet sitt naar ho gjeng.
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
// sone. Sjaa handlers/tid.go — drivaren les desse som UTC, so dei lyt
// skrivast ut med dei tali som faktisk skal staa der.
func veggtekst(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
