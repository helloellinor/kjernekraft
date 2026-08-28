package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"kjernekraft/models"

	"github.com/gorilla/sessions"
)

const (
	sessionName = "kjernekraft-session"

	// Nykelen ligg i umgjevnaden, aldri i koda. Han lyt vera minst 32
	// oktettar: kortare nykel gjev kortare HMAC og eit mindre arbeid for
	// den som vil skriva sin eigen kake.
	sessionKeyEnv = "KJERNEKRAFT_SESSION_KEY"
	minKeyLen     = 32

	// Set KJERNEKRAFT_ENV=development for aa sleppa Secure-flagget medan
	// ein arbeider paa http://localhost. Alt anna vert handsama som drift.
	envEnv         = "KJERNEKRAFT_ENV"
	envDevelopment = "development"
)

var sessionStore *sessions.CookieStore

// userCtxKey ber brukaren gjenom eit einskilt kall. Brukaren vert lesen
// or basen ein gong per soknad og ikkje ein gong per handsamar.
type ctxKey int

const (
	userCtxKey ctxKey = iota
	csrfCtxKey
)

// IsDevelopment segjer um me køyrer utanfor drift.
func IsDevelopment() bool {
	return os.Getenv(envEnv) == envDevelopment
}

// InitializeSessionStore set upp kakebui. Han feilar heller enn aa falla
// attende paa ein fast nykel: ein fast nykel i kjelda tyder at kven som
// helst kann smi ei gyldig økt for kven som helst.
func InitializeSessionStore() error {
	raw := os.Getenv(sessionKeyEnv)
	if raw == "" {
		return fmt.Errorf("%s er ikkje sett — lag ein med `openssl rand -base64 32`", sessionKeyEnv)
	}

	key := decodeKey(raw)
	if len(key) < minKeyLen {
		return fmt.Errorf("%s er for stutt: %d oktettar, minst %d trengst", sessionKeyEnv, len(key), minKeyLen)
	}

	sessionStore = sessions.NewCookieStore(key)
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   !IsDevelopment(),
		SameSite: http.SameSiteLaxMode,
	}
	return nil
}

// decodeKey tek imot base64, hex eller raa tekst, so det er lett aa lima
// inn det `openssl rand` skriv ut, same kva flagg ein gav han.
func decodeKey(raw string) []byte {
	raw = strings.TrimSpace(raw)
	if b, err := base64.StdEncoding.DecodeString(raw); err == nil && len(b) >= minKeyLen {
		return b
	}
	if b, err := hex.DecodeString(raw); err == nil && len(b) >= minKeyLen {
		return b
	}
	return []byte(raw)
}

// sessionUserID hentar berre kjennemerket or kaka. Alt anna um brukaren —
// namn, løyve — kjem or basen. Eit løyve som ligg i kaka er eit løyve
// lesaren kann skriva um att.
func sessionUserID(r *http.Request) (int64, bool) {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		return 0, false
	}
	id, ok := session.Values["user_id"].(int64)
	if !ok || id <= 0 {
		return 0, false
	}
	return id, true
}

// GetUserFromSession gjev brukaren for denne soknaden, eller nil.
func GetUserFromSession(r *http.Request) *models.User {
	if u, ok := r.Context().Value(userCtxKey).(*models.User); ok {
		return u
	}
	id, ok := sessionUserID(r)
	if !ok {
		return nil
	}
	user, err := DB.GetUserByID(id)
	if err != nil {
		return nil
	}
	return user
}

// SetUserInSession skriv ei ny økt for brukaren. Den gamle økti vert
// kasta fyrst, so eit kjennemerke ein motstandar hev planta fyre
// innloggingi ikkje fylgjer med inn.
func SetUserInSession(w http.ResponseWriter, r *http.Request, user *models.User) error {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		// Ei kaka me ikkje kann lesa er ei kaka me byggjer paa nytt.
		session, _ = sessionStore.New(r, sessionName)
	}

	for k := range session.Values {
		delete(session.Values, k)
	}
	session.Values["user_id"] = int64(user.ID)

	// Nytt CSRF-kjennemerke ved innlogging: eit kjennemerke ein
	// motstandar hev sett fyre innloggingi skal ikkje fylgja med inn.
	//
	// Men daa lyt kaka fylgja med same vegen. Gjorde ho ikkje det, stod
	// økti med eitt kjennemerke og kaka med eit anna, og den fyrste
	// endringi etter innlogging fekk 403. I lesaren gjekk det godt av di
	// umleidingi er ein GET som stiller deim likt att — men det er hell,
	// ikkje ein maate aa gjera det paa.
	nytt := newToken()
	session.Values[csrfSessionKey] = nytt
	if err := session.Save(r, w); err != nil {
		return err
	}
	settCSRFKaka(w, nytt)
	return nil
}

// ClearUserSession sletter økti.
func ClearUserSession(w http.ResponseWriter, r *http.Request) error {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		return err
	}
	for k := range session.Values {
		delete(session.Values, k)
	}
	session.Options.MaxAge = -1
	return session.Save(r, w)
}

// IsLoggedIn segjer um soknaden ber ei gyldig økt.
func IsLoggedIn(r *http.Request) bool {
	return GetUserFromSession(r) != nil
}

// IsAdmin les løyvet or brukaren basen gav oss.
func IsAdmin(user *models.User) bool {
	if user == nil {
		return false
	}
	for _, role := range user.Loyve {
		if role == "admin" {
			return true
		}
	}
	return false
}

// newToken lagar eit tilfeldig kjennemerke. crypto/rand feilar ikkje i
// praksis; gjer han det, er det ikkje noko me kann halda fram med.
func newToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("kjernekraft: kunde ikkje lesa tilfeldige oktettar: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// tokensEqual samanliknar i konstant tid.
func tokensEqual(a, b string) bool {
	return len(a) > 0 && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// withUser legg brukaren i kall-samanhengen.
func withUser(r *http.Request, user *models.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userCtxKey, user))
}
