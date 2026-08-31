// hent-timeplan fetches the schedule from Yogo and puts it in the
// database.
//
// The studio books in Yogo today. This is the tool that moves the plan
// here — not a sync that runs continuously, but a fetch you do a few times
// while moving in.
//
//	go run ./cmd/hent-timeplan -veker 4            # shows what would happen
//	go run ./cmd/hent-timeplan -veker 4 -skriv     # does it
//
// It writes nothing without -skriv. An import cannot be undone — classes
// get ids and signups hang off them — so the first run should be a list
// you read, not a change you discover (ARKET §7: the warning comes before
// the press).
//
// It can be run several times. An occurrence already in the database is
// skipped, and a run it recognises gets its new classes added to the
// existing series rather than a new one with the same name.
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

	// The span reaches both ways from today.
	//
	// Forward is the schedule; backward is the *history*, and that is not
	// decoration: without past classes there is nothing to mark anyone present
	// on, and "classes this year" on the person card is zero for everyone.
	// Half a year back is 26 weeks.
	no := time.Now().In(k.Sone)
	frå := no.AddDate(0, 0, -7*attende)
	til := no.AddDate(0, 0, 7*veker-1)

	ctx, stopp := context.WithTimeout(context.Background(), 3*time.Minute)
	defer stopp()

	fmt.Printf("Hentar %s – %s (%d vikor attende, %d fram) …\n",
		frå.Format("2.1.2006"), til.Format("2.1.2006"), attende, veker)
	komande, err := k.Timar(ctx, frå, til, yogo.Val{MedAvlyste: medAvlyste})
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

	// Rooms. Yogo has its own, the house has its own, and they meet on the
	// name — "Salen" and "Reformer" stand in both. A room we do not recognise
	// is not guessed at: the class gets room_id = 0, meaning no room, and the
	// name stays in location.
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

	// ---- how many the room holds ----
	//
	// Yogo has no room capacity. It has seats per class, and that is two
	// things mixed: how many the room *holds*, and how many the studio lets in
	// *this* time.
	//
	// One can be read from the other: a room cannot hold fewer than the
	// largest class that has run in it. So the highest number we see is the
	// room's, and anything below is a choice made for that class.
	//
	// Upward only. Seeing no class that fills the room does not mean the room
	// is smaller — only that nobody filled it in the span we asked about.
	// Setting the number *down* on that basis would make classes full that are
	// not.
	maks, fordeling := maksPerRom(komande)

	romPlassar := map[int]int{}
	fmt.Println("ROMMET      HELD  YOGO SITT STØRSTE  FORDELING")
	fmt.Println(strings.Repeat("─", 78))
	type romlyft struct {
		id, frå, til int
		namn         string
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

	// A class carries its own capacity only when it *differs* from the room's.
	//
	// If they are equal the number should be zero — "the room decides" — and
	// then the class follows if the room ever changes. Written onto every
	// class, there would be hundreds of places to fix the day the hall gained
	// two more mats, and the classes would have stayed on the old number
	// unnoticed.
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

	// ---- the writing ----
	//
	// One run at a time, each in its own transaction (LagSerie and UtvidSerie
	// each open one). If the fifth fails, the first four stand — deliberately:
	// the tool can be run again and what is already in is skipped. One
	// transaction around the whole import would give all-or-nothing, but also
	// a long lock on a database the server is reading from at the same time.
	fmt.Println()
	for _, l := range lyft {
		if err := db.SetRomPlassar(l.id, l.til); err != nil {
			return fmt.Errorf("kunde ikkje setja plassane i %s: %w", l.namn, err)
		}
		fmt.Printf("  %s held %d no (stod %d)\n", l.namn, l.til, l.frå)
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

// maksPerRom gives the largest number seen in each room, and the whole
// distribution alongside.
//
// The largest is the room's: a room cannot hold fewer than the largest
// class that has run in it. The distribution is the evidence — it shows
// what is the room and what is a choice ("18×250  10×12  5×24").
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

// latArva zeroes the capacity on classes that hold the room's number, and
// returns how many carry their own.
//
// Zero in capacity means "the room decides" (COALESCE(NULLIF(...))), which
// is what you want for an ordinary class: if the hall gains two mats,
// every class follows. A class deliberately set *lower* — five on the
// apparatus in the hall — carries its own number and should not follow.
//
// A class without a room cannot inherit anything and always carries its
// own.
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
