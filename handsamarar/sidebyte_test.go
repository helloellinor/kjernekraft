package handsamarar

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// lesMal les ei malfil raatt. Dei tri prøvone under ser etter attributt
// og ikkje etter det som vert teikna: det er attributti som er avgjerdi.
func lesMal(t *testing.T, delar ...string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(delar...))
	if err != nil {
		t.Fatalf("las ikkje %v: %v", delar, err)
	}
	return string(b)
}

// Sida skal koma inn paa toppen, ikkje 71 piksel nedanfor.
//
// `show:top` rullar *maalet* — <main> — til toppen av glaset. Hovudet er
// klistra og stend dei fyrste 71 pikslane, so sida hamna 71 piksel
// nedrulla med yverskrifti attum hovudet. Kvar einaste boosta lenkja
// gjorde det; eit friskt kall paa den same adressa gjorde det ikkje, og
// difor saag det ut som um det kom an paa kvar ein kom fraa.
//
// Maalt i ein nettlesar fyre og etter: skroll 71 → 0.
func TestSidebyteKjemInnPaaToppen(t *testing.T) {
	nav := lesMal(t, "templates", "components", "navigation", "navigation.html")

	if !strings.Contains(nav, "show:window:top") {
		t.Error("leidingi rullar ikkje glaset til toppen ved eit sidebyte")
	}
	if strings.Contains(nav, `hx-swap="outerHTML show:top"`) {
		t.Error("`show:top` er attende: han rullar <main> under det klistra hovudet")
	}
}

// Utloggingi kann ikkje boostast.
//
// Boosten byter <main> og let resten standa — og resten er nettupp det
// som ikkje lenger skal finnast: hovudet med namnet ditt og, for ein
// sjef, lenkja til administrasjonen. Ei boosta utlogging skreiv
// «/logout» i adressa og let førre brukaren staa i hovudet, so det saag
// ut som um den neste hadde arva løyvi hans.
func TestUtloggingiErEiHeilSidelasting(t *testing.T) {
	nav := lesMal(t, "templates", "components", "navigation", "navigation.html")

	i := strings.Index(nav, `action="/logout"`)
	if i < 0 {
		t.Fatal("fann ikkje utloggingsskjemaet")
	}
	skjema := nav[i:]
	if slutt := strings.Index(skjema, ">"); slutt > 0 {
		skjema = skjema[:slutt]
	}
	if !strings.Contains(skjema, `hx-boost="false"`) {
		t.Error("utloggingsskjemaet er boosta — daa stend hovudet til den " +
			"førre brukaren att etter at han hev gjenge ut")
	}
}

// Ingen `autofocus` inni ein fanebolk.
//
// Attributtet gjorde ingen skade so lenge bolkarne stod `hidden` i
// dokumentet: nettlesaren gjekk forbi eit felt han ikkje kunde sjaa.
// No teiknar tenaren den valde bolken, feltet stend synlegt fraa fyrste
// stund, og daa gjer nettlesaren det `autofocus` alltid hev gjort — han
// rullar det inn i glaset. Paa ei skjerm 600 piksel høg hamna
// administrasjonen 451 piksel nedrulla med fanerekkja utanfor glaset.
//
// Skal eit felt ha fokus, set skriptet det med `preventScroll`. Sjaa
// folk.js.
func TestIngenAutofocusIEinFanebolk(t *testing.T) {
	for _, fil := range [][]string{
		{"templates", "modules", "admin", "admin-users-table.html"},
		{"templates", "pages", "admin.html"},
		{"templates", "pages", "klippekort.html"},
		{"templates", "pages", "medlemskapet.html"},
		{"templates", "pages", "betaling.html"},
	} {
		mal := lesMal(t, fil...)
		// Kommentarane forklarar kvifor attributtet er burte, og dei
		// nemner honom. Det er berre attributtet sjølv som tel.
		utan := malkomm.ReplaceAllString(mal, " ")
		if strings.Contains(utan, "autofocus") {
			t.Errorf("%s hev autofocus i ein fanebolk — sida lastar nedrulla",
				filepath.Join(fil...))
		}
	}
}

// Eit byte skal ikkje føra deg ut or huset.
//
// Skjemaet var eit blankt `<form method="post">` mot ei rute som svara
// med JSON. Nettlesaren gjorde det eit skjema gjer — han navigerte — og
// du hamna paa ei kvit sida med `{"success":true,"message":"Medlemskap
// endret!"}`. Gjekk bytet ikkje i gjenom, fekk du grunnen som naki
// tekst, paa bokmaal, utan veg attende.
//
// Prøva vaktar baae halvdelane: at skjemaet spør gjenom htmx, og at det
// finst ein stad svaret kann standa.
func TestByteskjemaetSvararPaaSida(t *testing.T) {
	mal := lesMal(t, "templates", "pages", "medlemskapet.html")
	utan := malkomm.ReplaceAllString(mal, " ")

	i := strings.Index(utan, `action="/api/membership/change"`)
	if i < 0 {
		t.Fatal("fann ikkje byteskjemaet")
	}
	skjema := utan[i:]
	if slutt := strings.Index(skjema, ">"); slutt > 0 {
		skjema = skjema[:slutt]
	}

	for _, krav := range []string{
		`hx-post="/api/membership/change"`,
		`hx-target="#bytesvar"`,
	} {
		if !strings.Contains(skjema, krav) {
			t.Errorf("byteskjemaet manglar %s — daa navigerer nettlesaren til API-et", krav)
		}
	}

	// Staden svaret skal standa. Utan henne byter htmx inn i inkje, og
	// grunnen til at bytet ikkje gjekk kjem ingen stad.
	if !strings.Contains(utan, `id="bytesvar"`) {
		t.Error("bytefana hev ingen `.svar` — grunnen til eit avslag hev ingen stad aa staa")
	}
}

// Eit fanebyte skal ikkje flytta sida.
//
// htmx set `show:top` av seg sjølv paa kvar boosta lenkja
// (`scrollIntoViewOnBoost`, paa i standard), og det tyder «dreg maalet
// upp til toppen av glaset». Maalet vaart er fanearket, so eit klikk paa
// ei fana rulla yverskrifti og helsingi ut og la rekkja rett attum det
// klistra hovudet. Maalt: skroll 265 → 0.
//
// `show:none` er ikkje ein verdi htmx handsamar serskilt — han berre
// treffer korkje «top» eller «bottom», og daa vert det ingi rulling.
func TestFanebyteFlytterIkkjeSida(t *testing.T) {
	for _, fil := range [][]string{
		{"templates", "components", "common", "faner.html"},
		{"templates", "modules", "admin", "admin-stats.html"},
	} {
		mal := lesMal(t, fil...)
		utan := malkomm.ReplaceAllString(mal, " ")
		if !strings.Contains(utan, "show:none") {
			t.Errorf("%s: eit boosta byte utan `show:none` rullar sida",
				filepath.Join(fil...))
		}
	}
}

// Kva sida du stend paa skal vera skrive éin stad.
//
// Leidingi bar ein `active`-klasse attaat `aria-current="page"`.
// `hx-boost` teiknar aldri hovudet um att, so klassen vart staaande paa
// den sida du kom fraa medan `aria-current` fylgde med (sideskift.js
// skriv honom). Daa lyste tvo lenkjor «her er du», og LED-stripa la seg
// under den gamle: ho spurde etter `.active` fyrst, og `querySelector`
// gjev fyrste treffet i dokumentrekkjefylgd. Ho traff difor rett berre
// naar ein gjekk attende i rekkja.
func TestBerreEinFasitForKvaSidaEinStendPaa(t *testing.T) {
	nav := lesMal(t, "templates", "components", "navigation", "navigation.html")
	utan := malkomm.ReplaceAllString(nav, " ")

	// Ordet skal ikkje finnast i malen i det heile — korkje som klasse
	// eller som vilkaar. Ei prøva som leita etter ei serskild skriving
	// («active}}») gjekk gjenom med `}}active{{end}}` staaande att, som
	// er nettupp den skrivingi malen hadde.
	for _, daud := range []string{"active", "paa-heim"} {
		if strings.Contains(utan, daud) {
			t.Errorf("leidingi skriv %q attaat aria-current — tvo fasitar driv fraa kvarandre", daud)
		}
	}
	if !strings.Contains(utan, `aria-current="page"`) {
		t.Error("leidingi seier ingen stad kva sida ein stend paa")
	}

	css, err := os.ReadFile(filepath.Join("..", "static", "css", "deler", "05-hovudet.css"))
	if err != nil {
		t.Fatalf("las ikkje stilarket: %v", err)
	}
	utanKomm := komm.ReplaceAllString(string(css), " ")
	for _, daud := range []string{".nav-item.active", ".user-name.active", ".merkeord.paa-heim"} {
		if strings.Contains(utanKomm, daud) {
			t.Errorf("%s er ein veljar att — han maalar den sida du kom fraa", daud)
		}
	}
}

// Vikebladingi skal byta eit stykke, ikkje heile dokumentet.
//
// Pilene var `<button onclick="navigateWeek(±1)">`, og funksjonen sette
// `location.href`. Det var den siste staden paa staden som bygde
// dokumentet upp fraa grunnen — ny <head>, nye skrifter, atten skript um
// att — for ei endring som gjeld eitt stykke. Difor kjendest ho treg
// medan resten er kvikk.
//
// Og difor kunde ho ikkje førehandshentast heller: `preload` treng ei
// lenkja htmx driv. Ein knapp med `onclick` er usynleg for honom.
func TestVikebladingiBytEitStykke(t *testing.T) {
	mal := lesMal(t, "templates", "pages", "timeplan.html")
	utan := malkomm.ReplaceAllString(mal, " ")

	for _, krav := range []string{
		`id="vekebolken"`,         // stykket som skifter
		`hx-target="#vekebolken"`, // pilene peikar paa det
		`preload="mouseover"`,     // og hentar det fyre du trykkjer
		`class="vekepil" href=`,   // pila er ei lenkja
	} {
		if !strings.Contains(utan, krav) {
			t.Errorf("timeplanen manglar %s", krav)
		}
	}

	// Ingen av dei tvo gamle vegane maa koma attende — korkje som
	// `onclick` i malen eller som `location.href` i skriptet.
	skript := lesMal(t, "templates", "components", "common", "timeplan-scripts.html")
	for _, daud := range []string{"navigateWeek", "applyFilters"} {
		if strings.Contains(utan, daud) ||
			strings.Contains(malkomm.ReplaceAllString(skript, " "), daud) {
			t.Errorf("%s er attende — daa lastar vikebladingi heile sida um att", daud)
		}
	}
}

// Eit skript som styrer noko inne i <main> kann ikkje fanga elementet.
//
// dagfokus.js gjorde det: `var plan = document.querySelector(".timeplan")`
// heilt ytst, med eit `if (!plan) return` under. Tvo ting fylgde, og
// baae saag ut som «dagbytet er ustøtt»:
//
//   - Lasta du ei sida utan timeplan, gav uppslaget null og skriptet gav
//     seg for godt. Gjekk du so til timeplanen gjenom leidingi — som
//     byter <main> og ikkje dokumentet — var dagbytet daudt heile økti.
//   - Naar vikebladingi vart eit byte, vart `.timeplan` bytt ut under
//     føtene paa variabelen, som daa peika paa eit element utanfor
//     dokumentet.
//
// Maalt i ein nettlesar: eit trykk paa tysdag gav `data-dag=2` og 17 av
// 26 rader sikta burt ved eit friskt sidekall — og ingen ting i det
// heile paa dei tvo hine vegane. Difor «lyt lasta sida um att fyrst».
func TestSkriptaIMainHeldIkkjePaaGamleElement(t *testing.T) {
	// Kvart av desse styrer noko som bur inne i <main>, og <main> vert
	// bytt: av leidingi ved eit sidebyte, og av vikepilene ved ei
	// blading. Fangar dei elementet ved lasting, peikar dei etterpaa paa
	// noko som ikkje stend i dokumentet.
	//
	// Uppslaget lyt liggja i ein funksjon, og funksjonen lyt liggja i
	// lista til sideskift.js.
	fanga := regexp.MustCompile(`(?m)^[ \t]{0,4}(var|let|const)\s+\w+\s*=\s*document\.(querySelector|getElementById)`)

	for _, namn := range []string{"dagfokus.js", "dokk.js"} {
		js, err := os.ReadFile(filepath.Join("..", "static", "js", namn))
		if err != nil {
			t.Fatalf("las ikkje %s: %v", namn, err)
		}
		kode := string(js)

		if fanga.MatchString(kode) {
			t.Errorf("%s fangar eit element ved lasting — det er burte etter fyrste sidebyte", namn)
		}
		if !strings.Contains(kode, "__sideskift") {
			t.Errorf("%s set ikkje stoda paa nytt etter eit sidebyte", namn)
		}
	}
}
