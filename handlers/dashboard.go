package handlers

import (
	"kjernekraft/handlers/config"
	"kjernekraft/models"
	"log"
	"net/http"
	"time"
)

var OsloLoc *time.Location

// ElevDashboardHandler serves the Elev dashboard home page
func ElevDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user := GetUserFromSession(r)
	if user == nil {
		http.Redirect(w, r, "/innlogging", http.StatusTemporaryRedirect)
		return
	}

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

	ledige, err := AvailableSessions(int64(user.ID), lang, now)
	if err != nil {
		http.Error(w, "Could not fetch today's events", http.StatusInternalServerError)
		return
	}

	paamelde, err := EnrolledSessions(int64(user.ID), lang, now)
	if err != nil {
		log.Printf("paameldingar for %d: %v", user.ID, err)
	}

	// Den fyrste timen han hev meldt seg paa. Han ber helsingi.
	var neste *models.Event
	if len(paamelde) > 0 {
		neste = &paamelde[0].Event
	}

	// Kor mange timar ein stend paa *denne veka*. Lista gjev alt som
	// kjem, so ho lyt klyppast ved vekeskiftet: «paameld 4 timar denne
	// veka» um tri av deim gjeng neste maanad er ikkje sant.
	//
	// Veggklokka, ikkje umrekning — same grunnen som i helsinga.
	maandag := VikeMaandag(now, 0)
	nesteMaandag := maandag.AddDate(0, 0, 7)
	iVeka := 0
	for _, s := range paamelde {
		d := veggklokka(s.Event.StartTime)
		if !d.Before(maandag) && d.Before(nesteMaandag) {
			iVeka++
		}
	}

	// Klipp ein *kann bruke*: utgjengne kort og tome kort er ikkje noko
	// ein hev att, og eit tal som tel deim med lovar meir enn det finst.
	klippAtt := 0
	if kort, err := DB.GetUserKlippekort(int64(user.ID)); err != nil {
		log.Printf("klippekort for %d: %v", user.ID, err)
	} else {
		for _, k := range kort {
			if k.RemainingKlipp > 0 && k.ExpiryDate.After(now) {
				klippAtt += k.RemainingKlipp
			}
		}
	}

	naar := HelsingNaar(lang, neste, now)

	// Var ho nett paa ein time? Krev kryss i vestibylen — ei paamelding
	// aaleine segjer berre at ho hadde tenkt seg dit.
	ferdig, feil := DB.NettFrammott(int64(user.ID), now, 30*time.Minute)
	if feil != nil {
		log.Printf("nett frammøtt for %d: %v", user.ID, feil)
	}

	// Aktiviteten: eit halvt aar attende, dag for dag. Grensa kjem fraa
	// AktivitetFraa, so baae bileti fær tal for heile tidsrommet sitt.
	const aktivitetVekor = 52 // eit heilt aar, ei vika per prikk
	fraa := ActivityStart(now, aktivitetVekor)
	perDag, err := DB.AktivitetPerDag(int64(user.ID), fraa, now)
	perType, feil := DB.AktivitetPerDagType(int64(user.ID), fraa, now)
	if feil != nil {
		log.Printf("aktivitet per type for %d: %v", user.ID, feil)
		perType = map[string]map[string]int{}
	}
	if err != nil {
		log.Printf("aktivitet for %d: %v", user.ID, err)
		perDag = map[string]int{}
	}

	// Medlemskapet stend som eitt merke her; forvaltningi ligg paa
	// medlemskapssida. Feilar uppslaget, syner merket «ikkje medlem» —
	// det er ikkje sant, men det er ikkje verre enn ei tom rute, og
	// lenkja ved sida av fører til sanningi.
	medlemskap, err := DB.GetUserMembership(int64(user.ID))
	if err != nil {
		log.Printf("medlemskap for %d: %v", user.ID, err)
	}

	// Namnet paa kortet lyt koma den same vegen som paa
	// medlemskapssida: gjenom `Namn`, so eit namn administrasjonen hev
	// overstyrt gjeld baae stader. Raanamnet i basen er «Basis» der
	// sida syner «Årskort», og tvo namn paa den same tingen er verre
	// enn ikkje aa syna namnet i det heile.
	medlemsnamn := ""
	if medlemskap != nil {
		overstyrte, feil := DB.Produktnamn("medlemskap")
		if feil != nil {
			log.Printf("produktnamn: %v", feil)
			overstyrte = nil
		}
		medlemsnamn = Namn(overstyrte[medlemskap.Membership.ID], lang,
			MedlemskapNamn(lang, medlemskap.Membership))
	}

	data := map[string]interface{}{
		"Activity":          NewActivity(lang, perDag, perType, now, aktivitetVekor),
		"Medlemskap":        medlemskap,
		"MedlemskapNamn":    medlemsnamn,
		"HelsingTittel":     HelsingTittel(lang, user.Name, naar, now, ferdig != nil),
		"Briefing":          NyBriefing(lang, neste, now, iVeka, klippAtt),
		"Title":             "Elev Dashboard",
		"AvailableSessions": ledige,
		"EnrolledSessions":  paamelde,
		"ExternalCSS":       []string{},
		"CurrentPage":       "hjem",
		"UserName":          user.Name,
		"User":              user,
		"Lang":              lang,
		"CSRFToken":         CSRFToken(r),
		"IsAdmin":           sessionIsAdmin(r),
	}

	// Use the new template system
	tm := GetTemplateManager()
	if tmpl, exists := tm.GetTemplate("pages/dashboard"); exists {
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
		return
	}

	// If template doesn't exist, return error
	http.Error(w, "Template not found", http.StatusInternalServerError)
}
