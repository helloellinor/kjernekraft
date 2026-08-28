package handlers

import (
	"os"
	"strings"
	"testing"
	"time"

	"kjernekraft/models"
)

// teiknTimestyringa teiknar timebolken paa administrasjonssida.
func teiknTimestyringa(t *testing.T, timar []models.Event) string {
	t.Helper()

	// Malsettet vert leita fram fraa arbeidskatalogen, og under `go
	// test` er han pakken sin. Steg upp til rota fyrst.
	gamal, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(gamal)

	tm := GetTemplateManager()
	tm.ReloadTemplates()
	mal, ok := tm.GetTemplate("pages/admin")
	if !ok {
		t.Fatal("malsettet for administrasjonssida lét seg ikkje lasta")
	}

	seriar := GrupperTimar(timar, time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC))

	var ut strings.Builder
	if err := mal.ExecuteTemplate(&ut, "admin_class_management", map[string]interface{}{
		"Lang": "nn", "Timereglar": seriar, "Siktval": SiktvalFor(seriar),
		"Teachers":  []string{"Leon", "Kristina"},
		"Rooms":     []models.Room{{ID: 1, Name: "Salen", Capacity: 12}, {ID: 2, Name: "Studio", Capacity: 8}},
		"CSRFToken": "x", "IsAdmin": true, "UserName": "prøve",
	}); err != nil {
		t.Fatalf("malen feila: %v", err)
	}
	return ut.String()
}

// tvoTimar gjev tvo seriar paa kvar sin dag, med kvar sin lærar.
func tvoTimar() []models.Event {
	return []models.Event{
		{ID: 1, SerieID: 10, Title: "Yoga", TeacherName: "Leon", RoomName: "Salen", RoomID: 1,
			Capacity: 12, RoomCapacity: 12,
			StartTime: time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)},
		{ID: 2, SerieID: 10, Title: "Yoga", TeacherName: "Leon", RoomName: "Salen", RoomID: 1,
			Capacity: 12, RoomCapacity: 12,
			StartTime: time.Date(2026, 9, 14, 18, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 9, 14, 19, 0, 0, 0, time.UTC)},
		{ID: 3, SerieID: 11, Title: "Pilates", TeacherName: "Kristina", RoomName: "Studio", RoomID: 2,
			Capacity: 8, RoomCapacity: 8,
			StartTime: time.Date(2026, 9, 9, 17, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 9, 9, 18, 0, 0, 0, time.UTC)},
	}
}

// dagliste gjev alle daglistone slegne saman. Kvar serie hev si eigi.
func dagliste(t *testing.T, html string) string {
	t.Helper()
	var bitar []string
	for i := 0; ; {
		j := strings.Index(html[i:], `class="daglista"`)
		if j < 0 {
			break
		}
		byrjing := i + j
		slutt := strings.Index(html[byrjing:], "</ul>")
		if slutt < 0 {
			t.Fatal("daglista er ikkje lukka")
		}
		bitar = append(bitar, html[byrjing:byrjing+slutt])
		i = byrjing + slutt
	}
	if len(bitar) == 0 {
		t.Fatal("daglista stend ikkje der")
	}
	return strings.Join(bitar, "\n")
}

// Seriane stend i den eine spalta og dagane i den andre.
func TestTimestyringaHevTvoSpaltor(t *testing.T) {
	html := teiknTimestyringa(t, tvoTimar())

	if !strings.Contains(html, `class="timestyring"`) {
		t.Fatal("spaltone stend ikkje der")
	}
	seriar := strings.Index(html, `class="timestyring-seriar"`)
	timar := strings.Index(html, `class="timestyring-timar"`)
	liste := strings.Index(html, `class="serieliste"`)
	if seriar < 0 || timar < 0 {
		t.Fatalf("fann ikkje spaltone: seriar=%d timar=%d", seriar, timar)
	}
	if !(seriar < liste && liste < timar) {
		t.Errorf("serielista stend ikkje i den fyrste spalta")
	}
}

// Sikti stend paa overskriftslina; søket stend under henne.
//
// Tri veljarar og eit søkjefelt braut lina i ei halv spalta. Rommet gjekk
// ut — studioet hev tvo, og ingen leitar etter eit rom — og søket flytte
// under lina, der det hev breidd til aa vera ein veg inn i lista.
func TestSiktiStendPaaLinaOgSoketUnder(t *testing.T) {
	html := teiknTimestyringa(t, tvoTimar())

	hovud := strings.Index(html, `class="bolkhovud"`)
	slutt := strings.Index(html[hovud:], "</div>\n        </div>")
	if hovud < 0 || slutt < 0 {
		t.Fatal("bolkhovudet stend ikkje der")
	}
	lina := html[hovud : hovud+slutt]

	for _, vil := range []string{`id="sikt-dag"`, `id="sikt-laerar"`} {
		if !strings.Contains(lina, vil) {
			t.Errorf("%s stend ikkje paa overskriftslina", vil)
		}
	}
	if strings.Contains(lina, `id="seriesok"`) {
		t.Error("søkjefeltet stend paa lina — det braut henne i tvo")
	}
	if !strings.Contains(html, `id="seriesok"`) {
		t.Error("søkjefeltet stend ikkje der i det heile")
	}
	// Rommet er ute for godt.
	if strings.Contains(html, "sikt-rom") {
		t.Error("romsikti stend der endaa")
	}
}

// Ein veljar med eitt einaste val siktar ingen ting, og kjem ikkje.
func TestVeljarMedEittValKjemIkkje(t *testing.T) {
	ein := []models.Event{
		{ID: 1, SerieID: 10, Title: "Yoga", TeacherName: "Leon", RoomName: "Salen", RoomID: 1,
			Capacity: 12, RoomCapacity: 12,
			StartTime: time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)},
	}
	html := teiknTimestyringa(t, ein)

	for _, skal := range []string{`id="sikt-dag"`, `id="sikt-laerar"`} {
		if strings.Contains(html, skal) {
			t.Errorf("%s stend der endaa det berre finst eitt val", skal)
		}
	}
	if !strings.Contains(html, `id="seriesok"`) {
		t.Error("søkjefeltet stend ikkje der")
	}
}

// Overskrifta segjer kva du gjer, ikkje kva timen heiter.
//
// Namnet stod som overskrift *og* i fyrste feltet under. Ei overskrift
// som skifter kvar gong du klikkar er dessutan ein etikett og ikkje ei
// overskrift.
func TestOverskriftaSegjerKvaDuGjer(t *testing.T) {
	html := teiknTimestyringa(t, tvoTimar())

	i := strings.Index(html, `class="timestyring-timar"`)
	slutt := strings.Index(html[i:], "</h2>")
	hovudet := html[i : i+slutt]

	if !strings.Contains(hovudet, t2("nn", "admin.edit_class")) {
		t.Errorf("overskrifta segjer ikkje kva ein gjer: %q", hovudet)
	}
	if strings.Contains(hovudet, "Yoga") {
		t.Error("overskrifta ber namnet paa timen")
	}
}

// Seriesetningi ber heile timen — namn, lærar, rom, dag, klokke, lengd
// og plassar.
func TestRegelsetningiBerHeileTimen(t *testing.T) {
	html := teiknTimestyringa(t, tvoTimar())

	for _, namn := range []string{"tittel", "laerar", "room_id", "vekedag", "klokke", "minutt", "plassar"} {
		if !strings.Contains(html, `name="`+namn+`"`) {
			t.Errorf("feltet %s stend ikkje i seriesetningi", namn)
		}
	}
}

// Plassane stend med rommet sitt tal som utgangspunkt.
//
// Det var eit tomt felt med rommet sitt tal i grått fyrr, so ein kunde
// sjaa skilnaden paa «arva» og «valt» — men det kravde ei hjelpeline for
// aa forklara seg, og eit felt som lyt forklarast er for lurt.
func TestPlassaneArvarRommet(t *testing.T) {
	html := teiknTimestyringa(t, tvoTimar())

	i := strings.Index(html, `name="plassar"`)
	if i < 0 {
		t.Fatal("plassfeltet stend ikkje der")
	}
	byrjing := strings.LastIndex(html[:i], "<input")
	feltet := html[byrjing : i+strings.Index(html[i:], ">")]

	if !strings.Contains(feltet, `value="12"`) {
		t.Errorf("rommet sitt tal stend ikkje i feltet: %q", feltet)
	}
}

// Ein time utan serie kann ikkje endrast paa serienivaa, og kvart felt i
// setningi stend stengt.
func TestUtanRegelStengjerHeileSetningi(t *testing.T) {
	utan := []models.Event{
		{ID: 1, SerieID: 0, Title: "Yoga", TeacherName: "Leon", RoomName: "Salen", RoomID: 1,
			Capacity: 12, RoomCapacity: 12,
			StartTime: time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)},
	}
	html := teiknTimestyringa(t, utan)

	for _, namn := range []string{"tittel", "laerar", "room_id", "vekedag", "klokke", "minutt", "plassar", "skildring"} {
		if !lestengd(t, html, `name="`+namn+`"`) {
			t.Errorf("%s stend open paa ein time som ikkje ber nokon serie", namn)
		}
	}
	if strings.Contains(html, `class="setning lengdsetning"`) {
		t.Error("serielina stend der endaa timen ikkje ber nokon serie")
	}
}

// Grunnen til at felti stend stengde er ein merknad, ikkje eit vink.
//
// Han stod i den veikaste skrifti paa sida medan han forklara heile
// ruta. `.merknad` ber aatvaringslina i margen.
func TestGrunnenTilStengdeFeltErEinMerknad(t *testing.T) {
	utan := []models.Event{
		{ID: 1, SerieID: 0, Title: "Yoga", TeacherName: "Leon", RoomID: 1,
			Capacity: 12, RoomCapacity: 12,
			StartTime: time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)},
	}
	html := teiknTimestyringa(t, utan)

	i := strings.Index(html, t2("nn", "admin.serie_none_hint"))
	if i < 0 {
		t.Fatal("grunnen stend ikkje der")
	}
	byrjing := strings.LastIndex(html[:i], "<p")
	if !strings.Contains(html[byrjing:i], "merknad") {
		t.Errorf("grunnen ber ikkje merknadsforma: %q", html[byrjing:i])
	}
	// Og aatvaringi um avlysing er den same forma.
	j := strings.Index(html, t2("nn", "admin.cancel_warning"))
	if j < 0 {
		t.Fatal("aatvaringi um avlysing stend ikkje der")
	}
	if !strings.Contains(html[strings.LastIndex(html[:j], "<p"):j], "merknad") {
		t.Error("aatvaringi um avlysing ber ikkje merknadsforma")
	}
}

// Serien kann leggjast til, men ikkje skrivast ned.
func TestSerienKannLeggjastTilOgIkkjeSkrivastNed(t *testing.T) {
	html := teiknTimestyringa(t, tvoTimar())

	if !strings.Contains(html, `name="veker"`) {
		t.Fatal("det gjeng ikkje an aa leggja veker til serien")
	}
	i := strings.Index(html, `class="lengdtal"`)
	if i < 0 {
		t.Fatal("serietalet stend ikkje der")
	}
	if strings.Contains(html[i:i+40], "input") {
		t.Error("serietalet er eit felt — daa kann ein skriva timar burt")
	}
}

// Dagrada er tekst med ei rute framfyre, ikkje eit skjema.
func TestDagradaErTekstOgIkkjeEitSkjema(t *testing.T) {
	html := teiknTimestyringa(t, tvoTimar())
	lista := dagliste(t, html)

	if n := strings.Count(lista, `class="dagrad"`); n != 3 {
		t.Errorf("fekk %d dagrader, venta 3", n)
	}
	if n := strings.Count(lista, "<input"); n != 3 {
		t.Errorf("fekk %d felt i daglista, venta 3 — eitt per rad", n)
	}
	if strings.Contains(lista, "<button") {
		t.Error("det stend knappar i daglista")
	}
}

// Dagrada ser ut som serierada i lista attmed: namn, kven, klokke, tal.
func TestDagradaLiknarRegelrada(t *testing.T) {
	lista := dagliste(t, teiknTimestyringa(t, tvoTimar()))

	for _, del := range []string{"dagveke", "dagmaanad", "daglaerar", "dagklokke", "dagplassar"} {
		if !strings.Contains(lista, del) {
			t.Errorf("dagrada ber ikkje %s", del)
		}
	}
	// Namnet ber vekedagen, som serierada gjer.
	if !strings.Contains(lista, t2("nn", "timeplan.monday")) {
		t.Error("dagrada segjer ikkje kva vekedag det er")
	}
}

// Det som verkar paa dei merkte stend yver lista og ikkje under henne.
//
// Ein serie kann vera femti dagar lang, og ein knapp under femti rader er
// ein knapp du lyt rulla for aa finna — etter at du alt hev merkt det du
// vil gjera noko med.
//
// Det var éi setning som bar baade vikaren og avlysingi fyrr. Dei er tvo
// bolkar med kvar si yverskrift no, og daa er det knappane sjølve prøva
// lyt fylgja — regelen galdt aldri eit klassenamn, han galdt kor langt du
// lyt rulla.
func TestHandlinganeStendYverLista(t *testing.T) {
	html := teiknTimestyringa(t, tvoTimar())

	lista := strings.Index(html, `class="daglista"`)
	if lista < 0 {
		t.Fatal("fann ikkje lista")
	}
	for _, handling := range []string{`class="felt val-vikar"`, `class="btn-danger val-avlys"`} {
		i := strings.Index(html, handling)
		if i < 0 {
			t.Errorf("fann ikkje %s", handling)
			continue
		}
		if i > lista {
			t.Errorf("%s stend under lista — daa lyt ein rulla for aa naa han", handling)
		}
	}
}

// Setningi som gjeld dei merkte ber handlingane, og stend stengd i kvile.
//
// Ho bar dato og klokke med fyrr. Dei er ute (30.8.2026): timane i ei
// rekkje gjeng til den same tidi — det er det ei rekkje *er* — so ein
// klokkeveljar spurde um noko som alt var svara, og ein datoveljar gav
// berre meining for éin rad um gongen. Regelen prøva held er den same:
// det som er reiskap stend stengt til noko er merkt, og vert ikkje talt
// som ei endring.
func TestValsetningaGjeldDeiMerkte(t *testing.T) {
	html := teiknTimestyringa(t, tvoTimar())

	for _, felt := range []string{"val-vikar", "val-avlys"} {
		if !strings.Contains(html, felt) {
			t.Errorf("%s stend ikkje i valsetningi", felt)
		}
		if !lestengd(t, html, felt) {
			t.Errorf("%s stend open endaa ingen dag er merkt", felt)
		}
	}
	for _, burte := range []string{"val-dato", "val-klokke"} {
		if strings.Contains(html, burte) {
			t.Errorf("%s stend att; timane i ei rekkje gjeng til den same tidi", burte)
		}
	}
	// Reiskapen er ikkje ei endring: dokka skal ikkje telja det du skriv
	// i verktyet som noko som skal lagrast.
	i := strings.Index(html, "val-vikar")
	byrjing := strings.LastIndex(html[:i], "<input")
	if !strings.Contains(html[byrjing:i+strings.Index(html[i:], ">")], "data-ikkje-endring") {
		t.Error("val-vikar vert talt som ei endring")
	}
}

// Dei felti serien eig finst berre éin stad.
func TestKvartFeltFinstBerreEinStad(t *testing.T) {
	lista := dagliste(t, teiknTimestyringa(t, tvoTimar()))

	for _, namn := range []string{`name="laerar"`, `name="klokke"`, `name="minutt"`, `name="vekedag"`} {
		if strings.Contains(lista, namn) {
			t.Errorf("%s stend i daglista og — serien eig det feltet", namn)
		}
	}
}

// Skjemaet lagrar gjenom dokka, ikkje felt for felt.
//
// Det gjorde ti kall, eitt per felt, sendt i det du forlét feltet. Daa
// fanst ikkje «eg ombestemte meg», og ei endring som heng saman kunde
// verta halvvegs teki imot.
func TestSkjemaetLagrarGjenomDokka(t *testing.T) {
	html := teiknTimestyringa(t, tvoTimar())

	if !strings.Contains(html, "data-endringar") || !strings.Contains(html, "data-dokk=") {
		t.Error("skjemaet er ikkje kopla til dokka")
	}
	if !strings.Contains(html, `class="dokk endringsdokk"`) {
		t.Error("dokka stend ikkje der")
	}
	if !strings.Contains(html, t2("nn", "profile.save")) {
		t.Error("dokka ber ingen lagreknapp")
	}
	if !strings.Contains(html, "data-angra") {
		t.Error("dokka ber ingen veg attende")
	}
	// Ingen av dei gamle einfeltskalli skal stå att i skriptet.
	for _, veg := range []string{
		"/api/admin/rule/tittel", "/api/admin/rule/teacher", "/api/admin/rule/klokke",
		"/api/admin/rule/lengd", "/api/admin/rule/rom", "/api/admin/rule/vekedag",
		"/api/admin/rule/utvid", "/api/admin/class/avlys-fleire", "/api/admin/class/flytt",
	} {
		if strings.Contains(html, veg) {
			t.Errorf("%s vert framleis kalla — lagringi skal gaa gjenom dokka", veg)
		}
	}
	if !strings.Contains(html, "/api/admin/serie/lagra") {
		t.Error("skjemaet peikar ikkje paa lagringi")
	}
}

// Talet i serierada les seg som ord, ikkje som eit kryss.
func TestTaletILestSomOrdOgIkkjeSomKryss(t *testing.T) {
	html := teiknTimestyringa(t, tvoTimar())

	i := strings.Index(html, `class="serietal"`)
	if i < 0 {
		t.Fatal("talet stend ikkje i serierada")
	}
	rada := html[i : i+strings.Index(html[i:], "</span>")]

	if strings.Contains(rada, "×") {
		t.Error("krysset stend der endaa — det tyder «slett» ein annan stad i den same administrasjonen")
	}
	if !strings.Contains(rada, t2("nn", "admin.times")) {
		t.Errorf("talet seier ikkje kva det tel: %q", rada)
	}
}

// Eitt er ikkje fleire. «1 gonger» er ikkje eit ord.
func TestEittOgFleireFaarKvartSittOrd(t *testing.T) {
	ein := []models.Event{
		{ID: 1, SerieID: 10, Title: "Yoga", TeacherName: "Leon", RoomID: 1,
			Capacity: 12, RoomCapacity: 12,
			StartTime: time.Date(2026, 9, 7, 18, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)},
	}
	eitt := teiknTimestyringa(t, ein)
	if !strings.Contains(eitt, ">1 "+t2("nn", "admin.times_one")+"<") {
		t.Errorf("éin time seier ikkje «1 %s»", t2("nn", "admin.times_one"))
	}
	if strings.Contains(eitt, "1 "+t2("nn", "admin.times")+"<") {
		t.Error("«1 gonger» stend der")
	}
	if !strings.Contains(teiknTimestyringa(t, tvoTimar()), ">2 "+t2("nn", "admin.times")+"<") {
		t.Errorf("tvo timar seier ikkje «2 %s»", t2("nn", "admin.times"))
	}
}

// Steget til ein ny vekedag gjeng alltid framover.
func TestVekedagsstegGjengAlltidFramover(t *testing.T) {
	for _, p := range []struct {
		fraa time.Weekday
		maal int
		vil  int
	}{
		{time.Monday, 3, 2},
		{time.Monday, 0, 6},
		{time.Saturday, 1, 2},
		{time.Monday, 1, 0},
		{time.Sunday, 6, 6},
	} {
		if fekk := vekedagssteg(p.fraa, p.maal); fekk != p.vil {
			t.Errorf("vekedagssteg(%v, %d) = %d, venta %d", p.fraa, p.maal, fekk, p.vil)
		}
	}
}
