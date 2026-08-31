// tryggjing tek ein trygg kopi av basen medan tenaren gjeng.
//
//	go run ./cmd/tryggjing                    # kopi til ./tryggjing/
//	go run ./cmd/tryggjing -til /sti -hald 14 # annan stad, 14 dagar att
//	go run ./cmd/tryggjing -sjå              # list det som ligg der
//
// Kvifor ikkje berre `cp kjernekraft.db`:
//
// Basen gjeng i WAL. Det som stend i .db-fila er *ikkje* heile basen —
// det som er skrive sidan siste sjekkpunkt ligg i .db-wal, og ein kopi
// av berre den eine fila er ein base som manglar det siste som hende.
// Verre: kopierer ein medan nokon skriv, fær ein ei fil som er halvvegs
// gjenom ei endring, og ho ser heil ut til den dagen ein prøver aa
// bruka henne.
//
// `VACUUM INTO` er SQLite sin eigen maate. Han skriv ein ny, heil base
// med det same innhaldet, teken i eitt einaste augneblink, og han gjer
// det utan aa stengja skrivarane ute meir enn ei stund. Kopien er ogso
// pakka saman — ingen tome sider — so han er mindre enn originalen.
//
// Kvar kopi vert prøvd med `PRAGMA integrity_check` fyre han vert lagd
// til side. Ei tryggjing ein ikkje hev opna er ikkje ei tryggjing; ho er
// ei fil ein trur på.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const namnform = "kjernekraft-2006-01-02-1504.db"

func main() {
	base := flag.String("base", "", "basefila (standard: KJERNEKRAFT_DB eller ./kjernekraft.db)")
	til := flag.String("til", "tryggjing", "kvar kopiane skal liggja")
	hald := flag.Int("hald", 14, "kor mange dagar kopiane skal liggja")
	sjå := flag.Bool("sjaa", false, "list kopiane som finst, og gjer ingen ting")
	flag.Parse()

	stig := *base
	if stig == "" {
		stig = os.Getenv("KJERNEKRAFT_DB")
	}
	if stig == "" {
		stig = "./kjernekraft.db"
	}

	if *sjå {
		if err := list(*til); err != nil {
			log.Fatal(err)
		}
		return
	}

	kopi, err := tryggja(stig, *til)
	if err != nil {
		log.Fatalf("tryggjing feila: %v", err)
	}
	info, _ := os.Stat(kopi)
	fmt.Printf("%s  (%d kB)\n", kopi, info.Size()/1024)

	fjerna, err := rydd(*til, *hald)
	if err != nil {
		log.Printf("ryddingi feila: %v", err)
	} else if fjerna > 0 {
		fmt.Printf("rydda burt %d kopiar eldre enn %d dagar\n", fjerna, *hald)
	}
}

// tryggja skriv ein heil kopi av basen og prøver honom fyre han vert
// lagd til side.
func tryggja(base, mappe string) (string, error) {
	if _, err := os.Stat(base); err != nil {
		return "", fmt.Errorf("fann ikkje basen %s: %w", base, err)
	}
	if err := os.MkdirAll(mappe, 0o700); err != nil {
		return "", err
	}

	db, err := sql.Open("sqlite3", base+"?_busy_timeout=10000")
	if err != nil {
		return "", err
	}
	defer db.Close()

	// Fyrst til ei fil med eit anna namn, so ei halvskrivi tryggjing
	// aldri ser ut som ei ferdig ein. Ho fær rett namn til slutt, og
	// namnebytet er det einaste steget som er udeleleg.
	ferdig := filepath.Join(mappe, time.Now().Format(namnform))
	mellom := ferdig + ".vert-skriven"
	_ = os.Remove(mellom)

	// VACUUM INTO toler ikkje eit spursmaalsteikn i strengen, so stigen
	// vert lagd inn beinveges. Han kjem frå oss og ikkje frå nettet.
	if _, err := db.Exec("VACUUM INTO '" + strings.ReplaceAll(mellom, "'", "''") + "'"); err != nil {
		_ = os.Remove(mellom)
		return "", fmt.Errorf("VACUUM INTO: %w", err)
	}

	if err := prov(mellom); err != nil {
		_ = os.Remove(mellom)
		return "", err
	}
	if err := os.Rename(mellom, ferdig); err != nil {
		return "", err
	}
	return ferdig, nil
}

// prov opnar kopien og spør honom um han er heil. Ei tryggjing ein ikkje
// hev opna er ei fil ein trur på, og ikkje ei ein veit noko um.
func prov(stig string) error {
	db, err := sql.Open("sqlite3", stig+"?mode=ro")
	if err != nil {
		return err
	}
	defer db.Close()

	var svar string
	if err := db.QueryRow("PRAGMA integrity_check").Scan(&svar); err != nil {
		return fmt.Errorf("integrity_check: %w", err)
	}
	if svar != "ok" {
		return fmt.Errorf("kopien er ikkje heil: %s", svar)
	}
	// Og ein spurnad som treng at skjemaet finst. `integrity_check` ser
	// på sidone; han ser ikkje um det er ein *kjernekraft*-base.
	var tal int
	if err := db.QueryRow("SELECT count(*) FROM users").Scan(&tal); err != nil {
		return fmt.Errorf("kopien hev ingi brukartabell: %w", err)
	}
	return nil
}

func rydd(mappe string, dagar int) (int, error) {
	if dagar <= 0 {
		return 0, nil
	}
	oppf, err := os.ReadDir(mappe)
	if err != nil {
		return 0, err
	}
	grensa := time.Now().AddDate(0, 0, -dagar)
	fjerna := 0
	for _, o := range oppf {
		if o.IsDir() || !strings.HasPrefix(o.Name(), "kjernekraft-") || !strings.HasSuffix(o.Name(), ".db") {
			continue
		}
		info, err := o.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(grensa) {
			if err := os.Remove(filepath.Join(mappe, o.Name())); err == nil {
				fjerna++
			}
		}
	}
	return fjerna, nil
}

func list(mappe string) error {
	oppf, err := os.ReadDir(mappe)
	if err != nil {
		return err
	}
	var namn []string
	for _, o := range oppf {
		if !o.IsDir() && strings.HasPrefix(o.Name(), "kjernekraft-") && strings.HasSuffix(o.Name(), ".db") {
			namn = append(namn, o.Name())
		}
	}
	sort.Strings(namn)
	if len(namn) == 0 {
		fmt.Printf("ingen tryggjingar i %s\n", mappe)
		return nil
	}
	for _, n := range namn {
		info, err := os.Stat(filepath.Join(mappe, n))
		if err != nil {
			continue
		}
		fmt.Printf("%s  %6d kB  %s\n", n, info.Size()/1024, info.ModTime().Format("2006-01-02 15:04"))
	}
	fmt.Printf("%d kopiar\n", len(namn))
	return nil
}
