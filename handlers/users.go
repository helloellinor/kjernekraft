package handlers

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

// InnloggingHandler serves the login page
func InnloggingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Check if user is already logged in
		if IsLoggedIn(r) {
			http.Redirect(w, r, "/elev/hjem", http.StatusTemporaryRedirect)
			return
		}

		// Get language from cookies/request (using new system)
		lang := GetLanguageFromRequest(r)

		data := map[string]interface{}{
			"Title":       "Innlogging",
			"CurrentPage": "innlogging",
			"Lang":        lang,
			"CSRFToken":   CSRFToken(r),
			"IsAdmin":     sessionIsAdmin(r),
			"UserName":    sessionUserName(r),
			"Error":       loginError(r.URL.Query().Get("error")),
		}

		// Use the new template system
		tm := GetTemplateManager()
		if tmpl, exists := tm.GetTemplate("pages/innlogging"); exists {
			w.Header().Set("Content-Type", "text/html")
			if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
				http.Error(w, "Template execution error", http.StatusInternalServerError)
			}
			return
		}

		// If template doesn't exist, return error
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		// Handle login form submission
		email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
		password := r.FormValue("password")

		// Bremsa tel per IP og per e-post. Ein motstandar som byter
		// e-post kvar gong vert stogga av IP-nykelen; ein som kjem fraa
		// mange adressor mot den same kontoen vert stogga av den hine.
		ipKey := "ip:" + clientKey(r)
		userKey := "e-post:" + email
		if bremsaPaa() && (loginAttempts.Blocked(ipKey) || loginAttempts.Blocked(userKey)) {
			http.Redirect(w, r, "/innlogging?error=blocked", http.StatusSeeOther)
			return
		}

		// Validate credentials
		user, err := DB.AuthenticateUser(email, password)
		if err != nil {
			// Eit gale passord og ein feil i basen er tvo ulike ting, og
			// dei var det same her fyrr: kvar feil vart til «ugyldig
			// e-post eller passord». Ein brukar som ikkje kunde logga inn
			// av di eit felt i basen var NULL fekk vita at han hugsa
			// gale, og han hugsa rett. Ingen ting nokon stad sa kva som
			// verkeleg hende — korkje til honom eller til oss.
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

		// Utviklarlista er ei fil, ikkje eit løyve nokon set — so det
		// finst ingen augneblink der nokon «vert» utviklar som me kunde
		// hekta oss paa. Innloggingi er den fyrste gongen me veit kven
		// dette er etter at fila kann ha skift, og difor er det her
		// faktureringi vert stogga for deim. Læraren er alt stogga i
		// SettLoyve; dette tek den hine vegen.
		if stogga, err := DB.SynkFriMedlemskap(int64(user.ID)); err != nil {
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

// loginError gjer spurnadsparameteren um til ein umsetjingsnykel. Han
// gjev tomt for alt han ikkje kjenner, so ingen kann skriva sin eigen
// bodskap inn i innloggingssida gjenom adressa.
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

// LogoutHandler handles user logout.
//
// Han tek berre imot POST. Ei utlogging paa GET kann ein framand sida
// utløysa med eit bilete-merke, og daa er du logga ut utan aa ha bede um
// det.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := ClearUserSession(w, r)
	if err != nil {
		http.Error(w, "Logout error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/innlogging", http.StatusSeeOther)
}

// AddPaymentMethodHandler er teken burt. Han laga ein betalingsmaate
// paa kva som helst user_id or spurnadsstrengen, utan innlogging. Naar
// Stripe vert kopla paa, kjem betalingsmaatarne derifraa — ikkje fraa
// ein GET-parameter.

// GetUserPaymentMethodsHandler gjev betalingsmaatarne aat den innlogga
// brukaren, og berre deim.
func GetUserPaymentMethodsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Ikkje innlogga", http.StatusUnauthorized)
		return
	}
	methods, err := DB.GetUserPaymentMethods(int64(user.ID))
	if err != nil {
		http.Error(w, "Could not fetch payment methods", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(methods)
}

// AddUserHandler er teken burt. Han tok imot vilkaarleg JSON, sette
// løyvi brukaren sjølv bad um, og skreiv passordet slik det kom inn —
// utan bcrypt. SignUpHandler er den einaste vegen inn no.

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
		Loyve:                  []string{"user"}, // Default role
	}

	// Create user in database
	userID, err := DB.CreateUser(user)
	if err != nil {
		// Dei tvo avslagi brukaren kann gjera noko med fær kvar sitt
		// ord; alt anna er vaart problem og skal ikkje ut paa skjermen.
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
