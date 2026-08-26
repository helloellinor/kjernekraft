// Vikefeltet.
//
// Ein skriv talet paa vika ein vil til, som eit sidetal i ei bok. Vart
// det ei nedfallsliste, laut ein fyrst opna henne og so finna raden —
// tvo steg for noko ein alt kunde skriva.
//
// Han gjeng fyrst naar tale er ferdig skrive: «3» er ikkje eit ynske um
// veke tri, det er halve veg til «35». Difor ventar han til feltet mistar
// fokus, til Enter, eller til tale ikkje kann verta lengre.
(function () {
    "use strict";

    var felt = document.getElementById("veke-felt");
    if (!felt) return;

    var naa = parseInt(felt.dataset.veke, 10);
    var naaOffset = parseInt(felt.dataset.offset, 10);

    function gyldig(v) {
        return !isNaN(v) && v >= 1 && v <= 53;
    }

    function gaa() {
        var v = parseInt(felt.value, 10);
        if (!gyldig(v) || v === naa) {
            felt.value = naa;   // ikkje eit tal, eller det same — attende
            return;
        }
        // Offset fraa vika ein staar i. Studioet syner ikkje vikor som er
        // gjengne, so eit tal attende endar paa den fyrste ein kann sjaa.
        var offset = naaOffset + (v - naa);
        if (offset < 0) { felt.value = naa; return; }

        var url = new URL(window.location);
        url.searchParams.set("week", offset);
        window.location.href = url.toString();
    }

    felt.addEventListener("focus", function () { felt.select(); });
    felt.addEventListener("blur", gaa);
    felt.addEventListener("keydown", function (e) {
        if (e.key === "Enter") { e.preventDefault(); felt.blur(); }
        if (e.key === "Escape") { felt.value = naa; felt.blur(); }
    });

    // Tvo siffer *er* ferdig skrive; daa treng ein ikkje trykkja noko.
    felt.addEventListener("input", function () {
        felt.value = felt.value.replace(/[^0-9]/g, "");
        if (felt.value.length === 2 && gyldig(parseInt(felt.value, 10))) gaa();
    });
})();
