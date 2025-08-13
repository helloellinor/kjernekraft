package handlers

import (
	"encoding/json"
	"kjernekraft/database"
	"kjernekraft/models"
	"net/http"
	"strconv"
)

var DB *database.Database // Set this from main

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
