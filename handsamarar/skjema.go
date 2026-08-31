package handsamarar

import (
	"net/http"
	"strconv"
)

// skjemaTal reads an integer from the form. On anything else it writes
// 400 itself and returns false, so the caller only has to return:
//
//	eventID, ok := skjemaTal(w, r, "event_id")
//	if !ok {
//		return
//	}
//
// The message names the field rather than explaining: it reaches whoever
// is calling the API, and the page forms always send numbers.
func skjemaTal(w http.ResponseWriter, r *http.Request, nykel string) (int64, bool) {
	n, err := strconv.ParseInt(r.FormValue(nykel), 10, 64)
	if err != nil {
		http.Error(w, "Ugyldig "+nykel, http.StatusBadRequest)
		return 0, false
	}
	return n, true
}
