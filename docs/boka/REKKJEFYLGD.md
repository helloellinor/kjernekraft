# Rekkjefylgd — kva som lyt koma fyre kva

*Retting etter fyrste gjenomlesingi. Del I vart skriven i feil rekkjefylgd.*

## Feilen

Boki vart ordna etter **arkitektur** — tenar, so rutor, so App — av di det er
slik programmet er lagdelt. Men det er ikkje slik ein lesar byggjer upp
kunnskap. Fylgja er at umgrep vert nytta fyre dei er forklara:

| Umgrep | Fyrst nytta | Forklara |
|---|---|---|
| struktur | kap 1 | kap 4 ✗ |
| peikar | kap 1 | kap 4 ✗ |
| metode / mottakar | kap 1 | kap 4 ✗ |
| funksjon som verd | kap 1 | kap 4 ✗ |
| nullverd | kap 1 | kap 6 ✗ |
| defer | kap 1 | kap 5 ✗ |
| **grensesnitt** | **kap 2** | **kap 7** ✗ |
| error | kap 2 | kap 5 ✗ |
| kart (map) | kap 2 | aldri ✗ |
| lukking (closure) | kap 3 | aldri ✗ |
| typepåstand | kap 3 | aldri ✗ |
| iota | kap 3 | aldri ✗ |

Elleve av fjorten. Den verste er `http.Handler`: kapittel 2 lærer deg det
viktigaste **grensesnittet** i Go, seks kapittel fyre boki fortel kva eit
grensesnitt er.

Og kapittel 1 §1.5 opnar med `app.go` som den fyrste fila lesaren ser. Ho er 37
liner, men ho inneheld struktur, peikar, felt som er ein funksjon, samansett
literal, adresse-av, nullverd og «ingen konstruktørar» — sju umgrep på ein
gong, ingen av dei forklara.

## Regelen som rettar det

> **Ingen ting vert nytta fyre det er innført.** Kvart kapittel endar på ei
> verkeleg fil frå repoet som lesaren no kann lesa heilt.

Det tyder at språket kjem fyrst og huset etterpå — men kvart språkkapittel vert
grunna i ei ekte fil, ikkje i oppdikta døme. Kapittel 1 syner programmet i gang,
so lesaren ser kvar han skal, og so byggjer me upp dit.

## Ny rekkjefylgd

### Del I — Språket, med huset som døme

| | Kapittel | Endar på fila | Innfører |
|---|---|---|---|
| 1 | Orientering | `go run .` | pakke, dei tri kommandoane, kvar me skal |
| 2 | Verdar, funksjonar og pakkar | `handsamarar/slag.go` | `:=`/`var`, typar etter namn, funksjonssignatur, snitt, `append`, **kart**, `range`, `_`, streng |
| 3 | Strukturar, felt og metodar | `models/event.go` | struktur, felt, verdmottakar, peikar, peikarmottakar, stor/liten fyrstebokstav |
| 4 | Nullverdet | `models/event.go` att | nullverd, `nil`, ingen konstruktørar, fyrste smaken av sentinel-problemet |
| 5 | Grensesnitt | `database/timekolonnar.go` | grensesnitt, implisitt stetting, `error` som det fyrste dømet, `radskannar` |
| 6 | Feil er verdar | `handsamarar/vike.go` + ein liten basefunksjon | fleire returverd, `if err != nil`, `defer`, `panic` kort |
| 7 | Lukkingar og funksjonar som verdar | `handsamarar/app.go` + resten av `slag.go` | funksjonstypar, lukkingar, `App.Nå` |

**Andre rettingi:** kart høyrde heime i kapittel 2 attmed snitt — dei er baae
grunnleggjande samlingar, og kapittel 4 treng kartet for å syna at eit
`nil`-kart ikkje kann skrivast til. Kapittel 7 er difor reine lukkingar no.

**Retting i sjølve planen:** `error` *er* eit grensesnitt —
`type error interface { Error() string }`. Lærer ein feilhandsaming fyre
grensesnitt, gjer ein den same feilen um att. Grensesnitt treng berre metodar
(kap 3), so dei kann koma tidleg, og `error` vert det fyrste dømet på eitt.

### Del II — Huset

| | Kapittel | Innfører |
|---|---|---|
| 8 | `http.Handler` | *no* er det berre eit grensesnitt du kjenner. `HandlerFunc` |
| 9 | Mellomvare og rutaren | lukking som gjev ein `Handler`; chi; `r.Group` og løyvemodellen |
| 10 | App-en og sømet | kvifor handsamarane er metodar; 501-soga; prøvbarleik |
| 11 | Soknaden si lomme | `context`, privat nykeltype, `iota`, `WithUser` |

### Del III — Dataene og utsynet

| | Kapittel |
|---|---|
| 12 | Å tala med SQLite |
| 13 | Nullverdet som lyg ★ |
| 14 | Malar, stykke og htmx |
| 15 | Samtidig utan at du skreiv det |

### Del IV — Tryggleik i eige arbeid

| | Kapittel |
|---|---|
| 16 | Prøvor |
| 17 | Feil, skikkeleg |
| 18 | Å leggja til noko åleine |

## Kva som skjer med det som alt er skrive

Ingen ting av stoffet gjeng tapt — det skal flyttast:

| Gamalt | Nytt |
|---|---|
| kap 1 §1.1–1.3, §1.6–1.7 | vert nye kap 1, mest urørt |
| kap 1 §1.4 (C++-lista) | delt upp og strødd der kvart punkt høyrer heime |
| kap 1 §1.5 (`app.go`) | vert nye kap 6, når lesaren hev det som trengst |
| kap 2 (`Handler`) | vert nye kap 8 |
| kap 3 (mellomvare, chi) | vert nye kap 9; `context`-bolken vert kap 11 |
| kap 4 (App, mottakarar) | delt: mottakarar → kap 3, App-soga → kap 10 |

Åtvaringi um målet i koden (gamle §1.6) vert verande i kapittel 1.
