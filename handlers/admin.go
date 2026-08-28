package handlers

import (
	"kjernekraft/database"
	"kjernekraft/handlers/config"
	"kjernekraft/handlers/modules"
	"log"
	"net/http"
)

var AdminDB *database.Database

func AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.
	// Ligg denne handsamaren nokon gong utanfor den gruppa, er
	// administrasjonen open att.

	events, err := AdminDB.GetAllEvents()
	if err != nil {
		log.Printf("admin: kunde ikkje henta timar: %v", err)
		http.Error(w, "Kunne ikke hente events", http.StatusInternalServerError)
		return
	}

	freezeRequests, err := AdminDB.GetPendingFreezeRequests()
	if err != nil {
		http.Error(w, "Kunne ikke hente frysingsforespørsler", http.StatusInternalServerError)
		return
	}

	memberships, err := AdminDB.GetAllMemberships()
	if err != nil {
		http.Error(w, "Kunne ikke hente medlemskap", http.StatusInternalServerError)
		return
	}

	// Same veg som alle hine sidone. Denne eine las berre ?lang= og
	// fall attende paa "nb", so administrasjonen stod på bokmaal same
	// kva brukaren hadde valt — og fanone og korti på same skjermen
	// kunde koma på kvar sitt maal.
	lang := GetLanguageFromRequest(r)

	folk, err := AdminDB.FolkOversyn()
	if err != nil {
		log.Printf("folkeoversyn: %v", err)
		http.Error(w, "Kunne ikke hente brukere", http.StatusInternalServerError)
		return
	}

	// FolkOversyn er både medlemslista og kjelda til PT-veljaren. Før
	// vart alle brukarar henta ein gong til berre for den vesle veljaren;
	// GetAllUsers slo dessutan opp løyve éin gong per brukar.
	statsModule, err := modules.NewAdminStatsModule(len(folk), len(events), len(freezeRequests), lang)
	if err != nil {
		log.Printf("Error creating admin stats module: %v", err)
		http.Error(w, "Error creating admin module", http.StatusInternalServerError)
		return
	}

	rooms, err := AdminDB.GetRooms()
	if err != nil {
		log.Printf("kunde ikkje henta rom: %v", err)
	}

	// Veljarane paa denne sida — ny time, vikarfeltet — les løyvi.
	// Datalista `laerarar` stod tom fyrr: ho las $.Teachers, og den
	// nykelen vart aldri sett paa administrasjonssida i det heile.
	laerarar, err := AdminDB.LaerarNamn()
	if err != nil {
		log.Printf("kunde ikkje henta lærarane: %v", err)
	}

	// Gruppone ein time kann vera open for.
	grupper, err := AdminDB.Grupper()
	if err != nil {
		log.Printf("grupper: %v", err)
	}

	// Krav um studentrabatt som ventar paa svar. Dei høyrer heime i
	// meldingsfana attmed frysingane: baae er noko nokon hev bede um og
	// som ingen ting hender med fyrr studioet svarar.
	rabattkrav, err := AdminDB.VentandeRabattkrav()
	if err != nil {
		log.Printf("rabattkrav: %v", err)
	}

	// Seriane vert rekna ein gong: baade lista og vali sikti tilbyd
	// kjem av deim, og tvo utrekningar av det same kann skilja lag.
	seriar := GrupperTimar(events, config.GetInstance().GetCurrentTime())

	data := map[string]interface{}{
		"Rooms":          rooms,
		"Folk":           folk,
		"Teachers":       laerarar,
		"Title":          t(lang, "admin.title"),
		"Events":         events,
		"Grupper":        grupper,
		"Rabattkrav":     rabattkrav,
		"Timereglar":     seriar,
		"Siktval":        SiktvalFor(seriar),
		"FreezeRequests": freezeRequests,
		"Memberships":    memberships,
		"Stats":          statsModule,
		"Lang":           lang,
		"CSRFToken":      CSRFToken(r),
		"IsAdmin":        sessionIsAdmin(r),
		"UserName":       sessionUserName(r),
		"CurrentPage":    "admin",
		"ExternalCSS":    []string{},
	}

	// Use template manager instead of inline template
	tm := GetTemplateManager()
	tmpl, exists := tm.GetTemplate("pages/admin")
	if !exists {
		// Try to reload templates in case they weren't loaded
		tm.ReloadTemplates()
		tmpl, exists = tm.GetTemplate("pages/admin")
		if !exists {
			log.Printf("Available templates: %v", tm.GetAvailableTemplates())
			http.Error(w, "Admin template not found", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("Error executing admin template: %v", err)
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

// Stub functions for API endpoints - these need to be implemented
func GetUsersAPIHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func ApproveFreezeRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func RejectFreezeRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
