package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"kjernekraft/models"
)

// RoomConflictHandler segjer um rommet alt er uppteke.
//
// Konflikten skal segjast *fyre* ein trykkjer. Ein tenar som svarar 400
// etter at du hev fylt ut alt og trykt «Legg til» hev alt kasta bort
// tidi di — og verre: han lærer deg at det er normalt aa prøva seg fram.
func RoomConflictHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	romID, err := strconv.ParseInt(q.Get("room_id"), 10, 64)
	if err != nil {
		http.Error(w, "ugyldig room_id", http.StatusBadRequest)
		return
	}

	start, err := time.ParseInLocation("2006-01-02 15:04", q.Get("date")+" "+q.Get("start"), OsloLoc)
	if err != nil {
		http.Error(w, "ugyldig starttid", http.StatusBadRequest)
		return
	}
	slutt, err := time.ParseInLocation("2006-01-02 15:04", q.Get("date")+" "+q.Get("end"), OsloLoc)
	if err != nil {
		http.Error(w, "ugyldig slutttid", http.StatusBadRequest)
		return
	}
	// Ein time som kryssar midnatt endar dagen etter.
	if !slutt.After(start) {
		slutt = slutt.AddDate(0, 0, 1)
	}

	kollisjon, err := AdminDB.RoomConflict(romID, start, slutt)
	if err != nil {
		http.Error(w, "kunde ikkje sjaa etter konflikt", http.StatusInternalServerError)
		return
	}

	svar := map[string]interface{}{"conflict": kollisjon != nil}
	if kollisjon != nil {
		lang := GetLanguageFromRequest(r)
		svar["message"] = fmt.Sprintf("%s %s–%s: %s (%s)",
			t(lang, "admin.room_busy"),
			veggklokka(kollisjon.StartTime).Format("15:04"),
			veggklokka(kollisjon.EndTime).Format("15:04"),
			kollisjon.Title, kollisjon.TeacherName)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(svar)
}

// Konfliktsvaret som type, so det ikkje er eit kart med strengnyklar
// spreidd i tvo filer.
var _ = models.Event{}
