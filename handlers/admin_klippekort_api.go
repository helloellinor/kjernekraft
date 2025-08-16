package handlers

import (
	"encoding/json"
	"kjernekraft/models"
	"net/http"
	"strconv"
)

// CreateKlippekortTypeRequest represents the request to create a new klippekort type
type CreateKlippekortTypeRequest struct {
	Name           string `json:"name"`
	Price          int    `json:"price"`           // Price in øre
	TotalPunches   int    `json:"total_punches"`
	ValidityMonths int    `json:"validity_months"`
	Description    string `json:"description"`
}

// CreateKlippekortTypeHandler handles POST /api/admin/klippekort-type
func CreateKlippekortTypeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	if !IsAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateKlippekortTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.Price <= 0 {
		http.Error(w, "Price must be positive", http.StatusBadRequest)
		return
	}
	if req.TotalPunches <= 0 {
		http.Error(w, "Total punches must be positive", http.StatusBadRequest)
		return
	}
	if req.ValidityMonths <= 0 {
		http.Error(w, "Validity months must be positive", http.StatusBadRequest)
		return
	}

	// Get database connection
	db := AdminDB.Conn

	// Create new klippekort package
	query := `
		INSERT INTO klippekort_packages (name, category, klipp_count, price, description, valid_days, active, is_popular)
		VALUES (?, 'Custom', ?, ?, ?, ?, true, false)
	`

	validDays := req.ValidityMonths * 30 // Convert months to approximate days
	
	result, err := db.Exec(query, req.Name, req.TotalPunches, req.Price, req.Description, validDays)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the created ID
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get created ID", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"id":      id,
		"message": "Klippekort type created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateKlippekortTypeHandler handles PUT /api/admin/klippekort-type/{id}
func UpdateKlippekortTypeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	if !IsAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get ID from URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req CreateKlippekortTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.Price <= 0 {
		http.Error(w, "Price must be positive", http.StatusBadRequest)
		return
	}
	if req.TotalPunches <= 0 {
		http.Error(w, "Total punches must be positive", http.StatusBadRequest)
		return
	}
	if req.ValidityMonths <= 0 {
		http.Error(w, "Validity months must be positive", http.StatusBadRequest)
		return
	}

	// Get database connection
	db := AdminDB.Conn

	// Update klippekort package
	query := `
		UPDATE klippekort_packages 
		SET name = ?, klipp_count = ?, price = ?, description = ?, valid_days = ?
		WHERE id = ?
	`

	validDays := req.ValidityMonths * 30 // Convert months to approximate days
	
	_, err = db.Exec(query, req.Name, req.TotalPunches, req.Price, req.Description, validDays, id)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message": "Klippekort type updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteKlippekortTypeHandler handles DELETE /api/admin/klippekort-type/{id}
func DeleteKlippekortTypeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	if !IsAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get ID from URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Get database connection
	db := AdminDB.Conn

	// Check if any users have active klippekort of this type
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM user_klippekort uk JOIN klippekort_packages kp ON uk.package_id = kp.id WHERE kp.id = ? AND uk.is_active = true", id).Scan(&count)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "Cannot delete klippekort type with active user klippekort", http.StatusBadRequest)
		return
	}

	// Mark as inactive instead of deleting (soft delete)
	query := `UPDATE klippekort_packages SET active = false WHERE id = ?`
	
	_, err = db.Exec(query, id)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message": "Klippekort type deactivated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetKlippekortTypesHandler handles GET /api/admin/klippekort-types  
func GetKlippekortTypesHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is admin
	if !IsAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get database connection
	db := AdminDB.Conn

	// Get all klippekort packages
	query := `
		SELECT id, name, category, klipp_count, price, description, valid_days, active, is_popular
		FROM klippekort_packages
		ORDER BY name
	`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var packages []models.KlippekortPackage
	for rows.Next() {
		var pkg models.KlippekortPackage
		err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Category, &pkg.KlippCount, &pkg.Price, &pkg.Description, &pkg.ValidDays, &pkg.Active, &pkg.IsPopular)
		if err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Calculate price per session
		if pkg.KlippCount > 0 {
			pkg.PricePerSession = pkg.Price / pkg.KlippCount
		}
		
		packages = append(packages, pkg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(packages)
}

// Helper function to check if user is admin
func IsAdminUser(r *http.Request) bool {
	// This would typically check session/JWT/etc.
	// For now, return true for development
	// TODO: Implement proper admin authentication
	return true
}