# Arket — stilbok for Kjernekraft

Denne boki er reglar, ikkje smak. Kvar regel hev ei grunngjeving, og
grunngjevingi er det som gjeld: finn du eit høve der ho ikkje held, er
det regelen som skal skrivast um, ikkje brjotast i stilla.

Ho avløyser den gamle stilboki, som sa «Shadows Over Borders» og gav
Bootstrap-fargane `#007cba`, `#28a745`, `#ff6b35` og `#dc3545`. Den
regelen laga eit produkt som saag ut som kva som helst anna. Det er
ikkje ein smaksdom: ein regel som gjev same resultatet same kva du
byggjer, er ein regel som ikkje veit noko um deg.

---

Korrekturen — forma der du skriv rett på teksten og det endra ber eit
merke i margen — har sitt eige skriv:
[KORREKTUREN.md](KORREKTUREN.md). Ho er brukt i profilen og i prisinga.

Fasetten — den eine djupna systemet hev — hev sitt eige skriv:
[FASETTEN.md](FASETTEN.md). Han er det einaste draget som gjeng att i
heile systemet utan aa vera ein farge, og den lettaste tingen aa gjera
gal.

## 1. Tokeni

**Alle fargar stend i `:root` i `static/css/kjernekraft.css`, og ingen
annan stad.** Finn du eit hex-tal ute i ein komponent — i eit stilark, i
eit `style`-attributt, i ein Go-streng — er det ein feil. Han verkar i
eitt av dei tvo temaom og lyg i det hine.

Prøva er ei linja:

```sh
grep -rn 'style="[^"]*#' handlers/templates/    # skal gjeva ingen ting
```

Tri tilstandar, alltid alle tri:

| Tilstand | Veljar | Kva han er |
|---|---|---|
| ljos | `:root` | heile paletten, som grunnlag |
| mørk etter systemet | `@media (prefers-color-scheme: dark) { :root:not([data-theme="light"]) }` | berre tokeni skrivne um att |
| mørk etter val | `:root[data-theme="dark"]` | dei same tokeni ein gong til |

Ein farge som *berre* er skriven inne i eit media- eller
`[data-theme]`-blokk gjeld aldri i den ustempla tilstanden — og det er
den tilstanden dei fleste ser. Det er den vanlegaste feilen i eit
tofarga tema.

Er ein token **rekna** av andre token, skal han staa éin gong.
`--faneflate` er `color-mix` av blekket og arket og fylgjer difor temaet
av seg sjølv.

---

## 2. Fargane

Dei er tekne av tingi som ligg i salen: den rosa Togu-ballen, den
turkise sprayflaska fraa Clas Ohlson, den lilla Tune Up-ballen.

| Token | Kva han tyder |
|---|---|
| `--klas` turkis | **medlemskapet** — det som gjeng og gjeng |
| `--togu` rosa | **klippekortet** — det som vert talt |
| `--tuneup` lilla | **kurs, PT og rettleiding** — det som hender ein gong |
| `--aatvaring` | noko gjekk gale, eller lèt seg ikkje gjera um att |

**Fargen segjer *kva*. Forma segjer *korleis*.** Dei ligg paa kvar sin
kanal og kolliderer difor aldri:

```
aktiv       ●  fylt merke        frosen     —  strek
ventande    ○  open ring         utgjengen  ⋯  prikka, i graatt
```

Difor er «aktivt medlemskap» turkist og fylt, medan «aktivt klippekort»
er rosa og fylt: same stoda, ulikt produkt, og du ser baae delar paa
eitt blikk.

Aatvaringsfargen er ikkje ein av dei tri. Vart han det, tydde ein farge
tvo ting.

Komponentar tek produktfargen gjenom `--produkt`, ikkje beinveges:

```css
.membership-card { --produkt: var(--klas); }
.klippekort-card { --produkt: var(--togu); }
```

---

## 3. Ljoset

Ljosbandet fraa ordboki. Fire lag radialt — ein utbrend kjerne, fargen
tett kring han, ein vid veik halo, og ein bloom under — og **ribbor
yver**: skiljelinor som ligg *paa* ljoset og skjer det upp. Det er
ribbone som gjer at det ser ut som ljos gjenom ei rist, og ikkje som ein
farga boks.

### Klippekortmaalaren

Der bandet i ordboki er *eitt* ljos som glir langs lina, er maalaren
mange: **eitt per klipp du hev att**.

Grunnen er at eit klippekort ikkje er ei mengd — det er eit tal. Ein
prosentstripe segjer «tri fjerdedelar att»; du vil vita «sju». Difor tel
ein rutor.

Eit brukt klipp er ikkje burte, berre ikkje tent. Ruta stend der
framleis, i same storleiken, so maalaren aldri skiftar breidd naar du
brukar eit klipp: **det er ljoset som slokknar, ikkje ruta som
forsvinn.** Sjaa §6.

Ribbone er teikna paa kvar rute si eigi høgre sida, ikkje rekna i
prosent paa hylsteret. Ei prosentrekning maa gissa kvar flexen la
rutorne, og bommar so snart tale skifter.

Alt ljos er slege av under `prefers-reduced-motion: reduce`.

---

## 4. Papir er flatt. Ting er det ikkje.

Denne bolken sa «linor, ikkje skuggar», og opningi av stilarket sa
«skuggar finst ikkje». Det var aldri sant. Ljosbandet under hovudet —
det som lyser under den valde lenkja, og som er det finaste i heile
huset — er **fire lag radialgradient med ein utbrend kjerne**. Det stod
på kvar einaste side medan doktrinen sa at slikt ikkje fanst.

Ein regel som blir broten av det ein er stoltast av, er ikkje ein regel.
Han er ei mistyding av kva ein sjølv held på med. Så her er kva som
faktisk gjeld:

**Arket er flatt.** Papir har inga djupn. Kort, bolkar, tabellar,
overskrifter, faner — alt som *er* ark eller er laga av ark — skil seg
frå kvarandre med ei hårline på éin piksel og ingen ting meir.
`--haarlina` teiknar eit kort; `--kant` skil to bolkar.

**Ting på arket har djupn.** Ein knapp du trykkjer på. Eit lys som
lyser. Eit felt du skriv i. Desse er ikkje papir — dei er gjenstandar
som ligg oppå det, og gjenstandar fangar ljos. Å nekte dei det gjer dei
ikkje reinare; det gjer dei berre til teikningar av seg sjølve.

**Ljoset kjem ovanfrå.** Det er den eine regelen som bind alt saman, og
den einaste ein aldri bryt:

| Ting | Kva ljoset gjer |
|---|---|
| felt | fell ned i ei grøft — mørk overkant (FASETTEN) |
| knapp | fell på glas — dropa øvst, tjukk vegg i endane, attkast i foten (§18) |
| tom dag i vika | fell ned i eit hòl — mørk overleppe, ljos underleppe |
| ljosbandet | er sjølv kjelda |
| kort | fell på flatt papir — ingen ting hender |

Bryt ein den, sluttar rommet å henge saman: ein skugge som fell til
sida seier at ljoset står skeivt, og då må *alt* i systemet bli samde om
kva veg.

**Fallskugge** finst framleis i éi utgåve, `--skugge-flytande`, og han
er berre til det som *faktisk* flyt fritt over sida: ei nedfallsliste,
ein dialog. Djupna på ein knapp er ikkje det — han ligg på arket, han
svevar ikkje over det.

---

## 5. Skrifti

Union Gothic ber yverskriftene, Ubuntu ber lesetekst, Ubuntu Mono ber
tal som skal staa i kolonne. Alle er sjølvvertsette som `woff2` under
`static/fonts/`. **Ikkje eit einaste kall ut paa netet:** sida skal sjaa
lik ut same kvar ho vert opna, og ho skal teikna seg ferdig jamvel naar
studioet sitt nett ligg nede.

> **Union Gothic er mellombils.** Fila er CC BY-NC-ND — korkje handel
> eller avleidde verk — og ligg difor i `.gitignore`. Manglar ho, fell
> alt attende paa Ubuntu Condensed av seg sjølv. Naar ei lisensiert
> skrift kjem, er det éi `@font-face`-blokk som vert bytt.

### Breiddi tyder rang

Union Gothic hev tvo aksar, breidd 50–200 og vekt 400–1000. Breiddi er
ikkje pynt — ho er hierarki:

| Token | Breidd | Kvar |
|---|---|---|
| `--vidd-tittel` | 125 % | sidetittelen — kvar du er |
| `--vidd-yver` | 105 % | yverskrifter i sida |
| `--vidd-leiding` | 90 % | leidingi |
| `--vidd-merkelapp` | 68 % | merkelappar, faner, tabellhovud |

Det vidaste stend øvst og det smalaste nedst, so ein ser kva som er kva
fyre ein hev lese eit einaste ord. **Ein merkelapp skal ikkje taka meir
plass enn han er verd.**

---

## 6. Ingen ting skal skuva seg

Dette er den viktigaste regelen i boki.

Klikkar du paa noko, skal du framleis sjaa paa den staden du saag paa.
Skuvar innhaldet seg, lyt auga finna seg att, og daa er handlingi
dyrare enn ho ser ut.

I praksis:

- **Kantar stend heile tidi og skifter berre farge.** Ein kant som
  kjem og gjeng flytter alt kring seg med ein piksel.
- **Bolkar vert ikkje bytte ut.** Alle fanebolkarne stend i dokumentet
  samstundes; det er `data-bolk` paa rommet som avgjer kva som er
  synleg. Vart dei bytte ut, laut nettlesaren leggja sida ut paa nytt
  for kvart klikk.
- **Skiftar ein ting storleik, skift fyllet og ikkje kanten.** Den opne
  fana stikk upp ved at `padding-top` aukar med `--loft`; teksti flyt
  ikkje.
- **Ei rute som vert tom skal enno vera ei rute.** Sjaa
  klippekortmaalaren.
- **Eit felt som byter til redigering skal halda breiddi si.** Talet
  ligg i `data-oere`; teksti er formatert for auga og vert aldri lesi
  attende.

---

## 7. Ingen rullestenger

**Sidelengs rulling er forbode.** Ei side rullar ned, og ikkje til
sidone.

**Spør fyre du lagar ein rullestong nokon stad, og gaa ut fraa at
svaret er nei.**

Naar noko ikkje fær plass, er svaret eitt av desse — aldri ein
rullestong:

| I staden for | Gjer |
|---|---|
| rutenet som renn ut | spaltor som *deler* breiddi (`minmax(0, 1fr)`), og fær bolkar under kvarandre under ein brotstad |
| tabell som renn ut | `table-layout: fixed` og tekst som bryt |
| fanor som renn ut | dei bryt ned paa ei rad til — som fanor i tvo lag i ei mappa |
| leiding som renn ut | ho bryt |

Timeplanen er sju spaltor som deler breiddi. Under 62rem vert kvar dag
ein bolk under den fyrre, med dagen som yverskrift, so ein time framleis
høyrer til ein dag.

---

## 8. Tekst held seg innanfor

Tekst skal aldri renna ut or det ho stend i. Norsk lagar lange
samansette ord heile tidi — «medlemskapsadministrasjon» — og eit slikt
ord skal brjota, ikkje skuva spalta ut or glaset.

Tvo reglar gjer det, og ein treng baae:

```css
:where(body *) { min-width: 0; }
:where(p, li, td, th, label, h1, h2, h3, h4, h5, h6) { overflow-wrap: break-word; }
```

`min-width: 0` er den som vert gløymd. Eit rute- eller flexbarn hev
`min-width: auto` som standard, og nektar daa aa verta smalare enn
innhaldet sitt same kva ein segjer um breidd. **Det er den eine regelen
som gjer at rutenet renn ut or sida.**

---

## 9. Fanone

Same systemet som i ordboki. Ei fana er eit blad i ei mappa: lina gjeng
attum dei lukka og er **broti** under den opne. Det er brotet som gjer at
rekkja les seg som fanor — ikkje fargen, ikkje skuggen.

Fire tal styrer alt. Stend dei kvar for seg i seks reglar, lyt dei vera
samde, og er det ikkje:

| Tal | Kva han er |
|---|---|
| `--kant` | sidekantarne og toppen |
| `--fot` | kor tjukk lina fana stend paa er |
| `--fotfarge` | kva den lina er naar fana er lukka |
| `--loft` | kor mykje den opne fana stikk upp yver dei lukka |

To fellor:

- **Loftet skal vera heile pikslar.** Ein rem-brøk vert sjeldan eit
  heilt tal pikslar, og daa hamnar underkanten aat den opne fana ein
  halv piksel fraa lina ho skal møta. Det er den halve pikselen som gjer
  at fanone ikkje ser ut som fanor.
- **`background-clip: padding-box`.** Utan han vert flata maala under
  foten, foten er gjenomsynleg med vilje, og daa stend ingi fana paa
  noko.

Fanone er ein **allmenn** komponent — `.faneark`, `.faner`, `.fane`,
`.fanerom` — og skal nyttast att kvar det er fleire syn paa den same
tingen. Valet stend i adressa, so ei lenkja peikar dit ho segjer og ein
omlasting kjem attende til same staden.

---

## 10. Merket

`static/img/kjernekraft.svg`. Aatte K-ar i ein ring, nett 45 grader fraa
kvarandre, teikna som **eitt** teikn snutt aatte gonger — ingen av deim
er teikna for seg. Maali er tekne av det upphavlege merket: stammen 20
brei, armarne 17,7 vinkelrett, 41 grader fraa loddrett.

Han er `fill: none; stroke: currentColor`, so han tek fargen fraa teksti
og verkar i baae temaom utan aa skrivast um.

---

## 11. Ord

Kvar streng er ein umsetjingsnykel. `{{t .Lang "bolk.nykel"}}`, og
nykelen finst i alle tri filone under `locales/`.

Hardkodar du ein streng i ein mal, finst han berre paa eitt maal, og
ingen oppdagar det fyrr nokon byter maal.

---

## 12. Prøva stilboki

```sh
grep -rn 'style="[^"]*#' handlers/templates/   # ingen fargar i attributt
grep -rn 'overflow-x' static/css/              # ingi sidelengs rulling
grep -c 'font-stretch' static/css/             # breiddetrappa er i bruk
```

Og i nettlesaren, paa kvar sida:

```js
document.documentElement.scrollWidth > window.innerWidth   // skal vera false
```

---

## 13. Merket i bruk

`.merke` er ei **maske**, ikkje eit bilete: flata er `currentColor` og
fila er `mask`. Difor finst det éi fil paa disken og ikkje ei per farge,
og han fylgjer temaet utan at nokon skriv honom um. Same grepet som
krossen paa sletteknappen i ordboki.

| Stad | Klasse | Storleik |
|---|---|---|
| hovudet, attmed namnet | `.merkeord .merke` | 1,7 rem |
| innlogging, registrering, gløymt passord | `.doermerke` | 4 rem |
| fane-ikon | `<link rel="icon">` | fila si eigi |

Fila ber sin eigen `color` med eit mørkt alternativ inni seg. Som maske
er den fargen likegyldig; som fane-ikon er han det som gjer at merket
finst i baae nettlesartema.

**Han skal ikkje veksa nokon annan stad.** Eit merke som stend stort
paa kvar sida er ikkje eit merke lenger — det er ein bakgrunn.

---

## 14. Teikningar — der emojiane stend i dag

Emoji er ikkje eit teiknesprog. Dei er teikna av Apple, Google og
Microsoft kvar for seg, dei ser ulike ut paa kvar maskina, dei ber
hudfarge og kjøn me ikkje hev bede um, og dei kann ikkje ta
produktfargen. **Alle ni skal ut.**

### Kva som stend der no

**A — kategoriteikningar.** Seks stykke, paa `elev/klippekort`, steg 1.
Dette er dei einaste som treng ei *teikning*; resten er ikon eller skal
berre burt.

| No | Kategori | Kva ho skal segja |
|---|---|---|
| 🧘‍♀️ | Gruppetimer Sal | ein sal med matter — det opne golvet |
| 💪 | Reformer/Apparatus | reformeren sjølv: vogna, fjørerne, fotstonga |
| 🏃‍♂️ | Self Practice | den same reformeren, utan lærar attmed |
| 💻 | Online Gruppetimer | skjermen, sedd fraa matta |
| 👨‍💼 | Personlig Trening | tvo, ikkje ein |
| 🧠 | Stressmestring | pusten, ikkje hjernen |

Merk at 💪 og 🏃‍♂️ i dag skil tvo ting som er den same maskina med og
utan lærar — teikningarne bør vera same reformeren, so skilnaden er den
som faktisk finst.

**B — ikon i teksten.** 📍 for rom og 👨‍🏫 for lærar, paa timekorti.
Desse er ikkje teikningar; dei er tvo strekikon paa ein cirka 16 px
boks. Dei kann ogso berre gaa ut: rommet stend alt i halvfeit og
læraren under, og korti les seg fint utan.

**C — pynt som skal burt.** 🧊 paa frose medlemskap og 🍂 paa
hausttilbodet. Isbiten gjer arbeidet livsmerket alt gjer — ein frosen
stod er ein **strek** (§2) — so han segjer det same tvo gonger, med tvo
ulike former. Haustlauvet er pynt paa ein kampanje.

### Krav til fila

Same reglar som merket, av same grunn:

- **SVG**, `stroke="currentColor"`, `fill="none"`. Ingen eigne fargar i
  fila. Daa verkar ho i baae temaom utan aa skrivast um, og ho kann
  bera produktfargen der ho stend i eit produkt.
- **Kvadratisk `viewBox`**, helst `0 0 64 64`.
- **Ein strekbreidd** gjenom heile settet. Merket hev 20/750 &asymp; 2,7 %
  av breiddi; ligg teikningarne i nærleiken, høyrer dei saman med honom.
- **`stroke-linecap="butt"`** som merket, ikkje runde endar.
- Ingen tekst i fila.

### Kvar ho stend

Boksen finst alt og er fast, so eit kort ikkje skiftar høgd den dagen
teikningi kjem:

```html
<div class="teikning">…svg…</div>
```

`--teikningstorleik` styrer storleiken (3,5 rem som standard), og
`color` kjem av `--produkt` der det finst — turkis i medlemskap, rosa i
klippekort, lilla i kurs.

> Medan ho manglar, stend emojien i den same faste boksen. Difor hoppar
> ingen ting naar ho kjem paa plass.

---

## 15. Rangstigen i dokumentet

Breiddi ber rangen for auga (§5). Men dokumentet hev sin eigen
rangstige — `h1` til `h6` — og han er for deim som ikkje ser sida:
skjermlesaren byd fram yverskriftene som eit innhaldsoversyn, og daa er
det *talet* som gjeld, ikkje breiddi. I dag stend dei tvo stigane og
kranglar: same rolla er `h2` paa ei sida og `h3` paa grannesida, og tvo
sidor hev tvo `h1` kvar.

Regelen: **steget fylgjer djupni i dokumentet, og ingen ting anna.**
Utsjaanaden kjem av rolleklassa; steget kjem av kvar i sida yverskrifti
stend. Dei tvo skal aldri veljast kvar for seg.

| Steg | Rolla | Klasse |
|---|---|---|
| `h1` | sidetittelen — kvar du er. Éin per sida, *alltid éin*. | `.page-title` |
| `h2` | ein bolk paa sida, med lina under | `.section-title` |
| `h3` | eit kort i bolken, eller ei gruppa utan lina | korttitlane, `.undertittel` |
| `h4` | ei rad eller eit felt inne i kortet — sjeldan; oftast gjer `dt` arbeidet | — |

Treng du `h5`, er ikkje svaret `h5`: daa ber sida for mykje, og skal
delast.

Av dette fylgjer:

- **Tvo gjeremaal er tvo bolkar, ikkje tvo sidor paa same sida.**
  «Behandle klipp» og «Kjøp klipp» er éin `h1` og tvo `h2`.
- **Éi rolla, eitt namn.** `.module-title` og `.step-title` *er*
  `.section-title`; stilarket hev alt slege deim saman i éin veljar, og
  daa skal malarne ogso gjera det. `.login-title` er ein `.page-title`
  som fekk fast storleik av di han stod for seg sjølv.
- **Ei yverskrift utan klasse er ein feil.** Ho fell attende paa
  elementstilen og liknar ingen ting anna paa sida. (Unnataket er
  vilkaari, der innhaldet kjem ferdigt fraa tenaren og `.vilkaar
  h1`–`h3` tek imot.)

Prøva:

```sh
grep -rnE '<h[1-6]( [^>]*)?>' handlers/templates/ | grep -v 'class='
# skal gjeva ingen ting

for f in handlers/templates/pages/*.html; do echo "$(grep -c '<h1' $f) $f"; done
# ingi lina skal byrja paa meir enn 1
```

Tvo kjende unnatak, baae ufarlege: ein `<h1>` som stend i ein
*merknad* vert talt med (min-profil), og tvo `h1` i kvar si grein av
`{{if}}/{{else}}` er framleis éin paa den ferdige sida (membership).
Prøva tel filer, ikkje sidor — stig talet, sjaa etter kvifor.

---

## 16. Eitt lag

Millom arket og innhaldet ligg **høgst éin maala boks**.

Eit kort er ein ting som ligg paa arket: flata, haarlina, fyllet (§4).
Ligg det ein boks inni boksen, er ingen av deim lenger ein ting — dei er
emballasje. Inne i eit kort finst berre tekst og linor; inne i eit
fanerom likeins, for fanerommet *er* boksen sin.

Mønsteret som er rett finst alt, paa dashbordet: `h2.section-title`
stend beint paa arket med lina si, og korti ligg i eit rutenet under.
Bolken er ikkje ein boks — han er ei yverskrift, ei lina og romet etter
§17. **Det som skil tvo bolkar er rom, ikkje vegger.** Det er dette som
gjer at korti les seg som ting som ligg paa arket, og ikkje som boksar i
boksar.

Og ein regel um sjølve divane: **ein div skal anten maala eller leggja
ut.** Hev han korkje bakgrunn, kant, fyll, flex, rute eller breidd — og
ingi klasse stilarket kjenner — er han att fraa eit gamalt stilark og
skal burt. Kvart lag som ikkje gjer arbeid er ein stad rom kann verta
lagt til utan at nokon ser kvar det kom fraa, og det er *det* som gjer
at ei sida kjennest tilfeldig jamvel naar kvart tal stend paa trappa.

---

## 17. Rommet

Trappa finst — `--rom-1` til `--rom-7` i `:root` — og ho er i bruk mest
yveralt. Det som ikkje fanst, var kva stegi *tyder*. Eit steg utan
tyding vert valt etter smak, og daa ser sida tilfeldig ut jamvel naar
kvart tal stend paa trappa.

| Steg | | Tyder |
|---|---|---|
| `--rom-7` | 3 rem | millom bolkar paa sida |
| `--rom-6` | 2 rem | fraa sidetittel og leiding ned til fyrste bolken; yver ein `.undertittel` |
| `--rom-5` | 1,5 rem | fyllet i ein boks; gapet millom spaltor |
| `--rom-4` | 1 rem | millom kort i eit rutenet; fraa yverskrift til innhaldet hennar; millom radene i eit skjema |
| `--rom-3` | 0,75 rem | inne i ei gruppa: millom merkelapp og felt, millom knappar som høyrer saman |
| `--rom-2` | 0,5 rem | millom ting som les seg som eitt: ikon og ord, tal og eining |
| `--rom-1` | 0,25 rem | den minste lufti, der null hadde klistra |

Lovi attum tabellen: **romet utanfor ein ting er alltid større enn romet
inni honom.** Det er slik auga ser kva som høyrer saman utan vegger. Og
det er dette som gjer §16 mogleg: naar romi held rangen, treng ikkje
flokkarne gjerde.

Til sist: **komponenten ber ikkje rom sjølv; foreldra set gapet.** Ein
komponent med eigen margin veit kvar han stend, og daa kann han ikkje
flytta paa seg. `.module` gjer dette rett i dag: `gap` paa foreldri,
ingen margin paa borni.

---

## 18. Knappen

Knappen er ein **kapsel**. Ikkje eit avrunda rektangel, ikkje ein boks
med broti hyrne — ein full kapsel, `border-radius: var(--rund-knapp)`.

### Kvifor rundingsregelen ikkje gjeld her

Stilboka seier ingen rundingar utyver det vesle, og det står ved lag.
Regelen gjeld **arket og alt som er laga av det**: kort, felt, faner,
merke. Alle desse er papir, og papir har broten kant, ikkje kurve.

Ein knapp er ikkje papir. Han er ein fysisk ting som ligg *på* arket og
ventar på fingeren, og forma hans skal seie det. Difor to tokens og to
tydingar:

```css
--rund:       2px;    /* arket og alt som er laga av det */
--rund-knapp: 999px;  /* kontrollar, som ikkje er ark */
```

Finn du eit tredje rundingsverd i stilarket, er det drift.

### Han er glas, ikkje plast

Pilleforma er henta frå gamle knappar, og **glansen med henne**. Ein
knapp er ein gjenstand (§4): ljoset fell på han ovanfrå.

Han var ei kvelv ei stund — eit glim som stogga brått på midten. Det er
rett fysikk for eit stykke plast, og det er den harde kanten som gjer at
gamle nettknappar ser billige ut. Ei dropa glas oppfører seg annleis, og
skilnaden er tre ting:

| Lag | Kva det er | Kvar det står |
|---|---|---|
| **dropa** | ljoset som legg seg oppå, inne frå kanten, uskarpt i alle endar | `::before` |
| **glasveggen** | endane er mørkare av di ein ser gjennom meir glas der | `::after`, fyrste laget |
| **attkastet** | ljos som har gått *gjennom* kroppen og kjem att nedanfrå | `::after`, andre laget |

Ingen av dei tre har ein hard kant nokon stad. Det er heile skilnaden på
malt og teikna: ein målar set ljoset ned vått og lèt det renne ut, ein
teiknar dreg ei line og fyller innanfor.

Kroppen sjølv ber ikkje ljos i det heile — han er ein rein farge som
mørknar litt mot foten. Alt ljoset ligg i dei to laga oppå, og difor kan
ein byte farge på ein knapp utan å teikne glaset på nytt.

```css
--glans:      color-mix(in srgb, #fff 55%, transparent);  /* 20 % i mørkt */
--glans-svak: color-mix(in srgb, #fff 14%, transparent);
--knappedjup: inset 0 1px 0 var(--glans),      /* kanten som fangar ljoset */
              inset 0 -1px 1px …,              /* attkastet i foten */
              inset 0 -4px 6px -3px …blekk 14 %…,
              0 1px 1px …blekk 11 %…,          /* kontaktskuggen */
              0 4px 9px -3px …blekk 18 %…;     /* fallet, mjukt */
```

Tala er låge med vilje: 55 % er ei dropa, 90 % er plast. Fallet er to
skuggar og ikkje ein — den tette seier at han ligg på noko, den vide og
veike seier kor høgt.

Glaset ligg mellom flata og ordet, ikkje over ordet. Knappen er difor
`isolation: isolate`, og dei to laga står på `z-index: -1`.

### Ei rolle er fem fargar, ikkje eit lag

Fyrr stod den same gradienten skriven tre gonger, ein gong for kvar
rolle, og tre stader er tre stader å gløyme. No set ei rolle fem verd
og ingen ting meir:

```css
--knapp-kropp     /* flata han er støypt i */
--knapp-tone      /* det mørke i glasveggen og i foten */
--knapp-ord
--knapp-kant
--knapp-attkast   /* ljoset som kjem att nedanfrå */
```

Ein tilstand — hover, trykt — er ein ny farge i eitt av dei fem, aldri
ein ny gradient.

**Hover byter lina og ikkje ordet.** Merkefargen flytte seg inn i skrifta
ei stund òg, og det var to lyte i eitt: turkis på ljos flate er under
4,5:1, og `:hover` er framleis på i det ein trykkjer — så den uleselege
fargen låg nett der flata er mørkast. Merkefargen er ei **linefarge**.
Han er rekna for å halde éin piksel mot arket, ikkje for å bere ei
setning.

### Trappa

Same kapselen i tre storleikar: `.btn-liten`, ingen klasse, `.btn-stor`.
Glaset er rekna i `em` og i prosent og ikkje i pikslar, så dropa, veggen
og attkastet fylgjer skrifta av seg sjølv. Må ein teikne glaset på nytt
for ein ny storleik, er laget feil rekna.

Storleiken seier kor ofte handlinga hender, ikkje kor viktig ho er — det
siste står i fargen. Den store er kjøpsknappen og ingen ting anna; den
vesle er handlinga som bur inne i ei rad.

**Trykt snur ljoset.** Dropa går ut — glas som er trykt inn har ikkje
noko ljos oppå seg — attkastet vert veikt, skuggen fell inn ovanfrå, og
knappen sig éin piksel. Det er den einaste staden i huset der noko
flyttar seg — og han flyttar seg av di han er ein ting ein trykkjer på.

Skuggen som fell inn ovanfrå er difor ei **line** og ikkje eit slør. Låg
han som ei vaske over den øvre halvdelen, låg han rett over ordet, og
teksta vart tung å lese i nett det augneblinket ein braut henne.

### Ordet er rispa ned i glaset

`--ordet-inn` er eit glim på éin piksel under bokstaven og ingen ting
over han. Ljoset kjem ovanfrå, så den nedre lippa i ei rispe er den
einaste kanten som fangar det — same fysikken som fordjupinga i eit
felt, berre på ein bokstav.

Han står **i kvila**, ikkje berre i trykket. Ei rispe er noko som er
*skore* i ei flate, og eit snitt går ikkje att av di ingen held fingeren
nede: knappen flyttar seg når ein trykkjer, bokstaven gjer ikkje det.
Attpå får ordet ein ljos kant å lesast mot nett i det augneblinket flata
under det er mørkast.

Avslegen har han ikkje. Ingen djupn tyder ingen djupn — `text-shadow`
går ut saman med skuggen og glaset.

**Avslegen har ingen djupn i det heile.** Ingen dropa, ingen vegg, ingen
skugge: dei to laga står på `display: none`. Han er ikkje ein ting ein
kan trykkje på lenger, og skal ikkje sjå ut som ein.

### Røysta

Knappen bar merkelappskrifta før — 68 % smal, store bokstavar, sperra.
Det er språket til bolkoverskrifter, faner, `dt` og merke: ting som
**namngjev** ein kategori. Ein knapp namngjev ingen ting. Han er eit
verb, og eit verb skriv seg som eit ord i ei setning: lesetekstskrifta,
vanleg setningsform, ingen sperring.

«Lag timen», ikkje «LAG TIMEN».

### Fargen ligg i lina og i ordet

Hovudknappen var ei fylt flate i merkefargen — den tyngste tingen på
sida, tyngre enn overskrifta, i eit hus som elles byggjer alt av
hårliner. Åtvaringsknappen hadde alt rett svar skrive attmed seg: *ein
raud klump dreg auga fraa alt anna*. Ein turkis klump gjer det same.

| | Line | Ord | Flate |
|---|---|---|---|
| vanleg | `--kant` | `--blekk` | `--flate` |
| hovudhandling | `--merke` mørkna | `--blekk-fast` | `--merke-lys` |
| uatterkalleleg | `--aatvaring` | `--aatvaring` | `--flate` |

Flata er `--flate` og ikkje `--ark`. Ein knapp malt i sidefargen er eit
hòl med ei line rundt.

Hovudhandlinga er unnataket, og ho er det åleine: ho ber merkefargen i
flata av di ho er den eine knappen på sida som skal svarast fyrst. Det
er difor det står **éi** per skjerm — regelen og fyllet er den same
setninga sagd to gonger.

Ordet hennar er `--blekk-fast` og ikkje kvitt. Kvitt på turkis er to
ljose ting oppå kvarandre; mørkt på turkis er eit ord på ei flate. Og
`--blekk-fast` snur ikkje med temaet, av di merkefargen er ljos i begge
— eit ord som snur ville vorte ljost på ljost når skjermen er mørk.

### Grunnstilen gjeld `.btn`-familien, ikkje `<button>`

Tjue stader i huset laut skrella grunnstilen av att — eit merke, ei
fane, ei rad som opnar seg. Ein grunnstil som må opphevast tjue gonger
er ikkje ein grunnstil; han er eit unnatak som står først.

Ein naken `<button>` ber difor skrifta rundt seg og ingen ting meir.
Skal han sjå ut som ein knapp, får han `.btn`, `.btn-primary` eller
`.btn-danger`.
