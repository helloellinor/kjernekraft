package yogo

import (
	"testing"
	"time"

	"kjernekraft/models"
)

func time_(dag int, klokke string, tittel, laerar, rom string, minutt int) models.Event {
	oslo, _ := time.LoadLocation("Europe/Oslo")
	t, err := time.ParseInLocation("2006-01-02 15:04",
		time.Date(2026, 8, dag, 0, 0, 0, 0, oslo).Format("2006-01-02")+" "+klokke, oslo)
	if err != nil {
		panic(err)
	}
	return models.Event{
		Title:       tittel,
		TeacherName: laerar,
		RoomName:    rom,
		StartTime:   t,
		EndTime:     t.Add(time.Duration(minutt) * time.Minute),
	}
}

// Fire utslag av den same timen er éi rekkje, ikkje fire timar. Det er
// heile poenget med importen: huset endrar rekkjor.
func TestSameTimenKvarVikeVertEiRekkje(t *testing.T) {
	// 17., 24. og 31. august er tri maandagar paa rad.
	inn := []models.Event{
		time_(31, "17:30", "Classical Pilates", "Cyrena", "Salen", 50),
		time_(24, "17:30", "Classical Pilates", "Cyrena", "Salen", 50),
		time_(17, "17:30", "Classical Pilates", "Cyrena", "Salen", 50),
	}

	seriar := GrupperISeriar(inn)
	if len(seriar) != 1 {
		t.Fatalf("venta éi rekkje, fann %d", len(seriar))
	}
	if n := len(seriar[0].Timar); n != 3 {
		t.Errorf("rekkja hev %d timar, venta tri", n)
	}
	// Utslagi stend etter dato, so «fyrste» er den næraste.
	if d := seriar[0].Fyrste().StartTime.Day(); d != 17 {
		t.Errorf("fyrste utslaget er den %d., venta den 17.", d)
	}
}

// Ein vikar er ikkje den same rekkja. Byter læraren, er det ei anna
// rekkje — og det er rett: rekkja er «yoga med Leon maandag 18:00».
func TestUlikLaerarGjevUlikRekkje(t *testing.T) {
	seriar := GrupperISeriar([]models.Event{
		time_(17, "17:30", "Pilates Apparatus", "Cyrena", "Salen", 50),
		time_(24, "17:30", "Pilates Apparatus", "Carla", "Salen", 50),
	})
	if len(seriar) != 2 {
		t.Errorf("venta tvo rekkjor naar læraren skil, fann %d", len(seriar))
	}
}

// Tvo timar med det same namnet til den same tidi er tvo timar.
//
// Kjernekraft hev det maandag 17:30: «Pilates Apparatus» med tvo
// lærarar samstundes. Nykelen lyt skilja deim, elles les ein import den
// andre som ein han alt hev og legg honom aldri inn.
func TestTvoLikeNamnSamstundesErTvoTimar(t *testing.T) {
	ein := time_(31, "17:30", "Pilates Apparatus", "Kolbjørn Vårdal", "Salen", 50)
	tvo := time_(31, "17:30", "Pilates Apparatus", "Anne Bull Enger", "Salen", 50)

	if UtslagNykel(ein) == UtslagNykel(tvo) {
		t.Fatal("dei tvo fekk den same nykelen; den eine hadde vorte burte i importen")
	}
	if seriar := GrupperISeriar([]models.Event{ein, tvo}); len(seriar) != 2 {
		t.Errorf("venta tvo rekkjor, fann %d", len(seriar))
	}
}

// Nykelen er den same som `GrupperTimar` i handlers-pakka nyttar paa
// timar utan serie-id. Driv dei fraa kvarandre, lagar importen ei anna
// gruppering enn den administrasjonen syner.
func TestSerieNykelenErDenSameSomListaBrukar(t *testing.T) {
	e := time_(31, "17:30", "Classical Pilates", "Cyrena", "Salen", 50)
	vil := "Classical Pilates|Cyrena|Salen|1|17:30|50"
	if got := SerieNykel(e); got != vil {
		t.Errorf("nykel = %q, venta %q", got, vil)
	}
}

// Slagtabellen dekkjer timeplanen slik han stend. Ho er skriven ut av
// di namni hev stade stille; denne prøva er det som seier ifraa den
// dagen dei ikkje gjer det lenger.
func TestSlagtabellenDekkjerDenTimeplanenSomGjeng(t *testing.T) {
	for namn, vil := range map[string]string{
		"Vinyasa Flow":                 SlagYoga,
		"Yin Yoga & Breathwork":        SlagYoga,
		"Yoga styrke 45":               SlagYoga,
		"Hatha Yoga ":                  SlagYoga, // mellomrom bak, som hjaa Yogo
		"Classical Pilates":            SlagPilates,
		"Klassisk Pilates Matte":       SlagPilates,
		"Pilates Apparatus":            SlagPilates,
		"Pilates Reformer + Apparatus": SlagReformer,
		"Reformer Trio":                SlagReformer,
		"Fascia Flyt ":                 SlagFascia,
		"Fascia Movement ":             SlagFascia,
		"Fascia Flyt Pilates":          SlagFascia, // fascia, ikkje pilates
	} {
		if got := Slag(namn); got != vil {
			t.Errorf("Slag(%q) = %q, venta %q", namn, got, vil)
		}
	}
}

// «Online Yin Yoga» er den same treningi som «Yin Yoga». Ordet segjer
// kvar timen gjeng, ikkje kva han er.
func TestOnlineErDetSameSlaget(t *testing.T) {
	if got := Slag("Online Yin Yoga"); got != SlagYoga {
		t.Errorf("Slag(«Online Yin Yoga») = %q, venta %q", got, SlagYoga)
	}
	if got := Slag("Pilates Online"); got != SlagPilates {
		t.Errorf("Slag(«Pilates Online») = %q, venta %q", got, SlagPilates)
	}
}

// Eit slag me ikkje kjenner fær ingen farge, og importen seier ifraa.
// Ein graa venge er sann; ein turkis er ei gjeting som ser ut som eit
// svar (ARKET §1).
func TestUkjendtSlagVertGraattOgVertSagtIfraaUm(t *testing.T) {
	if got := Slag("Power Plate"); got != "" {
		t.Errorf("Slag(«Power Plate») = %q, venta tom streng", got)
	}
	ukjende := UkjendeSlag([]string{
		"Vinyasa Flow", "Power Plate", "Vinyasa Flow", "Shake the dust", "",
	})
	if len(ukjende) != 2 {
		t.Fatalf("venta tvo ukjende, fann %v", ukjende)
	}
	for _, n := range ukjende {
		if n != "Power Plate" && n != "Shake the dust" {
			t.Errorf("uventa ukjend: %q", n)
		}
	}
}
