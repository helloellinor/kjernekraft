# Ettersynet — dokumentstrukturen

*26. august 2026. Alle malar under `handlers/templates/`, den lause
`templates/events.html`, og heile `static/css/kjernekraft.css`.*

Dette er granskingi som ligg attum §15–§17 i stilboki. Stilboki segjer
kva som skal gjelda; dette arket segjer kvar det ikkje gjeld i dag, med
fil og lina, og i kva rekkjefylgd det løner seg aa retta.

Hovudfunnet fyrst: **kjensla av tilfeldig rom kjem ikkje av tali.**
Trappa `--rom-1`–`--rom-7` er i bruk mest yveralt; i heile maltreet
finst det nett tvo harde romverde i `style=""` (den eine med token).
Kjensla kjem av *strukturen*: same rolla paa ulike steg, boksar i
boksar, og daude lag som ingen ser men alle kjenner.

---

## Friskt

- Romtrappa er tokenfest og trappar sjølv ned under 48 rem.
- Kort er lina-paa-arket alt (§4): ingen skuggar, inga runding utyver
  det vesle.
- Reinaste sidone, som alt fylgjer §15/§16 utan aa vita det:
  `timeplan.html`, `min-profil.html`, `vilkaar.html`,
  `innlogging.html`, `gloymt-passord.html`, `registrering.html` — og
  dashbordet, som er *mønsteret*: yverskrift beint paa arket, rutenet
  under, ingen boks kring.
- Berre `dagmerke` og `language_selector` er ekte attbruk i dag — men
  dei er rett gjorde baae tvo.

---

## Funni

### F1 — éi rolla, tri namn, tvo steg

`.section-title`, `.module-title` og `.step-title` stend i same veljaren
(`kjernekraft.css:467`) og er piksel-like. I malarne:

| Skrivemaate | Kvar |
|---|---|
| `h2.section-title` | dashbord-modulane, `admin-events-table`, `admin-freeze-requests-table`, arket |
| **`h3`**`.section-title` | `admin-class-management`, `admin-membership-rules`, `admin-pricing-management`, `admin_settings` |
| `h2.module-title` | `charges-container`, `payment-methods-container`, `klippekort.html`, `membership.html` |
| `h2.step-title` | `klippekort.html` ×7 |

Inne i *same* faneark paa admin-sida stend «oversyn»-bolken paa `h2` og
«prisar»-bolken paa `h3`, i same storleik. → Alt vert `h2.section-title`.

### F2 — korttitlar paa h3 og h4 um einannan

`h3.package-name` (klippekort-sida) men `h4.package-name` (arket);
`h3.membership-name` men `h4.card-title` — tvo kort side um side i
verkstaden med kvar sitt steg. → Kort i ein bolk er `h3`, alltid.

### F3 — tvo h1 paa same sida

`klippekort.html:5,26` («Behandle klipp» / «Kjøp klipp») og
`membership.html:7,29` («Behandle medlemskap» / `{{.PageTitle}}`).
→ Éin `h1`, og gjeremaal nummer tvo vert ein `h2`-bolk.

### F4 — dørsidone hev sin eigen tittelklasse

`.login-title` er fast 2,4 rem (`kjernekraft.css:1418`) der
`.page-title` er `clamp(2rem, 5vw, 3.4rem)`. Same rolla, onnor klasse,
onnor aatferd naar glaset krympar. → `.page-title` ogso her.

### F5 — nakne yverskrifter

Fell attende paa elementstilen og liknar ingen ting anna paa sida:

| Stad | Kva |
|---|---|
| `klippekort.html:307` | `h2.purchase-title` — klassen finst ikkje i stilarket; 1,8 rem millom seks `h2.step-title` paa 1,05 rem |
| `klippekort.html:309` | `h3` utan klasse (`#selected-package-name`) |
| `registrering.html:85` | `h2` i vilkaars-dialogen |
| `modules/membership/membership.html:80` | `h3` i `.no-membership` |

### F6 — boksen i boksen

`.module` (flata + haarlina + `--rom-5` fyll) kring kort som sjølve er
flata + haarlina + `--rom-5`:

- `klippekort.html`: `.module.no-border` > `.klippekort-card`
- `membership.html`: `.module.no-border` > `.membership-card`
- `betaling.html`: `.module` > `.charge-item`
- admin: `.fanerom` (boksen) > `.rule-section` (boks til)

Og `no-border` (`kjernekraft.css:581`) tek berre *botnkanten* — boksen
stend att med tri kantar, bakgrunn og fyll. → §16: yverskrifti og
rutenetet beint paa arket, som paa dashbordet. `.rule-section` vert
gruppa med `h3.undertittel` og detaljrader; fanerommet er boksen.

### F7 — daude lag

**28 klassar i malarne hev ingen CSS-regel i det heile**, leivningar
fraa det gamle stilarket, m.a.: `.sida`, `.step`, `.purchase-section`,
`.checkout-container`, `.question-form`, `.categories-grid`,
`.top-section`, `.current-membership-section`, `.admin-section`,
`.rules-container`, `.rule-item`, `.class-container`,
`.settings-form`, `.membership-module`, `.charges-module`,
`.special-offer-notice`, `.selected-details` …

Verste rekkja er fire daude lag i rad, tvo stader:
`admin-class-management.html` (`.class-management-section` >
`.class-container` > `.upcoming-classes` > `.classes-table-container`)
og `admin-membership-rules.html` (`.admin-section` > `.rules-container`
> … > `.rule-item`). → §16: maala eller leggja ut, elles burt.

### F8 — abstraksjonen som skulde hindra alt dette er daud

`components/common/standard-module.html` (`standard_module`) er den
allmenne kort-malen — og hev **null kallarar**. Kvar brukar handskriv
`<div class="module"><h2 class="module-title">` i staden, seks stader,
og difor dreiv dei fraa einannan. Sidan modul-boksen gjeng ut med §16,
er svaret aa *sletta* malen, ikkje taka honom i bruk.

### F9 — events.html stend utanfor huset

`templates/events.html` er eit eige dokument med eige `<style>`:
`#ccc`, `15px`, `800px` — null token, null tema, null i18n, nakne
yverskrifter. → Anten inn under `base` med rolleklassane, eller burt um
sida ikkje lenger er i bruk.

### F10 — hardkoda yverskriftstekst (§11)

Alle 26 yverskriftene paa `klippekort.html`, 5 av 7 paa
`membership.html`, og «Systeminnstillinger» i `admin_settings.html:3`
(som ogso er den einaste fila med understrek i namnet millom
bindestrek-grannar). Ingen av deim gjeng gjenom `{{t}}`.

---

## Arbeidslista

I rekkjefylgd; kvart steg kann gjerast for seg og sida verkar etterpaa.

**1. Stegi og klassane (berre semantikk, ingen synleg skilnad)**
- `module-title`/`step-title`/`purchase-title` → `section-title`;
  synonymkjedone i `kjernekraft.css:467,552,1498` kortast inn.
- `h3.section-title` → `h2` (dei fire admin-modulane);
  `.undertittel` → `h3` yveralt (ni stader `h4`, ein stad `h3`).
- Korttitlar → `h3` (`card-title` i klippekort-modulen og arket;
  `package-name` i arket).
- Dei fire nakne yverskriftene (F5) fær rolleklasse.
- `klippekort.html` og `membership.html`: andre `h1` → `h2.section-title`.
- `.login-title` → `.page-title`; regelen `kjernekraft.css:1418` gjeng ut.
- Arket (verkstaden) rettar prøvone sine til dei same stegi.

**2. Boksarne (F6 — det synlege)**
- `betaling.html`, `klippekort.html`, `membership.html`: `.module`-laget
  kring AJAX-innhaldet burt; `h2.section-title` + behaldaren beint paa
  arket. `no-border` gjeng ut or stilarket.
- `admin-membership-rules.html`: `.rule-section`-boksane vert grupper
  med `h3.undertittel`; fanerommet er boksen.

**3. Daude lag (F7)**
- Alle 28 klasselause divane burt der dei korkje maalar eller legg ut;
  dei tvo fire-i-rad-kjedone fyrst. `.sida` gjeng ut or alle malar
  (`main` er sida).

**4. Husreinsking**
- `standard-module.html` slettast (F8).
- `admin_settings.html` → `admin-settings.html`, og «Systeminnstillinger»
  gjenom `{{t}}` (F10 — resten av F10 er sitt eige, større arbeid).
- `templates/events.html`: inn i huset eller ut or repoet (F9).
- `membership.html:92`: `margin-top: 1rem` inline → token i stilarket.

Prøvone som vaktar alt dette stend i §15; køyr deim etter kvart steg.
