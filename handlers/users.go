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
		data := map[string]interface{}{
			"Title":       "Innlogging",
			"CurrentPage": "innlogging",
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
		email := r.FormValue("email")
		password := r.FormValue("password")

		// TODO: Implement actual authentication logic
		// For now, redirect to dashboard
		if email != "" && password != "" {
			http.Redirect(w, r, "/elev/hjem", http.StatusTemporaryRedirect)
		} else {
			// Redirect back to login with error
			http.Redirect(w, r, "/innlogging?error=invalid", http.StatusTemporaryRedirect)
		}
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

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

func GetUserRolesHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}
	roles, err := DB.GetUserRoles(userID)
	if err != nil {
		http.Error(w, "Could not fetch roles", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(roles)
}

func AddPaymentMethodHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	provider := r.URL.Query().Get("provider")
	providerID := r.URL.Query().Get("provider_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || provider == "" || providerID == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	pmID, err := DB.AddPaymentMethod(userID, provider, providerID)
	if err != nil {
		http.Error(w, "Could not add payment method", http.StatusInternalServerError)
		return
	}
	if err := DB.AssignPaymentMethodToUser(userID, pmID); err != nil {
		http.Error(w, "Could not assign payment method", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payment method assigned"))
}

func GetUserPaymentMethodsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}
	methods, err := DB.GetUserPaymentMethods(userID)
	if err != nil {
		http.Error(w, "Could not fetch payment methods", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(methods)
}

func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid user data", http.StatusBadRequest)
		return
	}
	// Example: create user in database (implement CreateUser in database.go)
	userID, err := DB.CreateUser(u)
	if err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}
	u.ID = int(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

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
