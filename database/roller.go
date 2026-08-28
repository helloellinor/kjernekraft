package database

import "strings"

// Rollone er dei einaste orsakene til at ein brukar er noko meir enn ein
// elev. Namni er engelske av di «admin» og «user» alt laag i basen daa
// denne fila kom; ei norsk rolla i den same spalta hadde vore verre enn
// tvo engelske. Det synlege ordet kjem gjenom {{t}}, ikkje herifraa.
const (
	RollaAdmin  = "admin"
	RollaLaerar = "teacher"
)

// RollaFinst seier um namnet er ei rolla systemet kjenner. Utan denne
// kunde eit kall skriva kva som helst inn i roles-tabellen, og daa er
// rollelista ikkje lenger ei lista yver noko.
// Utviklarrolla stend *ikkje* her, og det er med vilje.
//
// Ho vert ikkje gjevi ut av nokon: ho vert lesi or ei fil paa tenaren
// (sjaa utviklar.go). RollaFinst er porten administrasjonsflata skriv
// gjenom, so ei rolla som stend her, er ei rolla ein administrator kann
// gjeva seg sjølv. Utviklarrolla gjev fri tilgang til huset; ho skal
// koma fraa den som eig maskini, ikkje fraa den som eig ein knapp.
func RollaFinst(rolla string) bool {
	return rolla == RollaAdmin || rolla == RollaLaerar
}

// SettRolla slær ei rolla av eller paa for ein brukar.
//
// AssignRoleToUser gjer berre det fyrste, og fell dessutan paa ein
// nykelkrasj naar rolla alt er der. Ein knapp ein kann trykkja tvo
// gonger treng baae vegar, og treng at den fyrste vegen toler aa
// gaast tvo gonger.
func (db *Database) SettRolla(userID int64, rolla string, paa bool) error {
	if !paa {
		_, err := db.Conn.Exec(`
			DELETE FROM user_roles
			WHERE user_id = ? AND role_id = (SELECT id FROM roles WHERE name = ?)`,
			userID, rolla)
		return err
	}

	roleID, err := db.GetOrCreateRole(rolla)
	if err != nil {
		return err
	}
	if _, err := db.Conn.Exec(
		"INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)",
		userID, roleID); err != nil {
		return err
	}

	// Forfremjingi skal gjelda med ein gong, og ho skal gjelda pengane
	// med. Sjaa SynkFriMedlemskap i svartmedlem.go.
	_, err = db.SynkFriMedlemskap(userID)
	return err
}

// HarRolla svarar for ein einskild brukar.
func (db *Database) HarRolla(userID int64, rolla string) (bool, error) {
	var tal int
	err := db.Conn.QueryRow(`
		SELECT COUNT(*) FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.name = ?`, userID, rolla).Scan(&tal)
	return tal > 0, err
}

// LaerarNamn gjev namni paa dei som hev lærarrolla.
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
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id
		WHERE r.name = ?
		ORDER BY u.name COLLATE NOCASE`, RollaLaerar)
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

// harRolla les den samanslegne rollestrengen FolkOversyn hentar, so
// lista slepp eit uppslag per person.
func harRolla(roller, rolla string) bool {
	for _, r := range strings.Split(roller, ",") {
		if strings.TrimSpace(r) == rolla {
			return true
		}
	}
	return false
}
