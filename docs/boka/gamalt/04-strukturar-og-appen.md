# 4. Strukturar, mottakarar og App-en

*Kjenn problemet fyrst. Løysingi er fire liner.*

---

## 4.1 Problemet, slik det stod

Fram til for kort tid sidan såg kvar einaste handsamar i Kjernekraft slik ut:

```go
var DB *database.Database        // sett frå main

func ElevTimeplan(w http.ResponseWriter, r *http.Request) {
	weekEvents, err := DB.GetEventsForWeek(måndag, brukarID)
	...
}
```

Ein pakkeglobal. `main` sette honom ein gong, og alle 68 handsamarane greip
etter honom.

Det verka. Det er verdt å segja beint ut: det er ingen ting *gale* med den
koden i drift. Han gjer det same som koden gjer no.

Men prøv å skriva ei prøve for honom.

Du kann ikkje. Ikkje av di det er vanskeleg — av di det er **umogleg**. Ein
prøve treng ei eigi base, og handsamaren kann berre nå den eine globale. Set du
henne, set du henne for alle prøvor i pakken samstundes.

Fylgja stod i repoet: **27 prøvefiler, alle grøne**, og ingen av dei rørte ein
einaste handsamar. Dei prøvde vikerekning, tidrekning, merkeform, malnyklar —
alt som *ikkje* var handsamarar, av di det var det einaste som kunde prøvast.

Og medan dei stod grøne, svara `/api/admin/freeze-requests/approve` med
**501 Not Implemented**. Administrasjonen talde ventande frysingar i
briefingen og baud fram ein knapp som ikkje gjorde noko. Basen hadde svaret
heile tidi — `ApproveFreezeRequest` låg der, ferdig skriven og prøvd.

Ingen såg det, av di det ikkje fanst ein måte å sjå det på.

## 4.2 Strukturar er berre data

Fyrr me fiksar det, treng du tri ting um typar i Go.

Ein `struct` er ein rein samling felt, som ein C++ `struct` utan `private:`:

```go
type App struct {
	DB *database.Database
	Nå func() time.Time
}
```

Ingen tilgangsstyring. Det som skal ut or pakken er skrive med stor
fyrstebokstav — `DB` og `Nå` er synlege utanfrå, eit felt som heitte `db` var
det ikkje. **Pakken er innkapslingseininga**, ikkje typen.

Det er ein av grunnane til at det tyder noko at `handsamarar` er éin flat pakke
med 81 filer: alt der inne ser alt anna.

## 4.3 Metodar er funksjonar med eit argument fyre namnet

Go hev ingen klassor. Han hev funksjonar som er *feste* til ein type:

```go
func (a *App) ElevTimeplan(w http.ResponseWriter, r *http.Request) {
	//   ^^^^^^^^ mottakaren
}
```

`a` er eit heilt vanleg argument som tilfeldigvis stend fyre namnet. **Det
finst ingen `this`**, og ingen ting er implisitt — vil du nå basen, skriv du
`a.DB`.

### Peikar eller verd

```go
func (a *App) ElevTimeplan(...)        // peikarmottakar — deler éin App
func (e Event) Full() bool             // verdmottakar — fær ein kopi
```

Tommelfingerregelen: peikarmottakar med mindre typen er liten og skal lesast.
Kjernekraft gjer nett det — `*App` og `*Database` er peikarar, medan dei små
hjelparane på modellane tek verd:

```go
// models/event.go
func (e Event) Full() bool { return e.Capacity > 0 && e.CurrentEnrolment >= e.Capacity }
```

Og her er ei nyttig innsikt for deg som saknar `const` frå kapittel 1.4: **ein
verdmottakar er det næraste Go kjem.** `func (e Event)` fær ein kopi, so han
*kann ikkje* endra originalen. Det stend ikkje i signaturen som eit lovnad, men
det er sant.

### Ingi arv

Ingen basisklassor, ingen `virtual`, ingen `override`. Vil du ha polymorfi,
nyttar du eit grensesnitt (kapittel 7). Vil du gjenbruka kode, legg du ein
struktur inn i ein annan, og metodane hans vert forfremja.

### Ingen konstruktørar

Dette er viktigare enn det ser ut. `App{}` er **alltid** lovleg og gjev ein
struktur der kvart felt stend på nullverdet sitt. Du kann ikkje hindra det.
Det finst ingen måte å segja «denne typen kann ikkje lagast utan ein base».

Skikken er ein vanleg funksjon med eit avtala namn:

```go
func NyApp(db *database.Database) *App {
	return &App{
		DB: db,
		Nå: func() time.Time { return config.GetInstance().GetCurrentTime() },
	}
}
```

Men det er ein *skikk*, ikkje ein regel. `&App{}` gjeng like godt, og gjev deg
ein App med `nil`-base. Kompilatoren seier ingen ting. Dette er den fyrste
smaken av kapittel 6.

## 4.4 Løysingi

```go
type App struct {
	DB *database.Database
	Nå func() time.Time
}
```

Det er alt. `var DB` er borte; basen er eit felt. Åtti funksjonar vart metodar,
`server.go` byggjer éin `app := handsamarar.NyApp(db)`, og rutone les

```go
r.Get("/elev/timeplan", app.ElevTimeplan)
```

i staden for `handsamarar.ElevTimeplan`.

Og no kann ei prøve gjera dette:

```go
a := &App{DB: prøvebase, Nå: func() time.Time { return fastTidspunkt }}
w := httptest.NewRecorder()
a.ApproveFreezeRequest(w, httptest.NewRequest("POST", "/...?user_id=1", nil))

if w.Code != http.StatusOK { ... }
```

Opna `handsamarar/frysing_test.go` og sjå det heile. Tri prøvor, og den fyrste
av dei ville ha fanga 501-feilen den dagen han vart skriven.

Det er heile poenget med kapittelet: **eit søm er ikkje ein arkitekturregel, det
er ein stad å setja inn noko anna.**

## 4.5 Funksjonar er verdar

`Nå` er ikkje eit tidspunkt. Han er ein **funksjon**, lagra i eit felt:

```go
Nå func() time.Time
```

`func() time.Time` er ein heilt vanleg type i Go, som `int` eller `string`.
Difor kann eit felt bera han, ein funksjon taka han som argument, og ei prøve
byta han ut:

```go
Nå: func() time.Time { return time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC) }
```

I C++ ville du gjort dette med `std::function` eller ein mal-parameter, eller
laga eit `IClock`-grensesnitt med éin metode og ei falsk implementering. Go
treng ingen av delane, av di funksjonstypen sjølv er nok.

Bruken ser slik ut, i `handsamarar/events.go`:

```go
now := a.Nå()
if event.StartTime.Sub(now).Hours() < 2 {
	http.Error(w, "Cannot sign up for classes within 2 hours of start time", http.StatusBadRequest)
	return
}
```

Ei prøve kann no setja klokka til fem minutt fyre timen og sjå at
påmeldingi vert avvist — utan å venta.

Merk at tri stader i huset framleis kallar `time.Now()` beint, og at dei hev
rett til det: innloggingsbremsen i `middleware.go` (veggklokka er det rette for
ein tryggleiksmekanisme), demo-datoane i `payment_api.go`, og frøet i
`test_data.go`. Regelen er ikkje «aldri `time.Now()`» — han er «avgjerder
brukaren ser, gjeng gjenom klokka til huset».

## 4.6 Kva det kosta, og korleis det gjekk

Åtti funksjonar skulde verta metodar, yver fyrti filer, pluss alle 68
innmeldingslinone i `server.go`.

I C++ ville det vore ein ubehageleg ettermiddag: hovudfiler, byggjesystem, og
ei kompilering på tvo minutt millom kvart freistnad.

I Go var det ei mekanisk endring med kompilatoren som søkjeverkty. Endra
signaturen, køyr `go build ./...`, og han listar kvar einaste stad som ikkje
lenger stemmer:

```
handsamarar/dashboard.go:32:17: undefined: AvailableSessions
handsamarar/innsjekk.go:192:6: undefined: innsjekkOpen
./server.go:149:31: undefined: handsamarar.SignUpPage
```

Ein feil per stad, alle på ein gong, på under eit sekund. Du rettar til lista
er tom, køyrer prøvone, og er ferdig.

Dette er kapittel 1.3 sitt argument i arbeid, og det er den beste grunnen til at
programmet vert verande i Go: **eit raskt og strengt byggjesteg gjer store
endringar trygge**, og det er akkurat det ein prototyp som skal ryddast treng.

## 4.7 Kva som *ikkje* vart løyst

Ærleg talt: `App` gjev pakken eit søm, ikkje ei grensa.

`handsamarar` er framleis éin flat pakke med 81 filer. Alt der inne når alt
anna. Det er difor `Session` kann tyda tvo ulike ting i same pakken —
`timeplanbolk.go:27` hev ein `type Session struct` som er ein *time på ein
serskild dag*, medan `session.go` heile vegen handlar um innloggingsøkti.

I tvo pakkar hadde baae namni vore greie. Sjå `ARCHITECTURE.md` §9.1 og §11.4;
det er eit ope spørsmål, ikkje ei avgjerd.

---

## Du kann no

- lesa ein metode og sjå kva mottakaren gjer
- velja millom peikar- og verdmottakar
- segja kvifor `App` finst, og kva han gjorde mogleg
- byta ut klokka i ei prøve
- nytta kompilatoren som søkjeverkty når du endrar ein type
- opna `handsamarar/app.go` og `handsamarar/frysing_test.go` og lesa baae heilt

**Neste — del II:** kapittel 5 gjeng ned i basen, der `database/sql` viser seg
å ikkje innehalda ein einaste database.
