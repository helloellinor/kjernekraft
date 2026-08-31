package handsamarar

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"kjernekraft/database"

	_ "github.com/mattn/go-sqlite3"
)

// frysingsbase set upp ein medlem som hev bede um frysing.
func frysingsbase(t *testing.T) (*App, int64) {
	t.Helper()
	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "frysing.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := database.Migrate(conn); err != nil {
		t.Fatal(err)
	}
	db := &database.Database{Conn: conn}

	res, err := conn.Exec(`INSERT INTO users (name, birthdate, email, phone, password)
	                       VALUES ('Ida', '1990-01-01', 'ida@do.me', '900', 'x')`)
	if err != nil {
		t.Fatal(err)
	}
	brukarID, _ := res.LastInsertId()

	res, err = conn.Exec(`INSERT INTO memberships (name, price, commitment_months, description, features, active)
	                      VALUES ('Aarskort', 69000, 12, '', '', TRUE)`)
	if err != nil {
		t.Fatal(err)
	}
	medlemskapID, _ := res.LastInsertId()

	_, err = conn.Exec(`INSERT INTO user_memberships
	    (user_id, membership_id, status, start_date, renewal_date, binding_end)
	    VALUES (?, ?, 'freeze_requested', ?, ?, ?)`,
		brukarID, medlemskapID,
		time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC),
		time.Date(2027, 3, 11, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	return &App{DB: db, Nå: func() time.Time { return time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC) }}, brukarID
}

func stoda(t *testing.T, a *App, brukarID int64) string {
	t.Helper()
	m, err := a.DB.GetUserMembership(brukarID)
	if err != nil || m == nil {
		t.Fatalf("las ikkje medlemskapet att: %v", err)
	}
	return m.Status
}

// Medlemen bad um frysing og vart ståande i «freeze_requested». Basen
// hadde svaret heile tidi, men ruta svara 501, so ingen kom seg vidare —
// korkje fram eller attende. Administrasjonen talde dei ventande i
// briefingen og baud fram ein knapp som ikkje gjorde noko.
func TestGodkjendFrysingSetMedlemskapetIPause(t *testing.T) {
	a, brukarID := frysingsbase(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/admin/freeze-requests/approve?user_id=1", nil)
	a.ApproveFreezeRequest(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("svara %d, venta 200: %s", w.Code, w.Body.String())
	}
	if got := stoda(t, a, brukarID); got != "paused" {
		t.Errorf("stoda vart %q, venta paused", got)
	}
}

// Avvist frysing skal setja medlemskapet i gang att, ikkje lata det stå.
func TestAvvistFrysingSetMedlemskapetAktivtAtt(t *testing.T) {
	a, brukarID := frysingsbase(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/admin/freeze-requests/reject?user_id=1", nil)
	a.RejectFreezeRequest(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("svara %d, venta 200: %s", w.Code, w.Body.String())
	}
	if got := stoda(t, a, brukarID); got != "active" {
		t.Errorf("stoda vart %q, venta active", got)
	}
}

// Ein brukar-id som ikkje er eit tal skal ikkje nå basen i det heile.
func TestFrysingUtanBrukarSvararVondSpurnad(t *testing.T) {
	a, _ := frysingsbase(t)

	for _, adresse := range []string{
		"/api/admin/freeze-requests/approve",
		"/api/admin/freeze-requests/approve?user_id=",
		"/api/admin/freeze-requests/approve?user_id=hallo",
		"/api/admin/freeze-requests/approve?user_id=0",
	} {
		w := httptest.NewRecorder()
		a.ApproveFreezeRequest(w, httptest.NewRequest(http.MethodPost, adresse, nil))
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s svara %d, venta 400", adresse, w.Code)
		}
	}
}
