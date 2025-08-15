package handlers

import (
	"encoding/gob"
	"kjernekraft/models"
	"net/http"

	"github.com/gorilla/sessions"
)

var (
	sessionStore *sessions.CookieStore
	sessionName  = "kjernekraft-session"
)

// InitializeSessionStore sets up the session store
func InitializeSessionStore() {
	// Use a secure key for production - this should be environment variable
	sessionStore = sessions.NewCookieStore([]byte("super-secret-key-change-in-production"))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	
	// Register the User type for session storage
	gob.Register(models.User{})
}

// GetUserFromSession retrieves the current user from the session
func GetUserFromSession(r *http.Request) *models.User {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		return nil
	}
	
	userInterface, ok := session.Values["user"]
	if !ok {
		return nil
	}
	
	user, ok := userInterface.(models.User)
	if !ok {
		return nil
	}
	
	return &user
}

// SetUserInSession stores the user in the session
func SetUserInSession(w http.ResponseWriter, r *http.Request, user *models.User) error {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		return err
	}
	
	session.Values["user"] = *user
	return session.Save(r, w)
}

// ClearUserSession removes the user from the session
func ClearUserSession(w http.ResponseWriter, r *http.Request) error {
	session, err := sessionStore.Get(r, sessionName)
	if err != nil {
		return err
	}
	
	session.Values["user"] = nil
	session.Options.MaxAge = -1 // Delete the session
	return session.Save(r, w)
}

// IsLoggedIn checks if a user is logged in
func IsLoggedIn(r *http.Request) bool {
	return GetUserFromSession(r) != nil
}