package database

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"time"
)

// Utviklarane.
//
// Lærarrolla vert gjevi ut av ein administrator gjenom flata. Denne gjer
// det ikkje. Ho stend i ei fil paa tenaren, og det er heile skilnaden:
// ho gjev fri tilgang til huset, so ho skal koma fraa den som eig
// maskini — ikkje fraa den som eig ein knapp i administrasjonen. Difor
// stend ho ikkje i RollaFinst, og difor ligg ho aldri i user_roles: det
// finst ingi skriveveg til henne gjenom nettet i det heile.
//
// Fila er ei liste yver e-postar, ei per lina. Tome liner og liner som
// byrjar med # er hoppa yver, so ein kann skriva kvifor nokon stend der.
//
//	# dei som byggjer huset
//	carl@cbcustomhom.es
//
// Stigen kjem or KJERNEKRAFT_UTVIKLARAR. Er han ikkje sett, vert
// ./utviklarar lesi. Finst fila ikkje, er det ingen utviklarar — det er
// ikkje ein feil, det er det vanlege i drift.

// UtviklarfilEnv segjer kvar lista ligg.
const UtviklarfilEnv = "KJERNEKRAFT_UTVIKLARAR"

var (
	utviklarLaas  sync.RWMutex
	utviklarSett  map[string]bool
	utviklarTid   time.Time
	utviklarStig  string
	utviklarLesen bool
)

func utviklarfil() string {
	if s := os.Getenv(UtviklarfilEnv); s != "" {
		return s
	}
	return "./utviklarar"
}

// lesUtviklarar les fila naar ho er endra sidan sist.
//
// Ho vert stat-a og ikkje lesi paa nytt kvar gong: dette vert spurt for
// kvar sida ein teiknar, og ei fillesing per sidevisning er ein pris ein
// ikkje treng betala for ei fil som skifter eit par gonger i aaret.
func lesUtviklarar() map[string]bool {
	stig := utviklarfil()

	var tid time.Time
	if info, err := os.Stat(stig); err == nil {
		tid = info.ModTime()
	}

	utviklarLaas.RLock()
	ferskt := utviklarLesen && stig == utviklarStig && tid.Equal(utviklarTid)
	sett := utviklarSett
	utviklarLaas.RUnlock()
	if ferskt {
		return sett
	}

	nytt := map[string]bool{}
	if f, err := os.Open(stig); err == nil {
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			lina := strings.TrimSpace(s.Text())
			if lina == "" || strings.HasPrefix(lina, "#") {
				continue
			}
			nytt[strings.ToLower(lina)] = true
		}
	}

	utviklarLaas.Lock()
	utviklarSett, utviklarTid, utviklarStig, utviklarLesen = nytt, tid, stig, true
	utviklarLaas.Unlock()
	return nytt
}

// ErUtviklar segjer um e-posten stend i lista.
func ErUtviklar(epost string) bool {
	if epost == "" {
		return false
	}
	return lesUtviklarar()[strings.ToLower(strings.TrimSpace(epost))]
}

// ErUtviklarID slær upp e-posten fyrst.
func (db *Database) ErUtviklarID(userID int64) (bool, error) {
	var epost string
	if err := db.Conn.QueryRow(`SELECT email FROM users WHERE id = ?`, userID).Scan(&epost); err != nil {
		return false, err
	}
	return ErUtviklar(epost), nil
}

// nullstillUtviklarbuffer tvingar ei ny lesing. Prøvone byter fil under
// føtene paa bufferet, og tvo filer skrivne i same augneblinken kann
// hava same tidsmerket.
func nullstillUtviklarbuffer() {
	utviklarLaas.Lock()
	utviklarSett, utviklarTid, utviklarStig, utviklarLesen = nil, time.Time{}, "", false
	utviklarLaas.Unlock()
}
