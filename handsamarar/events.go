package handsamarar

import (
	"encoding/json"
	"kjernekraft/models"
	"net/http"
)

// CreateEvent handles creating new events
func (a *App) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid event data", http.StatusBadRequest)
		return
	}

	eventID, err := a.DB.CreateEvent(event)
	if err != nil {
		http.Error(w, "Could not create event", http.StatusInternalServerError)
		return
	}

	event.ID = int(eventID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

// GetAllEvents returns all events as JSON
//
// Ruta er innlogga, so ho svarar den som spør — og ei privat PT-økt er
// ikkje hans um ho ikkje er sett av til honom. Fyrr gav ho alt til alle.
func (a *App) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)
	events, err := a.DB.EventsSynlegeFor(int64(user.ID))
	if err != nil {
		http.Error(w, "Could not fetch events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// EventSignup handles user signup for events
func (a *App) EventSignup(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	eventID, ok := skjemaTal(w, r, "event_id")
	if !ok {
		return
	}

	// Get event details to check timing restrictions
	event, err := a.DB.GetEventByID(eventID)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Check if event is within 2 hours. The clock is the house's, not the
	// machine's: a window the studio can move in development has to be
	// movable here too, or the two disagree about what "two hours" means.
	now := a.Nå()
	if event.StartTime.Sub(now).Hours() < 2 {
		http.Error(w, "Cannot sign up for classes within 2 hours of start time", http.StatusBadRequest)
		return
	}

	// Sign up user for event
	err = a.DB.SignupUserForEvent(int64(user.ID), eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully signed up for event"))
}

// EventCancelSignup handles user cancellation of event signup
func (a *App) EventCancelSignup(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	eventID, ok := skjemaTal(w, r, "event_id")
	if !ok {
		return
	}

	// Get event details to check timing restrictions
	event, err := a.DB.GetEventByID(eventID)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Check if event is within 2 hours. The clock is the house's, not the
	// machine's: a window the studio can move in development has to be
	// movable here too, or the two disagree about what "two hours" means.
	now := a.Nå()
	if event.StartTime.Sub(now).Hours() < 2 {
		http.Error(w, "Cannot cancel signup for classes within 2 hours of start time", http.StatusBadRequest)
		return
	}

	// Ei privat økt gjeng ein annan veg ut. Ho er tinga og klippa når
	// ho vart sett upp (BokPrivatTime), og læraren hev sett av tidi si —
	// so klippet lyt attende, og nokon lyt faa vita det. Ei vanleg
	// paamelding er berre ei paamelding: ho gjeng, og det er heile saki.
	privat, err := a.DB.ErPrivatTime(eventID)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}
	if privat {
		if err := a.DB.AvlysPrivatTime(eventID, int64(user.ID), now); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Successfully cancelled signup"))
		return
	}

	// Cancel user signup for event
	err = a.DB.CancelUserSignupForEvent(int64(user.ID), eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully cancelled signup for event"))
}
