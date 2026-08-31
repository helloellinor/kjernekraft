package handsamarar

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

	// Set KJERNEKRAFT_ENV=development to drop the Secure flag while working
	// on http://localhost. Anything else is treated as production.
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

// IsDevelopment says whether we are running outside production.
func IsDevelopment() bool {
	return os.Getenv(envEnv) == envDevelopment
}

// InitializeSessionStore sets up the cookie store. It fails rather than
// falling back to a fixed key: a fixed key in the source means anyone can
// forge a valid session for anyone.
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

// decodeKey accepts base64, hex or raw text, so it is easy to paste what
// openssl rand prints whatever flag you gave it.
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

// sessionUserID takes only the id from the cookie. Everything else about
// the user — name, permissions — comes from the database. A permission
// sitting in the cookie is a permission the browser can rewrite.
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
// Brukaren kjem or samanhengen og ingen annan stad. WithUser er lagd på
// rutaren fyre kvar einaste rute, so han hev alt slege upp brukaren når
// dette vert kalla — og gjorde han det ikkje, er svaret «ingen innlogga»,
// ikkje eit nytt basesøk.
//
// Fallet attende til basen som stod her las brukaren um att på eigi hand.
// Det gav funksjonen ei basetilkopling, og av di han vert kalla frå
// brukaren(), sessionIsAdmin(), sessionUserName() og IsLoggedIn(), drog
// den tilkoplingi seg gjenom heile økt- og mellomvare-laget.
func GetUserFromSession(r *http.Request) *models.User {
	u, _ := r.Context().Value(userCtxKey).(*models.User)
	return u
}

// brukaren gives the logged-in user, assuming RequireAuth or RequireAdmin
// has already said yes. It is for handlers behind that middleware and only
// for those. There used to be an `if user == nil` check in every handler,
// twenty-four of them, and none could fire: the middleware had already
// turned away anyone without a session. Should a route ever be registered
// outside the groups,
func brukaren(r *http.Request) *models.User {
	user := GetUserFromSession(r)
	if user == nil {
		panic("brukaren() kalla utanfor RequireAuth/RequireAdmin")
	}
	return user
}

// SetUserInSession writes a new session for the user. The old session is
// discarded first, so a token an attacker planted before the login does not
// travel in with them.
func SetUserInSession(w http.ResponseWriter, r *http.Request, user *models.User) error {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		// A cookie we cannot read is a cookie we rebuild.
		session, _ = sessionStore.New(r, sessionName)
	}

	for k := range session.Values {
		delete(session.Values, k)
	}
	session.Values["user_id"] = int64(user.ID)

	// A new CSRF token on login: a token an attacker set before the login must
	// 	// not travel in.
	// 	//
	// 	// But then the cookie has to follow the same way. If it did not, the
	// 	// session held one token and the cookie another, and the first change
	// 	// after login got 403. In the browser it worked because the redirect is
	// 	// a GET that levels them again — but that is luck, not a way to do
	// 	// it.
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

// IsLoggedIn says whether the request carries a valid session.
func IsLoggedIn(r *http.Request) bool {
	return GetUserFromSession(r) != nil
}

// IsAdmin les løyvet or brukaren basen gav oss.
func IsAdmin(user *models.User) bool {
	if user == nil {
		return false
	}
	for _, role := range user.Løyve {
		if role == "admin" {
			return true
		}
	}
	return false
}

// newToken makes a random token. crypto/rand does not fail in practice; if
// it does, there is nothing we can carry on with.
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
