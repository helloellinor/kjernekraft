package handsamarar

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	csrfSessionKey = "csrf_token"
	csrfCookieName = "kjernekraft-csrf"
	csrfHeaderName = "X-CSRF-Token"
	csrfFormField  = "csrf_token"
)

// WithUser reads the user from the session once and puts them in the
// context, so the fifteen handlers that ask do not make fifteen database
// queries.
func (a *App) WithUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Static files never need the user. The browser still sends the
		// session cookie with every CSS, JS and image request, so without
		// this guard each file cost a query for a logged-in user.
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		// In development everything on disk is reread — templates,
		// stylesheet and translations. Otherwise you see the key itself
		// and think you mistyped it.
		if IsDevelopment() {
			LastUmsetjingarPåNytt()
		}
		if id, ok := sessionUserID(r); ok {
			if user, err := a.DB.GetUserByID(id); err == nil {
				r = withUser(r, user)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth closes the route to anyone not logged in. An API route
// gets 401 so htmx can tell the difference; a page gets a redirect to
// the login.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetUserFromSession(r) == nil {
			denyUnauthenticated(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin requires the admin permission, read from the database —
// not from the cookie the browser sends.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromSession(r)
		if user == nil {
			denyUnauthenticated(w, r)
			return
		}
		if !IsAdmin(user) {
			svarFeil(w, r, http.StatusForbidden, "feil.tilgang")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireDevelopment keeps the test-data routes out of production. They
// overwrite the database, and a stranger must not be able to reach them.
func RequireDevelopment(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsDevelopment() {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func denyUnauthenticated(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") || r.Header.Get("HX-Request") == "true" {
		svarFeil(w, r, http.StatusUnauthorized, "feil.innlogging")
		return
	}
	http.Redirect(w, r, "/innlogging", http.StatusSeeOther)
}

// CSRF requires a token on anything that changes something. The token
// lives in the session — that is the truth — and is mirrored into a
// cookie the page can read, so htmx and fetch can send it back in a
// header.
//
// The session decides. Plain double-submit, comparing cookie against
// header, accepts a token an attacker set themselves.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ensureCSRFToken(w, r)

		// The token travels with the request rather than being looked up
		// again in the template. A second lookup means reading the session
		// twice, and if the cookie is unreadable — after a key change, say
		// — the second lookup comes back empty while the cookie holds a
		// fresh token. The form then got value="" and the user could never
		// get through the login again.
		r = withCSRFToken(r, token)

		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			next.ServeHTTP(w, r)
			return
		}

		sent := r.Header.Get(csrfHeaderName)
		if sent == "" {
			sent = r.FormValue(csrfFormField)
		}
		if !tokensEqual(token, sent) {
			svarFeil(w, r, http.StatusForbidden, "feil.kjennemerke")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ensureCSRFToken takes the token from the session, makes a new one if
// it is missing, and keeps the cookie in step with it.
func ensureCSRFToken(w http.ResponseWriter, r *http.Request) string {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		session, _ = sessionStore.New(r, sessionName)
	}

	token, _ := session.Values[csrfSessionKey].(string)
	if token == "" {
		token = newToken()
		session.Values[csrfSessionKey] = token
		if err := session.Save(r, w); err != nil {
			return token
		}
	}

	if cookie, err := r.Cookie(csrfCookieName); err != nil || cookie.Value != token {
		settCSRFKaka(w, token)
	}
	return token
}

// settCSRFKaka mirrors the token into a cookie the page can read.
//
// Not HttpOnly: the script has to read it and put it in the header. The
// token is not a secret from the page itself — it is proof the request
// came from there.
func settCSRFKaka(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   !IsDevelopment(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})
}

// CSRFToken gives templates the token so forms can carry it in a hidden
// field. It comes from the request context, where the CSRF middleware
// put it.
func CSRFToken(r *http.Request) string {
	if token, ok := r.Context().Value(csrfCtxKey).(string); ok {
		return token
	}
	// Without the middleware — in a test, say — read the session directly.
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		return ""
	}
	token, _ := session.Values[csrfSessionKey].(string)
	return token
}

// withCSRFToken puts the token in the request context.
func withCSRFToken(r *http.Request, token string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), csrfCtxKey, token))
}

// ---- login throttle ----
//
// Without one, the password field is an oracle: you can guess as fast as
// the machine allows. It counts failed attempts per IP and per address,
// and locks out for fifteen minutes after ten.

const (
	maxLoginAttempts = 10
	loginWindow      = 15 * time.Minute

	// Off by default; switch it on with KJERNEKRAFT_INNLOGGINGSBREMS=på.
	//
	// Decided 2026-08-28: leave it off while the house is new. A throttle
	// that locks out the wrong people is worse than one that locks out
	// nobody — the first looks like the site is down, and the person hit
	// is a customer who did everything right.
	//
	// Turn it on once there are real customers. Without it a password can
	// be guessed as fast as the machine allows.
	bremsEnv = "KJERNEKRAFT_INNLOGGINGSBREMS"
)

// bremsaPå says whether the login throttle is on.
func bremsaPå() bool {
	return os.Getenv(bremsEnv) == "paa"
}

type attemptCounter struct {
	mu       sync.Mutex
	attempts map[string]*attemptRecord
}

type attemptRecord struct {
	count int
	until time.Time
}

var loginAttempts = &attemptCounter{attempts: make(map[string]*attemptRecord)}

// Blocked says whether the key is locked out, clearing expired entries
// as it goes.
func (c *attemptCounter) Blocked(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sweep()
	rec, ok := c.attempts[key]
	return ok && rec.count >= maxLoginAttempts && time.Now().Before(rec.until)
}

func (c *attemptCounter) Fail(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sweep()
	rec, ok := c.attempts[key]
	if !ok || time.Now().After(rec.until) {
		rec = &attemptRecord{}
		c.attempts[key] = rec
	}
	rec.count++
	rec.until = time.Now().Add(loginWindow)
}

func (c *attemptCounter) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.attempts, key)
}

// sweep must be called with the lock held.
func (c *attemptCounter) sweep() {
	now := time.Now()
	for k, rec := range c.attempts {
		if now.After(rec.until) {
			delete(c.attempts, k)
		}
	}
}

// clientKey is where the request came from.
//
// Trust only who has the right to say. Behind our own proxy, RemoteAddr
// is the proxy — reading it alone gave every user the same key, so ten
// wrong passwords from anyone locked out the whole house for fifteen
// minutes. But X-Forwarded-For is a header anyone can invent, so
// trusting it from anyone throttles nobody.
//
// Both are true, so: if the request comes from
// loopback it came through our own proxy and the header is ours; from
// outside, RemoteAddr is the real sender and that is what counts. An
// attacker reaching the server directly does not get their own header
// believed, and one coming through the tunnel cannot write it —
// Cloudflare sets CF-Connecting-IP over whatever the client sent.
func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if !eigenMellomtenar(host) {
		return host
	}
	if v := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); v != "" {
		return v
	}
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		// The first is the client; the rest are proxies it passed through.
		if i := strings.IndexByte(v, ','); i >= 0 {
			v = v[:i]
		}
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	// No header to trust. Falling back is better than dropping the
	// throttle: the address key still stands, and that is what stops
	// guessing against one account.
	return host
}

// eigenMellomtenar says whether the request came through something we
// set up on this machine. Only then are its headers worth anything.
func eigenMellomtenar(vert string) bool {
	ip := net.ParseIP(vert)
	return ip != nil && ip.IsLoopback()
}

// sessionUserName gives the logged-in user's name, or empty. The nav is
// only drawn when there is a name to draw it for.
func sessionUserName(r *http.Request) string {
	if user := GetUserFromSession(r); user != nil {
		return user.Name
	}
	return ""
}

// sessionIsAdmin says whether the logged-in user has the admin
// permission, so the template only shows the link to those who can use
// it.
func sessionIsAdmin(r *http.Request) bool {
	return IsAdmin(GetUserFromSession(r))
}

// t is the translation, called from Go rather than from a template.
func t(lang, key string) string {
	return GetLocalization().T(lang, key)
}

// renderPage draws a page through the template manager. Eight handlers
// each had their own version of it, and one of them forgot the charset.
func renderPage(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	tm := GetTemplateManager()
	tmpl, ok := tm.GetTemplate(name)
	if !ok {
		// In development templates change while the server runs.
		tm.ReloadTemplates()
		tmpl, ok = tm.GetTemplate(name)
	}
	if !ok {
		log.Printf("malen %q finst ikkje", name)
		http.Error(w, "Malen finst ikkje", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("teikning av %q: %v", name, err)
		http.Error(w, "Feil ved teikning av sida", http.StatusInternalServerError)
	}
}
