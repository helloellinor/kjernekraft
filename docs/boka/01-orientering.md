# 1. Orientering

*Kva du hev, kvar me skal, og korleis boki er lagd upp.*

---

## 1.1 Kva som stend her

Kjernekraft er eit ferdig program. Det gjeng, det er rett, og det er ditt eige.
Det er verdt å segja fyrst, av di resten av boki peikar på ting som kunde vore
betre, og då er det lett å gløyma kva som alt er sant:

| | |
|---|---|
| 68 handsamarar | ein per rute |
| 109 spurningar | alle i éin pakke, ingen SQL utanfor honom |
| 53 malar | ei sida per rute, og bolkar dei deler |
| 35 stilarkfiler | som vert til éi fil for lesaren |
| 708 nyklar × 3 mål | nynorsk, bokmål, engelsk |
| 18 tabellar | og éin til som kjem etterpå |

Fasiten er ikkje ein spesifikasjon i eit skriv. Fasiten er programmet som
gjeng. Når boki segjer «slik skal det vera», tyder det «slik *er* det, og her
er kvifor».

## 1.2 Set han i gang, fyrst av alt

```
go run .
```

So `http://localhost:18108/innlogging` i lesaren.

Det er heile byggjesteget. Ikkje eit steg *fyre* — det er det heile. Ingen
CMake, ingen vcpkg, ingi hovudfiler, ingen `#include`-vakter, ingi
framleis-erklæringar.

Vil du sjå at tenaren set saman stilarket sitt sjølv, hent det beint:

```
curl -s http://localhost:18108/static/css/kjernekraft.css | head -30
```

Det er 35 filer under `static/css/deler/`, sette saman i talrekkjefylgd av
`handsamarar/stilark.go`, medan soknaden gjeng. Endra ei av dei, last sida på
nytt, og endringi er der.

## 1.3 Tri kommandoar, og so er det ikkje meir

```
go build ./...      byggjer alt
go test ./...       køyrer alle prøvone
gofmt -w .          rettar formateringi
```

`./...` tyder «denne mappa og alt under henne».

Kor lenge tek det? Heile programmet byggjer på under eit sekund, og alle
prøvone gjeng på kring tri.

Det er ikkje ein detalj. Det er ein føresetnad for arbeidsmåten resten av boki
kviler på: når byggjesteget er raskare enn tanken, vert kompilatoren eit
**søkjeverkty**. Du endrar ein type, køyrer `go build ./...`, og han listar
kvar einaste stad som ikkje lenger stemmer — alle på ein gong. Me nyttar det
med vilje seinare.

## 1.4 Kvar me skal

Boki fylgjer éin veg gjenom huset: **timeplanen og påmeldingi**. Adressone er
`/elev/timeplan` og `POST /api/events/signup`.

Dei tvo rører ved alt som finst i programmet:

- rutaren og løyvet — du lyt vera innlogga
- `App`-en, som ber basen og klokka
- fire spurningar mot basen
- kapasiteten, som er det vanskelegaste talet i heile programmet
- tvo malar: ei heil sida og eit stykke
- tri slag feil: 404, 400 med grunn, og 500
- prøvor av baae slagi

Du skal ikkje skriva eit nytt program. Du skal opna dei filone som alt ligg
her, og etter kvart kapittel skal du kunna lesa ei fil til utan hjelp.

## 1.5 Korleis boki er lagd upp

Éin regel, og han er strengare enn han ser ut:

> **Ingen ting vert nytta fyre det er innført.**

Fyrste utkastet av desse kapitli braut den regelen elleve gonger. Det verste
dømet: kapittel 2 lærde deg `http.Handler` — som *er* eit grensesnitt — seks
kapittel fyre boki fortalde kva eit grensesnitt er. Det er retta no, og
prisen er at huset kjem seint: du ser ikkje ein handsamar fyre kapittel 8.

Til gjengjeld skal ingen ting i dei kapitli vera uforklara når du kjem dit.

**Del I** (kapittel 2–7) er språket, men aldri med oppdikta døme — kvart
kapittel endar på ei ekte fil frå dette repoet som du då kann lesa heilt.
**Del II** (8–11) er sjølve huset. **Del III og IV** er dataene, utsynet og
prøvone.

Sjå `REKKJEFYLGD.md` um du vil ha heile kartet, med grunnane.

## 1.6 Ei åtvaring um målet i koden

Denne boki er på nynorsk. Kommentarane i koden er ikkje det — ikkje alle.

Opna `server.go` og les kommentaren yver `standardPort`:

> 8080 is the port most things take first, so "address already in use" is not
> an accident, it is the normal state. Worse is when something *else* is
> already listening: the page answers, but it is not this page, and you sit
> wondering why your changes do not show.
>
> The number is 108, the beads on a mala and the sun salutations in a full
> round. A port you remember is a port you do not guess at.

Engelsk. Men i `handsamarar/vike.go`, som du opnar i kapittel 6, stend det
nynorsk.

**Kommentarane stend midt i eit skifte.** I `handsamarar` blandar 22 filer dei
tvo måli, 29 er reint nynorske og 17 reint engelske. Det er ikkje ein tilstand
nokon valde — det er ei umsetjing som stogga på halvvegen. `ARCHITECTURE.md`
§9.4 og R1 handlar um kva ein skal gjera med henne.

Fram til det er avgjort: ver budd på baae, og lat deg ikkje uroa av at same
slaget kommentar er skriven på tvo mål i tvo filer attmed kvarandre.

## 1.7 Regelen kommentarane fylgjer

Same kva mål dei stend på, gjer dei det same: **dei segjer *kvifor*, ikkje
*kva*.**

Koden segjer kva. Ein kommentar som segjer det same som koden er ei fil å
halda ved like tvo gonger. Ein kommentar som ber grunnen er det einaste minnet
avgjerdi hev.

Portkommentaren yver er eit døme: han fortel ikkje at talet er 18108 — det
stend i koden. Han fortel kvifor det ikkje er 8080.

Du kjem til å sjå dette att i kvar einaste fil, og det er den beste eigenskapen
kodebasen hev.

---

## Du kann no

- byggja, prøva og formatera programmet
- setja tenaren i gang og sjå sida
- segja kva boki skal gjenom, og i kva rekkjefylgd
- vita kvifor du møter tvo mål i kommentarane

**Neste:** kapittel 2 tek dei minste delane — verdar, funksjonar og pakkar — og
endar med at du kann lesa halve `handsamarar/slag.go` utan hjelp.
