package yogo

import "strings"

// ---- what kind of training a Yogo class is ----
//
// Yogo does not know the kind. It has *class types* — 66 of them,
// everything the studio has ever set up — and the name is all that
// separates them: "Vinyasa Flow", "Klassisk Pilates Matte", "Fascia Flyt".
// The house has four kinds, and they carry the wing colour (ARKET §1):
//
//	fascia · yoga · pilates · reformer
//
// So the names have to be translated, and it is a translation and not a
// computation: "Vinyasa Flow" is yoga, but the word "yoga" is not in the
// name. No rule can know that. A table can.
//
// Written out rather than derived, because the names have stood still for
// two years and a table you can read is a table you can correct. What is
// *not* here gets no colour — deliberately: an unknown class must not pass
// itself off as yoga.
const (
	SlagFascia   = "fascia"
	SlagYoga     = "yoga"
	SlagPilates  = "pilates"
	SlagReformer = "reformer"
)

// The table. The key is the Yogo name washed: lower case, no surrounding
// space. Yogo's names carry it — "Fascia Flyt ", "Hatha Yoga ",
// "Reformer " — and a name with a trailing space is a different name.
//
// Alphabetical so a line can be found. A new name makes the import say so
// (UkjendeSlag), and then one line here is correction enough.
var slagtabellen = map[string]string{
	"akroyoga":                            SlagYoga,
	"basic yoga":                          SlagYoga,
	"classical pilates":                   SlagPilates,
	"dharma yoga i - 2":                   SlagYoga,
	"dharma yoga ii":                      SlagYoga,
	"dharma yoga iii":                     SlagYoga,
	"drop-in mamma og baby yoga":          SlagYoga,
	"embodied flow yoga":                  SlagYoga,
	"fascia flyt":                         SlagFascia,
	"fascia flyt pilates":                 SlagFascia,
	"fascia movement":                     SlagFascia,
	"folding mat":                         SlagPilates,
	"forrest yoga":                        SlagYoga,
	"fusion pilates":                      SlagPilates,
	"gravidyoga":                          SlagYoga,
	"hatha yoga":                          SlagYoga,
	"introduksjon til pilates apparatene": SlagPilates,
	"joe's gym":                           SlagPilates,
	"klassisk pilates intermediate":       SlagPilates,
	"klassisk pilates matte":              SlagPilates,
	"neuro pilates":                       SlagPilates,
	"pilates":                             SlagPilates,
	"pilates apparatus":                   SlagPilates,
	"pilates e-motion":                    SlagPilates,
	"pilates fusion":                      SlagPilates,
	"pilates reformer + apparatus":        SlagReformer,
	"pilates slow flow":                   SlagPilates,
	"prøv yoga for første gang":           SlagYoga,
	"reformer apparatus intro session":    SlagReformer,
	"reformer trio":                       SlagReformer,
	"self practice pilates apparatus":     SlagPilates,
	"semiprivate folding mat":             SlagPilates,
	"semiprivate folding mat express":     SlagPilates,
	"slow flow & meditasjon":              SlagYoga,
	"studentpraksis yin yoga":             SlagYoga,
	"vinyasa flow":                        SlagYoga,
	"yang & yin":                          SlagYoga,
	"yin yoga":                            SlagYoga,
	"yin yoga & breathwork":               SlagYoga,
	"yoga basic":                          SlagYoga,
	"yoga styrke 45":                      SlagYoga,
}

// Slag gives the kind a Yogo name belongs to, or the empty string.
//
// The empty string is an answer, not an error: some of Yogo's class types
// are neither yoga, pilates, reformer nor fascia — "Power Plate",
// "Stressmestringsgruppe", "Nevro-atletisk trening" — and a kind we do not
// have is better said in grey than guessed at.
//
// What *is* computed are the online variants and the one-offs: "Online
// Vinyasa Flow" is the same training as "Vinyasa Flow". Hence one attempt
// on the whole name and one on the name without the "online" prefix.
func Slag(namn string) string {
	reint := strings.ToLower(strings.TrimSpace(namn))
	if s, ok := slagtabellen[reint]; ok {
		return s
	}
	// «Online Yin Yoga» og «Yin Yoga» er det same slaget. Ordet segjer
	// kvar timen gjeng, ikkje kva han er.
	if utan := strings.TrimSpace(strings.TrimPrefix(reint, "online")); utan != reint {
		if s, ok := slagtabellen[utan]; ok {
			return s
		}
	}
	if utan := strings.TrimSpace(strings.TrimSuffix(reint, "online")); utan != reint {
		if s, ok := slagtabellen[utan]; ok {
			return s
		}
	}
	return ""
}

// UkjendeSlag gives the names in the list the table does not know, once
// each.
//
// The import prints them. An import that quietly gives grey wings to a
// whole week is an import that looks like it went fine; a list of what it
// did not know is the difference between a correction and a riddle.
func UkjendeSlag(namn []string) []string {
	sett := map[string]bool{}
	var ut []string
	for _, n := range namn {
		n = strings.TrimSpace(n)
		if n == "" || Slag(n) != "" || sett[strings.ToLower(n)] {
			continue
		}
		sett[strings.ToLower(n)] = true
		ut = append(ut, n)
	}
	return ut
}
