# 4. Nullverdet

*Kva stend i eit felt du ikkje sette? Og kvifor er det eit heilt kapittel?*

---

## 4.1 Regelen

> **Kvar variabel i Go hev alltid ein verd. Det finst ikkje uinitialisert
> minne.**

Erklærer du noko utan å setja det, får det **nullverdet** for typen sin:

| Type | Nullverd |
|---|---|
| `int`, `int64`, `float64` | `0` |
| `string` | `""` |
| `bool` | `false` |
| peikar | `nil` |
| snitt (`[]T`) | `nil` |
| kart (`map[K]V`) | `nil` |
| funksjon | `nil` |
| grensesnitt | `nil` |
| struktur | kvart felt på sitt nullverd, heilt ned |

```go
var e models.Event
// e.Title == ""   e.Capacity == 0   e.StartTime er nullpunktet i tid
```

## 4.2 Kva C++ mista her, og det er ein siger

I C++ er dette udefinert:

```cpp
int n;
std::cout << n;      // kva som helst
```

Ein heil klasse av feil — les-fyre-skriv, uinitialiserte medlemer, `struct`ar
med søppel i seg — er **borte** i Go. Ikkje redusert; borte. Kompilatoren
nullstiller alt.

Det er ein av dei tydelegaste gevinstane ved språket, og det er verdt å seia
fyrst, av di resten av kapittelet handlar um kostnaden.

## 4.3 Kostnaden

Nullverdet er alltid *ein* verd. Men han er ikkje alltid *eit svar*.

Sjå på `Event` frå kapittel 3:

```go
Capacity int      // kor mange plassar timen hev
```

Kva tyder `Capacity == 0`?

Det kann tyda **tri** ulike ting:

1. Timen hev null plassar. (Nesten aldri sant.)
2. Timen set inga eigi kapasitet, og rommet sitt tal gjeld. (Det vanlege.)
3. Spurningi som henta timen **spurde ikkje etter** kapasiteten i det heile.

Typen `int` kann ikkje skilja dei. Ein lesar heller.

Det er ikkje eit tenkt problem. Det hev slege ut tvo gonger i dette
programmet, og kapittel 13 tek heile soga. Her held me oss til kva du treng å
sjå etter.

## 4.4 Namnet på lyten: sentinelverd

Ein **sentinel** er ein vanleg verd som er lånt ut til å tyda «ikkje noko».
`0` for «inga kapasitet». `-1` for «ikkje funne». `""` for «ikkje sett».

Du kjenner det frå C++. Skilnaden er at C++ hev eit svar sidan C++17:

```cpp
std::optional<int> capacity;
if (capacity) { use(*capacity); }
```

Typen ber sjølv skilnaden på «noko» og «ingen ting», og kompilatoren tvingar
deg til å spyrja.

**Go hev ikkje `optional`.** Han hev tvo måtar som gjer det same, og båe er
med vilje mindre lettvinte:

```go
var n *int              // nil tyder fråverande
var n sql.NullInt64     // {Int64: 0, Valid: false}
```

Ein peikar til `int` kann vera `nil`, og då er han fråverande. Er han ikkje
`nil`, er talet han peikar på svaret — jamvel um det er `0`.

Det er `optional`, skrive med det verktyet Go hev.

## 4.5 Slik ser det ut i huset

`models.Event` hev **tvo** kapasitetsfelt, og etter kapittel 3 veit du kvifor
strukturen er slik. No kann du sjå kvifor typane er ulike:

```go
Capacity     int      // det utrekna talet: timen sitt, elles rommet sitt
EigenPlassar *int     // timen sitt eige — nil tyder «set ikkje noko sjølv»
```

`Capacity` er alltid eit tal, og det er greitt: han er *rekna ut*, og eit
utrekna svar finst alltid.

`EigenPlassar` er ein **peikar**, og det er heile poenget. `nil` tyder at timen
ikkje set noko sjølv. Eit tal tyder at han gjer det.

Fyrr dette var han ein `int` der `0` bar den tydingi — ein sentinel — og då
kunde ingen sjå skilnad på «timen set ikkje noko» og «spurningi henta ikkje
feltet».

Kommentaren på feltet segjer det beint ut:

> Han var eit `int` der 0 tydde «ingi eigi». Det er den same feilen som gjer at
> ein spurning utan romkopling gjev 0 plassar og ser rett ut: talet 0 kann ikkje
> segja frå om det er eit svar eller eit fråver. Ein peikar kann.

## 4.6 `nil` er meir enn nullpeikaren

I C++ er `nullptr` berre for peikarar. I Go er `nil` nullverdet for fem ting:
peikar, snitt, kart, funksjon og grensesnitt.

To ting som overraskar:

**Eit `nil`-snitt er brukande.**

```go
var ut []models.Event      // nil
ut = append(ut, e)         // heilt greitt
len(ut)                    // 0 fyre, 1 etter
```

Du treng ikkje å laga det fyrst. Difor kann kvar funksjon i `database/` som
byggjer ei liste byrja med ein tom `var ut []models.Event` og so `append`-a i
veg, utan eit `make` fyrst.

**Eit `nil`-kart kann lesast, men ikkje skrivast.**

```go
var m map[string]int
n := m["x"]        // greitt — gjev 0
m["x"] = 1         // brest
```

«Brest» tyder her ein **panikk**: programmet stoggar, skriv ut kva som hende og
kvar, og — i denne tenaren — vert teke imot og gjort um til eit 500-svar.
Kapittel 6 tek det skikkeleg.

Dette er ei av dei få tingi i Go som brest ved køyring i staden for ved
kompilering, og det tek nye Go-skrivarar kvar gong. Skal kartet skrivast til,
lyt det lagast:

```go
m := make(map[string]int)     // no gjeng skrivingi
```

## 4.7 Ingen konstruktørar — no ser du kvifor det tyder noko

I kapittel 3 sa eg at `models.Event{}` alltid er lovleg. No kann du sjå kva
det kostar.

```go
a := &App{}          // heilt lovleg
a.DB                 // nil
a.Nå                 // nil — og eit kall panikkar
```

Det finst **ingen måte** i Go å segja «denne typen kann ikkje lagast utan ein
base». Ingen privat konstruktør, ingen `= delete`, ingi invariant kompilatoren
handhevar.

Det Go hev i staden er ein skikk: ein vanleg funksjon med eit avtala namn, som
lagar ein ferdig utfylt verd og gjev honom attende.

```go
func NyApp(db *database.Database) *App {
	// ... set felti ...
}
```

(Kroppen hans set eit felt som er ein *funksjon*, og det er kapittel 7. Opna
`handsamarar/app.go` når du kjem dit.)

Men det er ein **skikk**, ikkje ein regel. `&App{}` gjeng like godt og gjev deg
ein App utan base. Kompilatoren segjer ingen ting.

Dette er ein ekte veikskap ved Go samanlikna med C++, og det er ærleg å kalla
det det. Måten ein lever med det på: gjer nullverdet **brukande** der du kann,
og gjer det **openbert gale** der du ikkje kann.

`App` er av det andre slaget. Difor panikkar `brukaren()` med ei melding i
staden for å gjeva `nil` vidare — kapittel 6.

## 4.8 Sjekklista

Når du ser eit felt eller ein returverd, spør:

1. **Er `0` / `""` / `false` eit lovleg svar for dette?**
   Nei → nullverdet er trygt som «ikkje sett».
   Ja → du hev ein sentinel, og du treng ein peikar.

2. **Kann feltet vera usett av di *ingen spurde etter det*?**
   Ja → typen lyg. Anten fyll det alltid, eller lag ein eigen type som ikkje
   hev feltet i det heile.

Punkt 2 er den subtile, og han er grunnen til at desse finst i `database/`:

```go
type Romkollisjon struct {
	Tittel string
	Lærar  string
	Start  time.Time
	Slutt  time.Time
}
```

Ein spurning som berre spør «er rommet uppteke» gav fyrr ein heil
`models.Event` med fire felt fylte og seksten på nullverdet sitt. Typen lova
ein heil time; verdet var det ikkje.

Ein eigen type med berre dei felti spurningi faktisk hentar **kann ikkje lyga
på den måten**. Det er den same medisinen som `*int`, gjeven ein type i staden
for eit felt.

---

## Du kann no

- segja kva kvar type inneheld når han ikkje er sett
- kjenna att ein sentinelverd og vita kvifor han er farleg
- lesa `*int` som «Go sin `std::optional<int>`»
- vita at eit `nil`-snitt er brukande og eit `nil`-kart ikkje er det
- segja kvifor `App{}` er lovleg, og kva ein gjer med det
- opna `models/event.go` og skjøna kvifor `Capacity` og `EigenPlassar` hev
  ulike typar

**Neste:** kapittel 5 tek grensesnitt — og det fyrste dømet er ein type du hev
sett i kvar einaste funksjonssignatur utan å vita at det var eitt: `error`.
