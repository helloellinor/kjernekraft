// hent-timeplan hentar timeplanen fraa Yogo og legg honom inn i basen.
//
// Studioet bokar i Yogo i dag. Dette er verktyet som flytter planen
// hit — ikkje ein synkronisering som gjeng jamt, men ei henting ein
// gjer nokre gonger medan ein flytter inn.
//
//	go run ./cmd/hent-timeplan -veker 4            # syner kva som vilde hendt
//	go run ./cmd/hent-timeplan -veker 4 -skriv     # gjer det
//
// Han skriv ingen ting utan `-skriv`. Ein import er ikkje noko ein kann
// gjera um att — timane fær id-ar, og paameldingar heng i deim — so
// fyrste gongen skal vera ei liste ein les, ikkje ei endring ein
// uppdagar (ARKET §7: aatvaringi stend fyre trykket).
//
// Han kann køyrast fleire gonger. Eit utslag som alt stend i basen vert
// hoppa yver, og ei rekkje han kjenner att fær dei nye timane sine lagde
// til den rekkja som finst — ikkje ei ny med det same namnet.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"kjernekraft/database"
	"kjernekraft/models"
	"kjernekraft/yogo"
)

func main() {
	log.SetFlags(0)

	veker := flag.Int("veker", 4, "kor mange heile vikor fram fraa i dag")
	attende := flag.Int("attende", 0, "kor mange vikor attende fraa i dag; 26 er eit halvt aar")
	skriv := flag.Bool("skriv", false, "skriv til basen; utan denne vert ingen ting endra")
	medAvlyste := flag.Bool("med-avlyste", false, "tak med avlyste timar")
	flag.Parse()

	if err := køyr(*attende, *veker, *skriv, *medAvlyste); err != nil {
		log.Fatalf("\nstogga: %v", err)
	}
}

func køyr(attende, veker int, skriv, medAvlyste bool) error {
	// ---- Yogo ----
	k, err := yogo.Ny()
	if err != nil {
		return err
	}
	if veker < 1 {
		return fmt.Errorf("-veker lyt vera minst 1")
	}
	if attende < 0 {
		return fmt.Errorf("-attende kann ikkje vera negativ")
	}

	// Spennet er tvo vegar ut fraa i dag.
	//
	// Framover er timeplanen; attende er *historia*, og ho er ikkje pynt:
	// utan timar som hev vore finst det ingen ting aa merkja nokon
	// frammøtt paa, og «kor mange timar i aar» paa folkekortet er null
	// for alle. Eit halvt aar attende er 26 vikor.
	no := time.Now().In(k.Sone)
	fraa := no.AddDate(0, 0, -7*attende)
	til := no.AddDate(0, 0, 7*veker-1)

	ctx, stopp := context.WithTimeout(context.Background(), 3*time.Minute)
	defer stopp()

	fmt.Printf("Hentar %s – %s (%d vikor attende, %d fram) …\n",
		fraa.Format("2.1.2006"), til.Format("2.1.2006"), attende, veker)
	komande, err := k.Timar(ctx, fraa, til, yogo.Val{MedAvlyste: medAvlyste})
	if err != nil {
		return err
	}
	if len(komande) == 0 {
		fmt.Println("Yogo hadde ingen timar i det spennet. Ingen ting aa gjera.")
		return nil
	}
	fmt.Printf("Yogo gav %d timar.\n\n", len(komande))

	// ---- basen ----
	kopling, err := database.Connect()
	if err != nil {
		return fmt.Errorf("fekk ikkje opna basen: %w", err)
	}
	defer kopling.Close()
	if err := database.Migrate(kopling); err != nil {
		return fmt.Errorf("migrering: %w", err)
	}
	db := &database.Database{Conn: kopling}

	// Romi. Yogo hev sine, huset hev sine, og dei møtest paa namnet —
	// «Salen» og «Reformer» stend i baae. Eit rom me ikkje kjenner att
	// vert ikkje gjeta paa: timen fær `room_id = 0`, som tyder «ikkje
	// noko rom», og namnet stend att i `location`.
	rom, err := db.GetRooms()
	if err != nil {
		return fmt.Errorf("kunde ikkje henta romi: %w", err)
	}
	romNr := map[string]int{}
	for _, r := range rom {
		romNr[strings.ToLower(strings.TrimSpace(r.Name))] = r.ID
	}

	var utanRom []string
	for i := range komande {
		if id, ok := romNr[strings.ToLower(komande[i].RoomName)]; ok {
			komande[i].RoomID = id
		} else if komande[i].RoomName != "" {
			utanRom = append(utanRom, komande[i].RoomName)
		}
	}

	// ---- kor mange rommet held ----
	//
	// Yogo hev ingi romkapasitet. Han hev `seats` per time, og det er
	// tvo ting blanda i eitt: kor mange rommet *held*, og kor mange
	// studioet slepp inn *denne* gongen.
	//
	// Det eine kann lesast av det andre: eit rom kann ikkje halda færre
	// enn den største timen som hev gjenge i det. Difor er det høgste
	// talet me ser rommet sitt tal, og alt under er eit val nokon hev
	// teke for den timen.
	//
	// Berre uppyver. Ser me ingen time som fyller rommet, tyder det
	// ikkje at rommet er mindre — berre at ingen fylte det i det spennet
	// me spurde um. Aa setja talet *ned* paa det grunnlaget hadde gjort
	// timar fulle som ikkje er det.
	maks, fordeling := maksPerRom(komande)

	romPlassar := map[int]int{}
	fmt.Println("ROMMET      HELD  YOGO SITT STØRSTE  FORDELING")
	fmt.Println(strings.Repeat("─", 78))
	type romlyft struct {
		id, fraa, til int
		namn          string
	}
	var lyft []romlyft
	for _, r := range rom {
		nytt := r.Capacity
		if maks[r.ID] > nytt {
			nytt = maks[r.ID]
			lyft = append(lyft, romlyft{r.ID, r.Capacity, nytt, r.Name})
		}
		romPlassar[r.ID] = nytt

		merke := ""
		if maks[r.ID] > r.Capacity {
			merke = fmt.Sprintf("  ← vert %d", nytt)
		} else if maks[r.ID] > 0 && maks[r.ID] < r.Capacity {
			merke = "  ← ingen time fylte rommet; talet stend"
		}
		fmt.Printf("%-12s %4d  %17d  %-24s%s\n",
			r.Name, r.Capacity, maks[r.ID], fordelingstekst(fordeling[r.ID]), merke)
	}
	fmt.Println()

	// Timen ber berre si eigi kapasitet naar ho *skil seg* fraa rommet.
	//
	// Er dei like, skal talet vera null — «rommet raar» — og daa fylgjer
	// timen med um rommet ein gong vert eit anna. Skreiv me talet inn
	// paa kvar time, var det attthundrad stader aa retta den dagen
	// Salen fekk tvo matter til, og timane hadde stade att paa det gamle
	// talet utan at nokon saag det.
	eigne := latArva(komande, romPlassar)

	// ---- det som alt stend her ----
	gamle, err := db.GetAllEvents()
	if err != nil {
		return fmt.Errorf("kunde ikkje lesa dei timane som alt finst: %w", err)
	}
	finst := map[string]bool{}    // utslag som alt er lagde inn
	serieAv := map[string]int64{} // rekkje -> serie-id i basen
	for _, e := range gamle {
		finst[yogo.UtslagNykel(e)] = true
		if e.SerieID != 0 {
			if _, teke := serieAv[yogo.SerieNykel(e)]; !teke {
				serieAv[yogo.SerieNykel(e)] = e.SerieID
			}
		}
	}

	// ---- planen ----
	seriar := yogo.GrupperISeriar(komande)

	var (
		nye, utvida, uroerde int
		nyeTimar, hoppa      int
	)
	fmt.Println("REKKJA                                                  NYE  FINST  SLAG")
	fmt.Println(strings.Repeat("─", 78))

	type arbeid struct {
		serie   yogo.Serie
		nye     []models.Event
		serieID int64 // 0 = ny rekkje
	}
	var kø []arbeid

	for _, s := range seriar {
		var friske []models.Event
		var kjende int
		for _, t := range s.Timar {
			if finst[yogo.UtslagNykel(t)] {
				kjende++
				continue
			}
			friske = append(friske, t)
		}
		hoppa += kjende
		nyeTimar += len(friske)

		f := s.Fyrste()
		slag := f.ClassType
		if slag == "" {
			slag = "— ukjend —"
		}
		fmt.Printf("%-3s %-5s %-28s %-16s %4d %6d  %s\n",
			yogo.Vekedagsnamn(f.StartTime.Weekday()),
			f.StartTime.Format("15:04"),
			stutt(f.Title, 28), stutt(f.TeacherName, 16),
			len(friske), kjende, slag)

		if len(friske) == 0 {
			uroerde++
			continue
		}
		id := serieAv[s.Nykel]
		if id == 0 {
			nye++
		} else {
			utvida++
		}
		kø = append(kø, arbeid{serie: s, nye: friske, serieID: id})
	}

	fmt.Println(strings.Repeat("─", 78))
	fmt.Printf("%d rekkjor: %d nye, %d som vert lengre, %d som alt er komplette.\n",
		len(seriar), nye, utvida, uroerde)
	fmt.Printf("%d timar aa leggja inn, %d stend her fraa fyrr.\n", nyeTimar, hoppa)
	fmt.Printf("%d av dei %d timane ber si eigi kapasitet; resten arvar rommet sitt tal.\n",
		eigne, len(komande))

	// ---- det me ikkje visste ----
	var namn []string
	for _, e := range komande {
		namn = append(namn, e.Title)
	}
	if ukjende := yogo.UkjendeSlag(namn); len(ukjende) > 0 {
		sort.Strings(ukjende)
		fmt.Printf("\nUtan slag (vengen vert graa) — legg deim i `slagtabellen` i yogo/slag.go:\n")
		for _, n := range ukjende {
			fmt.Printf("  %q\n", n)
		}
	}
	if len(utanRom) > 0 {
		fmt.Printf("\nRom huset ikkje kjenner (timen fær ikkje noko rom):\n")
		for _, n := range ulike(utanRom) {
			fmt.Printf("  %q\n", n)
		}
	}

	if !skriv {
		fmt.Println("\nIngen ting er endra. Køyr med -skriv for aa leggja det inn.")
		return nil
	}
	if len(kø) == 0 {
		fmt.Println("\nAlt stend her alt. Ingen ting aa skriva.")
		return nil
	}

	// ---- skrivinga ----
	//
	// Ei rekkje um gongen, kvar i si eigi økt (`LagSerie` og
	// `UtvidSerie` opnar kvar sin). Fell den femte, stend dei fire
	// fyrste — og det er med vilje: verktyet kann køyrast ein gong til,
	// og det som alt kom inn vert hoppa yver. Ei økt kring heile
	// importen hadde gjeve alt-eller-ingen ting, men ogso ei lang laasing
	// paa ein base tenaren les fraa samstundes.
	fmt.Println()
	for _, l := range lyft {
		if err := db.SetRomPlassar(l.id, l.til); err != nil {
			return fmt.Errorf("kunde ikkje setja plassane i %s: %w", l.namn, err)
		}
		fmt.Printf("  %s held %d no (stod %d)\n", l.namn, l.til, l.fraa)
	}

	var lagde int
	for _, a := range kø {
		f := a.serie.Fyrste()
		if a.serieID == 0 {
			id, ider, err := db.LagSerie(a.nye)
			if err != nil {
				return fmt.Errorf("«%s %s %s»: %w", f.Title,
					yogo.Vekedagsnamn(f.StartTime.Weekday()), f.StartTime.Format("15:04"), err)
			}
			lagde += len(ider)
			fmt.Printf("  ny rekkje %d: %s %s %s — %d timar\n", id,
				yogo.Vekedagsnamn(f.StartTime.Weekday()), f.StartTime.Format("15:04"),
				f.Title, len(ider))
			continue
		}
		ider, err := db.UtvidSerie(a.serieID, a.nye)
		if err != nil {
			return fmt.Errorf("«%s»: %w", f.Title, err)
		}
		lagde += len(ider)
		fmt.Printf("  rekkje %d fekk %d timar til: %s %s %s\n", a.serieID, len(ider),
			yogo.Vekedagsnamn(f.StartTime.Weekday()), f.StartTime.Format("15:04"), f.Title)
	}

	fmt.Printf("\n%d timar lagde inn i %d rekkjor.\n", lagde, len(kø))
	return nil
}

// stutt klipper eit namn so tabellen held spaltone sine.
func stutt(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// ulike gjev kvar verdi ein gong, sortert.
func ulike(inn []string) []string {
	sett := map[string]bool{}
	var ut []string
	for _, s := range inn {
		if sett[s] {
			continue
		}
		sett[s] = true
		ut = append(ut, s)
	}
	sort.Strings(ut)
	return ut
}

// fordelingstekst syner kor mange timar som gjekk med kvart tal, so ein
// ser kva som er rommet og kva som er eit val: «18×252  10×12  5×24».
func fordelingstekst(m map[int]int) string {
	if len(m) == 0 {
		return "—"
	}
	tal := make([]int, 0, len(m))
	for s := range m {
		tal = append(tal, s)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(tal)))
	var b []string
	for _, s := range tal {
		b = append(b, fmt.Sprintf("%d×%d", s, m[s]))
	}
	return strings.Join(b, "  ")
}

// maksPerRom gjev det største talet me hev sett i kvart rom, og heile
// fordelingi attaat.
//
// Det største talet er rommet: eit rom kann ikkje halda færre enn den
// største timen som hev gjenge i det. Fordelingi er beviset — ho syner
// kva som er rommet og kva som er eit val («18×250  10×12  5×24»).
func maksPerRom(timar []models.Event) (map[int]int, map[int]map[int]int) {
	maks := map[int]int{}
	fordeling := map[int]map[int]int{}
	for _, e := range timar {
		if e.RoomID == 0 {
			continue
		}
		if e.Capacity > maks[e.RoomID] {
			maks[e.RoomID] = e.Capacity
		}
		if fordeling[e.RoomID] == nil {
			fordeling[e.RoomID] = map[int]int{}
		}
		fordeling[e.RoomID][e.Capacity]++
	}
	return maks, fordeling
}

// latArva nullar kapasiteten paa dei timane som held rommet sitt tal, og
// gjev att kor mange som ber si eigi.
//
// Null i `capacity` tyder «rommet raar» (`COALESCE(NULLIF(...))`), og
// det er det me vil ha for ein vanleg time: fær Salen tvo matter til,
// fylgjer alle timane med. Ein time som er sett *lægre* med vilje —
// fem paa apparati i Salen — ber talet sitt sjølv og skal ikkje fylgja.
//
// Ein time utan rom kann ikkje arva noko og ber alltid sitt eige.
func latArva(timar []models.Event, romPlassar map[int]int) (eigne int) {
	for i := range timar {
		if timar[i].RoomID != 0 && timar[i].Capacity == romPlassar[timar[i].RoomID] {
			timar[i].Capacity = 0
			continue
		}
		if timar[i].Capacity > 0 {
			eigne++
		}
	}
	return eigne
}
