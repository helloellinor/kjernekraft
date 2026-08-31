package handsamarar

import (
	"time"

	"kjernekraft/database"
	"kjernekraft/handsamarar/config"
)

// App er huset handsamarane bur i.
//
// Dei var frie funksjonar som greip etter ein pakkeglobal `DB`, og det er
// grunnen til at ingen av dei sytti handsamarane hadde ei einaste prøva:
// det fanst ingen måte å setja ein av dei upp mot ei prøvebase. Tjuesju
// prøvefiler i denne pakken prøvde alt *utanum* handsamarane — vike- og
// tidrekning, merkeform, malnyklar — og medan dei stod grøne, svara
// /api/admin/freeze-requests/approve 501 og ingen såg det.
//
// Klokka bur her av same grunnen. Huset held seg til den injiserte klokka
// i `config`, men nokre stader kalla `time.Now()` beint, og tvo av dei
// sat i vindaugs-sjekkar brukaren merkar. Ligg klokka på App, kann ei
// prøva flytta tidi i staden for å venta på henne.
type App struct {
	DB *database.Database

	// Nå er klokka huset går etter. Sjå config.GetCurrentTime — ho kann
	// stillast i utvikling, og det er heile grunnen til at ho finst.
	Nå func() time.Time
}

// NyApp byggjer huset med den vanlege klokka.
func NyApp(db *database.Database) *App {
	return &App{
		DB: db,
		Nå: func() time.Time { return config.GetInstance().GetCurrentTime() },
	}
}
