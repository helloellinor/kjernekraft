package yogo

import "strings"

// ---- kva slag trening ein Yogo-time er ----
//
// Yogo kjenner ikkje slaget. Han hev *timetypar* — 66 av deim, alt
// studioet hev sett upp gjenom aari — og namnet er det einaste som
// skil deim: «Vinyasa Flow», «Klassisk Pilates Matte», «Fascia Flyt».
// Huset her hev fire slag, og dei ber vengefargen (§1 i ARKET, kartet
// stend i 00-token.css):
//
//	fascia · yoga · pilates · reformer
//
// Difor lyt namni umsetjast, og det er ei umsetjing og ikkje ei
// utrekning: «Vinyasa Flow» er yoga, men ordet «yoga» stend ikkje i
// namnet. Ingi regel kann vita det. Tabellen kann.
//
// Han er skriven ut i staden for rekna av di namni hev stade stille i
// tvo aar, og av di ei tabell ein kann lesa er ei tabell ein kan retta.
// Det som *ikkje* stend her fær ingen farge — og det er meiningi:
// «Ein graa venge seier «ingen farge», som er sant.» Ein ukjend time
// skal ikkje utgje seg for yoga.
const (
	SlagFascia   = "fascia"
	SlagYoga     = "yoga"
	SlagPilates  = "pilates"
	SlagReformer = "reformer"
)

// slagtabellen. Nykelen er Yogo-namnet vaska: smaa bokstavar, ingen
// mellomrom i endane. Namni hjaa Yogo ber deim — «Fascia Flyt »,
// «Hatha Yoga », «Reformer » — og eit namn med eit mellomrom bak er
// eit anna namn enn det same utan.
//
// Rekkjefylgda her er alfabetisk so det gjeng an aa finna ein linje.
// Kjem det eit nytt namn, seier importen ifraa (`UkjendeSlag`), og daa
// er det ei linje her som er retting nok.
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

// Slag gjev slaget eit Yogo-namn høyrer til, eller tom streng.
//
// Tom streng er eit svar og ikkje ein feil: nokre av timetypane hjaa
// Yogo er korkje yoga, pilates, reformer eller fascia — «Power Plate»,
// «Stressmestringsgruppe», «Nevro-atletisk trening» — og eit slag me
// ikkje hev er betre sagt med graatt enn gjeta paa.
//
// Det som *er* rekna, er dei nettbaserte og dei eingongshøvi: «Online
// Vinyasa Flow» er den same treningi som «Vinyasa Flow», og eit namn
// som endar paa eit aarstal eller ei hending er framleis slaget sitt.
// Difor eit fyrste forsøk paa heile namnet, og so eit paa namnet utan
// «online»-føreleddet.
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

// UkjendeSlag gjev dei namni i lista som tabellen ikkje kjenner, ein
// gong kvar.
//
// Importen skriv deim ut. Ein import som stilt gjev graa vengar til ei
// heil vika er ein import som ser ut til aa ha gjenge bra; ei liste
// yver kva han ikkje visste er skilnaden paa ei retting og ei gaata.
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
