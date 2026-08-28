# Skråkanten

Fasetten er **ei** lippa: ei myrk lina i yverkanten, og ingen ting i
foten. Skråkanten er **tvo**: den myrke lippa uppe, og ei ljos nede.

Tvo lippor er ikkje ein sterkare fasett. Det er ei anna form. Éi lippa
segjer *«her byrjar noko»*; tvo segjer *«her er eit hol med ein botn du
ser ned i»* — kanten er ein kutta flate og ikkje ei lina, og det er difor
han vert kjend att som ein skråkant og ikkje som ein skugge.

Dette skrivet stend attmed [FASETTEN.md](FASETTEN.md) og ikkje i staden
for han. Fasetten er framleis djupna. Skråkanten er kva som hender med
kanten kring henne.

---

## 1. Kvifor han fekk eit eige skriv

Han fanst frå fyrr. Han stod tvo stader — vekefeltet i timeplantittelen
og den tome dagen i rutenettet — og FASETTEN §5 sa rett ut at han skulde
verta verande der:

> **hovdinga høyrer til vika og til ingen ting anna.** Kjem det ein
> tredje stad som ikkje er timeplanen, er regelen brote — ein fasett som
> stend paa alt, tyder ikkje noko paa noko.

Den regelen er teken upp att her, med vilje, og han er ikkje kasta. Han
er **skriven om**: skiljet skal ikkje lenger gå mellom *å ha* skråkant og
*å ikkje ha* han. Det skal gå på **kor djupt han ligg**.

Grunnen er at den gamle regelen løyste feil oppgåve. Han var meint å
verna om at vika er noko for seg — men det ho er noko for seg *i*, er
storleiken og tittelrolla, ikkje kanten. Eit felt du skriv i er det same
slaget ting kvar det står, og huset hev alt sagt det ein gong: éin
grunnstil, og klassane som stend att er berre mål (`.soke`).

---

## 2. Tokeni

Tvo lippor, og tvo former bygde av deim. Fargen på lippa seier kva veg
det gjeng; **geometrien** seier kor djupt.

```css
/* i :root, uppe attmed --glans */
--ned: #000;
--upp: #fff;
--lippe-djup: color-mix(in srgb, var(--ned) 22%, transparent);
--lippe-ljos: color-mix(in srgb, var(--upp) 55%, transparent);

/* i det myrke temaet */
--lippe-djup: color-mix(in srgb, var(--ned) 40%, transparent);
--lippe-ljos: color-mix(in srgb, var(--upp) 26%, transparent);

/* formene */
--skraakant:
  inset 0  1px 2px      var(--lippe-djup),
  inset 0 -1px 1px      var(--lippe-ljos);
--skraakant-djup:
  inset 0  2px 2px -1px var(--lippe-djup),
  inset 0 -2px 1px -1px var(--lippe-ljos);
--skraakant-fasett: var(--skraakant), 0 0 0 1px var(--merke);
```

| Tal | Kva han er | Kvifor han er slik |
|-----|-----------|--------------------|
| `0 1px` / `0 -1px` | lippone fell **rett ned** og **rett upp** | Same ljoset som overalt elles. Skeive lippor gjev tvo ljoskjeldor. |
| `2px` mot `1px` | uskarpleiken, ulik i dei tvo | Den myrke lippa er ein skugge og skal vera mjuk. Den ljose er ei *flate som fangar ljos* — ho er kanten sjølv og ikkje noko han kastar, og difor skarpare. |
| `22 %` | den myrke i kvila | Ti meir enn fasetten sine 12. Ho lyt bera meir no, av di ho ikkje lenger er åleine um å seia kva veg det gjeng. |
| `55 %` | den ljose i kvila | Høgt tal, låg verknad: ho ligg på éin piksel og på ei flate som alt er ljos. Målt gjev ho ΔL +0.073 — eit drag, ikkje ei lina. |
| `-1px` | spreidingi i den djupe | Held draget smalt når det fyrst ligg tvo pikslar inn. Utan han vert det ei stripa i staden for ein kant. |

**Djupni er geometri og ikkje farge.** `--skraakant-djup` nyttar dei same
tvo lippone, berre lenger inn. Var djupni farge med, hadde ho trunge sitt
eige tema-tal, og då er det tvo ting å halda i takt der det held med
eitt.

---

## 3. `--ned` og `--upp` — og kvifor dei ikkje snur

Dette er heile grunnen til at skråkanten treng eit skriv, og det er ei
**retting av FASETTEN §7**, ikkje berre eit tillegg.

Fasetten er rekna av `--blekk`. §7 stod ope um det:

> I det myrke temaet snur `--blekk`: han vert ljos. Skuggen vert daa ei
> **ljos** lina i yverkanten, og fysisk tyder det motsette.

Med éi lippa var det til å leva med, og §7 valde å la det stå. Med tvo
lippor er det ikkje det lenger, og det er ikkje ei smaksak: snur baae,
snur **retningi**.

```
ljost tema    myrk uppe, ljos nede   → eit hol
myrkt tema    ljos uppe, myrk nede   → ein knott
```

Feltet sluttar ikkje å vera tydeleg — det vert tydeleg som det motsette.
Ei einsleg lippa er tvitydig nok til at auga les henne velviljug; eit par
er det ikkje, av di det er nettopp *paret* som ber kva veg det gjeng.

### Det var ikkje eit tenkt problem

Den tome dagen hadde det alt, og ingen hadde målt det. Fyrr stod ho slik:

```css
.dauddjup  { fill: color-mix(in srgb, var(--blekk) 28%, transparent); }
.daudljos  { fill: color-mix(in srgb, var(--flate) 80%, transparent); }
```

med denne grunngjevingi attmed: *`--flate` er alltid den ljosaste flata i
temaet, so ho lyser i baae*. Det er sant um **rangeringi** millom tokeni
og usant um **ljoset**. Åtti prosent av ei mest svart flate er framleis
mest svart. Rekna ut i relativ luminans:

| tema | øvre lippa | nedre lippa | les seg som |
|------|-----------|------------|-------------|
| ljost | L=0.404 | L=0.884 | hol |
| myrkt | L=0.101 | **L=0.014** | **knott** |

Den tome dagen i timeplanen stod altso og las seg som ein **knott** i det
myrke temaet, og hadde gjort det heile tidi.

### Løysingi, og at ho ikkje er ny

```css
--ned: #000;
--upp: #fff;
```

Dei tvo er ikkje fargar. Dei er ei **retning**: ljoset kjem ovanfrå i
baae temaom, og tyngdi snur ikkje når fargane gjer det.

Dette er ikkje eit nytt påfunn i huset. `--glans` — glimet på knappen —
gjer alt det same, og av same grunnen:

```css
--glans: color-mix(in srgb, #fff 55%, transparent);   /* ljost */
--glans: color-mix(in srgb, #fff 20%, transparent);   /* myrkt */
```

Han er `#fff`-rekna i baae, og han fell frå 55 % til 20 % i det myrke, av
di ei mest svart flate treng mindre kvitt for å syna det same. Skråkanten
fylgjer honom — og han gav ogso talet: 26 % gjev ΔL +0.093 i det myrke,
mot +0.073 i det ljose. Det er den same lippa, ikkje ei sterkare.

> Det er framleis `color-mix`. Ingen skriv eit ferdig-rekna hex-tal inn i
> ein skugge; ein skriv kor mykje ned og kor mykje upp, og let
> nettlesaren rekna resten mot flata som ligg under.

### Ei fella i fila sjølv

`--ned`, `--upp` og lippone lyt stå i `:root` **uppe**, attmed `--glans`.
`[data-theme="dark"]` og `:root` hev same spesifisitet, so eit `:root` som
kjem *etter* temablokki i fila vinn yver henne, og tema-tali vert daude
utan at nokon ser det. Formene (`--skraakant` og dei andre) kann stå kvar
som helst, av di `var()` vert løyst der han vert *nytta* og ikkje der han
vert skriven.

---

## 4. Kvar han høyrer heime

**Ja** — alt du skriv i, altso nøyaktig det same settet som fasetten
hadde:

`input[type=text]`, `email`, `password`, `date`, `tel`, `number`,
`search`, `time`, `textarea`, `.lappfelt`, `.namnefelt`, `.setning .felt`.

**Djup utgåve** (`--skraakant-djup`): `.bobla` / `.vekefelt`, og den tome
dagen i rutenettet (`.dauddjup` / `.daudljos`).

**Nei:**

- **Veljarar.** Uendra frå FASETTEN §3. Ein veljar er ein knapp med ei
  liste attum, ikkje ei renna. Han hev kant og inga djupn.
- **Knappar.** Uendra. Ein knapp stend *paa* arket og hev `--knappedjup`.
  Gjev ein honom ein skråkant med, tyder skråkant baae vegar på ein gong.
- **Kort, avkryssingar, radioknappar.** Uendra frå FASETTEN §3.

Regelen bak er den same som fyrr, og det er verdt å seia han om att av di
han ikkje vart endra: **djupn tyder «her kann du leggja noko inn med
eigne ord».** Skråkanten gjer henne lettare å sjå. Han gjer henne ikkje
til noko anna.

---

## 5. Fokus

Fokus legg merkefargen attaat, som fyrr — men djupni vert **ikkje**
djupare. Det var fasetten sitt grep (tolv til seksten prosent), og det
grepet fanst av di ringen åleine var for lite. Med tvo lippor er forma
alt tydeleg i kvila, og eit felt som *ogso* skifter djupn når du tek i
det, rører seg meir enn det seier.

```
kvila   skråkanten            → «her kann du skriva»
fokus   skråkanten + ringen   → «og her er du no»
```

> **Fella står ved lag.** `.bobla` tek aldri ringen — korkje i kvila
> eller i fokus (FASETTEN §5). Ho lyt difor skriva `box-shadow` ut i
> *baae* tilstandane. Sløyfer ein henne i fokus, tek grunnstilen yver og
> ringen kjem attende bakvegen. Det hev hendt ein gong fyrr.

---

## 6. Prøvone

**Laust `box-shadow`.** Same prøva som FASETTEN §8, med skråkanten lagd
til i sila:

```sh
grep -rn 'box-shadow' static/css/kjernekraft.css \
  | grep -v 'var(--skraakant\|var(--fasett)\|var(--fordjuping)\|var(--skugge-flytande)' \
  | grep -v 'var(--knappedjup' \
  | grep -v 'box-shadow: none'
```

Han gjev **fire** treff, og tvo av deim er ikkje skuggar i det heile —
ein `transition` som nemner `box-shadow`, og ei line inne i ein
kommentar. Dei tvo verkelege er fana i administrasjonen og halo-en kring
`.ljos-paa`, og baae er grunngjevne i fila.

Det er tvo færre enn fyrr. `.vekefelt` stod med skråkanten skriven ut for
hand i baae tilstandane; no er han eit token, og FASETTEN §8 sin
«fire meinte treff» er difor **tvo** meinte treff.

**Faste fargar.** Denne prøva er *ikkje* verdt å skriva, og det skal stå
kvifor: huset hev alt tjuge `#fff`/`#000` frå fyrr — glansen, glødkjernen,
utskriftsstilen, `::backdrop`. Ein grep etter faste fargar seier difor
ingen ting. Regelen er ikkje «ingen faste fargar»; han er:

> Ein fast farge er lovleg berre der han tyder ei **retning** eller eit
> **medium** — kva veg ljoset kjem, eller at dette skal på papir. Tyder
> han ein *farge*, skal han vera eit token.

---

## 7. Eit spursmaal som stend ope

Sjå på tali i det myrke temaet:

```
myrkt   flate L=0.019   øvre lippa L=0.008 (Δ-0.010)   nedre L=0.111 (Δ+0.093)
```

Den øvre lippa er så godt som borte. Ho *kann* ikkje vera anna: ei svart
lippa på ei mest svart flate hev ingen stad å gå. I praksis er skråkanten
difor **einlippa i det myrke temaet** — berre den ljose foten ber honom —
og det er den motsette einlippa av den fasetten hadde.

Det er to ærlege svar, og eg hev ikkje teke noko av deim:

1. **Lat det staa.** Ein fot som fangar ljos er ei like sann lesing av
   eit hol som ein skugge i toppen, og i eit myrkt rom er det den ein
   faktisk ser. Forma les seg rett, og tali segjer det same.
2. **Lat flata bera det i staden.** Gjer feltflata eit hakk *myrkare* enn
   arket i det myrke temaet, i staden for eit hakk ljosare som no
   (`--blekk` 3 % yver `--flate` gjev ei ljosare flate når blekket er
   ljost). Då fær holet ein botn ein ser ned i, og den øvre lippa treng
   ikkje bera noko.

Prøva er ei line i ein nettlesar: set temaet til myrkt, sjå på
søkjefeltet i **Folk**, og spør um det ser ut som eit hol eller som ei
flate med ein strek under. Er svaret «strek under», er det val 2 som
gjeld.
