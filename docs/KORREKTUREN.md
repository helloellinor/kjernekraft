# Korrekturen

Dette er den venlegaste forma i systemet, og ho har ikkje hatt eit namn
før no.

Ho ser slik ut: teksten står der han står, du skriv rett på han, og det
du har endra får eit merke i margen. Ingenting er gjort før du seier frå.

Det er korrektur. Du les eit ark, du rettar i det, og rettinga står i
margen til nokon set det på nytt. Difor heiter forma **korrekturen**, og
difor er merket ein strek i margen og ikkje ei ramme rundt feltet.

Ho er brukt tre stader i dag: **Min profil**, **Prising → Medlemskap**
og **Prising → Reglar**.

---

## Dei fem reglane

### 1. Ingen redigeringsmodus

Alt som kan endrast, kan endrast der det står. Det finst ingen
«Rediger»-knapp som byter ut sida med ei anna utgåve av henne.

Før hadde prislista tre knappar per rad — *Rediger pris*, *Lagre*,
*Avbryt* — der dei to siste stod gøymde til du trykte den første. Tre
knappar per rad, ganga med talet på rader, for å endre eitt tal. Og
medan du heldt på med éin, kunne du ikkje sjå kva du hadde gjort med dei
andre.

### 2. Ingen merkelappar

Feltet står i ei setning, ei adresse eller ei liste som alt seier kva
han er. «Medlemer **[kan ▾]** oppgradere til eit dyrare medlemskap» treng
inga overskrift som seier «Oppgraderingsreglar» og inga skildring under
som seier det ein gong til.

Regelen: **er merkelappen det same ordet som feltet står i, er han eit
ord som ikkje arbeider.** Ein merkelapp er berre rett når feltet ikkje
kan seie det sjølv.

### 3. Fordjupinga seier at det er skrivbart

Feltet ligg *i* arket. Det er den vesle djupna — og ikkje ei ramme, ikkje
ein understrek — som seier at her kan du skrive, og ho står der heile
tida, ikkje berre når du er borti feltet. Sjå [FASETTEN.md](FASETTEN.md).

### 4. Det endra feltet ber merket sjølv

Ein strek i margen i `--togu`, den rosa. Ingen tekst seier «du har ulagra
endringar»: feltet *ser* endra ut.

Fargen er ikkje merkefargen. Merkefargen tyder «dette er i live», og ei
ulagra endring er ikkje det — ho er noko som ventar. Og ho er ikkje
åtvaringsfargen heller: du har ikkje gjort noko gale.

Streken står i margen og ikkje rundt feltet, av di ein ramme er lett å
forveksle med fokus. Margen står att når du har gått vidare til neste
felt.

### 5. Dokka tel, og ho listar ikkje

Nedst kjem ei dokk fram med **kor mange** endringar som ventar, og to
handlingar: *Angre* og *Lagre*. Ho seier ikkje kva som er endra — det
står i felta, og dei er framleis på skjermen.

Ein ny rad tel som **éi** endring og ikkje som fire felt. Marker
containeren `data-eitt`.

---

## Det som ikkje kan gjerast om att, vert markert — ikkje gjort

Å slette ein rad slettar han ikkje. Raden vert **merkt**: gjennomstroken,
veik, med åtvaringsfargen i margen, og eit lite «vert sletta» attmed.
Han forsvinn når du lagrar.

Det er dette som gjer forma trygg nok til å sløyfe stadfestingsdialogar.
Den ugjenkallelege handlinga er **å lagre**, og ho er éin knapp på éin
stad, ikkje eitt kryss på kvar rad.

---

## Slik tek du henne i bruk

```html
<form id="mitt-skjema" method="post" action="…"
      data-endringar data-dokk="mi-dokk">

  <input class="lappfelt" name="…" value="…" aria-label="…">

  <aside class="dokk endringsdokk" id="mi-dokk" hidden>
    <div class="dokkinnhald">
      <span class="endringstal" aria-live="polite"
            data-ein="…" data-fleire="…"></span>
      <div class="dokkhandling">
        <button type="button" class="btn" data-angra>Angre</button>
        <button type="submit" class="btn-primary">Lagre</button>
      </div>
    </div>
  </aside>
</form>
```

`endringar.js` tek seg av resten. Dei fire haldepunkta:

| Merke | Kva han gjer |
|---|---|
| `data-endringar` | skjemaet er ein korrektur |
| `data-dokk="id"` | kva dokk som høyrer til |
| `data-eitt` | alt inni tel som éi endring |
| `data-tel` | eit løynt felt skal telje med (til dømes eit slettemerke) |

Kjem det felt til etter at sida er lasta — ein ny rad — send
`endringar:nye` på skjemaet, so tek dokka dei med.

---

## Kvar ho ikkje høyrer heime

- **Der eit steg må gjerast ferdig før det neste gjev meining.** Ei
  registrering er ikkje ein korrektur; ho er ei rekkje.
- **Der lagringa ikkje er eitt kall.** Dokka lovar at alt går saman.
- **Der du ikkje eig arket.** Ein korrektur er noko du gjer i ditt eige
  dokument. Ein søknad til nokon andre er ikkje det.
