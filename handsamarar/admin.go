package handsamarar

import (
	"encoding/json"
	"kjernekraft/handsamarar/config"
	"kjernekraft/handsamarar/modules"
	"log"
	"net/http"
	"strconv"
)

func (a *App) AdminPage(w http.ResponseWriter, r *http.Request) {
	// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.
	// Ligg denne handsamaren nokon gong utanfor den gruppa, er
	// administrasjonen open att.

	events, err := a.DB.GetAllEvents()
	if err != nil {
		log.Printf("admin: kunde ikkje henta timar: %v", err)
		http.Error(w, "Kunne ikke hente events", http.StatusInternalServerError)
		return
	}

	freezeRequests, err := a.DB.GetPendingFreezeRequests()
	if err != nil {
		http.Error(w, "Kunne ikke hente frysingsforespørsler", http.StatusInternalServerError)
		return
	}

	memberships, err := a.DB.GetAllMemberships()
	if err != nil {
		http.Error(w, "Kunne ikke hente medlemskap", http.StatusInternalServerError)
		return
	}

	// Same veg som alle hine sidone. Denne eine las berre ?lang= og
	// fall attende på "nb", so administrasjonen stod på bokmål same
	// kva brukaren hadde valt — og fanone og korti på same skjermen
	// kunde koma på kvar sitt mål.
	lang := GetLanguageFromRequest(r)

	folk, err := a.DB.FolkOversyn()
	if err != nil {
		log.Printf("folkeoversyn: %v", err)
		http.Error(w, "Kunne ikke hente brukere", http.StatusInternalServerError)
		return
	}

	// FolkOversyn er både medlemslista og kjelda til PT-veljaren. Før
	// vart alle brukarar henta ein gong til berre for den vesle veljaren;
	// GetAllUsers slo dessutan opp løyve éin gong per brukar.
	statsModule := modules.NewAdminStatsModule(len(folk), len(events), len(freezeRequests), lang)

	rooms, err := a.DB.GetRooms()
	if err != nil {
		log.Printf("kunde ikkje henta rom: %v", err)
	}

	// The selects on this page — new class, the substitute field — read the
	// 	// permissions. The `laerarar` datalist stood empty: it read $.Teachers,
	// 	// and that key was never set on the admin page at all.
	laerarar, err := a.DB.LærarNamn()
	if err != nil {
		log.Printf("kunde ikkje henta lærarane: %v", err)
	}

	// The kinds already in use. The field is free text, so without a list to
	// 	// pick from "Yoga" and "yoga" become two kinds — and then they carry
	// 	// different wings in the list.
	slagsortar, err := a.DB.Slagsortar()
	if err != nil {
		log.Printf("kunde ikkje henta slagi: %v", err)
	}

	// The groups a class can be open to.
	grupper, err := a.DB.Grupper()
	if err != nil {
		log.Printf("grupper: %v", err)
	}

	// Student-discount claims waiting for an answer. They belong in the notices
	// 	// tab beside the freezes: both are something somebody has asked for and
	// 	// nothing happens with until the studio answers.
	rabattkrav, err := a.DB.VentandeRabattkrav()
	if err != nil {
		log.Printf("rabattkrav: %v", err)
	}

	// The series are computed once: both the list and the options the filter
	// 	// offers come from them, and two computations of the same thing can
	// 	// drift apart.
	seriar := GrupperTimar(events, config.GetInstance().GetCurrentTime())

	// Meldingane som ventar. Ein feil her skal ikkje taka heile
	// administrasjonssida: fana stend tom, og resten verkar.
	meldingar, err := a.DB.VentandeMeldingar()
	if err != nil {
		log.Printf("meldingar: %v", err)
	}

	// Fanone. Nyklarne stod i malen fyrr — fem `data-bolk` skrivne for
	// hand — og tenaren visste ikkje um deim. No er dei her, av di det er
	// tenaren som avgjer kva bolk som vert teikna, og ein nykel han ikkje
	// kjenner er ei lenkja som fell attende paa fyrste fana.
	//
	// Prisrekkja hev sitt eige namn i adressa (`prisfane`) av di ho stend
	// inne i fana «Prisar». Den ytre rekkja nullstiller henne: er ein
	// fyrst gjengen ut or prisane, hev det ingi meining aa slæpa med seg
	// kva prisfana ein stod paa.
	fanor := fanerekkje(r, "faneark-admin", "fane", t(lang, "admin.title"), []Tab{
		{Key: "meldingar", Name: t(lang, "admin.tab_messages")},
		{Key: "timeplan", Name: t(lang, "admin.tab_schedule")},
		{Key: "prisar", Name: t(lang, "admin.tab_prices")},
		{Key: "folk", Name: t(lang, "admin.tab_people")},
		{Key: "innstillingar", Name: t(lang, "admin.tab_settings")},
	}, "prisfane")

	prisfanor := fanerekkje(r, "faneark-prisar", "prisfane", t(lang, "admin.tab_prices"), []Tab{
		{Key: "medlemskap", Name: t(lang, "admin.tab_membership")},
		{Key: "klippekort", Name: t(lang, "admin.tab_klippekort")},
		{Key: "reglar", Name: t(lang, "admin.tab_rules")},
	})

	data := sidedata(r, SidaAdmin, "Administrasjon", map[string]any{
		"Faner":      fanor,
		"Prisfaner":  prisfanor,
		"Rooms":      rooms,
		"Folk":       folk,
		"Teachers":   laerarar,
		"Slagsortar": slagsortar,
		"Title":      t(lang, "admin.title"),
		"Events":     events,
		"Grupper":    grupper,
		"Rabattkrav": rabattkrav,
		"Timereglar": seriar,
		"Siktval":    SiktvalFor(seriar),
		// Vekefelti i timebolken tel i dei same ISO-vikone som
		// timeplanen. Talet kjem herifrå og ikkje frå lesaren:
		// klokka i innstillingane er den huset held seg til.
		"VekeNo":         veketalNo(),
		"VikorIAaret":    VikorIAaret(config.GetInstance().GetCurrentTime()),
		"FreezeRequests": freezeRequests,
		"Meldingar":      meldingar,
		"Memberships":    memberships,
		"Stats":          statsModule,
	})

	renderPage(w, r, "pages/admin", data)
}

// Frysingi vert avgjord her. Basen hev gjort arbeidet heile tidi —
// ApproveFreezeRequest set `frozen_at` so utlaupet stoggar aa telja, og
// RejectFreezeRequest set medlemskapet attende til 'active' — men
// rutorne svara 501, so medlemen vart staaande i 'freeze_requested' og
// kom seg korkje fram eller attende. Administrasjonen talde deim i
// briefingen og baud fram ein knapp som ikkje gjorde noko.
//
// Brukaren kjem i spurnadsstrengen av di det er admin som handlar paa
// vegner av ein annan; dei hine frysingsrutorne les honom or økti.
func (a *App) ApproveFreezeRequest(w http.ResponseWriter, r *http.Request) {
	frysingssvar(w, r, a.DB.ApproveFreezeRequest, "Frysingi er godkjend.")
}

func (a *App) RejectFreezeRequest(w http.ResponseWriter, r *http.Request) {
	frysingssvar(w, r, a.DB.RejectFreezeRequest, "Frysingi er avvist.")
}

// frysingssvar les brukaren, gjer handlingi og svarar. Dei tvo rutorne
// skil seg i eitt kall og éi setning, so dei deler resten.
func frysingssvar(w http.ResponseWriter, r *http.Request, gjer func(int64) error, kvittering string) {
	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "Ugyldig brukar", http.StatusBadRequest)
		return
	}
	if err := gjer(userID); err != nil {
		log.Printf("frysing for brukar %d: %v", userID, err)
		http.Error(w, "Kunde ikkje endra frysingi", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true, "message": kvittering})
}
