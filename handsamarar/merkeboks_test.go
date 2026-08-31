package handsamarar

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Skiva og datovindauga er teikna av stilarket no, ikkje av SVG-en.
// Geometrien deira er difor skrivi tvo stader: som konstantar i
// merkeform.go, og som `calc(N * var(--me))` i 30-timeplanen.css.
//
// To stader tyder at dei kann driva frå kvarandre, og drifta ville
// ikkje synast som ein feil — berre som eit vindauga som ligg ein
// hårsbreidd for høgt. Denne prøva held dei i takt, på same vis som
// TestSlagfargarKjennerDeiSameSlagi held slagfargane i takt.
func TestMerkeboksaneFylgjerMerkeform(t *testing.T) {
	ark, err := os.ReadFile(filepath.Join("..", "static", "css", "deler", "30-timeplanen.css"))
	if err != nil {
		t.Fatalf("las ikkje stilarket: %v", err)
	}
	css := string(ark)

	// Kva merkeform.go seier. Namni er dei same som konstantane der.
	skiveB := kroppB - 2*skiveMarg
	skiveH := kroppH - 2*skiveMarg
	vil := []struct {
		vel, eig string
		tal      float64
	}{
		{".merkeskive", "left", kroppX + skiveMarg},
		{".merkeskive", "top", kroppY + skiveMarg},
		{".merkeskive", "width", skiveB},
		{".merkeskive", "height", skiveH},
		{".merkeskive", "border-radius", kroppR - 1},
		{".merkerute", "left", ruteX},
		{".merkerute", "top", ruteY},
		{".merkerute", "width", ruteB},
		{".merkerute", "height", ruteH},
		{".merkerute", "border-radius", ruteR},
	}

	for _, v := range vil {
		fekk, ok := calcTal(css, v.vel, v.eig)
		if !ok {
			t.Errorf("%s { %s } stend ikkje som calc(N * var(--me)) i stilarket", v.vel, v.eig)
			continue
		}
		if math.Abs(fekk-v.tal) > 0.001 {
			t.Errorf("%s { %s }: stilarket seier %g, merkeform.go seier %g",
				v.vel, v.eig, fekk, v.tal)
		}
	}

	// Og stabelen sitt høgd/breidd-tilhøve er e^(1/4)-steget, det same
	// som MarkHeight/MarkWidth i merkeform.go.
	if fekk, ok := calcFaktor(css, ".merkestabel", "height"); ok {
		vil := MarkHeight / MarkWidth
		if math.Abs(fekk-vil) > 0.0001 {
			t.Errorf(".merkestabel høgd: stilarket gjev %g gonger breiddi, merkeform.go %g", fekk, vil)
		}
	} else {
		t.Error(".merkestabel { height } fann eg ikkje")
	}
}

// calcTal hentar N ut or `eig: calc(N * var(--me))` inne i ein veljar.
func calcTal(css, vel, eig string) (float64, bool) {
	blokk, ok := cssBlokk(css, vel)
	if !ok {
		return 0, false
	}
	re := regexp.MustCompile(regexp.QuoteMeta(eig) + `:\s*calc\(([0-9.]+)\s*\*\s*var\(--me\)\)`)
	m := re.FindStringSubmatch(blokk)
	if m == nil {
		return 0, false
	}
	f, err := strconv.ParseFloat(m[1], 64)
	return f, err == nil
}

// calcFaktor hentar faktoren ut or `height: calc(58 * var(--me) * F)`.
func calcFaktor(css, vel, eig string) (float64, bool) {
	blokk, ok := cssBlokk(css, vel)
	if !ok {
		return 0, false
	}
	re := regexp.MustCompile(regexp.QuoteMeta(eig) + `:\s*calc\(58\s*\*\s*var\(--me\)\s*\*\s*([0-9.]+)\)`)
	m := re.FindStringSubmatch(blokk)
	if m == nil {
		return 0, false
	}
	f, err := strconv.ParseFloat(m[1], 64)
	return f, err == nil
}

func cssBlokk(css, vel string) (string, bool) {
	i := strings.Index(css, vel+" {")
	if i < 0 {
		i = strings.Index(css, vel+"{")
	}
	if i < 0 {
		return "", false
	}
	j := strings.Index(css[i:], "}")
	if j < 0 {
		return "", false
	}
	return css[i : i+j], true
}

// Kroppen, fana og indeksane stend ein gong i merke_defs og vert henta
// med <use>. Fyllet deira kann difor ikkje veljast med ein veljar —
// dokumentet sine veljarar naar ikkje inn i eit skuggetre — so det
// stend som arva token på `.dagmerke`.
//
// Denne prøva vaktar tvo ting: at formene *er* delte, og at ingen
// skriv tilstandane attende som etterkomar-veljarar. Fell det siste
// attende, ser merket rett ut i grunnstoda og sluttar å svara på
// «vald», «påmeld» og «i dag» — utan at noko brest.
func TestMerkeformeneErDelteOgFyltAvToken(t *testing.T) {
	mal, err := os.ReadFile(filepath.Join("templates", "components", "common", "dagmerke.html"))
	if err != nil {
		t.Fatalf("las ikkje malen: %v", err)
	}
	m := string(mal)
	for _, id := range []string{"#merkekropp", "#merkeindeks"} {
		if !strings.Contains(m, `<use href="`+id+`"/>`) {
			t.Errorf("merket hentar ikkje %s — formene er teikna per merke att", id)
		}
	}

	ark, err := os.ReadFile(filepath.Join("..", "static", "css", "deler", "30-timeplanen.css"))
	if err != nil {
		t.Fatalf("las ikkje stilarket: %v", err)
	}
	css := string(ark)
	for _, tok := range []string{"--fane-kant", "--fane-flate", "--kropp-kant",
		"--kropp-flate", "--kropp-djup", "--ring-syn"} {
		if !strings.Contains(css, tok+":") {
			t.Errorf("%s er ikkje sett — tilstanden naar ikkje inn i <use>", tok)
		}
	}
	for _, daud := range []string{".form-flate", ".form-kant", ".fanekant", ".form-djup"} {
		if strings.Contains(css, daud+" ") || strings.Contains(css, daud+"{") ||
			strings.Contains(css, daud+",") {
			t.Errorf("%s er ein veljar att — han treffer ikkje inni eit <use>", daud)
		}
	}
}

// hx-boost byter berre <main>. Det er heile grunnen til at skripti kann
// liggja der dei ligg: dei stend utanfor, og vert difor aldri køyrde um
// att. Byter nokon dette til å swappa heile <body>, vert htmx sjølv
// lasta på nytt for kvart sidebyte, og kvar lyttar kjem oppå den fyrre.
//
// Prøva vaktar tri ting: at boosten finst, at han er avgrensa til
// <main>, og at skripti vert lasta utan vilkår — eit `{{if eq
// .CurrentPage …}}` kring dei tyder at sida du *gjeng til* aldri fær
// skriptet sitt, av di me ikkje hentar ein ny <head>.
func TestBoostenByterBerreInnhaldet(t *testing.T) {
	les := func(delar ...string) string {
		b, err := os.ReadFile(filepath.Join(delar...))
		if err != nil {
			t.Fatalf("las ikkje %v: %v", delar, err)
		}
		return string(b)
	}
	base := les("templates", "layouts", "base.html")
	nav := les("templates", "components", "navigation", "navigation.html")

	// Boosten skal stå på leidingi og ingen annan stad.
	//
	// htmx *arvar* hx-target og hx-select nedyver treet. Stod dei på
	// <body>, tok dei kvart einaste hx-get på sida med seg: dei fire som
	// ikkje hev sitt eige mål — /api/charges og dei tri admin-listone —
	// slutta å byta seg sjølve og skreiv yver <main> i staden.
	// Medlemskapssida og administrasjonen vart tome, og ingen ting brast
	// synleg. Det er heile grunnen til at denne prøva finst.
	if strings.Contains(base, "<body hx-boost") || strings.Contains(base, "hx-target=\"main\"") {
		t.Error("boosten stend på <body> — då arvar kvart hx-get målet hans, " +
			"og stykke som skal byta seg sjølve skriv yver <main>")
	}
	if !strings.Contains(nav, `hx-boost="true"`) {
		t.Fatal("leidingi hev ingen hx-boost — sidebyta byggjer dokumentet um att")
	}
	for _, krav := range []string{`hx-target="main"`, `hx-select="main"`} {
		if !strings.Contains(nav, krav) {
			t.Errorf("%s manglar på leidingi", krav)
		}
	}

	// Skripti skal lastast utan vilkår: sida du *gjeng til* fær aldri ein
	// ny <head>, so eit {{if eq .CurrentPage …}} tyder at skriptet hennar
	// aldri kjem.
	i := strings.Index(base, "js/htmx.min.js")
	j := strings.Index(base, "{{if .JS}}")
	if i < 0 || j < 0 || j < i {
		t.Fatal("fann ikkje skript-blokka")
	}
	if strings.Contains(base[i:j], "CurrentPage") {
		t.Error("skripti vert lasta etter kva sida ein stend på — " +
			"då fær sida du gjeng til aldri sitt eige skript")
	}
	if !strings.Contains(base[i:j], "js/sideskift.js") {
		t.Error("sideskift.js er ikkje lasta — leidingi og bindingane fylgjer ikkje med")
	}
}
