package handlers

import (
	"encoding/json"
	"html"
	"kjernekraft/database"
	"kjernekraft/models"
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
		if loginAttempts.Blocked(ipKey) || loginAttempts.Blocked(userKey) {
			http.Redirect(w, r, "/innlogging?error=blocked", http.StatusSeeOther)
			return
		}

		// Validate credentials
		user, err := DB.AuthenticateUser(email, password)
		if err != nil {
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

// AssignRoleToUserHandler ligg bak RequireAdmin i rutaren. Han var open
// fyrr, og daa var «gjer meg til admin» eitt einaste kall.
func AssignRoleToUserHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	roleName := r.URL.Query().Get("role")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || roleName == "" {
		http.Error(w, "Invalid user_id or role", http.StatusBadRequest)
		return
	}
	roleID, err := DB.AddRole(roleName)
	if err != nil {
		http.Error(w, "Could not add role", http.StatusInternalServerError)
		return
	}
	if err := DB.AssignRoleToUser(userID, roleID); err != nil {
		http.Error(w, "Could not assign role", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Role assigned"))
}

// GetUserRolesHandler gjev rollone aat den innlogga brukaren. Han tok
// eit user_id or spurnadsstrengen fyrr, og daa kunde kven som helst
// lesa kven som helst.
func GetUserRolesHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Ikkje innlogga", http.StatusUnauthorized)
		return
	}
	roles, err := DB.GetUserRoles(int64(user.ID))
	if err != nil {
		http.Error(w, "Could not fetch roles", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(roles)
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
// rollone brukaren sjølv bad um, og skreiv passordet slik det kom inn —
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
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`<div class="error">Ugyldig skjemadata</div>`))
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="error">Alle påkrevde felt må fylles ut</div>`))
		return
	}

	if !termsAccepted {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="error">Du må akseptere handelsbetingelsene</div>`))
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<div class="error">Feil ved behandling av passord</div>`))
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
		Roles:                  []string{"user"}, // Default role
	}

	// Create user in database
	userID, err := DB.CreateUser(user)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		errorMsg := html.EscapeString(err.Error())
		w.Write([]byte(`<div class="error">` + errorMsg + `</div>`))
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="success">Bruker registrert med suksess! Bruker ID: ` + strconv.FormatInt(userID, 10) + `</div>`))
}
