package handlers

import (
	"encoding/json"
	"kjernekraft/models"
	"net/http"
	"strconv"
	"time"
)

// CreateClassHandler creates a new class/event
func CreateClassHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add admin authentication check here

	var classData struct {
		Title          string `json:"title"`
		ClassType      string `json:"class_type"`
		TeacherName    string `json:"teacher_name"`
		Location       string `json:"location"`
		Date           string `json:"date"`
		StartTime      string `json:"start_time"`
		EndTime        string `json:"end_time"`
		Capacity       int    `json:"capacity"`
		Color          string `json:"color"`
		Description    string `json:"description"`
		IsRecurring    bool   `json:"is_recurring"`
		RecurringWeeks int    `json:"recurring_weeks"`
	}

	if err := json.NewDecoder(r.Body).Decode(&classData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Parse date and times
	classDate, err := time.Parse("2006-01-02", classData.Date)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse("15:04", classData.StartTime)
	if err != nil {
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse("15:04", classData.EndTime)
	if err != nil {
		http.Error(w, "Invalid end time format", http.StatusBadRequest)
		return
	}

	// Combine date and times
	startDateTime := time.Date(classDate.Year(), classDate.Month(), classDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, classDate.Location())
	endDateTime := time.Date(classDate.Year(), classDate.Month(), classDate.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, classDate.Location())

	// Create events based on recurring settings
	weeksToCreate := 1
	if classData.IsRecurring {
		weeksToCreate = classData.RecurringWeeks
	}

	var createdEventIDs []int64

	for week := 0; week < weeksToCreate; week++ {
		weekOffset := time.Duration(week) * 7 * 24 * time.Hour
		weekStartTime := startDateTime.Add(weekOffset)
		weekEndTime := endDateTime.Add(weekOffset)

		event := models.Event{
			Title:            classData.Title,
			Description:      classData.Description,
			StartTime:        weekStartTime,
			EndTime:          weekEndTime,
			Location:         classData.Location,
			Organizer:        "Kjernekraft",
			ClassType:        classData.ClassType,
			TeacherName:      classData.TeacherName,
			Capacity:         classData.Capacity,
			CurrentEnrolment: 0,
			Color:            classData.Color,
		}

		eventID, err := AdminDB.CreateEvent(event)
		if err != nil {
			http.Error(w, "Could not create event", http.StatusInternalServerError)
			return
		}
		createdEventIDs = append(createdEventIDs, eventID)
	}

	response := map[string]interface{}{
		"success":    true,
		"message":    "Class(es) created successfully",
		"event_ids":  createdEventIDs,
		"events_created": len(createdEventIDs),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteClassHandler deletes a class/event
func DeleteClassHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add admin authentication check here

	// Extract class ID from URL path
	// Expected format: /api/admin/class/{id}
	path := r.URL.Path
	classIDStr := path[len("/api/admin/class/"):]
	
	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	if err := AdminDB.DeleteEvent(classID); err != nil {
		http.Error(w, "Could not delete class", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Class deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateClassHandler updates a class/event
func UpdateClassHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add admin authentication check here

	// Extract class ID from URL path
	path := r.URL.Path
	classIDStr := path[len("/api/admin/class/"):]
	
	classID, err := strconv.ParseInt(classIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	var updateData struct {
		Title       string `json:"title"`
		ClassType   string `json:"class_type"`
		TeacherName string `json:"teacher_name"`
		Location    string `json:"location"`
		Date        string `json:"date"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		Capacity    int    `json:"capacity"`
		Color       string `json:"color"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Parse date and times (similar to create)
	classDate, err := time.Parse("2006-01-02", updateData.Date)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse("15:04", updateData.StartTime)
	if err != nil {
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse("15:04", updateData.EndTime)
	if err != nil {
		http.Error(w, "Invalid end time format", http.StatusBadRequest)
		return
	}

	startDateTime := time.Date(classDate.Year(), classDate.Month(), classDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, classDate.Location())
	endDateTime := time.Date(classDate.Year(), classDate.Month(), classDate.Day(),
		endTime.Hour(), endTime.Minute(), 0, 0, classDate.Location())

	event := models.Event{
		ID:               int(classID),
		Title:            updateData.Title,
		Description:      updateData.Description,
		StartTime:        startDateTime,
		EndTime:          endDateTime,
		Location:         updateData.Location,
		ClassType:        updateData.ClassType,
		TeacherName:      updateData.TeacherName,
		Capacity:         updateData.Capacity,
		Color:            updateData.Color,
	}

	if err := AdminDB.UpdateEvent(event); err != nil {
		http.Error(w, "Could not update class", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Class updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}