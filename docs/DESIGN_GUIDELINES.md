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

## 4. Linor, ikkje skuggar

Det som skil eit kort fraa arket er **ei lina paa éin piksel**.

`--haarlina` teiknar eit kort. `--kant` skil tvo bolkar.

Skugge finst i **éi** utgaava, `--skugge-flytande`, og han er berre til
det som *faktisk* flyt: ei nedfallsliste, ein dialog. Ligg ein ting paa
arket, hev han ikkje skugge.

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
| `--vidd-merkelapp` | 68 % | merkelappar, knappar, tabellhovud |

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

