package handlers

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"kjernekraft/models"
)

// Fyrenamn. «Vi sest torsdag morgon, Anna Larsen!» er ikkje ei helsing,
// det er ei innkalling.
func fyrenamnet(namn string) string {
	if i := strings.IndexByte(namn, ' '); i > 0 {
		return namn[:i]
	}
	return namn
}

// HelsingTittel er sidetittelen: «Sest i morgon tidleg, Anna».
//
// Tittelen bar «Hei, Anna» ei stund. Det er ei helsing utan innhald —
// sida veit *naar* du skal hit att, og det er det fyrste ein lurer paa.
// Naar-et er vagt med vilje («i morgon tidleg», ikkje «07:15»): det
// presise slaget stend i briefingen under, og ein tittel som ber eit
// klokkeslett les seg som ei innkalling.
//
// `naar` tom tyder at det ikkje stend noko framfor deg. Daa kann
// tittelen ikkje lova eit attersyn, og fell attende paa helsingi.
func HelsingTittel(lang, namn, naar string, naa time.Time, nettFerdig bool) string {
	// Nett ferdig med ein time ho *var paa*. Daa er det ikkje neste gong
	// som er det ho lurer paa — ho stend i garderoben. Dette gjeng fyre
	// alt anna, av di det er det ferskaste som hev hendt.
	if nettFerdig {
		return fmt.Sprintf(t(lang, "greeting.takk_for_no"), fyrenamnet(namn))
	}
	if naar == "" {
		// Ingen time framfor seg: daa kann tittelen ikkje lova eit
		// attersyn, og «Hei» er det einaste han hev aa segja. Men
		// klokka veit meir enn det — ho veit um det er morgon eller
		// natt — og ei helsing som veit kva tid paa døgnet det er, er
		// den same helsingi ein fær i døri.
		return fmt.Sprintf(t(lang, "greeting.hi_fmt"), doegnhelsing(lang, naa), fyrenamnet(namn))
	}
	return fmt.Sprintf(t(lang, "greeting.title_see_you"), naar, fyrenamnet(namn))
}

// doegnhelsing gjev «God morgon», «God dag», «God eftan», «God kveld»
// eller «God natt» etter kva klokka er.
//
// Skili er der ein sjølv byter ord, ikkje der klokka gjer det: morgonen
// varer til ein er komen i gang, dagen til ettermiddagen tek yver,
// eftanen til det er kveld, og natti byrjar naar ein ikkje ventar
// fleire.
func doegnhelsing(lang string, naa time.Time) string {
	switch h := naa.In(OsloLoc).Hour(); {
	case h < 5:
		return t(lang, "greeting.god_natt")
	case h < 10:
		return t(lang, "greeting.god_morgon")
	case h < 17:
		return t(lang, "greeting.god_dag")
	case h < 20:
		return t(lang, "greeting.god_eftan")
	case h < 23:
		return t(lang, "greeting.god_kveld")
	default:
		return t(lang, "greeting.god_natt")
	}
}

// HelsingNaar gjev det vage naar-et tittelen ber: «i morgon tidleg»,
// «tysdag kveld», «2. september». Tom streng naar ingen time stend att.
//
// Ein seier ikkje «vi sest torsdag 07:15» til nokon — ein seier
// «torsdag morgon». Klokkeslettet stend i briefingen for den som treng
// det paa minuttet.
func HelsingNaar(lang string, neste *models.Event, naa time.Time) string {
	if neste == nil {
		return ""
	}

	// Tidi i databasen er klokka på veggen; ho lyt byggjast opp att og
	// ikkje reknast om. Sjå tid.go — det var ei umrekning her som gjorde
	// at helsinga sa to timar for seint.
	start := veggklokka(neste.StartTime)
	naa = naa.In(OsloLoc)

	iDag := naa.Format("2006-01-02")
	iMorgon := naa.AddDate(0, 0, 1).Format("2006-01-02")
	dagen := start.Format("2006-01-02")

	switch {
	// Kvarteret fyre. Daa er ikkje spursmaalet *naar* lenger — det veit
	// ho, ho stend og gjer seg klaar — men *kvar*. Alle timar gjeng i
	// andre høgd, so det er det einaste som er att aa segja.
	case start.Sub(naa) <= 15*time.Minute:
		return t(lang, "greeting.andre_hogd")
	case dagen == iDag:
		// I dag treng ein slaget, ikkje dagen. Men slaget aaleine vart
		// «Sest 18:00, Anna» i tittelen, og det er ikkje ei setning —
		// difor ber det med seg ordet som gjer det til ei.
		return fmt.Sprintf(t(lang, "greeting.today_at"), start.Format("15:04"))
	case dagen == iMorgon:
		// «i morgon morgon» er ikkje noko nokon seier. Morgonen fær si
		// eigi vending; dei andre bolkarne fylgjer etter «i morgon».
		if start.Hour() < 10 {
			return t(lang, "greeting.tomorrow_early")
		}
		return t(lang, "greeting.tomorrow") + " " + tidbolk(lang, start)
	case start.Sub(naa) < 7*24*time.Hour:
		return vekedag(lang, start) + " " + tidbolk(lang, start)
	default:
		// Lenger fram enn ei vika: daa er dagen viktigare enn tidi.
		return fmt.Sprintf(t(lang, "greeting.date_fmt"), start.Day(), maanad(lang, start))
	}
}

// NestePresis gjev det presise naar-et briefingen ber: «i dag, 18:00»,
// «tysdag, 18:00», «2. september, 18:00».
//
// Tittelen er vag av di han er ei helsing; denne er presis av di ho er
// eit svar. Dei tvo skal ikkje segja det same med ulike ord — difor ber
// denne alltid klokkeslettet, ogso naar tittelen alt sa «i morgon».
func NestePresis(lang string, neste *models.Event, naa time.Time) string {
	if neste == nil {
		return ""
	}
	start := veggklokka(neste.StartTime)
	naa = naa.In(OsloLoc)

	iDag := naa.Format("2006-01-02")
	iMorgon := naa.AddDate(0, 0, 1).Format("2006-01-02")
	dagen := start.Format("2006-01-02")

	var dag string
	switch {
	case dagen == iDag:
		dag = t(lang, "greeting.today")
	case dagen == iMorgon:
		dag = t(lang, "greeting.tomorrow")
	case start.Sub(naa) < 7*24*time.Hour:
		dag = vekedag(lang, start)
	default:
		dag = fmt.Sprintf(t(lang, "greeting.date_fmt"), start.Day(), maanad(lang, start))
	}
	return dag + ", " + start.Format("15:04")
}

// Briefing er setninga under tittelen, delt i dei bitane malen set
// saman. Logikken bur her og ikkje i malen av di det er her han kann
// prøvast: det er samansetjinga som kann segja noko rart, ikkje ordi.
//
// Tri lekkar, og kvar av deim kann falla burt:
//
//	«Den neste timen din er tysdag, 18:00.»          alltid, um noko stend att
//	«Du er påmeld 3 timar denne veka»                 berre naar det er fleire enn éin
//	« og har 30 klipp å bruke att.»                   berre naar det finst klipp
//
// Fell den midtarste burt, kann den siste ikkje byrja med «og» — daa
// heng bindeordet i lause lufti. Difor tvo former av klippelekken, og
// difor vel Go kva nykel malen skal be um: malen kann ikkje sjaa kva
// som stend fyre honom.
type Briefing struct {
	Naar       string // «tysdag, 18:00»; tom naar ingen time stend att
	Med        string // læraren; tom naar timen ikkje hev nokon
	Stad       string // rommet; tom naar det ikkje er sett
	VekeTal    int
	VekeNykel  string // tom = lekken vert ikkje sagd
	KlippTal   int
	KlippNykel string // tom = lekken vert ikkje sagd
	// Setningi for «ingen timar, men klipp». Ho er heil i seg sjølv og
	// avløyser baade tom-lekken og klippelekken.
	TomNykel string
}

// NyBriefing vel lekkane og formene deira.
//
// `iVeka` er kor mange timar ein stend paa denne veka — den neste
// medrekna. Éin time er den ein alt hev fenge vita um i setningi fyre,
// so han vert ikkje sagd ein gong til; fyrst fraa tvo er det ei
// upplysning.
func NyBriefing(lang string, neste *models.Event, naa time.Time, iVeka, klipp int) Briefing {
	b := Briefing{Naar: NestePresis(lang, neste, naa)}

	// Kven og kvar. «Den neste timen din er i morgon, 07:45» segjer naar
	// ein skal, men ikkje kva ein gjeng til: ein time er ein lærar i eit
	// rom. Baae er valfrie — stend dei tome, fell lekken burt og
	// setningi held seg heil.
	if neste != nil {
		b.Med = neste.TeacherName
		b.Stad = neste.RoomName
		if b.Stad == "" {
			b.Stad = neste.Location
		}
	}

	// Utan ein time framfor seg finst det ingi vike aa telja i, og daa
	// kann vekelekken ikkje standa. Regelen lyt bu her og ikkje i
	// malen: gjorde malen det aaleine, valde Go framleis «og har …»-
	// forma av klippelekken, og setningi vart «Du har ingen timar
	// framfor deg. og har 30 klipp å bruke att.» Bindeordet peika paa
	// noko som aldri vart teikna.
	if b.Naar == "" {
		iVeka = 0
	}

	if iVeka >= 2 {
		b.VekeTal = iVeka
		b.VekeNykel = "greeting.week_count_many"
	}

	// Ingen timar, men klipp paa boki: daa er det ikkje tvo setningar
	// som stend attmed kvarandre — «Du har ingen timar framfor deg.»
	// og so «Du har 30 klipp å bruke att.» — det er *ei* setning med
	// eit «men» i, og eit «men» er heile skilnaden. Det fyrste er ein
	// mangel, det andre er ein veg ut av honom, og det siste er ei
	// oppmoding.
	if b.Naar == "" && klipp > 0 {
		b.KlippTal = klipp
		if klipp == 1 {
			b.TomNykel = "greeting.ingen_timar_men_eitt"
		} else {
			b.TomNykel = "greeting.ingen_timar_men_klipp"
		}
		return b
	}

	if klipp > 0 {
		b.KlippTal = klipp
		eitt := klipp == 1
		switch {
		case b.VekeNykel != "" && eitt:
			b.KlippNykel = "greeting.klipp_join_one"
		case b.VekeNykel != "":
			b.KlippNykel = "greeting.klipp_join_many"
		case eitt:
			b.KlippNykel = "greeting.klipp_solo_one"
		default:
			b.KlippNykel = "greeting.klipp_solo_many"
		}
	}
	return b
}

// vekedag og maanad gjev namnet paa det maalet brukaren hev valt.
//
// Dei stod som tvo faste norske tabellar fyrr, og daa sa den engelske
// sida «tysdag». Vekedagane bur alt i `timeplan.*` av di timeplanen
// treng deim; her vert dei berre sette i rett kasus.
//
// Kasusen er data og ikkje kode: engelsk skriv vekedagar og maanader
// med stor fyrstebokstav midt i ei setning, norsk gjer det ikkje.
func vekedag(lang string, t0 time.Time) string {
	nykel := [...]string{
		"timeplan.sunday", "timeplan.monday", "timeplan.tuesday",
		"timeplan.wednesday", "timeplan.thursday", "timeplan.friday",
		"timeplan.saturday",
	}[t0.Weekday()]
	return iSetning(lang, t(lang, nykel))
}

func maanad(lang string, t0 time.Time) string {
	nykel := [...]string{
		"month.jan", "month.feb", "month.mar", "month.apr",
		"month.may", "month.jun", "month.jul", "month.aug",
		"month.sep", "month.oct", "month.nov", "month.dec",
	}[int(t0.Month())-1]
	return t(lang, nykel)
}

// iSetning set ordet i det kasuset maalet nyttar midt i ei setning.
func iSetning(lang, ord string) string {
	if ord == "" || t(lang, "greeting.day_case") == "title" {
		return ord
	}
	r := []rune(ord)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// tidbolk gjev «morgon», «føremiddag», «ettermiddag» eller «kveld».
//
// Ein seier ikkje «vi sest torsdag 07:15» til nokon — ein seier
// «torsdag morgon». Klokkeslettet stend i timeplanen for den som treng
// det paa minuttet.
func tidbolk(lang string, t0 time.Time) string {
	switch h := t0.Hour(); {
	case h < 10:
		return t(lang, "greeting.morning")
	case h < 12:
		return t(lang, "greeting.forenoon")
	case h < 17:
		return t(lang, "greeting.afternoon")
	default:
		return t(lang, "greeting.evening")
	}
}

// Kvalifisert segjer um brukaren fær sjaa student- og honnørprisane.
//
// Honnør kjem av fødselsdagen — det er eit tal systemet alt hev, og
// ingen skal krysse av for at dei hev vorte 67. Studentbevis er noko
// ein fortel, og studioet ser det i resepsjonen.
func Kvalifisert(u *models.User) bool {
	if u == nil {
		return false
	}
	if u.StudentSenior {
		return true
	}
	if fodd, err := time.Parse("2006-01-02", u.Birthdate); err == nil {
		aar := time.Since(fodd).Hours() / 24 / 365.25
		if aar >= 67 {
			return true
		}
	}
	return false
}
