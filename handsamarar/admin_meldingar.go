package handsamarar

import (
	"net/http"
	"strconv"

	"kjernekraft/handsamarar/config"
)

// HandsamaMelding takes a notice out of the tab.
//
// It is not deleted. An empty tab should mean "nothing is waiting" and
// not "nothing has happened" — and the day an email sender arrives, it
// is the same row it looks at: `sendt` is its column, this one is the
// humans'.
func (a *App) HandsamaMelding(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("melding_id"), 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "Invalid melding_id", http.StatusBadRequest)
		return
	}

	if err := a.DB.HandsamaMelding(id, config.GetInstance().GetCurrentTime()); err != nil {
		http.Error(w, "Could not update melding", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
