package yogo

import (
	"context"
	"os"
	"testing"
	"time"
)

// Ei prøva som *verkeleg* ringjer Yogo.
//
// Ho stend av som standard. Prøvone i huset skal kunna køyra utan nett
// og utan aa plaga ein tenar me ikkje rår yver — men når skjemaet
// hjaa Yogo skiftar, er det berre eit ekte kall som ser det. Difor:
//
//	YOGO_LIVE=1 go test ./yogo/ -run Live -v
//
// Ho hentar éi vike og krev berre at svaret er timar med namn og klokke.
// Ho seier ingen ting um *kva* timar det er — det er studioet sitt val,
// ikkje ein paastand koden kann halda.
func TestLiveHentingFråYogo(t *testing.T) {
	if os.Getenv("YOGO_LIVE") == "" {
		t.Skip("set YOGO_LIVE=1 for aa ringja api.yogo.dk")
	}

	k, err := Ny()
	if err != nil {
		t.Fatalf("Ny: %v", err)
	}

	ctx, stopp := context.WithTimeout(context.Background(), 45*time.Second)
	defer stopp()

	timar, err := k.NesteVeker(ctx, time.Now(), 1, Val{})
	if err != nil {
		t.Fatalf("NesteVeker: %v", err)
	}
	if len(timar) == 0 {
		t.Fatal("ingen timar den komande vika — anten er studioet stengt, eller so hev API-en skift")
	}

	for _, e := range timar {
		if e.Title == "" {
			t.Errorf("time utan namn: %+v", e)
		}
		if e.StartTime.IsZero() || !e.EndTime.After(e.StartTime) {
			t.Errorf("%s: ugild klokke %v–%v", e.Title, e.StartTime, e.EndTime)
		}
	}
	t.Logf("henta %d timar, fyrste: %s %s %s med %s i %s",
		len(timar), timar[0].StartTime.Format("Mon 2.1. 15:04"),
		timar[0].Title, timar[0].Description, timar[0].TeacherName, timar[0].RoomName)
}
