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

### Dei to fasettane

Dette skrivet har heile tida handla om djupna og aldri om **hårlina rundt
henne** — og det er dei to *saman* som blir kjende att som ein fasett.
Difor finst det to utgåver i huset, og berre den eine har vore skriven ned:

| | Kant | Djupn | Kvar |
|---|---|---|---|
| **Den kanta** | `1px solid var(--kant)` | `--fordjuping` | grunnstilen: søkjefeltet i Folk, innstillingar, registrering — og setningsfelta i timeplanen |
| **Den nakne** | inga | `--fordjuping` | korrekturen: `.lappfelt` på Min profil og i Prising |

Den kanta les seg tydelegast som eit *hol* — kanten er lippa. Den nakne
er lettare og går inn i ei tett liste utan å lage eit rutenett av seg
sjølv, som er heile grunnen til at korrekturen har henne.

**Fillet er ikkje ein del av fasetten.** Ein kant og ei djupn gjer
fasetten; kva som ligg i botnen av holet er fritt. `--flate` gjer holet
tydeleg og litt hardt; ei gjennomsynleg flate
(`color-mix(in srgb, var(--flate) 70%, transparent)`) lèt arket under
skine gjennom og er rettare der feltet står inne i noko anna — i ei
setning, i ei overskrift.

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

Dette var **den einaste** staden i systemet med eit hovding fram til 27.
august 2026. No er det to, og den andre er vekefeltet i timeplantittelen
(`.vekefelt`).

Grunngjevinga, so ho stend skriven: vekeveljaren *er* overskrifta til
det rutenettet. Talet og den tomme dagen er den same forma i to
storleikar, og då kan dei låne kvarandre si form. Feltet låg heilt nake
før — `mix-blend-mode` og ingen ting anna — og det var ikkje lesbart nok.

Regelen er framleis den same, berre med eit tal til: **hovdinga høyrer
til vika og til ingen ting anna.** Kjem det ein tredje stad som ikkje er
timeplanen, er regelen brote — ein fasett som stend paa alt, tyder ikkje
noko paa noko.

### Bobla

Vekefeltet er den fyrste av ei form som har fått namn: **bobla**,
`.bobla` i stilarket.

Ei bobla er eit skrivefelt som ligg *i* noko anna — i ein tittel, i ei
line — og ikkje på eit ark. Ho har det tveegga sporet over, og ei
**mjølkete** flate: `--flate` på 62 % med ei uskarpleik bak. Ei tett
flate gjer sporet til ein boks lagd oppå tittelen; ei gjennomsynleg lèt
det som ligg under skine gjennom, og då les det seg som eit hol i noko.

**Ei bobla tek aldri ringen i merkefargen.** Ikkje i kvile og ikkje i
fokus. Ein ring legg ei hard line tvert over den gjennomsynlege kanten,
og då er det ikkje ei bobla lenger — det er ein boks. At bobla er der er
heile hintet om at ein kan skrive i henne. Fokus seier frå ved at flata
blir eit hakk mindre gjennomsynleg, og ved at markøren blinkar i
`--klas`. På ein telefon kjem talpanelet opp attåt.

Dette er den eine staden i huset der fokus **ikkje** ber merkefargen.

> **Fella.** `box-shadow` lyt stå skriven ut i *begge* tilstandane.
> Sløyfer ein henne i fokus, tek grunnstilen over: `input:focus` set
> `box-shadow: var(--fasett)`, og fasetten har ringen i seg. Ringen vart
> teken bort her ein gong og kom rett tilbake derifrå — ein regel som
> ikkje seier noko, seier det grunnstilen seier.

### Framlegget, ikkje verdien

Talet i bobla er ein `placeholder` og ikkje ein `value`. Det er vika du
*er* i — eit framlegg om kva som ville stått der om du ikkje skreiv noko
— og difor blinkar markøren frå fyrste stund, og fyrste tasten du slår
skriv talet ditt i staden for å leggje seg attmed eit tal som alt var
der.

Feltet valde seg sjølv når det fekk fokus før (`felt.select()`), og då
låg det eit turkist drag over talet i staden for ein markør. Eit merkt
tal ser ikkje ut som noko som ventar på deg; det ser ut som noko du
nettopp gjorde.

Ho stod som ein stipla kasse fyrr. Ein stipla kasse er eit *umriss*, og
eit umriss les seg som noko som skal koma — det motsette av det sanne.

---

## 6. Tvo maatar aa gjera henne gal paa

Baae vart gjorde medan timemerket vart teikna, og baae er lette aa gjera
om att.

**Mjuk skugge og kvitt hovding — paa eit felt.** Fyrste utgaava av
merket hadde ein skugge ovanfraa *og* eit ljost drag nedanfraa, med
fire pikslar uskarpleik, og det saag uppblaast ut. Fasetten her hev
**ikkje** noko hovding: ho er grunn, hard og einsidig, og det er difor
ho les seg som trykt og ikkje som stempla.

> Dette gjeld **felt og merke** — ting som er laga av arket. Det gjeld
> *ikkje* knappen. Ein knapp er ein gjenstand som ligg paa arket, og han
> hev glans med vilje (DESIGN_GUIDELINES §4 og §18). Denne bolken vart
> ei stund lesen som eit ålment forbod mot glans i heile huset, og det
> var aldri det han sa — ljosbandet under hovudet hev vore fire lag
> gradient med utbrend kjerne heile tidi.

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

> Treng ein ny kontroll djupn, skal han nytta eit av dei fire som finst
> — `--fordjuping`, `--fasett`, `--bobla`, `--knappedjup` — og ikkje ein
> ny skugge. Finn du eit `box-shadow` i ein komponent som ikkje er ein
> av dei, eller `--skugge-flytande` paa noko som ikkje faktisk flyt, er
> det ein feil.

Prøva er ei linja:

```sh
grep -rn 'box-shadow' static/css/kjernekraft.css \
  | grep -v 'var(--fasett)\|var(--fordjuping)\|var(--skugge-flytande)' \
  | grep -v 'var(--knappedjup\|var(--bobla' \
  | grep -v 'box-shadow: none'
```

`box-shadow: none` er ikkje ein ny skugge, det er ein som vert teken
burt, so han tel ikkje med.

I dag gjev prøva **fem** treff, og alle fem er meinte:

1. Fana i administrasjonen kastar ein skugge upp og ut, av di ei fana
   faktisk ligg yver arket.
2.–3. Hovdinga i `.vekefelt`, i kvila og i fokus — §5. Ho let seg ikkje
   skriva som eit token, av di ho er tvo `inset`-ar som lyt liggja i den
   same eigenskapen.
4.–5. Halo-en kring lampa som brenn (`.ljos-paa`, `.status-active`) og
   den veike halo-en kring lampa som ventar (`.status-freeze_requested`).
   Dei er *ljoskjelder* og ikkje skuggar: lag med glød utyver, ikkje eitt
   lag mørke nedyver. Ei lampe utan halo lyser ikkje.

   Dei kan ikkje bli token. Halo-en er rekna av `--lampe`, og eit token
   på `:root` får fargen sin *der* — ein custom property blir substituert
   éin gong der han står, og så arva ferdig utrekna. Skulle han fylgje
   lampa, laut han bu i regelen som veit kva ho brenn i. Same grunnen
   til at `--faneflate` og dei andre rekna verdiane står skrivne ut att i
   kvart tema.

Knappen sin djupn stend i `--knappedjup` og `--knappedjup-inn`, og
bobla i feltet i `--bobla` og `--bobla-fokus`. Baae er difor tekne ut av
prøva, som seg hør og bør: dei er token, ikkje lause skuggar.

`--bobla` er fire lag, og alle fire er den same fysikken lesen nedanfrå
og opp. Feltet er ei renne med noko i: fordjupinga er lippa som kastar
ein skugge ned i vatnet, meniskusen i foten er ljoset som kjem ut att
under, glastjukna kring er linsa som pressar det ho ser mot kanten, og
det siste laget er ljoset som slepp ut på arket under henne. Ho er ei
**form**, ikkje ein skugge nokon fann på — og det er difor ho får stå.

Grunngjevingi stend i fila attmed kvar av deim. Alt som kjem ut av
prøva, skal ha ei slik grunngjeving skriven attmed seg — elles er det ein
feil.
