package handsamarar

import (
	"kjernekraft/handsamarar/config"
	"kjernekraft/handsamarar/modules"
	"kjernekraft/models"
	"log"
	"net/http"
	"time"
)

var OsloLoc *time.Location

// ElevDashboard serves the Elev dashboard home page
func (a *App) ElevDashboard(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	settings := config.GetInstance()
	now := settings.GetCurrentTime()

	// Timar med ledig plass, i dag og i morgon.
	//
	// Bolken spurde «kva går i dag» før. Men ein time som er full er
	// ikkje eit tilbod, og i dag åleine er for kort: er klokka sju om
	// kvelden, står bolken tom kvar kveld. Han spør «kvar er det plass»
	// no, og då må i morgon vere med.
	//
	// Lista vert bygd av den same funksjonen som brotstykket nyttar, so
	// sida og ei henting midt i økta aldri kan seie kvar sitt.
	lang := GetLanguageFromRequest(r)

	ledige, err := a.AvailableSessions(int64(user.ID), lang, now)
	if err != nil {
		http.Error(w, "Could not fetch today's events", http.StatusInternalServerError)
		return
	}

	paamelde, err := a.EnrolledSessions(int64(user.ID), lang, now)
	if err != nil {
		log.Printf("paameldingar for %d: %v", user.ID, err)
	}

	// Klipp ein *kann bruke*: utgjengne kort og tome kort er ikkje noko
	// ein hev att, og eit tal som tel deim med lovar meir enn det finst.
	// An empty list is a valid answer: the user has no cards.
	var kort []models.KlippekortWithDetails
	klippAtt := 0
	if k, err := a.DB.GetUserKlippekort(int64(user.ID)); err != nil {
		log.Printf("klippekort for %d: %v", user.ID, err)
	} else {
		kort = k
		klippAtt = klargjerKlippekort(kort, now)
	}
	klippekortModul := modules.NewKlippekortModule(kort, lang)

	// Var ho nett på ein time? Krev kryss i vestibylen — ei paamelding
	// aaleine segjer berre at ho hadde tenkt seg dit.
	ferdig, feil := a.DB.NettFrammott(int64(user.ID), now, 30*time.Minute)
	if feil != nil {
		log.Printf("nett frammøtt for %d: %v", user.ID, feil)
	}

	// Aktiviteten: eit halvt år attende, dag for dag. Grensa kjem frå
	// AktivitetFrå, so båe bileti fær tal for heile tidsrommet sitt.
	const aktivitetVekor = 52 // eit heilt år, ei vika per prikk
	frå := ActivityStart(now, aktivitetVekor)
	perType, err := a.DB.AktivitetPerDagType(int64(user.ID), frå, now)
	if err != nil {
		log.Printf("aktivitet per type for %d: %v", user.ID, err)
		perType = map[string]map[string]int{}
	}
	perDag := make(map[string]int, len(perType))
	for dag, typar := range perType {
		for _, tal := range typar {
			perDag[dag] += tal
		}
	}

	// Medlemskapet stend som eitt merke her; forvaltningi ligg på
	// medlemskapssida. Feilar uppslaget, syner merket «ikkje medlem» —
	// det er ikkje sant, men det er ikkje verre enn ei tom rute, og
	// lenkja ved sida av fører til sanningi.
	medlemskap, err := a.DB.GetUserMembership(int64(user.ID))
	if err != nil {
		log.Printf("medlemskap for %d: %v", user.ID, err)
	}

	// Namnet på kortet lyt koma den same vegen som på
	// medlemskapssida: gjenom `Namn`, so eit namn administrasjonen hev
	// overstyrt gjeld båe stader. Raanamnet i basen er «Basis» der
	// sida syner «Årskort», og tvo namn på den same tingen er verre
	// enn ikkje aa syna namnet i det heile.
	// «Medlem sidan» kjem frå kontoen, ikkje frå avtalen. Sjå
	// MedlemSidan i database/svartmedlem.go.
	var medlemSidan *time.Time
	if medlemskap != nil {
		if d, feil := a.DB.MedlemSidan(int64(user.ID)); feil != nil {
			log.Printf("medlem sidan for %d: %v", user.ID, feil)
		} else if !d.IsZero() {
			medlemSidan = &d
		}
	}

	medlemsnamn := ""
	if medlemskap != nil {
		overstyrte, feil := a.DB.Produktnamn("medlemskap")
		if feil != nil {
			log.Printf("produktnamn: %v", feil)
			overstyrte = nil
		}
		medlemsnamn = Namn(overstyrte[medlemskap.ID], lang,
			MedlemskapNamn(lang, medlemskap.Membership))
	}

	helsingTittel, briefing := Heimehovudet(lang, user.Name, paamelde, klippAtt, ferdig, now)

	data := sidedata(r, SidaHeim, "Elev Dashboard", map[string]any{
		"Activity":          NewActivity(lang, perDag, perType, now, aktivitetVekor),
		"Medlemskap":        medlemskap,
		"MedlemSidan":       medlemSidan,
		"MedlemskapNamn":    medlemsnamn,
		"HelsingTittel":     helsingTittel,
		"Briefing":          briefing,
		"KlippekortModul":   klippekortModul,
		"AvailableSessions": ledige,
		"EnrolledSessions":  paamelde,
		"User":              user,
	})

	renderPage(w, r, "pages/dashboard", data)
}
