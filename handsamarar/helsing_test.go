package handsamarar

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"kjernekraft/models"
)

// Helsinga sa to timar for seint.
//
// Tidene i databasen er klokka på veggen — «2026-08-27 16:30:00» — og
// drivaren merkjer dei UTC av di strengen ikkje ber noka sone. Helsinga
// rekna dei om til Oslo og la difor to timar på: «Sest 18:30» om ein
// time som gjekk 16:30, og ingen var påmeld noko halv sju.
//
// Prøva byggjer tidi slik drivaren gjev henne, og krev at helsinga
// seier den same klokka som lista under.
func TestHelsingaSegjerDenSameKlokkaSomTimen(t *testing.T) {
	oslo, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Skip("ingi sonefil:", err)
	}
	gamal := OsloLoc
	OsloLoc = oslo
	defer func() { OsloLoc = gamal }()

	// Slik drivaren les «2026-08-27 16:30:00»: rett klokke, feil merkelapp.
	start := time.Date(2026, 8, 27, 16, 30, 0, 0, time.UTC)
	nå := time.Date(2026, 8, 27, 9, 0, 0, 0, oslo)

	fekk := HelsingNår("nn", &models.Event{StartTime: start}, nå)

	if !strings.Contains(fekk, "16:30") {
		t.Errorf("naar-et sa %q, venta klokka 16:30", fekk)
	}
	if strings.Contains(fekk, "18:30") {
		t.Errorf("naar-et flytte timen to timar: %q", fekk)
	}
}

// Fyrenamnet budde i Helsing fyrr. Det gjer det ikkje lenger — tittelen
// ber namnet no, og setningi ber timen — so regelen fylgjer med dit han
// gjeld. «Hei, Anna Larsen» er ikkje ei helsing, det er ei innkalling.
func TestTittelenNyttarBerreFyrenamnet(t *testing.T) {
	nå := time.Date(2026, 8, 27, 9, 0, 0, 0, time.UTC)
	fekk := HelsingTittel("nn", "Ellinor Linnea Stokke", "i morgon tidleg", nå, false)
	if !strings.Contains(fekk, "Ellinor") {
		t.Errorf("tittelen sa %q, venta fyrenamnet", fekk)
	}
	if strings.Contains(fekk, "Stokke") {
		t.Errorf("tittelen nytta heile namnet: %q", fekk)
	}
}

// Ein time seinare same døgeret skal framleis lesast som «i dag», og
// ikkje skuvast yver midnatt av ei umrekning.
func TestSeinTimeIDagVertIkkjeIMorgon(t *testing.T) {
	oslo, err := time.LoadLocation("Europe/Oslo")
	if err != nil {
		t.Skip("ingi sonefil:", err)
	}
	gamal := OsloLoc
	OsloLoc = oslo
	defer func() { OsloLoc = gamal }()

	start := time.Date(2026, 8, 27, 23, 0, 0, 0, time.UTC) // 23:00 på veggen
	nå := time.Date(2026, 8, 27, 18, 0, 0, 0, oslo)

	fekk := HelsingNår("nn", &models.Event{StartTime: start}, nå)
	if !strings.Contains(fekk, "23:00") {
		t.Errorf("naar-et sa %q, venta 23:00 i dag", fekk)
	}
}

// Briefingen skal aldri segja noko rart.
//
// Ho er sett saman av tri lekkar som kvar kann falla burt, og det er
// samansetjinga — ikkje ordi — som kann ryka: eit «og» utan noko fyre
// seg, tvo mellomrom, eit punktum som stend aaleine, eller ei setning
// som endar utan punktum. Prøva teiknar den verkelege malen for kvar
// kombinasjon på alle tri maali og les etter nettupp det.
func TestBriefingenSegjerAldriNokoRart(t *testing.T) {
	// Umsetjingane fyrst: `lastMaali` leitar etter «../mål» og lyt
	// difor lesa medan me endå stend i pakkekatalogen. Etterpå gjeng
	// me eit hakk upp, av di malane vert leita etter frå rota.
	lastMaali(t)

	gamal, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(gamal)

	tm := GetTemplateManager()
	tm.ReloadTemplates()
	mal, ok := tm.GetTemplate("pages/dashboard")
	if !ok {
		t.Fatal("malen pages/dashboard vart ikkje lasta")
	}

	oslo := OsloLoc
	if oslo == nil {
		oslo = time.UTC
	}
	nå := time.Date(2026, 8, 27, 9, 0, 0, 0, oslo)
	neste := &models.Event{StartTime: time.Date(2026, 8, 28, 18, 0, 0, 0, time.UTC)}

	for _, lang := range []string{"nn", "nb", "en"} {
		for _, harTime := range []bool{true, false} {
			for _, iVeka := range []int{0, 1, 3} {
				for _, klipp := range []int{0, 1, 30} {
					var e *models.Event
					if harTime {
						e = neste
					}
					b := NyBriefing(lang, e, nå, iVeka, klipp)

					var ut bytes.Buffer
					if err := mal.ExecuteTemplate(&ut, "briefing", map[string]interface{}{
						"Lang": lang, "Briefing": b,
					}); err != nil {
						t.Fatalf("%s time=%v veka=%d klipp=%d: %v", lang, harTime, iVeka, klipp, err)
					}
					// Berre teksti: merkelappane er malen sitt, ikkje setningi si.
					tekst := strings.TrimSpace(fjernMerke(ut.String()))

					kva := fmt.Sprintf("%s time=%v veka=%d klipp=%d: %q", lang, harTime, iVeka, klipp, tekst)
					switch {
					case tekst == "":
						t.Errorf("%s — tom", kva)
					case strings.Contains(tekst, "  "):
						t.Errorf("%s — tvo mellomrom", kva)
					case strings.Contains(tekst, " ."), strings.Contains(tekst, " ,"):
						t.Errorf("%s — mellomrom fyre teiknsetjing", kva)
					case strings.Contains(tekst, ".."):
						t.Errorf("%s — dobbelt punktum", kva)
					case strings.Contains(tekst, "%!"), strings.Contains(tekst, "%s"), strings.Contains(tekst, "%d"):
						t.Errorf("%s — uutfylt formatstreng", kva)
					case strings.Contains(tekst, "<no value>"):
						t.Errorf("%s — manglande verd", kva)
					// Punktum eller utropsteikn: båe sluttar ei setning.
					// «Matta kallar på deg!» er ei oppmoding, og ei
					// oppmoding endar ikkje på punktum.
					case !strings.HasSuffix(tekst, ".") && !strings.HasSuffix(tekst, "!"):
						t.Errorf("%s — endar utan teiknsetjing", kva)
					}
					// Eit bindeord lyt ha noko fyre seg i si eigi setning.
					for _, og := range []string{". og ", ". and ", ". og har", ". and have"} {
						if strings.Contains(tekst, og) {
							t.Errorf("%s — bindeord utan noko fyre seg", kva)
						}
					}
					// Lekkane skal berre stande når dei hev noko aa segja.
					if iVeka < 2 && b.VekeNykel != "" {
						t.Errorf("%s — vekelekken stend med %d timar", kva, iVeka)
					}
					if klipp == 0 && b.KlippNykel != "" {
						t.Errorf("%s — klippelekken stend med null klipp", kva)
					}
					if klipp > 0 && !strings.Contains(tekst, fmt.Sprint(klipp)) {
						t.Errorf("%s — klippetalet mangla", kva)
					}
				}
			}
		}
	}
}

// fjernMerke tek burt HTML-merkelappane so prøva les setningi og ikkje malen.
func fjernMerke(s string) string {
	var b strings.Builder
	inne := false
	for _, r := range s {
		switch {
		case r == '<':
			inne = true
		case r == '>':
			inne = false
		case !inne:
			b.WriteRune(r)
		}
	}
	// Malen kann leggja att linjeskift; setningi hev berre mellomrom.
	return strings.Join(strings.Fields(b.String()), " ")
}
