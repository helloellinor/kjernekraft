# Bandbreidda — kva ei sida kostar, og kva som er verdt aa gjera

Maalt 31.8.2026 mot ein tenar med `middleware.Compress(5)` paa, innlogga
brukar, `/elev/timeplan`. Alle tal er **det som gjeng yver netet**, ikkje
kjeldefiler paa disk. Sjaa «Maateb» nedst — skilnaden er stor nok til aa
snu ei tilraading.

## Kva ei fyrste lasting kostar

| | yver netet | del |
|---|---|---|
| skrifter (tvo woff2) | 133 244 B | **65 %** |
| htmx.min.js | 16 485 B | 8 % |
| stilarket | 18 078 B | 9 % |
| dei sytten hine skripti | ~26 000 B | 13 % |
| sjølve sida (HTML) | 9 120 B | 4 % |
| **til saman** | **209 209 B** | |

Og til samanlikning, den tyngste sida i huset:

| | raatt | gzip | brotli |
|---|---|---|---|
| `/admin?fane=timeplan` | 1 208 883 | 37 008 | 20 649 |

Ein million tekn kollapsar til sju og tretti kilobyte av di dei tvo og
sytti skjemai er nesten like. Det er ikkje eit bandbreiddeproblem.

## Kva ei *andre* lasting kostar — og kvifor ho ikkje er null

Ho skulde vore null. Kvar ressurs hev alt eit innhaldsavtrykk i adressa
(`?v=1aba632d0c2d`), som er heile grunnlaget for aa bufra for alltid.
Men:

| ressurs | `Cache-Control` som vert sendt |
|---|---|
| skrifter | `public, max-age=31536000, immutable` ✅ |
| stilarket | `public, max-age=300` — fem minutt |
| **alle skripti** | **ingen** |

Utan `Cache-Control` gjeter nettlesaren sjølv, og gjetinga endar som
oftast i ein spurnad med `If-Modified-Since`. Attende kjem `304 Not
Modified` med nesten ingen kropp — men **atten rundturar**, ein per
skript, fyre sida kann teikna ferdig. Paa eit nett med 40 ms tur-retur er
det ein knapp sekund som ikkje syner i noko byte-tal.

## Kva som er verdt aa gjera, etter vinst

**1. Bufra dei avtrykte adressone for alltid.** `max-age=31536000,
immutable` paa stilarket og skripti, slik skriftene alt hev. Avtrykket i
adressa gjer det trygt: endrar fila seg, endrar adressa seg. Ei andre
lasting fell fraa ~60 kB og atten rundturar til **null**. Faa liner i
`stilark.go` og filtenaren. Minst arbeid, størst vinst.

**2. Slaa dei atten skripti saman til eitt.** Dei vert alle send til alle
sidone alt (sjaa `base.html`), so det er ingi utveljing aa mista. Éin
spurnad i staden for atten paa fyrste lasting; med (1) er andre lasting
null uansett. Rekkjefylgda lyt haldast — htmx fyrst, csrf.js etter.

**3. Brotli attaat gzip.** Maalt paa desse filone:

| | gzip | brotli | spart |
|---|---|---|---|
| stilarket | 18 078 | ~14 995 | 17 % |
| skripti | 42 659 | 37 791 | 11 % |
| `/admin?fane=timeplan` | 37 008 | 20 649 | **44 %** |

Mest verdt paa HTML, som er det einaste som ikkje kann bufrast.
`middleware.Compress` tek ein eigen kodar; brotli er ikkje med i chi av
seg sjølv.

**4. Skriftene, med atterhald.** 133 kB er to tridjepartar av fyrste
lastingi, so det er her tale ligg. Men dei er alt skorne ned til 336
teikn kvar, og UnionGothic er variabel med tvo aksar (`wdth` 50–200,
`wght` 400–1000) som huset *brukar* — `--vidd-yver` og
breidd-tyder-rang er ikkje pynt. Ein kann skjera til dei teikni som
verkeleg stend paa sidone (kring 120?) og kanskje henta 30 kB. Ein kann
ikkje flata dei ut til faste snitt utan aa missa formsproget.

## Kva som *ikkje* er verdt aa gjera

**Teikna stykke paa tenaren i staden for `hx-select`.** Det ser ut som
den store vinsten: kvart fanebyte og kvart vikesteg hentar heile sida og
kastar det meste. Men maalt:

| | heile svaret | stykket | kasta |
|---|---|---|---|
| fanebyte i administrasjonen | 36 999 B | ~34 272 B | **7 %** |
| vikesteg i timeplanen | 8 270 B | ~5 808 B | **30 %** |

Stykket *er* mest heile sida. Sju prosent er ikkje verdt eit
fragment-endepunkt per side, og tretti prosent av aatte kilobyte er tvo
og eit halvt. `hx-select` kostar lite her av di sidone er magre kring
innhaldet.

## Maateb

Maal svaret, ikkje kjelda. To gonger i dag tok eg feil av nettupp det:

* `/admin?fane=timeplan` er 1,2 MB paa disk og 37 kB yver netet. Eg baud
  fram eit risikabelt umskrive av eit skript paa 300 liner for aa retta
  eit bandbreiddeproblem som ikkje fanst.
* Stilarket er 318 kB i `deler/`, og 70 % av det er kommentarar. Eg var i
  ferd med aa tilraada at dei vart strokne — men `byggStilark()` strøk
  dei alt (`utanKommentar`, `stilark.go:127`). Det som vert sendt er 94
  kB raatt, 18 kB pakka.

```sh
curl -s -H "Accept-Encoding: gzip" -o /dev/null -w '%{size_download}\n' <adressa>
```
