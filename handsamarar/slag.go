package handsamarar

import "strings"

// slagKlassane are the kinds of training the house draws a colour for,
// in the order the activity bar stacks them, bottom to top.
//
// Written out with the prefix rather than composed from "slag-" + kind:
// scripts/daude-klassar.sh matches class names as they appear in the
// stylesheet, and a name that only exists composed reads as dead.
var slagKlassane = []string{
	"slag-fascia",
	"slag-yoga",
	"slag-pilates",
	"slag-reformer",
}

// Slagi are the same kinds without the prefix — the values
// events.class_type carries.
var Slagi = slagUtanKrok()

func slagUtanKrok() []string {
	ut := make([]string, len(slagKlassane))
	for i, k := range slagKlassane {
		ut[i] = strings.TrimPrefix(k, "slag-")
	}
	return ut
}

// SlagKlasse gives the CSS hook for a class type, or "" when the house
// does not know it.
//
// class_type is free text, so it is washed to lower-case a–z and then
// looked up. An unknown kind gets no class and the wing falls back to
// grey: a wing that lies about what the class is, is worse than none.
// TestSlagfargarKjennerDeiSameSlagi keeps this list and the stylesheet
// in step.
func SlagKlasse(slag string) string {
	reint := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r
		}
		return -1
	}, strings.ToLower(slag))
	for i, s := range Slagi {
		if s == reint {
			return slagKlassane[i]
		}
	}
	return ""
}
