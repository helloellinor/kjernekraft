package database

import "strings"

// Løyvi er dei einaste orsakene til at ein brukar er noko meir enn ein
// elev. Namni er engelske av di «admin» og «user» alt laag i basen daa
// denne fila kom; ei norsk løyvet i den same spalta hadde vore verre enn
// tvo engelske. Det synlege ordet kjem gjenom {{t}}, ikkje herifraa.
const (
	LoyveAdmin  = "admin"
	LoyveLaerar = "teacher"
)

// LoyveFinst seier um namnet er eit løyve systemet kjenner. Utan denne
// kunde eit kall skriva kva som helst inn i roles-tabellen, og daa er
// løyvelista ikkje lenger ei lista yver noko.
// Utviklarløyvet stend *ikkje* her, og det er med vilje.
//
// Ho vert ikkje gjevi ut av nokon: ho vert lesi or ei fil paa tenaren
// (sjaa utviklar.go). LoyveFinst er porten administrasjonsflata skriv
// gjenom, so eit løyve som stend her, er eit løyve ein administrator kann
// gjeva seg sjølv. Utviklarløyvet gjev fri tilgang til huset; ho skal
// koma fraa den som eig maskini, ikkje fraa den som eig ein knapp.
func LoyveFinst(løyvet string) bool {
	return løyvet == LoyveAdmin || løyvet == LoyveLaerar
}

// SettLoyve slær eit løyve av eller paa for ein brukar.
//
// GjevLoyve gjer berre det fyrste, og fell dessutan paa ein
// nykelkrasj naar løyvet alt er der. Ein knapp ein kann trykkja tvo
// gonger treng baae vegar, og treng at den fyrste vegen toler aa
// gaast tvo gonger.
func (db *Database) SettLoyve(userID int64, løyvet string, paa bool) error {
	if !paa {
		_, err := db.Conn.Exec(`
			DELETE FROM brukarloyve
			WHERE user_id = ? AND loyve_id = (SELECT id FROM loyve WHERE name = ?)`,
			userID, løyvet)
		return err
	}

	loyveID, err := db.LoyveIDFor(løyvet)
	if err != nil {
		return err
	}
	if _, err := db.Conn.Exec(
		"INSERT OR IGNORE INTO brukarloyve (user_id, loyve_id) VALUES (?, ?)",
		userID, loyveID); err != nil {
		return err
	}

	// Forfremjingi skal gjelda med ein gong, og ho skal gjelda pengane
	// med. Sjaa SynkFriMedlemskap i svartmedlem.go.
	_, err = db.SynkFriMedlemskap(userID)
	return err
}

// HarLoyve svarar for ein einskild brukar.
func (db *Database) HarLoyve(userID int64, løyvet string) (bool, error) {
	var tal int
	err := db.Conn.QueryRow(`
		SELECT COUNT(*) FROM brukarloyve ur
		JOIN loyve r ON r.id = ur.loyve_id
		WHERE ur.user_id = ? AND r.name = ?`, userID, løyvet).Scan(&tal)
	return tal > 0, err
}

// LaerarNamn gjev namni paa dei som hev lærarløyvet.
//
// Dette er det som skil steg 1 fraa det som var: veljarane spurde fyrr
// GetDistinctTeachers, som les `SELECT DISTINCT teacher_name FROM
// events`. Den lista er *historia* — kven som faktisk heldt ein time —
// og ho veks med kvar skrivefeil nokon nokon gong hev gjort i feltet.
// Ho er framleis rett i timeplanfilteret, der ein filtrerar yver timar
// som hev vore. Ho er ikkje rett i ein veljar, der ein peikar paa
// nokon som skal halda ein time som ikkje hev vore.
func (db *Database) LaerarNamn() ([]string, error) {
	rows, err := db.Conn.Query(`
		SELECT u.name FROM users u
		JOIN brukarloyve ur ON ur.user_id = u.id
		JOIN loyve r ON r.id = ur.loyve_id
		WHERE r.name = ?
		ORDER BY u.name COLLATE NOCASE`, LoyveLaerar)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ut []string
	for rows.Next() {
		var namn string
		if err := rows.Scan(&namn); err != nil {
			return nil, err
		}
		ut = append(ut, namn)
	}
	return ut, rows.Err()
}

// harLoyve les den samanslegne løyvestrengen FolkOversyn hentar, so
// lista slepp eit uppslag per person.
func harLoyve(løyve, løyvet string) bool {
	for _, r := range strings.Split(løyve, ",") {
		if strings.TrimSpace(r) == løyvet {
			return true
		}
	}
	return false
}
