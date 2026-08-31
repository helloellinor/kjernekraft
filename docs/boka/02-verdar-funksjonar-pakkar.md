# 2. Verdar, funksjonar og pakkar

*Dei minste delane, og ei ekte fil du kann lesa når kapittelet er ute.*

---

Me skal lesa `handsamarar/slag.go`. Ho er 51 liner, ho rører korkje HTTP eller
basen, og ho inneheld nett dei umgrepi dette kapittelet treng. Ved slutten kann
du lesa dei fyrste tvo tridjepartane av henne heilt; resten ventar til
kapittel 7.

Her er det me skal fram til:

```go
package handsamarar

import "strings"

var slagKlassane = []string{
	"slag-fascia",
	"slag-yoga",
	"slag-pilates",
	"slag-reformer",
}

var Slagi = slagUtanKrok()

func slagUtanKrok() []string {
	ut := make([]string, len(slagKlassane))
	for i, k := range slagKlassane {
		ut[i] = strings.TrimPrefix(k, "slag-")
	}
	return ut
}
```

Fire slag trening, skrivne tvo gonger: ein gong med `slag-` framfyre, som er
det stilarket nyttar, og ein gong utan, som er det basen ber.

## 2.1 Pakken

```go
package handsamarar
```

Fyrste lina i kvar Go-fil segjer kva **pakke** ho høyrer til.

Det er `namespace`, men strengare på tvo måtar:

- **Éin pakke per mappe.** Du kann ikkje ha tvo pakkar i same mappa.
- **Ein pakke kann ikkje spreiast yver tvo mapper.** Alle filone i
  `handsamarar/` er den same pakken, og dei ser kvarandre utan `import`.

Difor tyder det noko at `handsamarar` er éin flat pakke med 81 filer: alt der
inne når alt anna. Me kjem attende til kva det kostar.

**Pakken er innkapslingseininga**, ikkje typen. Det finst ingen `private:`.
I staden:

> **Stor fyrstebokstav = synleg utanfor pakken. Liten = ikkje.**

Sjå det i fila: `Slagi` og `SlagKlasse` kann nyttast frå andre pakkar.
`slagKlassane` og `slagUtanKrok` kann ikkje. Det er heile tilgangsstyringi i
Go, og du les henne av namnet åleine.

## 2.2 Å henta noko utanfrå

```go
import "strings"
```

Som `#include`, men på pakkenivå og **aldri relativt**. Det finst ingen
`../`-stiar.

Fleire vert skrivne i ei blokk, og `gofmt` sorterer dei:

```go
import (
	"time"

	"kjernekraft/handsamarar/config"
)
```

Standardbiblioteket øvst, ei tom line, so det som høyrer til denne modulen.
`kjernekraft/handsamarar/config` er full sti frå modulrota — namnet `kjernekraft`
stend i `go.mod`.

Ein ting som kjem til å irritera deg fyrste dagen: **ein `import` du ikkje
nyttar er ein kompilatorfeil**, ikkje ei åtvaring. Det same gjeld ein variabel
du ikkje les. Det kjennest fiendtleg i ein dag, og so sluttar du å skriva daud
kode.

## 2.3 Typen stend etter namnet

Dette er den syntaktiske skilnaden du snublar i mest:

```go
var count int              // C++: int count;
var names []string         // C++: std::vector<std::string> names;
var p *User                // C++: User* p;
```

Namnet fyrst, typen etter. Det ser bakvendt ut i ein dag og so les det seg
sjølv — særleg for funksjonar, der C++ vert stygg og Go ikkje:

```go
func slagUtanKrok() []string
```

«Funksjonen `slagUtanKrok` tek ingen ting og gjev ei liste med strengar.»
Venstre til høgre, i den rekkjefylgdi du les.

### `var` og `:=`

```go
var slagKlassane = []string{ ... }   // på pakkenivå
ut := make([]string, len(slagKlassane))   // inne i ein funksjon
```

`:=` erklærer **og** gissar typen, og han kann berre nyttast inne i ein
funksjon. `var` kann nyttast baae stader, men på pakkenivå hev du ikkje noko
val.

Regelen er kort: **`var` ute, `:=` inne.**

### Grunntypane

`int`, `int64`, `float64`, `bool`, `string`, `byte`, `rune`.

To ting verdt å vita med ein gong:

- **`string` er uforanderleg og UTF-8.** Du kann ikkje endra ein bokstav i han.
  Set du tvo saman, får du ein ny.
- **Det finst ingi implisitt umvending.** Ein `int` vert ikkje ein `int64` av
  seg sjølv; du skriv `int64(x)`. Det er difor du ser `int64(user.ID)` strødd
  yver handsamarane — `ID` er `int`, og basen vil ha `int64`.

## 2.4 Snitt: Go si liste

```go
var slagKlassane = []string{
	"slag-fascia",
	"slag-yoga",
	"slag-pilates",
	"slag-reformer",
}
```

`[]string` er eit **snitt** (slice) av strengar — Go sin `std::vector<T>`, med
éin viktig skilnad me kjem attende til.

Merk komma etter det siste elementet. Det er ikkje valfritt når klammen stend
på eigi line; `gofmt` set det inn.

### `make` og `len`

```go
ut := make([]string, len(slagKlassane))
```

`make([]T, n)` lagar eit snitt med *n* element. `len(x)` gjev talet på element.

Her lagar me altso eit snitt med fire **tomme strengar**, klart til å fyllast.
Merk at dei er tomme og ikkje søppel: i Go er ingen ting nokon gong
uinitialisert. Kva kvar type inneheld når han ikkje er sett, er kapittel 4 —
og det er eit heilt kapittel av di det tyder meir enn det ser ut.

### `append`

Snittet yver hev fast lengd. Skal det veksa, nyttar du `append`:

```go
var ut []string          // tomt, og heilt brukande
ut = append(ut, "yoga")
ut = append(ut, "pilates")
```

**Du lyt taka imot returverdet.** `append` kann verta nøydd til å flytta heile
snittet til ein større stad, so han gjev deg eit nytt snitt attende. Skriv du
berre `append(ut, x)` og kastar svaret, hender ingen ting — og kompilatoren
segjer frå um akkurat det, av di returverdet vert ubrukt.

Dette er den klassiske nybyrjarfeilen i Go, og han er den einaste av dei som
kompilatoren tek for deg.

### `range`

```go
for i, k := range slagKlassane {
	ut[i] = strings.TrimPrefix(k, "slag-")
}
```

`range` gjev deg **tvo** ting: indeksen og verdet. `i` er 0, 1, 2, 3; `k` er
`"slag-fascia"` og so vidare.

Vil du berre ha verdet, skriv du `_` for indeksen:

```go
for _, s := range Slagi { ... }
```

`_` er **den blanke identifikatoren**. Han kastar det du ikkje vil ha, og han
er svaret på regelen i §2.2 um ubrukte variablar: `_` tel ikkje som ubrukt.

Du kjem til å sjå honom overalt.

### Ei felle du skal vita um no

`k` i lykkja yver er ein **kopi**. Endrar du honom, hender ingen ting med
snittet. Difor skriv me `ut[i] = ...` og ikkje `k = ...`.

For strengar spelar det inga rolle. For strukturar gjer det det, og då vert
det ei ekte felle. Kapittel 3 syner korleis huset kjem seg rundt henne.

## 2.5 Kart

Det andre samlingsslaget er kartet — Go sin `std::unordered_map`:

```go
var dagKort = map[time.Weekday]string{
	time.Monday:  "MÅ",
	time.Tuesday: "TY",
}
```

`map[NØKKELTYPE]VERDTYPE`. Du finn det i `handsamarar/timeplanbolk.go`.

Oppslag hev **tvo** former:

```go
s := dagKort[time.Monday]           // "MÅ"
s, finst := dagKort[time.Sunday]    // "", false
```

Den andre er den du vil ha når fråver tyder noko. `finst` er `false` når
nykelen ikkje er der, og `s` er då den tomme strengen.

Tre ting til:

- **`len(m)`** gjev talet på par.
- **`delete(m, nykel)`** tek ut eitt.
- **Rekkjefylgdi ved `range` er tilfeldig med vilje**, so du ikkje uforvarande
  kjem til å byggja på henne. Vil du ha orden, lyt du sortera nyklane sjølv.

Eit kart du ikkje hev laga, kann lesast men ikkje skrivast. Kvifor, og kva som
hender, er kapittel 4.

## 2.6 Funksjonar

```go
func slagUtanKrok() []string {
	...
	return ut
}
```

`func`, namn, argument i parentes, returtype **etter** parentesen, kropp i
klammer.

Argument av same typen kann slåast saman:

```go
func f(a, b int) int
```

Ein funksjon utan returtype gjev ingen ting attende:

```go
func veketalNo() int          // gjev eit tal
func setUpp()                 // gjev ingen ting
```

Og Go kann gjeva **fleire** verdar attende — det er slik feilhandsaming
verkar, og me tek det i kapittel 6.

### Klammene hev ikkje eit val

```go
func f() {        // slik
}

func f()          // ikkje slik — dette er ein kompilatorfeil
{
}
```

Go set inn semikolon på slutten av lina automatisk, so ein `{` som stend
åleine på neste line vert til ein funksjon utan kropp. Difor ser all Go-kode
lik ut, og difor er `gofmt` ikkje stillbar.

## 2.7 Ein pakkevariabel som kallar ein funksjon

```go
var Slagi = slagUtanKrok()
```

Dette er lovleg, og det er verdt eit augneblink: ein pakkenivå-variabel kann
setjast av eit funksjonskall, og kallet gjeng **når programmet startar**, fyre
`main`.

Go finn sjølv ut rekkjefylgdi. Her lyt `slagKlassane` vera ferdig fyre
`slagUtanKrok()` gjeng, og det ordnar han utan at du segjer frå.

I C++ er dette «static initialization order fiasco» — same greia, men der er
rekkjefylgdi millom umsetjingseiningar udefinert. Go definerer henne.

## 2.8 Les fila

No kann du lesa dei fyrste 28 linone av `handsamarar/slag.go` heilt. Opna
henne.

Kommentaren yver `slagKlassane` fortel deg noko koden ikkje kann:

> Written out with the prefix rather than composed from "slag-" + kind:
> scripts/daude-klassar.sh matches class names as they appear in the
> stylesheet, and a name that only exists composed reads as dead.

Altso: ein kunde ha skrive `"slag-" + slag` og spart fire strengar. Grunnen til
at ein ikkje gjorde det, er eit skript som leitar etter daude CSS-klassor med å
søkja etter namnet slik det stend i stilarket. Eit namn som berre finst
samansett, finn han ikkje — og då meiner han klassa er daud.

Det er §1.7 i arbeid: kommentaren ber grunnen, av di grunnen ikkje kann lesast
noko anna stad.

**Resten av fila** — `SlagKlasse`, frå line 38 — nyttar ein funksjon som vert
send som argument til ein annan funksjon. Det heiter ei lukking, og det er
kapittel 7. Lat henne liggja.

---

## Du kann no

- segja kva ein pakke er, og lesa av eit namn um det er synleg utanfrå
- lesa ein funksjonssignatur: namn, argument, returtype
- skilja `var` frå `:=` og vita kvar kvar av dei høyrer heime
- laga og fylla eit snitt med `make`, `len`, `append` og `range`
- slå upp i eit kart, og nytta tvo-verd-forma når fråver tyder noko
- nytta `_` og vita kvifor han finst
- lesa `handsamarar/slag.go` line 1–28

**Neste:** kapittel 3 gjev verdane eit namn og ei form — strukturar — og syner
korleis ein funksjon vert festa til ein type.
