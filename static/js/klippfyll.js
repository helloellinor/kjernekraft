// klippfyll.js — «Fyll på» paa eit klippekort.
//
// Knappen fører til kjøpsbolken med den kategorien alt vald, so ein
// slepp aa finna henne att sjølv: `?fane=kjop-klipp&fill=<kategori>`.
//
// Fana var ein emneknagg fyrr — `?fill=…#kjop-klipp` — og daa laut
// spyrjedelen staa fyre knaggen, av di «#kjop-klipp?fill=» legg
// spyrjinga inni knaggen og sida ser henne aldri. Den fella er burte:
// fana er `?fane=` som alle hine no, og tenaren les henne. Sjaa
// handsamarar/faner.go.
//
// KVIFOR HAN BUR HER, OG IKKJE I MALEN:
//
// Han stod som `onclick="fillUpKlippekort(...)"`, og funksjonen stod i
// eit `<script>` inni den same malen som knappen. Paa heimesida gjekk
// det: bolken vert teikna av tenaren, og daa køyrer skriptet med sida.
//
// Paa klippekortsida gjorde det ikkje det. Der vert bolken henta etter
// at sida er lasta og sett inn med `innerHTML` — og `innerHTML` køyrer
// *aldri* `<script>`. Det er ikkje ein lyte i nettlesaren; det stend i
// standarden. Knappane synte seg, funksjonen fanst ikkje, og eit trykk
// gav `ReferenceError` i konsollen og ingen ting paa skjermen.
//
// Difor: éin lyttar paa `document`, sett ein gong naar sida kjem. Han
// bryr seg ikkje um naar knappen kom eller kor mange gonger bolken vert
// bytt ut — han fangar trykket paa vegen upp. Ein lyttar paa sjølve
// knappen hadde hatt den same lyten som skriptet: han fanst ikkje for
// dei knappane som kom seinare.
//
// Han vert send til kvar side og ikkje til ei lista av dei som treng
// honom. Fire hundre teikn er billegare enn den lista: bolken stend paa
// tvo sidor i dag, og den dagen han stend paa ei tridje er det ingen som
// hugsar at det finst ei lista aa føra honom inn i. Same grunnen som
// `namnemeny.js`.
(function () {
    "use strict";

    document.addEventListener("click", function (e) {
        var knapp = e.target.closest("[data-fyll-kategori]");
        if (!knapp) return;
        var kategori = knapp.getAttribute("data-fyll-kategori");
        if (!kategori) return;
        window.location.href =
            "/elev/klippekort?fane=kjop-klipp&fill=" + encodeURIComponent(kategori);
    });
})();
