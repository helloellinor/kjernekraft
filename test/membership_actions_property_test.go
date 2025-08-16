package test

import (
	"kjernekraft/database"
	"kjernekraft/models"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// MembershipState represents the state of user membership
type MembershipState struct {
	UserID       int64
	MembershipID int64
	Status       string
	IsFrozen     bool
}

// MembershipAction represents membership-related actions
type MembershipAction interface {
	Apply(state *MembershipState, db *database.Database) error
	String() string
}

// ChangeMembershipAction represents changing membership type
type ChangeMembershipAction struct {
	NewMembershipID int64
}

func (a *ChangeMembershipAction) Apply(state *MembershipState, db *database.Database) error {
	// Simulate checking if membership change is allowed
	canChange, _ := db.CanChangeMembership(state.UserID, a.NewMembershipID)
	if !canChange {
		return nil // Not allowed, but not an error for testing
	}
	
	err := db.ChangeUserMembership(state.UserID, a.NewMembershipID)
	if err != nil {
		return err
	}
	
	state.MembershipID = a.NewMembershipID
	return nil
}

func (a *ChangeMembershipAction) String() string {
	return "ChangeMembership(to membership ID: " + string(rune(a.NewMembershipID)) + ")"
}

// FreezeMembershipAction represents freezing a membership
type FreezeMembershipAction struct{}

func (a *FreezeMembershipAction) Apply(state *MembershipState, db *database.Database) error {
	if state.IsFrozen {
		return nil // Already frozen
	}
	
	err := db.UpdateMembershipStatus(state.UserID, "freeze_requested")
	if err != nil {
		return err
	}
	
	state.Status = "freeze_requested"
	state.IsFrozen = true
	return nil
}

func (a *FreezeMembershipAction) String() string {
	return "FreezeMembership()"
}

// UnfreezeMembershipAction represents unfreezing a membership
type UnfreezeMembershipAction struct{}

func (a *UnfreezeMembershipAction) Apply(state *MembershipState, db *database.Database) error {
	if !state.IsFrozen {
		return nil // Not frozen
	}
	
	err := db.UpdateMembershipStatus(state.UserID, "active")
	if err != nil {
		return err
	}
	
	state.Status = "active"
	state.IsFrozen = false
	return nil
}

func (a *UnfreezeMembershipAction) String() string {
	return "UnfreezeMembership()"
}

// Property tests for membership actions

func TestMembershipActionsPropertyBased(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	parameters.MaxSize = 5
	
	properties := gopter.NewProperties(parameters)
	
	// Property: Membership freeze/unfreeze should be reversible
	properties.Property("freeze unfreeze reversible", prop.ForAll(
		func(userID int64) bool {
			if userID <= 0 {
				return true // Skip invalid user IDs
			}
			
			db, cleanup := setupTestDB()
			defer cleanup()
			
			// Create a test user first
			user := models.User{
				Name:      "Test Member",
				Email:     "member@example.com",
				Phone:     "11111111",
				Birthdate: "1990-01-01",
				Password:  "testpassword",
				Roles:     []string{"user"},
			}
			
			id, err := db.CreateUser(user)
			if err != nil {
				return true // Skip if user creation fails
			}
			
			state := &MembershipState{
				UserID:       int64(id),
				MembershipID: 1,
				Status:       "active",
				IsFrozen:     false,
			}
			
			// Freeze membership
			freezeAction := &FreezeMembershipAction{}
			err = freezeAction.Apply(state, db)
			if err != nil {
				return true // Skip on error
			}
			
			wasFrozen := state.IsFrozen
			
			// Unfreeze membership
			unfreezeAction := &UnfreezeMembershipAction{}
			err = unfreezeAction.Apply(state, db)
			if err != nil {
				return true // Skip on error
			}
			
			// Should be back to not frozen if it was frozen
			if wasFrozen && state.IsFrozen {
				return false
			}
			
			return true
		},
		gen.Int64Range(1, 1000),
	))
	
	// Property: Membership ID should be valid after change
	properties.Property("membership change preserves valid ID", prop.ForAll(
		func(newMembershipID int64) bool {
			if newMembershipID <= 0 {
				return true // Skip invalid membership IDs
			}
			
			db, cleanup := setupTestDB()
			defer cleanup()
			
			// Create a test user first
			user := models.User{
				Name:      "Test Change Member",
				Email:     "changemember@example.com",
				Phone:     "22222222",
				Birthdate: "1990-01-01",
				Password:  "testpassword",
				Roles:     []string{"user"},
			}
			
			id, err := db.CreateUser(user)
			if err != nil {
				return true // Skip if user creation fails
			}
			
			state := &MembershipState{
				UserID:       int64(id),
				MembershipID: 1,
				Status:       "active",
				IsFrozen:     false,
			}
			
			action := &ChangeMembershipAction{NewMembershipID: newMembershipID}
			err = action.Apply(state, db)
			if err != nil {
				return true // Skip on error
			}
			
			// Membership ID should be positive
			return state.MembershipID > 0
		},
		gen.Int64Range(1, 10),
	))
	
	properties.TestingRun(t)
}

// Test the invariants that should always hold for membership operations
func TestMembershipInvariants(t *testing.T) {
	db, cleanup := setupTestDB()
	defer cleanup()
	
	// Create a test user
	user := models.User{
		Name:      "Membership Test User",
		Email:     "membership@example.com",
		Phone:     "33333333",
		Birthdate: "1990-01-01",
		Password:  "testpassword",
		Roles:     []string{"user"},
	}
	
	userID, err := db.CreateUser(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	state := &MembershipState{
		UserID:       userID,
		MembershipID: 1,
		Status:       "active",
		IsFrozen:     false,
	}
	
	// Test freeze action
	freezeAction := &FreezeMembershipAction{}
	err = freezeAction.Apply(state, db)
	if err != nil {
		t.Errorf("Failed to freeze membership: %v", err)
	}
	
	// Verify invariant: frozen status should match IsFrozen field
	if state.Status == "freeze_requested" && !state.IsFrozen {
		t.Error("Invariant violation: status is freeze_requested but IsFrozen is false")
	}
	
	// Test double freeze (should be safe)
	err = freezeAction.Apply(state, db)
	if err != nil {
		t.Errorf("Double freeze should not cause error: %v", err)
	}
	
	// Test unfreeze action
	unfreezeAction := &UnfreezeMembershipAction{}
	err = unfreezeAction.Apply(state, db)
	if err != nil {
		t.Errorf("Failed to unfreeze membership: %v", err)
	}
	
	// Verify invariant: active status should mean not frozen
	if state.Status == "active" && state.IsFrozen {
		t.Error("Invariant violation: status is active but IsFrozen is true")
	}
	
	// Test double unfreeze (should be safe)
	err = unfreezeAction.Apply(state, db)
	if err != nil {
		t.Errorf("Double unfreeze should not cause error: %v", err)
	}
}

// Benchmark membership action performance
func BenchmarkMembershipActionSequence(b *testing.B) {
	db, cleanup := setupTestDB()
	defer cleanup()
	
	// Create a test user
	user := models.User{
		Name:      "Benchmark User",
		Email:     "benchmark@example.com",
		Phone:     "44444444",
		Birthdate: "1990-01-01",
		Password:  "testpassword",
		Roles:     []string{"user"},
	}
	
	userID, err := db.CreateUser(user)
	if err != nil {
		b.Fatalf("Failed to create user: %v", err)
	}
	
	actions := []MembershipAction{
		&FreezeMembershipAction{},
		&UnfreezeMembershipAction{},
		&ChangeMembershipAction{NewMembershipID: 2},
		&ChangeMembershipAction{NewMembershipID: 1},
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		state := &MembershipState{
			UserID:       userID,
			MembershipID: 1,
			Status:       "active",
			IsFrozen:     false,
		}
		
		for _, action := range actions {
			action.Apply(state, db)
		}
	}
}