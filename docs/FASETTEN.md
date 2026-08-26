# Fasetten

Dette er den eine djupna Kjernekraft hev, og ho er tvo pikslar stor.

Alt anna i stilboki er flatt med vilje: hårlinor i staden for skuggar,
éin farge paa heile arket, ingen rundingar utyver det vesle. Skugge er
teke av for det som *faktisk* flyt — ei nedfallsliste, ein dialog. Difor
finst det berre éin stad til der noko hev djupn, og det er denne: eit
felt du kann skriva i ligg **i** arket og ikkje paa det.

Ho er verd eit eige skriv av tvo grunnar. Ho er det einaste draget som
gjeng att i heile systemet utan aa vera ein farge, og ho er den lettaste
tingen i verdi aa gjera gal.

---

## 1. Tokeni

Baae stend i `:root` i `static/css/kjernekraft.css`, og ingen annan stad.

```css
--fordjuping: inset 0 1px 2px color-mix(in srgb, var(--blekk) 12%, transparent);
--fasett:
  inset 0 1px 2px color-mix(in srgb, var(--blekk) 16%, transparent),
  0 0 0 1px var(--merke);
```

Kvart tal tyder noko:

| Tal | Kva han er | Kvifor han er slik |
|-----|-----------|--------------------|
| `0 1px` | skuggen fell **rett ned**, éin piksel | Ljoset kjem ovanfraa. Ein skugge som fell til sida segjer at ljoset stend skeivt, og daa lyt *alt* i systemet vera samde um kva veg. |
| `2px` | uskarpleiken | To pikslar er so vidt mjukt. Fire ser vaatt ut, null ser ut som ein feil i teikningi. |
| `12 %` | styrken i kvila | Nok til aa sjaa i sidesynet, ikkje nok til aa lesa som ei ramma. |
| `16 %` | styrken i fokus | Fire prosentpoeng. Du ser det ikkje naar du leitar etter det; du ser det naar det hender. |
| `0 0 0 1px` | ringen utanpaa | Ikkje uskarp. Han doblar kanten fraa éin piksel til tvo utan aa flytta noko. |

Fargen er rekna av `--blekk` og ikkje skriven som eit hex-tal. Ein skugge
er blekket som ligg tjukkare, ikkje ein eigen farge.

---

## 2. Kvifor ho stend paa alt heile tidi

Fordjupingi stend paa **alle** skrivefelt, ikkje berre paa det ein held
paa med.

Var ho berre i fokus, laut ein klikka rundt paa sida for aa finna ut kva
som let seg endra. Djupna er ikkje ei tilbakemelding — ho er ei
*upplysning*: her kann du skriva. Ho lyt difor vera der fyre du gjer noko.

Ved fokus kjem merkefargen attaat, og djupna vert litt djupare:

```
kvila   djupn 12 %                       → «her kann du skriva»
fokus   djupn 16 % + ring i merkefargen  → «og her er du no»
```

Tvo eigenskapar, tvo spursmaal. Hadde fargen bore baae, hadde ein ikkje
kunna sjaa kva som var skrivbart utan aa fyrst taka i det.

---

## 3. Kvar ho høyrer heime

**Ja:** `input[type=text]`, `email`, `password`, `date`, `tel`, `number`,
`textarea`, søkjefeltet, adresselappen, namnefeltet.

**Nei:**

- **Veljarar.** Ein veljar plukkar eitt av nokre faae — han er ein knapp
  med ei liste attum, ikkje ei renna ein hell noko nedi. Reglane segjer
  det rett ut: `:where(input:not([type=checkbox]):not([type=radio]), textarea)`.
- **Knappar.** Ein knapp stend *paa* arket. Gjev han djupn, tyder djupn
  tvo ting.
- **Kort.** Eit kort er ei hårfin lina og ingen ting meir.
- **Avkryssingar og radioknappar.** Dei er for smaa; tvo pikslar
  uskarpleik paa ei rute paa sekstan gjer henne berre grumsete.

Regelen bak: **fordjupingi tyder «her kann du leggja noko inn med eigne
ord».** Ho er ikkje pynt, og ho er ikkje ein maate aa gjera ein kontroll
viktigare paa.

---

## 4. Den same fasetten utanfor CSS

Timemerket i timeplanen er teikna som SVG, og der finst ikkje
`box-shadow`. Fordjupingi er difor skriven ein gong til som eit filter,
og ho skal tyda det same:

```xml
<filter id="merkedjup">
  <feOffset in="SourceAlpha" dy="1" result="ned"/>
  <feGaussianBlur in="ned" stdDeviation=".85" result="mjuk"/>
  <feComposite in="SourceAlpha" in2="mjuk" operator="out" result="band"/>
  <feComposite in="SourceGraphic" in2="band" operator="in"/>
</filter>
```

Lese paa norsk: *tak silhuetten, skuv han eitt hakk ned, mjuka honom
upp, og drag han fraa silhuetten sjølv. Det som vert att, er ei tunn
stripa inne i yverkanten. Klypp elementet til den stripa.*

Det er `inset 0 1px 2px` skrive med andre ord, og tali er dei same, av di
merket er teikna i eit viewBox der éi eining er éin piksel ved
normalstorleik.

To ting er verde aa merkja seg:

- **Filteret ber ingen farge.** Det arbeider paa `SourceGraphic`, so
  fargen kjem fraa CSS paa elementet. Fyrste utgaava nytta `feFlood` med
  `flood-color: var(--kant)`, og det ser rett ut heilt til ein set tvo
  tema paa den same sida: fargen vert løyst der *filteret* stend, ikkje
  der det vert nytta. Alle merki fekk fargane til rot-temaet.
- **Filteret arbeider paa silhuetten aat heile figuren.** Merket er tri
  former under éi lina. Ein strek langs kvar form teikna skøytane deira
  med, so skiltet fekk ei lina tvert yver kroppen.

---

## 5. Det eine holet som hev tvo kantar

Regelen ovanfor — *ingen hovding i botnen* — gjeld felt. Eit felt er ei
**grunn** fordjuping i arket: ljoset kjem inn ovanfraa og fell paa ein
botn du ikkje ser. Difor hev det berre ein myrk yverkant.

Den tome ruta i vika er noko anna. Der gjeng det ingen time den dagen,
og plassen skal lesast som eit **tomt spor** — eit hol med ein botn du
faktisk ser ned i. Eit slikt hol hev tvo kantar: skuggen fraa den øvre
lippa, og ljoset som slær i botnen og kastar att paa den nedre.

```css
.daudflate { fill: color-mix(in srgb, var(--blekk) 6%, transparent); }
.dauddjup  { fill: color-mix(in srgb, var(--blekk) 28%, transparent); }
.daudljos  { fill: color-mix(in srgb, var(--flate) 80%, transparent); }
```

Ljoset er `--flate` og ikkje kvitt. `--flate` er alltid den ljosaste
flata i temaet — kalkstein i det ljose, ei tone yver arket i det myrke —
so den nedre lippa lyser i baae utan at nokon skriv henne om att.

Dette er **den einaste** staden i systemet med eit hovding. Kjem det
fleire, er regelen brote: ein fasett som stend paa alt, tyder ikkje noko
paa noko.

Ho stod som ein stipla kasse fyrr. Ein stipla kasse er eit *umriss*, og
eit umriss les seg som noko som skal koma — det motsette av det sanne.

---

## 6. Tvo maatar aa gjera henne gal paa

Baae vart gjorde medan timemerket vart teikna, og baae er lette aa gjera
om att.

**Mjuk skugge og kvitt hovding.** Fyrste utgaava av merket hadde ein
skugge ovanfraa *og* eit ljost drag nedanfraa, med fire pikslar
uskarpleik. Det er den klassiske utgaava av ein fasett — og det ser
uppblaast ut, som ein knapp fraa 2008. Fasetten her hev **ikkje** noko
hovding. Ho er grunn, hard og einsidig, og det er difor ho les seg som
trykt og ikkje som stempla.

**Fordjupingi som ein glød i det myrke.** Sjaa bolk 7.

---

## 7. Eit spursmaal som stend ope

`--fordjuping` er rekna av `--blekk`. I det ljose temaet er blekket
myrkt, og skuggen vert ei myrk lina i yverkanten — rett.

I det myrke temaet snur `--blekk`: han vert ljos. Skuggen vert daa ei
**ljos** lina i yverkanten, og fysisk tyder det motsette — ein lyst
yverkant er ein kant som stikk **upp**, ikkje eit hol.

Det er tvo ærlege svar, og eg hev ikkje teke noko av deim:

1. **Lat det staa.** Paa ein myrk skjerm les ei veik ljos lina i
   yverkanten som eit glimt av ljos som fangar kanten, og feltet ser
   framleis ut som eit felt. Det er ogso slik det stend i dag, og det
   stend fint.
2. **Gjer skuggen svart i baae temaom.** `rgb(0 0 0 / .45)` i staden for
   blekket. Daa er fysikken den same i baae, men i det myrke vert
   fordjupingi mest usynleg, av di flata alt er nesten svart.

Prøva er ei linja i ein nettlesar: set temaet til myrkt, sjaa paa
søkjefeltet i **Folk**, og spør um det ser ut som eit hol eller som ein
kant. Er svaret «kant», er det val 2 som gjeld.

---

## 8. Regelen

> Treng ein ny kontroll djupn, skal han nytta `--fordjuping` eller
> `--fasett` — ikkje ein ny skugge. Finn du eit `box-shadow` i ein
> komponent som ikkje er ein av dei tvo, eller `--skugge-flytande` paa
> noko som ikkje faktisk flyt, er det ein feil.

Prøva er ei linja:

```sh
grep -rn 'box-shadow' static/css/kjernekraft.css \
  | grep -v 'var(--fasett)\|var(--fordjuping)\|var(--skugge-flytande)' \
  | grep -v 'box-shadow: none'
```

`box-shadow: none` er ikkje ein ny skugge, det er ein som vert teken
burt, so han tel ikkje med.

I dag gjev prøva **eitt** treff, og det er meint: fana i
administrasjonen kastar ein skugge upp og ut, av di ei fana faktisk ligg
yver arket. Grunngjevingi stend i fila attmed regelen. Alt som kjem ut av
prøva, skal ha ei slik grunngjeving skriven attmed seg — elles er det ein
feil.
