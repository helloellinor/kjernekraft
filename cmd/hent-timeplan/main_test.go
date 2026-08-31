package main

import (
	"testing"

	"kjernekraft/models"
)

func time_(romID, plassar int) models.Event {
	return models.Event{RoomID: romID, Capacity: plassar}
}

// Rommet held so mange som den største timen i det. Yogo hev ingi
// romkapasitet — han hev `seats` per time — so det er der talet lyt
// koma frå.
//
// Kjernekraft: reformer-rommet stod på 4 i basen, medan 36 timar i det
// hadde fem paameldingsplassar. Eit rom som er sett for laagt gjer
// timar fulle fyre dei er det.
func TestRommetHeldSoMangeSomDenStorsteTimenIDet(t *testing.T) {
	maks, fordeling := maksPerRom([]models.Event{
		time_(2, 4), time_(2, 4), time_(2, 5), time_(2, 3),
		time_(1, 18), time_(1, 10),
		time_(0, 99), // utan rom — skal ikkje telja
	})

	if maks[2] != 5 {
		t.Errorf("reformer = %d, venta 5", maks[2])
	}
	if maks[1] != 18 {
		t.Errorf("salen = %d, venta 18", maks[1])
	}
	if _, finst := maks[0]; finst {
		t.Error("ein time utan rom sette eit tal paa rom 0")
	}
	if fordeling[2][4] != 2 {
		t.Errorf("fordelingi seier %d timar med fire plassar, venta tvo", fordeling[2][4])
	}
}

// Ein vanleg time arvar rommet; ein som er sett lægre ber talet sitt.
//
// Det er skilnaden på «kor mange rommet held» og «kor mange me slepp
// inn denne gongen», og han lyt vera i basen: null tyder «rommet rår»,
// so fær Salen tvo matter til, fylgjer dei vanlege timane med av seg
// sjølve. Skreiv me talet inn på kvar time, var det tvo hundrad stader
// aa retta, og dei hadde stade att på det gamle talet.
func TestTimenArvarRommetNårHanIkkjeErSettLægre(t *testing.T) {
	timar := []models.Event{
		time_(1, 18), // Salen full — arvar
		time_(1, 10), // Klassisk Pilates Matte, sett lægre — ber sitt eige
		time_(1, 5),  // apparat i Salen — ber sitt eige
		time_(2, 5),  // reformer full — arvar
		time_(2, 4),  // reformer sett lægre — ber sitt eige
		time_(0, 12), // utan rom — kann ikkje arva
	}
	eigne := latArva(timar, map[int]int{1: 18, 2: 5})

	if eigne != 4 {
		t.Errorf("%d timar ber si eigi kapasitet, venta fire", eigne)
	}
	vil := []int{0, 10, 5, 0, 4, 12}
	for i, v := range vil {
		if timar[i].Capacity != v {
			t.Errorf("time %d hev kapasitet %d, venta %d", i, timar[i].Capacity, v)
		}
	}
}

// Eit rom huset ikkje kjenner att kann ikkje arva noko. Timen ber talet
// sitt sjølv, og det er rett: ingen ting aa arva frå.
func TestTimeUtanRomBerAlltidSittEigeTal(t *testing.T) {
	timar := []models.Event{time_(0, 12)}
	if eigne := latArva(timar, map[int]int{1: 18}); eigne != 1 {
		t.Errorf("eigne = %d, venta ein", eigne)
	}
	if timar[0].Capacity != 12 {
		t.Errorf("kapasiteten vart %d, venta 12", timar[0].Capacity)
	}
}
