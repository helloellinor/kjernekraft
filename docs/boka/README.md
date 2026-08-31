# Kjernekraft, bygd upp att

*Ei bok um Go, skrivi for ein som kann C++, med dette programmet som døme.*

---

## Kva dette er

Ei lesebok, ikkje ei øvingsbok. Du skriv ikkje eit nytt program undervegs — du
opnar dei filone som alt ligg her, og etter kvart kapittel skal du kunna lesa
ei fil til utan hjelp.

Tri lovnader:

1. **Kvart nytt umgrep opnar med det du alt gjer i C++**, og syner so kva Go
   gjer i staden — og kvifor det vart gjort ulikt.
2. **Éin veg gjenom huset:** timeplanen og påmeldingi, `/elev/timeplan` og
   `POST /api/events/signup`. Dei tvo rører ved alt som finst i programmet.
3. **Kvart kapittel endar med kva du kann no**, og kva fil du skal kunna opna.

Fasiten er programmet som gjeng. Når boki segjer «slik skal det vera», tyder
det «slik *er* det, og her er kvifor».

## Kapitli

### Del I — Språket, med huset som døme

| | | |
|---|---|---|
| 1 | [Orientering](01-orientering.md) | kva du hev, kvar me skal, og korleis boki er lagd upp |
| 2 | [Verdar, funksjonar og pakkar](02-verdar-funksjonar-pakkar.md) | pakke, `:=`, snitt, kart, `range` — endar på `slag.go` |
| 3 | [Strukturar, felt og metodar](03-strukturar-og-metodar.md) | struktur, mottakar, peikar — endar på `models/event.go` |
| 4 | [Nullverdet](04-nullverdet.md) | kva som stend i eit felt du ikkje sette, og kvifor `*int` finst |
| 5 | Grensesnitt *(kjem)* | implisitt stetting; `error` er det fyrste dømet |
| 6 | Feil er verdar *(kjem)* | fleire returverd, `defer`, `panic` |
| 7 | Lukkingar og funksjonar som verdar *(kjem)* | endar på `handsamarar/app.go` |

### Del II — Huset *(kjem)*

| | | |
|---|---|---|
| 8 | `http.Handler` | éi form, og alt er den formi |
| 9 | Mellomvare og rutaren | fire liner chi som er heile løyvemodellen |
| 10 | App-en og sømet | 27 grøne filer og ein knapp som svara 501 |
| 11 | Soknaden si lomme | `context`, privat nykeltype, `iota` |

### Del III — Dataene og utsynet *(kjem)*

| | | |
|---|---|---|
| 12 | Å tala med SQLite | `database/sql` som ein sokkel utan database i |
| 13 | Nullverdet som lyg ★ | heile soga um kapasiteten |
| 14 | Malar, stykke og htmx | escaping etter posisjon |
| 15 | Samtidig utan at du skreiv det | null goroutinar, og likevel treng du lås |

### Del IV — Tryggleik i eige arbeid *(kjem)*

| | | |
|---|---|---|
| 16 | Prøvor | `httptest`, og sømet frå kapittel 10 |
| 17 | Feil, skikkeleg | `%w`, `errors.Is` |
| 18 | Å leggja til noko åleine | sjekklista, ende til ende |

### Tillegg *(kjem)*

- **A.** C++ → Go, oppslagstavla
- **B.** Kompilatorfeil du kjem til å møta, og kva dei tyder
- **C.** Ordtilfanget — peikar til `ARCHITECTURE.md` §8
- **D.** Kvar den verkelege koden skil seg frå boki si, og kvifor

## Fyrr du byrjar

```
go run .
```

og so `http://localhost:18108/innlogging`. Alt i boki viser til kode som ligg i
dette repoet; sti og linetal er skrivne ut kvar gong det trengst.

## Skyldfolk

- **`docs/GO.md`** — det same stoffet som eit oppslagsverk i staden for eit
  kurs. Nyttig når du veit kva du leitar etter.
- **`docs/ARCHITECTURE.md`** — kva programmet er, og kva alt heiter. §2 er
  Go- og chi-maskineriet i kort form; §8 er ordlista.
- **`docs/ARKET.md`** — **bindande** for alt som er visuelt.
