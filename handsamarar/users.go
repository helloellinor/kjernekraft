package handsamarar

import (
	"encoding/json"
	"errors"
	"kjernekraft/database"
	"kjernekraft/models"
	"log"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var DB *database.Database // Set this from main

// Innlogging serves the login page
func (a *App) Innlogging(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Check if user is already logged in
		if IsLoggedIn(r) {
			http.Redirect(w, r, "/elev/hjem", http.StatusTemporaryRedirect)
			return
		}

		renderPage(w, r, "pages/innlogging", sidedata(r, SidaInnlogging, "Innlogging", map[string]any{
			"Error": loginError(r.URL.Query().Get("error")),
		}))
		return
	}

	if r.Method == "POST" {
		// Handle login form submission
		email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
		password := r.FormValue("password")

		// Bremsa tel per IP og per e-post. Ein motstandar som byter
		// e-post kvar gong vert stogga av IP-nykelen; ein som kjem frå
		// mange adressor mot den same kontoen vert stogga av den hine.
		ipKey := "ip:" + clientKey(r)
		userKey := "e-post:" + email
		if bremsaPå() && (loginAttempts.Blocked(ipKey) || loginAttempts.Blocked(userKey)) {
			http.Redirect(w, r, "/innlogging?error=blocked", http.StatusSeeOther)
			return
		}

		// Validate credentials
		user, err := a.DB.AuthenticateUser(email, password)
		if err != nil {
			// A wrong password and a database error are two different things, and
			// 			// they were the same here: every failure became "invalid email or
			// 			// password". A user who could not log in because a field was NULL
			// 			// was told they misremembered, and they had not. Nothing anywhere
			// 			// said what actually happened — not to them, not to us.
			if !errors.Is(err, database.ErrUgyldigInnlogging) {
				log.Printf("innlogging for %q: %v", email, err)
				http.Redirect(w, r, "/innlogging?error=teknisk", http.StatusSeeOther)
				return
			}
			loginAttempts.Fail(ipKey)
			loginAttempts.Fail(userKey)
			http.Redirect(w, r, "/innlogging?error=invalid", http.StatusSeeOther)
			return
		}

		loginAttempts.Reset(ipKey)
		loginAttempts.Reset(userKey)

		// Set user in session
		err = SetUserInSession(w, r, user)
		if err != nil {
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		// The developer list is a file, not a permission somebody grants — so
		// 		// there is no moment where someone "becomes" a developer that we could
		// 		// hook onto. Login is the first time we know who this is after the
		// 		// file may have changed, so this is where their billing is stopped.
		// 		// The teacher is already stopped in SettLøyve; this covers the other
		// 		// direction.
		if stogga, err := a.DB.SynkFriMedlemskap(int64(user.ID)); err != nil {
			log.Printf("fritt medlemskap for %d: %v", user.ID, err)
		} else if stogga {
			log.Printf("brukar %d er forfremja; den betalte avtalen er avslutta", user.ID)
		}

		// Redirect to dashboard
		http.Redirect(w, r, "/elev/hjem", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// loginError turns the query parameter into a translation key. It returns
// empty for anything it does not know, so nobody can write their own
// message into the login page through the URL.
func loginError(code string) string {
	switch code {
	case "invalid":
		return "login.error_invalid"
	case "blocked":
		return "login.error_blocked"
	case "teknisk":
		return "login.error_technical"
	default:
		return ""
	}
}

// Logout handles user logout.
//
// It accepts POST only. A logout on GET can be triggered by a foreign page
// with an image tag, and then you are logged out without asking.
func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	err := ClearUserSession(w, r)
	if err != nil {
		http.Error(w, "Logout error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/innlogging", http.StatusSeeOther)
}

// AddPaymentMethodHandler is gone. It created a payment method on any
// user_id from the query string, without a login. When Stripe is connected
// the payment methods come from there — not from a GET parameter.

// GetUserPaymentMethods gives the logged-in user's payment methods,
// and only theirs.
func (a *App) GetUserPaymentMethods(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)
	methods, err := a.DB.GetUserPaymentMethods(int64(user.ID))
	if err != nil {
		http.Error(w, "Could not fetch payment methods", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(methods)
}

// AddUserHandler is gone. It accepted arbitrary JSON, set whatever
// permissions the user asked for, and stored the password as it arrived —
// without bcrypt. SignUp is the only way in now.

func (a *App) SignUp(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form data
	err := r.ParseMultipartForm(32 << 20) // 32 MB max memory
	if err != nil {
		// Fallback to regular form parsing
		err = r.ParseForm()
		if err != nil {
			svarFeil(w, r, http.StatusBadRequest, "feil.skjemadata")
			return
		}
	}

	// Extract form values
	name := strings.TrimSpace(r.FormValue("name"))
	birthdate := strings.TrimSpace(r.FormValue("birthdate"))
	email := strings.TrimSpace(r.FormValue("email"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	address := strings.TrimSpace(r.FormValue("address"))
	postalCode := strings.TrimSpace(r.FormValue("postal_code"))
	city := strings.TrimSpace(r.FormValue("city"))
	country := strings.TrimSpace(r.FormValue("country"))
	password := r.FormValue("password")
	newsletter := r.FormValue("newsletter") == "on"
	termsAccepted := r.FormValue("terms_accepted") == "on"

	// Validate required fields
	if name == "" || birthdate == "" || email == "" || phone == "" || password == "" {
		svarFeil(w, r, http.StatusBadRequest, "feil.pakravde")
		return
	}

	if !termsAccepted {
		svarFeil(w, r, http.StatusBadRequest, "feil.vilkaar")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		svarFeil(w, r, http.StatusInternalServerError, "feil.krasj")
		return
	}

	// Create user object
	user := models.User{
		Name:                   name,
		Birthdate:              birthdate,
		Email:                  email,
		Phone:                  phone,
		Address:                address,
		PostalCode:             postalCode,
		City:                   city,
		Country:                country,
		Password:               string(hashedPassword),
		NewsletterSubscription: newsletter,
		TermsAccepted:          termsAccepted,
		Løyve:                  []string{"user"}, // Default role
	}

	// Create user in database
	userID, err := a.DB.CreateUser(user)
	if err != nil {
		// The two refusals the user can do something about each get their own
		// 		// word; everything else is our problem and does not belong on the
		// 		// screen.
		switch {
		case errors.Is(err, database.ErrEpostIBruk):
			svarFeil(w, r, http.StatusConflict, "feil.epost_i_bruk")
		case errors.Is(err, database.ErrTelefonIBruk):
			svarFeil(w, r, http.StatusConflict, "feil.telefon_i_bruk")
		default:
			log.Printf("CreateUser: %v", err)
			svarFeil(w, r, http.StatusInternalServerError, "feil.krasj")
		}
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="success">Bruker registrert med suksess! Bruker ID: ` + strconv.FormatInt(userID, 10) + `</div>`))
}
