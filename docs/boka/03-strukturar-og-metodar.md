# 3. Strukturar, felt og metodar

*Å gjeva ein verd ei form, og å festa ein funksjon til henne.*

---

Me skal lesa `models/event.go`. Ho er den mest lesne fila i huset — typen `Event`
er timen, og han fer gjenom kvar einaste del av programmet.

Ho rører korkje HTTP eller basen. Ho er berre form.

## 3.1 Ein struktur er berre felt

```go
type Event struct {
	ID          int
	Title       string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
	ClassType   string
	TeacherName string
	Capacity    int
	CurrentEnrolment int
}
```

`type NAMN struct { ... }`. Kvart felt er namn so type, som i kapittel 2.

Det er ein C++ `struct` utan `private:`, og — viktigare — **utan
konstruktørar, utan destruktorar, utan arv og utan `virtual`**. Han er data,
og ingen ting anna.

Å laga ein:

```go
e := models.Event{Title: "Vinyasa", Capacity: 18}
```

Felti du ikkje nemner får nullverdet sitt. Meir um det i kapittel 4 — det er
eit heilt kapittel, av di det er viktigare enn det ser ut.

Merk at du kann skriva `models.Event{}` heilt tomt, og det er **alltid**
lovleg. Det finst ingen måte i Go å segja «denne typen kann ikkje lagast utan
ein tittel».

### Synlegskap, att

Same regelen som i kapittel 2, no på felt:

```go
type Event struct {
	ID    int        // synleg utanfor models
	Title string     // synleg
	// eit felt som heitte `id` hadde ikkje vore det
}
```

Alle felti i `Event` er store, av di både `handsamarar` og malane les dei.

### Merkelappar

```go
ID    int    `json:"id"`
Title string `json:"title"`
```

Strengen etter typen er ein **merkelapp** (struct tag). Han er metadata som
bibliotek les med refleksjon — her `encoding/json`, som nyttar honom til å
avgjera kva feltet heiter i JSON.

Kompilatoren ser ikkje på innhaldet. Skriv du `json:"titel"` med skrivefeil,
er det ingen som segjer frå.

## 3.2 Metodar er funksjonar med ein mottakar

Her er heile forskjellen frå C++, og han er mindre enn du ventar:

```go
func (e Event) Full() bool {
	return e.Capacity > 0 && e.CurrentEnrolment >= e.Capacity
}
```

`(e Event)` fyre namnet er **mottakaren**. Det er eit heilt vanleg argument som
tilfeldigvis stend på ein annan stad.

**Det finst ingen `this`.** Ingen ting er implisitt. Vil du nå eit felt, skriv
du `e.Capacity`.

Kalla ser ut som du ventar:

```go
if event.Full() { ... }
```

### Metoden treng ikkje stå i typen

I C++ stend metodane inne i klassa. I Go stend dei kvar som helst i **den same
pakken** som typen. `models/event.go` har dei attmed kvarandre av di det er
ryddig, ikkje av di det er kravd.

Fylgja: du kann ikkje leggja ein metode på ein type frå ein annan pakke. Vil du
det, lyt du laga din eigen type.

## 3.3 Mottakar på verd

```go
func (e Event) Ledige() int {
	if n := e.Capacity - e.CurrentEnrolment; n > 0 {
		return n
	}
	return 0
}

func (e Event) Full() bool { return e.Capacity > 0 && e.CurrentEnrolment >= e.Capacity }

func (e Event) LengdMin() int { return int(e.EndTime.Sub(e.StartTime).Minutes()) }
```

`(e Event)` — utan `*` — tyder at metoden fær ein **kopi** av strukturen.

To ting fylgjer:

1. **Han kann ikkje endra originalen.** Skriv han `e.Capacity = 5`, endrar han
   kopien sin, og so er han burte.
2. Kopien kostar. `Event` hev tjue felt, so kvart kall kopierer tjue felt.
   Her er det uråd å merka; for store strukturar i tronge lykkjor er det ikkje.

Og her er ei innsikt for deg som saknar `const` frå C++:

> **Ein mottakar på verd er det næraste Go kjem `const`.**

Det stend ikkje i signaturen som eit lovnad kompilatoren handhevar utanfrå, men
det er sant: `func (e Event) Full()` *kann ikkje* endra timen din.

### `if` med ei setning fyrst

```go
if n := e.Capacity - e.CurrentEnrolment; n > 0 {
	return n
}
```

`if SETNING; VILKÅR { }`. `n` finst berre inne i `if`-en — og i ein `else` um
det hadde vore ein.

Dette er mykje nytta i Go, og du kjem til å sjå det mest i feilhandsaming
(kapittel 6). Poenget er å halda ein variabel so nær bruken som råd.

## 3.4 Peikarar

```go
p := &event        // adresse-av, som i C++
p.Title            // IKKJE p->Title
*p                 // eksplisitt avpeiking, sjeldan naudsynt
```

**Go hev ingen `->`.** Han peikar av seg sjølv når du nyttar `.` på ein peikar.
Det er ein ting mindre å hugsa.

Kva som er borte frå C++:

- **Ingi peikararitmetikk.** `p + 1` finst ikkje.
- **Ingen `delete`.** Det vert samla inn.
- **Ingen hengjande peikarar.** Å gjeva att `&lokalVariabel` er **trygt** —
  kompilatoren flytter honom på haugen sjølv. Det er ulovleg i C++ og dagleg
  kost her.
- **Ingi referansar** som eige umgrep. Peikar eller verd, ferdig.

`nil` er nullpeikaren. Han er òg nullverdet for ein del andre ting, og det tek
me i kapittel 4.

## 3.5 Mottakar på peikar

```go
func (a *App) ElevTimeplan(...)          // peikarmottakar
func (e Event) Full() bool               // verdmottakar
```

`*App` tyder at metoden fær **peikaren**, ikkje ein kopi. To grunnar til å velja
det:

1. **Han skal endra noko.** Ein verdmottakar kann ikkje.
2. **Kopien er for dyr**, eller typen skal delast.

Kjernekraft gjer nett dette: `*App` og `*Database` er peikarmottakarar av di
det finst **éin** App og **éi** base som alle deler. Dei små hjelparane på
`Event` tek verd, av di dei berre les.

Tommelfingerregelen: **peikarmottakar med mindre typen er liten og du berre
les.**

Ein ting du ikkje treng å tenkja på: du kann kalla ein peikarmetode på ein
vanleg variabel, og Go tek adressa sjølv. `app.ElevTimeplan(...)` verkar anten
`app` er ein `App` eller ein `*App`.

## 3.6 Ingi arv — og kva du gjer i staden

Ingen basisklassor. Ingen `virtual`. Ingen `override`. Ingen `protected`.

To ting tek plassen:

**Vil du gjenbruka kode:** legg ein struktur *inn i* ein annan utan feltnamn,
og metodane hans vert forfremja:

```go
type Base struct{ ID int }
type Time struct {
	Base            // innlagd
	Title string
}
// no hev Time både .ID og alle metodane til Base
```

Kjernekraft nyttar ikkje dette. Det er verdt å vita at det finst, og verdt å
vera varsam med: det ser ut som arv og er det ikkje.

**Vil du ha polymorfi:** eit grensesnitt. Det er kapittel 5.

## 3.7 Fella frå kapittel 2, no på ekte

I kapittel 2 sa eg at `k` i ein `range` er ein kopi, og at det vert ei ekte
felle for strukturar. Her er henne:

```go
for _, e := range hendingar {
	e.Capacity = 20        // gjer ingen ting
}
```

`e` er ein kopi av strukturen. Du endrar kopien, lykkja gjeng vidare, og
kopien er burte.

Slik gjer huset det i staden — `handsamarar/dashboard_components.go`:

```go
for i := range klippekort {
	k := &klippekort[i]        // peikar inn i snittet
	k.DaysUntilExpiry = int(k.ExpiryDate.Sub(now).Hours() / 24)
	k.IsExpiring = k.DaysUntilExpiry <= 30 && k.DaysUntilExpiry > 0
}
```

`range` med berre indeks, og so `&snitt[i]` for å få ein peikar inn i sjølve
snittet. No stend skrivingane.

Dette er verdt å læra ein gong skikkeleg, av di kompilatoren **ikkje** hjelper
deg. Den fyrste versjonen kompilerer fint og gjer ingen ting.

## 3.8 Les fila

Opna `models/event.go`. Du kann henne no, frå fyrste til siste line — struktur,
felt, merkelappar, tri metodar med mottakar på verd, og enno ein struktur
(`Room`) til slutt.

Legg merke til kommentaren yver `Room`:

> Room er eit rom i studioet. Kapasiteten er ein eigenskap ved rommet og ikkje
> noko ein skriv inn per time.

Den setningi er grunnen til at `Event` hev **tvo** kapasitetsfelt. Kvifor det
er slik, og kvifor det har vore ei feilkjelde tvo gonger, er kapittel 4 og
kapittel 13.

---

## Du kann no

- lesa og skriva ein `struct` med felt og merkelappar
- festa ein metode til ein type med ein mottakar
- velja millom mottakar på verd og på peikar, og segja kvifor
- nytta `&` og `.` på peikarar, og vita kva C++ hev som Go ikkje hev
- skriva `if n := ...; n > 0`
- unngå range-kopi-fella med `&snitt[i]`
- lesa `models/event.go` heilt

**Neste:** kapittel 4 spør kva som eigenleg stend i eit felt du ikkje sette —
og kvifor det er det farlegaste spørsmålet i heile boki.
