package handlers

import (
	"html/template"
	"kjernekraft/database"
	"net/http"
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
