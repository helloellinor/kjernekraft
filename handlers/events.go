package handlers

import (
	"encoding/json"
	"html/template"
	"kjernekraft/database"
	"kjernekraft/models"
	"net/http"
	"strconv"
	"time"
)

type EventHandler struct {
	DB       *database.Database
	Template *template.Template
}

func NewEventHandler(db *database.Database) *EventHandler {
	return &EventHandler{DB: db}
}

func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	location := r.URL.Query().Get("location")

	events, err := h.DB.GetFilteredEvents(startDate, endDate, location)
	if err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}
	h.Template.Execute(w, events)
}

// CreateEventHandler handles creating new events
func CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid event data", http.StatusBadRequest)
		return
	}

	eventID, err := DB.CreateEvent(event)
	if err != nil {
		http.Error(w, "Could not create event", http.StatusInternalServerError)
		return
	}

	event.ID = int(eventID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

// GetAllEventsHandler returns all events as JSON
//
// Ruta er innlogga, so ho svarar den som spør — og ei privat PT-økt er
// ikkje hans um ho ikkje er sett av til honom. Fyrr gav ho alt til alle.
func GetAllEventsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	events, err := DB.EventsSynlegeFor(int64(user.ID))
	if err != nil {
		http.Error(w, "Could not fetch events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// EventSignupHandler handles user signup for events
func EventSignupHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	eventIDStr := r.FormValue("event_id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Get event details to check timing restrictions
	event, err := DB.GetEventByID(eventID)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Check if event is within 2 hours
	now := time.Now()
	if event.StartTime.Sub(now).Hours() < 2 {
		http.Error(w, "Cannot sign up for classes within 2 hours of start time", http.StatusBadRequest)
		return
	}

	// Sign up user for event
	err = DB.SignupUserForEvent(int64(user.ID), eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully signed up for event"))
}

// EventCancelSignupHandler handles user cancellation of event signup
func EventCancelSignupHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	eventIDStr := r.FormValue("event_id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Get event details to check timing restrictions
	event, err := DB.GetEventByID(eventID)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Check if event is within 2 hours
	now := time.Now()
	if event.StartTime.Sub(now).Hours() < 2 {
		http.Error(w, "Cannot cancel signup for classes within 2 hours of start time", http.StatusBadRequest)
		return
	}

	// Ei privat økt gjeng ein annan veg ut. Ho er tinga og klippa naar
	// ho vart sett upp (BokPrivatTime), og læraren hev sett av tidi si —
	// so klippet lyt attende, og nokon lyt faa vita det. Ei vanleg
	// paamelding er berre ei paamelding: ho gjeng, og det er heile saki.
	privat, err := DB.ErPrivatTime(eventID)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}
	if privat {
		if err := DB.AvlysPrivatTime(eventID, int64(user.ID), now); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Successfully cancelled signup"))
		return
	}

	// Cancel user signup for event
	err = DB.CancelUserSignupForEvent(int64(user.ID), eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully cancelled signup for event"))
}
