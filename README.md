# Kjernekraft

Studiosystem for eit yogastudio i Oslo: timeplan, påmelding, medlemskap
og klippekort, med ei administrasjonsside for dei som driv huset.

Go og [chi](https://github.com/go-chi/chi) på serversida, SQLite som
base, `html/template` og htmx på framsida. Ingen byggjesteg for
frontenden — malar, CSS og JS blir lesne frå disken.

## Kom i gang

```sh
export KJERNEKRAFT_SESSION_KEY=$(openssl rand -base64 32)
export KJERNEKRAFT_ENV=development

go run .          # http://localhost:8080
```

Tenaren nektar å starte utan `KJERNEKRAFT_SESSION_KEY`. `KJERNEKRAFT_ENV`
er valfri, men alt anna enn `development` — det tome med — blir handsama
som drift: `Secure` på kakene, og testdata-rutene svarar 404. Sjå
`.env.example`.

Basen blir laga og migrert ved oppstart, så det trengst ikkje noko
oppsett før første køyring.

Til dagleg arbeid er `./køyr` betre enn `go run .` — sjå
[Utviklingstenaren](#utviklingstenaren) nedanfor.

## Det som held påstandane sanne

Dette er den eine regelen som er verd å lese før resten.

Ein påstand i eit dokument veit ikkje når han blir usann. Difor er dei
påstandane som betyr noko her ikkje skrivne ned i prosa — dei er prøver
som brest, og namna deira er sjølve påstanden:

```
TestAlleMaalHevDeiSameNyklane          språkfilene har ikkje drive frå kvarandre
TestMalaneBedBerreUmNyklarSomFinst     ingen mal ber om ein nykel som ikkje finst
TestIngenUmsetjingErTom                ingen nykel står att som tom streng
TestStandardmaaletErNynorsk            nynorsk er det du får utan å velje
TestDeiTvoMyrkeBlokkaneErSamde         kvart myrkt token har eit ljost motstykke
TestSundagenHoyrerTilVikaSomGjengUt    vikerekninga er samd med seg sjølv
TestHelsingaSegjerDenSameKlokkaSomTimen  helsinga flyt ikkje to timar
```

`go test ./...` køyrer alle. Legg du til ein streng i ein mal utan å
setje han i alle tre språkfilene, fell `./handlers` — han renderer ikkje
berre bort nykelen i stilla.

Skriv du ein ny påstand i dette dokumentet, skriv prøva òg. Elles er han
sann berre den dagen han blei skriven.

## Tilgang

Autorisasjonen ligg i rutaren, ikkje i handsamarane. `server.go` set opp
fire grupper, og det er plasseringa til ruta som gjev løyvet:

| Gruppe | Middleware | Inneheld |
|---|---|---|
| open | — | `/`, `/signup`, `/terms`, `/innlogging`, `/logout`, `/static/*` |
| innlogga | `RequireAuth` | `/elev/*` og dei `/api/*`-endane ein elev nyttar |
| administrasjon | `RequireAdmin` | `/admin`, `/api/admin/*`, `/users/assign-role` |
| utvikling | `RequireDevelopment` | testdata-rutene og `/arket`; 404 utan `KJERNEKRAFT_ENV=development` |

To reglar held dette saman:

- **Rolla kjem frå basen, ein gong per soknad.** Øktkaka held ein
  brukar-id og ikkje noko meir. `WithUser` hentar brukaren, og
  `RequireAdmin` les rolla av han. Ei rolla i kaka er ei rolla nettlesaren
  kan skrive om.
- **Kvar mutasjon ber ein CSRF-nykel.** `CSRF` legg han i økta og speglar
  han i ei lesbar kake; `static/js/csrf.js` heng han på htmx- og
  `fetch`-kall, og skjema ber han i eit gøymt `csrf_token`-felt. Ein
  `POST` utan han får 403.

Ei rute lagd utanfor ei gruppe er open. Det er den eine tingen som må
sitje i denne fila.

Rollene systemet kjenner står i `database/roller.go`: `admin` og
`teacher`. Namna er engelske av di dei låg i basen frå før; det synlege
ordet kjem gjennom `{{t}}`.

## Huset

```
server.go              rutene, og dermed tilgangen
handlers/              ein fil per område; malane under templates/
  ├── modules/         Go-sida av malmodulane (klippekort, charges, admin-stats)
  ├── config/          innstillingar, m.a. tidssona
  └── templates/
      ├── layouts/     base.html — hovud, <main>, botnline
      ├── pages/       ei fil per side
      ├── components/  attbrukande bitar (navigasjon, dagmerke, week-grid)
      └── modules/     funksjonsbolkar etter område (dashboard, admin, membership)
database/              spurningar og migrering; ei fil per område
models/                dei formene som går mellom base og mal
locales/               nn, nb, en
static/css|js|img|fonts
scripts/               eingongsskript: seed (medlemskap og prisar), testbrukar
cmd/, test/            prøvedata og prøvene som treng ei eiga base
docs/                  teikningane og stilboka
```

Skiljet mellom `components/` og `modules/`: ein **komponent** er ein bit
fleire sider treng (navigasjonen, språkveljaren, dagmerket). Ein **modul**
høyrer til eitt forretningsområde (klippekortbolken, folkelista). All stil
ligg under `static/css/deler/` — ein bolk per fil, med eit tal framfor
namnet som set rekkjefylgda (`00-token` før alt som les tokena). Dei blir
sette saman til éi adresse, `/static/css/kjernekraft.css`, av
`handlers/stilark.go`. Det finst ikkje stilmalar per komponent.

Nokre delar det er verd å vite om:

- **`handlers/timeregel.go`** — timeplanen er reglar, ikkje enkelttimar.
  «Yoga med Leon, måndag 18:00» er regelen; timane er utslaga hans og ber
  regel-id-en sin. Administrasjonen endrar regelen.
- **`handlers/tid.go`** — tidene i basen er veggklokka i Oslo, lagra utan
  sone, og drivaren les dei som UTC. `veggklokka` byggjer tida opp att av
  tala som faktisk står der. Rekn ikkje om ei tid frå basen utan han.
- **`handlers/vike.go`** — vikerekninga, teken ut av timeplanhandsamaren
  så ho kan prøvast. Ho har eit motstykke i `static/js/timeplan-veke.js`
  som går andre vegen; dei to må vere samde.
- **`handlers/merkeform.go`** og **`handlers/aktivitet.go`** — figurar blir
  rekna ut i Go og ikkje i malen. Ein mal syner kva som er i figuren; kvar
  kvart punkt ligg er ein annan slags kunnskap.
- **`handlers/klippekjop.go`** — pakkene kjem frå basen. Dei stod i HTML-en
  før, og dei to sanningane var ikkje samde.

## Språka

Nynorsk (`nn`, standard), bokmål (`nb`) og engelsk (`en`). Valet ligg i ei
kake med eitt års levetid og verkar òg før innlogging; `?lang=en` på kva
adresse som helst byter.

All synleg tekst går gjennom `{{t .Lang "nykel.namn"}}`. Filene ligg i
`locales/`. Sjå `docs/LOCALIZATION.md` for korleis ein legg til eit språk.

Legg du til ein nykel, legg han i alle tre filene. Prøvene i
`./handlers` er det som seier frå.

## Arket

Stilboka er `docs/DESIGN_GUIDELINES.md`, og ho er reglar med grunngjeving:
held ikkje grunngjevinga i eit høve, er det regelen som skal skrivast om.
`docs/FASETTEN.md` skildrar den eine djupna i systemet — to pikslar, på
felt ein kan skrive i — og `docs/SKRAAKANTEN.md` kanten kring henne: tvo
lippor i staden for ei, og kvifor dei ikkje snur når temaet gjer det. `docs/KORREKTUREN.md` skildrar redigeringsforma:
du skriv rett på teksten, og det du har endra får eit merke i margen.

`/arket` er verkstaden — heile stilarket på éi side, i begge tema. Han
svarar berre i utvikling, og han teiknar seg sjølv i prøvene.

## Utviklingstenaren

```sh
./køyr              # startar, og byggjer på nytt når Go-filer endrar seg
./køyr --ein-gong   # startar utan vaktar
./køyr --stogg      # stoggar
```

Han lyder på <http://localhost:8080> og skriv til `.køyr.logg`.

| Du endrar | Kva som skjer |
|---|---|
| `.go` | vaktaren byggjer og startar på nytt, kring eitt sekund |
| malar i `handlers/templates/` | ingenting — dei blir lesne per soknad. Oppdater sida. |
| `static/css`, `static/js` | ingenting. Oppdater sida. |

Dei to siste gjeld berre når `KJERNEKRAFT_ENV=development`, som `./køyr`
set. I utvikling går både sidene og dei statiske filene ut med
`Cache-Control: no-store` — elles sit ein og ser på si eiga gamle CSS og
lurer på kvifor endringa ikkje slo inn. Skriftene er unntekne og blir
bufra hardt i begge; skiftar ei skriftfil namn, skiftar ho òg adresse.

Skriptet set ein fast `KJERNEKRAFT_SESSION_KEY`, så ein ikkje blir logga
ut kvar gong tenaren startar. Han er **berre** til dette.

### Prøvebrukar

Til å sjå på — grafar, medlemskort, klippekort — er `scripts/testbrukar`
betre enn eit tomt oppsett:

```sh
go run ./scripts/testbrukar          # lagar Solfrid
go run ./scripts/testbrukar -slett   # tek henne bort att
```

`solfrid@test.local` / `password123`. Ho har rolla `user` og ikkje meir,
så `/admin` svarar 403 for henne — det er heile poenget med henne.

Ho er ikkje tilfeldig. Eit varmekart med jamn støy i ser ut som ein feil;
ei rekkje ein kjenner att les ein med ein gong. Difor har ho ei soge —
varsam i mars, tre i veka i mai, topp i juni, to veker borte i juli, og
attende i august heilt fram til den veka du står i. Timane er verkelege
timar frå basen, så aktivitetsbolken har noko å teikne.

Det minimale oppsettet finst framleis:
`POST /api/setup-test-data` (eller `/elev/testdata` i nettlesaren) lagar
`anna@example.com` / `password123`.

Ho får rolla `user` og **ikkje** `admin` — `/admin` svarar 403 for henne
rett etter oppsettet. Rutene som deler ut roller ligg sjølve bak
`RequireAdmin`, så den første administratoren må setjast i basen:

```sh
sqlite3 kjernekraft.db "
  INSERT OR IGNORE INTO roles (name) VALUES ('admin');
  INSERT OR IGNORE INTO user_roles (user_id, role_id)
  SELECT u.id, r.id FROM users u, roles r
  WHERE u.email = 'anna@example.com' AND r.name = 'admin';"
```

Etter det kan resten delast ut frå `/admin`.

## Dokumenta

| Fil | Kva han er |
|---|---|
| `docs/BLUEPRINT.md` | grunnteikninga — heile huset, skjerm for skjerm. Ei teikning, ikkje ein byggjeplass |
| `docs/DESIGN_GUIDELINES.md` | stilboka: reglar med grunngjeving |
| `docs/FASETTEN.md` | den eine djupna i systemet |
| `docs/SKRAAKANTEN.md` | kanten kring djupna — tvo lippor, og retningi som ikkje snur |
| `docs/KORREKTUREN.md` | redigeringsforma — retting i margen |
| `docs/LOCALIZATION.md` | språk: legge til, omsetje, halde ved lag |
| `docs/STRUCTURE_AUDIT.md` | ettersynet av dokumentstrukturen, med funn som står att |
