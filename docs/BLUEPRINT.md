# Grunnteikningi — heile huset, skjerm for skjerm

*26. august 2026. Kva me hev, kva dei andre hev (YOGO, Momence, TeamUp,
Punchpass, Gymdesk — granska same dagen), og korleis huset skal sjaa ut
naar det er ferdigt. Dette er ei teikning, ikkje ein byggjeplass: ingen
ting her er bygt enno.*

Fyrebiletet er YOGO: minst mogleg system, rein funksjon, og kvar ting
gjord éin gong. Der dei store (Momence, TeamUp) sel marknadsføring og
appar, sel YOGO at studioet verkar. Det er den lina me legg oss paa —
og so skal me vera *smalare* enn YOGO der me kann.

---

## 1. Stoda — kva som verkar, kva som er fasade

Fyrst ærlegdomen. Sumt av det som ser ferdigt ut, er det ikkje:

| Omraade | Stoda |
|---|---|
| Timeplan, booking med kapasitet | **verkar** |
| Timereglar i administrasjonen (lærar, klokke, lengd, skildring, vikar, avlys) | **verkar** |
| Medlemskapsval, byte, fryse-*førespurnad* | verkar for eleven |
| Frysing: godkjenning/avslag hjaa admin | **fasade** — knappane finst, API-et svarar «Not implemented» |
| Venteliste | **fasade** — knappen segjer «Venteliste» naar timen er full; tenaren segjer berre nei |
| Klipp-økonomien | **finst ikkje** — booking trekkjer aldri eit klipp og spør aldri um medlemskap; alt er gratis for den som er innlogga |
| Betaling | **fasade** — betalingsmaatane er mock, klippekortkjøp skriv i basen utan aa ta pengar, `last_billed` finst men ingen krev inn |
| Lærarar | **fri tekst** paa timen. `roles`-tabellen og tildelings-API finst, men ingi flata styrer deim |
| Frammøte, skranke/POS, kvitteringar, rapportar, e-post | **finst ikkje** |
| Klippekortpakkor | ligg i basen (`klippekort_packages`) **og** hardkoda i HTML — tvo sanningar |

## 2. Bordpengane — det alle fem systemi hev

Fraa granskingi: dette hev kvart einaste av dei fem systemi, ogso dei
minste. Det er lista yver kva eit studio *ventar* seg:

| Bordpengar | Me i dag |
|---|---|
| Attkomande timar med einskild-unnatak og avlysing | ✔ (timereglane) |
| Sjølvbooking med kapasitet, bookingvindauga og avmeldingsfrist | delvis — kapasitet ja, vindauga/frist nei |
| Venteliste med opprykk og melding til den som rykkjer upp | ✘ |
| Medlemskap (attkomande trekk) + klippekort med utløp + drop-in | delvis — produkti finst, pengane gjer ikkje |
| Prøvetilbod og frysing/pause | delvis (hausttilbod finst; frysing halvveges) |
| Kortbetaling paa nett + lagra kort + attkomande trekk | ✘ |
| Kundeprofil med frammøtehistorie; eigne lærar-innloggingar | ✘ |
| Automatisk e-post: stadfesting, avlysing, opprykk, paaminning | ✘ |
| Enkel rapport: pengar og frammøte, med eksport | ✘ |
| Open timeplan utan innlogging | ✘ (alt ligg attum innlogging) |
| Handsaming av sein avmelding / ikkje-møtt (golvet er: mogleg aa merkja) | ✘ |

## 3. Det me hoppar yver — med vilje

Det berre dei store hev, og som eit lite studio ikkje treng:

- **Marknadsføringspakkor** — tovegs SMS-innboks, «AI-leads»,
  nyhendebrev med segmentering. E-posten vaar er transaksjonell, punktum.
- **Merkevare-app** — tillegg hjaa alle (≈100 €/mnd); nettsida er app-en.
- **Video/livestream-bibliotek.**
- **Vareteljing og strekkodar i kassa** — skranken sel dei tri produkti
  vaare, ikkje t-skjorter med lagerstyring.
- **Dørlaas-integrasjon.**
- **Automatisk gebyr-innkrevjing med gjenforsøk** (Momence prøver 3×) —
  golvet er aa *merkja* ikkje-møtt; gebyr kann koma seinare, um nokon
  gong.
- **Fleirmodus-ventelistor** (Momence hev tri strategiar) — YOGO og
  Gymdesk greier seg med fyrstemann-i-køen + melding. Me òg.

## 4. Grunnsteinane — det som gjer at alt klikkar

Fem grunnsteinar. Kvar ny funksjon skal staa paa ein av deim; treng ho
ein sjette, er ho truleg feil tenkt.

### 4.1 Personen — éin tabell, roller uppaa

Ein lærar er ikkje eit namn i eit fritekstfelt; han er ein **brukar med
rolla lærar**. Same innlogging, same profil, same folk-liste.

- Rolla vert slegen av og paa i folk-lista (API-et
  `/users/assign-role` finst alt — det manglar berre knappen).
- Alle stader som spør etter ein lærar — ny time, vikarfeltet — hentar
  **brukarar med lærarrolla**, ikkje fri tekst. `teacher_name` paa
  timen stend att som historisk sanning (den som heldt timen, den
  dagen), men *veljarane* les rollone.
- Den som hev lærarrolla ser sine eigne timar og frammøtelista si
  (§6). Ikkje eit eige system: same hus, eitt rom til.

### 4.2 Produktet — tri slag, tri fargar

Stilboki §2 hev alt teikna dette: **medlemskapet** (turkis, gjeng og
gjeng), **klippekortet** (rosa, vert talt), **kurs/PT** (lilla, hender
ein gong). Alt som vert selt er eitt av dei tri. Drop-in er ikkje eit
fjerde produkt — det er eit klippekort med eitt klipp (pakkorne finst
alt i basen). Skranken (§6) sel dei same tri produkti gjenom dei same
kodevegane som nettkjøpet; skilnaden er berre betalingsmaaten.

Pakkorne og prisane bur i **basen aaleine**; klippekort-sida sluttar aa
hardkoda deim (F-punkt i §7).

### 4.3 Bookingi — éin rad, fire tilstandar

Ei paamelding er éi rad i `event_signups` med ein tilstand, og
tilstandane ber formene fraa stilboki §2:

| Tilstand | Form | Tyder |
|---|---|---|
| paameld | ● fylt | plassen er din, dekt av medlemskap eller klipp |
| ventande | ○ open ring | paa venteliste; ingen ting trekt |
| frammøtt | ✔ | du kom; klippet er endeleg brukt |
| ikkje møtt | ⋯ prikka | du kom ikkje; klippet gjekk tapt (regel i §5) |

**Dekkjinga** er kjernen: aa booka *kostar* noko. Ved paamelding spør
systemet i denne rekkjefylgdi: (1) aktivt medlemskap som dekkjer
kategorien → book; (2) klippekort med klipp att i kategorien → trekk
eitt klipp og book; (3) ingen av delane → vis vegen til kjøp. Rettidig
avmelding gjev klippet attende; sein avmelding (etter fristen, §5) gjer
det ikkje. Kva eit medlemskap *dekkjer* er ein eigenskap paa
medlemskapstypen (nytt felt: kategoriar).

**Ventelista** er den enklaste som finst (YOGO/Gymdesk-modellen): full
time → «Venteliste»-knappen (som alt stend der) legg deg til som
ventande. Naar nokon melder seg av, rykkjer fyrstemann med gyldig
dekkjing upp automatisk: klippet vert trekt *daa*, e-post gjeng ut, og
den som ikkje lenger hev dekkjing vert hoppa yver. Ingen krav um svar,
ingen fristar med nattestenging — det er TeamUp-kompleksitet.

**Avlysingi** knyter alt saman: naar admin avlyser ein time (knappen
finst alt), fær kvar paameld e-post, og kvart trekt klipp gjeng
attende av seg sjølv. I dag slettar knappen radi og ingen fær vita
noko — det er den største samanhengsfeilen i huset.

### 4.4 Pengane — éi bok

`charges` er **hovudboki**. Alt som kostar noko vert ei rad der, same
kvar det hende: nettkjøp (Stripe), skranken (kontant, terminal, Vipps),
maanadstrekket, eit seinare gebyr. Rada ber kjelda og betalingsmaaten.

- **Kvitteringi** er ei charge-rad synt fint (og skrivbar ut —
  `@media print` finst alt). Eleven ser sine under Betaling; skranken
  kann skriva ut den siste.
- **Maanadstrekket** er ein jobb som gjeng dagleg: finn medlemskap der
  `last_billed` + syklusen er forfalle, trekk kortet, skriv rada.
  Feilar trekket → medlemskapet vert merkt (form: prikka), eleven fær
  e-post, admin ser det i oversynet. Ingen automatisk purring — lista
  til admin *er* purringi, i fyrste umgang.
- Stripe er einaste nett-integrasjonen. Skranken *registrerer*
  betalingsmaaten (kontant/terminal/Vipps) utan integrasjon — YOGO
  sin SMS-betalingslenkje er fin, men det er ein seinare luksus.

### 4.5 Reglane og meldingane

**Reglane** bur alle paa same staden og i same forma: timereglane
(finst), medlemskapsreglane (finst), og dei nye **bookingreglane** —
kor tidleg ein kann booka, avmeldingsfristen, kva som hender med
klippet ved ikkje-møtt. Tri tal og eit val, i Innstillingar-fana.

**Meldingane** er éin modul med éi liste utløysingar, alle
transaksjonelle: stadfesting ved booking, avlysing av time, opprykk
fraa venteliste, paaminning (valfri), «tvo klipp att / kortet gjeng
snart ut» (Punchpass-regelen: det som kjem fyrst), mislukka trekk.
E-post fyrst; SMS er ein pengeslukande luksus me ventar med. (Gløymt
passord vert verande mailto til studioet — det *er* barebones, og det
verkar.)

---

## 5. Skjerm for skjerm — elevsida

**Timeplanen** (`/elev/timeplan`) — som i dag, pluss: dokka syner
**skildringi** (ho finst no, men ingen ser henne), dekkjinga di («dekt
av medlemskapet» / «brukar 1 klipp — 6 att» / «kjøp klipp»), og
fristane. Full time → venteliste-knappen verkar (§4.3). Dagmerket ber
alt teikni for full/umme; ventande fær den opne ringen. Timeplanen
vert dessutan **open utan innlogging** (bordpengar hjaa alle fem):
same sida, berre med «logg inn for aa booka» i dokka.

**Hjem** (`/elev/hjem`) — som i dag, pluss: ventande plassar synte med
open ring, og eit varsel naar eit trekk hev feila.

**Klippekort** (`/elev/klippekort`) — pakkorne kjem fraa basen (ikkje
HTML), kjøpet gjeng gjenom Stripe og skriv i hovudboki, og maalaren
varslar naar det er tvo klipp att eller utløpet nærmar seg.

**Medlemskap** (`/elev/medlemskap`) — som i dag, pluss: frysing med
start- og sluttdato (YOGO-forma; førespurnaden finst alt), og synleg
bindings- og oppseiingstid.

**Betaling** (`/elev/betaling`) — ekte kort (Stripe), ekte
belastningar, og kvar rad kann opnast som kvittering.

**Min profil / vilkaar / innlogging** — ferdige som dei er.

## 6. Skjerm for skjerm — lærarsida

Ingen ny stad: lærar-rolla laasar upp **Mine timar** (lekkja i
leidingi). Lista er reglane hans, same form som admin ser, men berre
hans — og kvar dag opnar **frammøtelista**: namni paa dei paamelde,
eitt trykk per namn: møtt ✔ / ikkje møtt ⋯. Det trykket er heile
frammøtesystemet, og det er det som gjer rapportane (§7) og
ikkje-møtt-regelen (§4.3) sanne. Vikarbyte melder han til studioet som
i dag — YOGO sitt lærar-til-lærar-byte er fint, men det er eit system
til, og me hev alt vikarfeltet hjaa admin.

## 7. Skjerm for skjerm — administrasjonen

Fanearket veks fraa fem til seks fanor. Alt anna er utviding av rom
som alt finst.

**Oversyn** — tali (finst), frysingar som ventar (finst — men
knappane skal *verka*), og nytt: mislukka trekk. Alt som krev eit svar
fraa deg, paa éin stad.

**Timeplan** — timereglane (ferdige). Nytt inni regelen: dagveljaren
fær «frammøte»-lekkja som opnar same lista som læraren ser — admin
skal ikkje ha eit eige frammøtesystem.

**Skranken** *(ny fane — POS)* — bygd av deler som finst:

1. **Søk** (same felt som folk-lista, autofokus) → treff paa namn.
   Ingen treff → hurtigregistrering: namn, e-post, telefon. Ferdig.
2. **Sel**: tri knappar i produktfargane — medlemskap, klippekort,
   drop-in (= 1-klipps-pakka + booking paa timen som gjeng no).
   Pakkeval fraa basen, same data som elevsida.
3. **Betaling**: kontant / terminal / Vipps — eit val, ikkje ein
   integrasjon. Rada gjeng i hovudboki med kjelda `skranken`,
   produktet vert aktivt med ein gong, kvitteringi kann skrivast ut.

Heile skranken er dei tri stegi. Ho deler kvar einaste kodeveg med
nettkjøpet; det einaste ho eig sjølv er skjermbiletet.

**Prisar** — som i dag (er alt under umbygging), pluss:
klippekortpakkorne fraa basen vert redigerbare her, og hausttilbodet
vert til eit allment prøvetilbod-felt.

**Folk** — folk-lista (finst), pluss i den opne rada: **rollone som
merke ein kann slaa av og paa** (lærar, admin — held-merki finst alt
som form), og historia: bookingar, frammøte, belastningar. Dette er
svaret paa «eg kann ikkje gjera nokon til lærar».

**Innstillingar** — tidssona (finst), pluss bookingreglane (§4.5) og
studioet sin e-postavsendar.

**Rapportar** *(inne i Oversyn, ikkje eiga fane fyrr det trengst)* —
tvo spursmaal, tvo svar: **pengane** (hovudboki summert per maanad og
produkt) og **frammøtet** (per regel og per lærar — det siste *er*
lønsgrunnlaget, Punchpass-modellen: rapporten er løns-input, ikkje
lønssystem). Alt kann lastast ned som CSV.

---

## 8. Byggjerekkjefylgdi

Kvart steg gjev verdi aaleine og byggjer paa det fyrre; ingen ting her
er starta.

1. **Lærarar som brukarar** — rolleknappane i folk, veljarane les
   rollone. Lite arbeid, laaser upp §6.
2. **Bookingøkonomien** — dekkjing, klippetrekk, ekte venteliste,
   avlysing som varslar og refunderer. (Krev 3 for varslingi.)
3. **E-postmodulen** — utløysingane i §4.5. Fyresetnad for alt etterpaa.
4. **Stripe og maanadstrekket** — ekte pengar inn, hovudboki vert sann.
5. **Skranken** — no er ho berre eit skjermbilete framfor ferdige vegar.
6. **Frammøtet** — lærarsida og eitt-trykks-lista.
7. **Rapportane** — les berre det som alt ligg der.

Og tvo ryddingar som ikkje treng venta paa noko: klippekort-sida skal
lesa pakkorne fraa basen, og frysingsknappane hjaa admin skal verka.
