package handlers

import (
	"encoding/json"
	"kjernekraft/database"
	"net/http"
	"strconv"
)

// SettLoyveHandler slær eit løyve av eller paa for ein brukar. Han ligg
// bak RequireAdmin i rutaren.
//
// Svaret ber lærarlista med seg attende. Eit løyve som er slegi paa her
// skal vera valbar i veljarane paa den same skjermen med ein gong; utan
// lista hadde ho lege der utan aa syna seg fyrr sida vart lasta paa
// nytt, og daa kjennest knappen som um han ikkje gjorde noko.
func SettLoyveHandler(w http.ResponseWriter, r *http.Request) {
	brukarID, err := strconv.ParseInt(r.URL.Query().Get("brukar"), 10, 64)
	if err != nil {
		http.Error(w, "Ugild brukar", http.StatusBadRequest)
		return
	}

	loyve := r.URL.Query().Get("loyve")
	if !database.LoyveFinst(loyve) {
		http.Error(w, "Ukjend loyve", http.StatusBadRequest)
		return
	}
	paa := r.URL.Query().Get("paa") == "1"

	// Den som tek administratorrolla si eigi, laaser seg ute or rommet
	// han stend i — og det er ingen veg attende gjenom flata. Løyvet
	// kann framleis takast av *ein annan* administrator.
	if loyve == database.LoyveAdmin && !paa {
		if sjolv := GetUserFromSession(r); sjolv != nil && int64(sjolv.ID) == brukarID {
			http.Error(w, "Du kann ikkje taka administratorrolla di eigi", http.StatusConflict)
			return
		}
	}

	if err := DB.SettLoyve(brukarID, loyve, paa); err != nil {
		http.Error(w, "Kunde ikkje setja loyve", http.StatusInternalServerError)
		return
	}

	laerarar, err := DB.LaerarNamn()
	if err != nil {
		http.Error(w, "Kunde ikkje henta lærarane", http.StatusInternalServerError)
		return
	}
	if laerarar == nil {
		laerarar = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"laerarar": laerarar})
}
