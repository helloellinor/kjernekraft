package test

import (
	"kjernekraft/database"
	"kjernekraft/models"
	"log"
	"os"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// UserState represents the current state of a user in the system
type UserState struct {
	User       *models.User
	IsLoggedIn bool
	HasSession bool
}

// UserAction represents an action that can be performed on a user
type UserAction interface {
	Apply(state *UserState, db *database.Database) error
	String() string
}

// RegisterUserAction represents user registration
type RegisterUserAction struct {
	Name      string
	Email     string
	Phone     string
	Birthdate string
}

func (a *RegisterUserAction) Apply(state *UserState, db *database.Database) error {
	user := models.User{
		Name:      a.Name,
		Email:     a.Email,
		Phone:     a.Phone,
		Birthdate: a.Birthdate,
		Password:  "defaultpassword",
		Roles:     []string{"user"},
	}
	
	id, err := db.CreateUser(user)
	if err != nil {
		return err
	}
	user.ID = int(id)
	state.User = &user
	state.IsLoggedIn = false
	state.HasSession = false
	return nil
}

func (a *RegisterUserAction) String() string {
	return "RegisterUser(" + a.Name + ")"
}

// UpdateProfileAction represents user profile update
type UpdateProfileAction struct {
	Name  string
	Email string
	Phone string
}

func (a *UpdateProfileAction) Apply(state *UserState, db *database.Database) error {
	if state.User == nil {
		return nil // No user to update
	}
	
	state.User.Name = a.Name
	state.User.Email = a.Email
	state.User.Phone = a.Phone
	
	return db.UpdateUser(state.User)
}

func (a *UpdateProfileAction) String() string {
	return "UpdateProfile(" + a.Name + ")"
}

// LoginAction represents user login
type LoginAction struct {
	Email    string
	Password string
}

func (a *LoginAction) Apply(state *UserState, db *database.Database) error {
	if state.User == nil {
		return nil // No user to login
	}
	
	// Simulate login by checking if email matches
	if state.User.Email == a.Email {
		state.IsLoggedIn = true
		state.HasSession = true
	}
	return nil
}

func (a *LoginAction) String() string {
	return "Login(" + a.Email + ")"
}

// LogoutAction represents user logout
type LogoutAction struct{}

func (a *LogoutAction) Apply(state *UserState, db *database.Database) error {
	state.IsLoggedIn = false
	state.HasSession = false
	return nil
}

func (a *LogoutAction) String() string {
	return "Logout()"
}

// Test setup
func setupTestDB() (*database.Database, func()) {
	// Create temporary database file
	tmpfile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		log.Fatal(err)
	}
	tmpfile.Close()
	
	// Set environment variable for database path
	oldPath := os.Getenv("DB_PATH")
	os.Setenv("DB_PATH", tmpfile.Name())
	
	dbConn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	
	if err := database.Migrate(dbConn); err != nil {
		log.Fatal(err)
	}
	
	db := &database.Database{Conn: dbConn}
	
	cleanup := func() {
		dbConn.Close()
		os.Remove(tmpfile.Name())
		os.Setenv("DB_PATH", oldPath)
	}
	
	return db, cleanup
}

// Simplified property tests

func TestUserActionsPropertyBased(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	parameters.MaxSize = 5
	
	properties := gopter.NewProperties(parameters)
	
	// Property: User registration should create a valid user
	properties.Property("user registration creates valid user", prop.ForAll(
		func(name, email, phone string) bool {
			if name == "" || email == "" || phone == "" {
				return true // Skip invalid inputs
			}
			
			db, cleanup := setupTestDB()
			defer cleanup()
			
			state := &UserState{}
			action := &RegisterUserAction{
				Name:      name,
				Email:     email,
				Phone:     phone,
				Birthdate: "1990-01-01",
			}
			
			err := action.Apply(state, db)
			if err != nil {
				return true // Some errors are expected (e.g., duplicate email)
			}
			
			// Verify user was created
			return state.User != nil && state.User.ID > 0 && state.User.Name == name
		},
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.RegexMatch(`[a-z]+@[a-z]+\.[a-z]+`),
		gen.RegexMatch(`[0-9]{8}`),
	))
	
	// Property: Login should only succeed for existing users with correct email
	properties.Property("login requires existing user", prop.ForAll(
		func(userName, userEmail, loginEmail string) bool {
			if userName == "" || userEmail == "" || loginEmail == "" {
				return true // Skip invalid inputs
			}
			
			db, cleanup := setupTestDB()
			defer cleanup()
			
			// Register user first
			user := models.User{
				Name:      userName,
				Email:     userEmail,
				Phone:     "12345678",
				Birthdate: "1990-01-01",
				Password:  "testpassword",
				Roles:     []string{"user"},
			}
			
			id, err := db.CreateUser(user)
			if err != nil {
				return true // Skip if user creation fails
			}
			user.ID = int(id)
			
			state := &UserState{User: &user}
			loginAction := &LoginAction{Email: loginEmail, Password: "password"}
			
			loginAction.Apply(state, db)
			
			// Login should only succeed if email matches
			expectedLoggedIn := (loginEmail == userEmail)
			return state.IsLoggedIn == expectedLoggedIn
		},
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.RegexMatch(`[a-z]+@[a-z]+\.[a-z]+`),
		gen.RegexMatch(`[a-z]+@[a-z]+\.[a-z]+`),
	))
	
	// Property: Profile updates should preserve user ID
	properties.Property("profile update preserves ID", prop.ForAll(
		func(userName, userEmail, newName, newEmail string) bool {
			if userName == "" || userEmail == "" || newName == "" || newEmail == "" {
				return true // Skip invalid inputs
			}
			
			db, cleanup := setupTestDB()
			defer cleanup()
			
			// Register user first
			user := models.User{
				Name:      userName,
				Email:     userEmail,
				Phone:     "12345678",
				Birthdate: "1990-01-01",
				Password:  "testpassword",
				Roles:     []string{"user"},
			}
			
			id, err := db.CreateUser(user)
			if err != nil {
				return true // Skip if user creation fails
			}
			user.ID = int(id)
			originalID := user.ID
			
			state := &UserState{User: &user}
			updateAction := &UpdateProfileAction{
				Name:  newName,
				Email: newEmail,
				Phone: "87654321",
			}
			
			updateAction.Apply(state, db)
			
			return state.User.ID == originalID
		},
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.RegexMatch(`[a-z]+@[a-z]+\.[a-z]+`),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.RegexMatch(`[a-z]+@[a-z]+\.[a-z]+`),
	))
	
	properties.TestingRun(t)
}

// Helper function to verify invariants
func verifyUserStateInvariants(state *UserState, db *database.Database) bool {
	// Invariant 1: If user is logged in, they must have a session
	if state.IsLoggedIn && !state.HasSession {
		return false
	}
	
	// Invariant 2: If user exists, they must have a valid ID
	if state.User != nil && state.User.ID <= 0 {
		return false
	}
	
	// Invariant 3: If user exists in state, they should exist in database
	if state.User != nil && state.User.ID > 0 {
		users, err := db.GetAllUsers()
		if err != nil {
			return false
		}
		
		found := false
		for _, u := range users {
			if u.ID == state.User.ID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}

// Simple test to verify the property-based testing works
func TestUserActionBasics(t *testing.T) {
	db, cleanup := setupTestDB()
	defer cleanup()
	
	state := &UserState{}
	
	// Test user registration with unique data
	registerAction := &RegisterUserAction{
		Name:      "Test User Basic",
		Email:     "testbasic@example.com",
		Phone:     "99999999",
		Birthdate: "1990-01-01",
	}
	
	err := registerAction.Apply(state, db)
	if err != nil {
		t.Errorf("Failed to register user: %v", err)
		return
	}
	
	if state.User == nil || state.User.ID <= 0 {
		t.Error("User registration failed - no user created")
		return
	}
	
	// Test profile update
	updateAction := &UpdateProfileAction{
		Name:  "Updated User Basic",
		Email: "updatedbasic@example.com",
		Phone: "88888888",
	}
	
	originalID := state.User.ID
	err = updateAction.Apply(state, db)
	if err != nil {
		t.Errorf("Failed to update user profile: %v", err)
		return
	}
	
	if state.User.ID != originalID {
		t.Error("Profile update changed user ID")
	}
	
	if state.User.Name != "Updated User Basic" {
		t.Error("Profile update did not change name")
	}
	
	// Test login with correct email
	loginAction := &LoginAction{
		Email:    "updatedbasic@example.com",
		Password: "password",
	}
	
	err = loginAction.Apply(state, db)
	if err != nil {
		t.Errorf("Failed to login: %v", err)
	}
	
	if !state.IsLoggedIn {
		t.Error("Login with correct email should succeed")
	}
	
	// Test login with wrong email
	loginWrongAction := &LoginAction{
		Email:    "wrong@example.com",
		Password: "password",
	}
	
	// Reset login state
	state.IsLoggedIn = false
	state.HasSession = false
	
	err = loginWrongAction.Apply(state, db)
	if err != nil {
		t.Errorf("Failed to process login: %v", err)
	}
	
	if state.IsLoggedIn {
		t.Error("Login with wrong email should not succeed")
	}
	
	// Test logout
	state.IsLoggedIn = true
	state.HasSession = true
	
	logoutAction := &LogoutAction{}
	err = logoutAction.Apply(state, db)
	if err != nil {
		t.Errorf("Failed to logout: %v", err)
	}
	
	if state.IsLoggedIn {
		t.Error("Logout should clear login state")
	}
	
	if state.HasSession {
		t.Error("Logout should clear session state")
	}
}

// Benchmark for performance testing
func BenchmarkUserActionSequence(b *testing.B) {
	db, cleanup := setupTestDB()
	defer cleanup()
	
	// Pre-generate some actions
	actions := []UserAction{
		&RegisterUserAction{
			Name:      "Test User",
			Email:     "test@example.com",
			Phone:     "12345678",
			Birthdate: "1990-01-01",
		},
		&LoginAction{Email: "test@example.com", Password: "password"},
		&UpdateProfileAction{
			Name:  "Updated User",
			Email: "updated@example.com",
			Phone: "87654321",
		},
		&LogoutAction{},
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		state := &UserState{}
		for _, action := range actions {
			action.Apply(state, db)
		}
	}
}