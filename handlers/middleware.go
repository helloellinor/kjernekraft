package handlers

import (
	"context"
	"log"
	"net"
	"net/http"
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

// WithUser les brukaren or økti ein gong og legg honom i samanhengen, so
// dei femten handsamarane som spør etter honom ikkje gjev femten
// basesoknader.
func WithUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := sessionUserID(r); ok {
			if user, err := DB.GetUserByID(id); err == nil {
				r = withUser(r, user)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth stengjer ruta for den som ikkje er innlogga. Ei API-rute
// fær 401 so htmx kann sjaa skilnad; ei side fær ei umleiding til
// innloggingi.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetUserFromSession(r) == nil {
			denyUnauthenticated(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin krev rolla «admin», og rolla vert lesi or basen — ikkje or
// kaka lesaren sender.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromSession(r)
		if user == nil {
			denyUnauthenticated(w, r)
			return
		}
		if !IsAdmin(user) {
			http.Error(w, "Ingen tilgang", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireDevelopment held testdata-rutorne utanfor drift. Dei skriv yver
// basen, og dei skal ikkje kunna naaast av ein framand.
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
		http.Error(w, "Ikkje innlogga", http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/innlogging", http.StatusSeeOther)
}

// CSRF krev eit kjennemerke paa alt som endrar noko. Kjennemerket ligg i
// økti — det er fasiten — og vert spegla ut i ei kaka lesaren kann lesa,
// so htmx og fetch kann senda det attende i ein hovudlinja.
//
// Det er økti som avgjer. Ei rein dubbel-innsending, der ein berre
// samanliknar kaka med hovudlinja, godtek eit kjennemerke ein motstandar
// hev sett sjølv.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ensureCSRFToken(w, r)

		// Kjennemerket fylgjer soknaden i staden for aa slaast upp paa
		// nytt i malen. Slær ein det upp paa nytt, lyt ein lesa økti ein
		// gong til — og er kaka uleseleg (til dømes av di nykelen er
		// skift), gjev det andre uppslaget tomt medan kaka hev eit
		// ferskt kjennemerke. Daa fekk skjemaet value="" og brukaren
		// kom aldri gjenom innloggingi att.
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
			http.Error(w, "Ugyldig eller manglande CSRF-kjennemerke", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ensureCSRFToken hentar kjennemerket or økti, lagar eit nytt um det
// manglar, og syter for at kaka stend i takt med det.
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
		http.SetCookie(w, &http.Cookie{
			Name:  csrfCookieName,
			Value: token,
			Path:  "/",
			// Ikkje HttpOnly: skriptet i base.html skal lesa han og
			// leggja honom i hovudlinja. Kjennemerket er ikkje ein
			// løyndom for sida sjølv — det er eit prov paa at soknaden
			// kjem derifraa.
			HttpOnly: false,
			Secure:   !IsDevelopment(),
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400 * 7,
		})
	}
	return token
}

// CSRFToken gjev malane kjennemerket, so skjema kann bera det i eit løynt
// felt. Det kjem or kall-samanhengen, der CSRF-mellomvara la det.
func CSRFToken(r *http.Request) string {
	if token, ok := r.Context().Value(csrfCtxKey).(string); ok {
		return token
	}
	// Utan mellomvara — i ein prøve, til dømes — les me økti beinveges.
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		return ""
	}
	token, _ := session.Values[csrfSessionKey].(string)
	return token
}

// withCSRFToken legg kjennemerket i kall-samanhengen.
func withCSRFToken(r *http.Request, token string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), csrfCtxKey, token))
}

// ---- innloggingsbremsa ----
//
// Utan ei bremsa er passordfeltet ein oråkall: ein kann gissa so fort
// maskini orkar. Bremsa tel mislukka forsøk per IP og per e-post, og
// stengjer i femtan minutt etter ti.

const (
	maxLoginAttempts = 10
	loginWindow      = 15 * time.Minute
)

type attemptCounter struct {
	mu       sync.Mutex
	attempts map[string]*attemptRecord
}

type attemptRecord struct {
	count int
	until time.Time
}

var loginAttempts = &attemptCounter{attempts: make(map[string]*attemptRecord)}

// Blocked segjer um nykelen er utestengd, og ryddjer burt utgjengne postar
// medan han gjeng gjenom.
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

// sweep lyt kallast med laasen halden.
func (c *attemptCounter) sweep() {
	now := time.Now()
	for k, rec := range c.attempts {
		if now.After(rec.until) {
			delete(c.attempts, k)
		}
	}
}

// clientKey er kjelda soknaden kjem fraa. Bak ein umvend mellomtenar er
// RemoteAddr adressa aat mellomtenaren, og daa bremsar me heile verdi
// under eitt. X-Forwarded-For er ikkje lesen med vilje: han er ei
// hovudlina kven som helst kann dikta upp, og daa bremsar me ingen.
func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// sessionUserName gjev namnet aat den innlogga brukaren, eller tomt.
// Leidingi vert teikna berre naar det er eit namn aa teikna henne for.
func sessionUserName(r *http.Request) string {
	if user := GetUserFromSession(r); user != nil {
		return user.Name
	}
	return ""
}

// sessionIsAdmin segjer um den innlogga brukaren hev admin-rolla, so
// malen kann syna administrasjonslenkja for dei som faktisk kjem inn.
func sessionIsAdmin(r *http.Request) bool {
	return IsAdmin(GetUserFromSession(r))
}

// t er umsetjingi, kalla fraa Go og ikkje fraa ein mal.
func t(lang, key string) string {
	return GetLocalization().T(lang, key)
}

// renderPage teiknar ei sida gjenom malstyraren.
//
// Dei aatte handsamarane gjorde kvar sin utgaava av det same: hent
// styraren, slaa upp malen, prøv aa lasta paa nytt um han manglar,
// skriv Content-Type, køyr «base», og logg feilen. Det stod ni gonger
// med smaa skilnader — ein av deim gløymde teiknsettet.
func renderPage(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	tm := GetTemplateManager()
	tmpl, ok := tm.GetTemplate(name)
	if !ok {
		// I utvikling vert malarne endra medan tenaren gjeng.
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
