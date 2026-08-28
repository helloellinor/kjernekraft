// Vikefeltet.
//
// Ein skriv talet paa vika ein vil til, som eit sidetal i ei bok. Vart
// det ei nedfallsliste, laut ein fyrst opna henne og so finna raden —
// tvo steg for noko ein alt kunde skriva.
//
// Han gjeng fyrst naar tale er ferdig skrive: «3» er ikkje eit ynske um
// veke tri, det er halve veg til «35». Difor ventar han til feltet mistar
// fokus, til Enter, eller til tale ikkje kann verta lengre.
//
// Feltet stend tomt. Vika ein er i er eit *framlegg* — han ligg i
// placeholder — og difor blinkar markøren fraa fyrste stund og fyrste
// tasten skriv tale ditt. Fyrr laag vika i verdien og feltet valde seg
// sjølv naar det fekk fokus, og daa laag det eit turkist drag yver tale i
// staden for ein markør: eit merkt tal ser ikkje ut som noko som ventar
// paa deg, det ser ut som noko du nettupp gjorde.
(function () {
    "use strict";

    var felt = document.getElementById("veke-felt");
    var knapp = document.getElementById("veke-knapp");
    var meny = document.getElementById("veke-meny");
    if (!felt || !knapp || !meny) return;

    function opna(vis) {
        meny.hidden = !vis;
        knapp.setAttribute("aria-expanded", vis ? "true" : "false");
        if (vis) felt.focus();
    }

    knapp.addEventListener("click", function () { opna(meny.hidden); });
    document.addEventListener("click", function (e) {
        if (!meny.hidden && !meny.contains(e.target) && e.target !== knapp) opna(false);
    });

    var naa = parseInt(felt.dataset.veke, 10);
    var naaOffset = parseInt(felt.dataset.offset, 10);
    var vekerIAar = parseInt(felt.dataset.veker, 10) || 52;

    // Kva `?week=`-offset skal til for aa koma til vike v?
    //
    // Tenaren tel vikor fram fraa den ein stend i, so det er skilnaden i
    // vikenummer. Golvet er minus den vika ein alt hev bladd fram til:
    // lenger attende enn den vika ein *er* i finst ikkje. Regelen og
    // rekninga stend i `veketal.js`.
    function offsetTil(v) {
        return naaOffset + Veketal.framover(v, naa, vekerIAar, naaOffset);
    }

    function gyldig(v) { return Veketal.gyldig(v, vekerIAar); }

    function gaa() {
        var v = parseInt(felt.value, 10);
        if (v === naa || (isNaN(v) && felt.value.trim() === "")) {
            felt.value = "";    // det same, eller ingen ting — attende til framlegget
            return;
        }
        if (!gyldig(v)) { Veketal.nekt(felt); return; }
        var offset = offsetTil(v);
        if (offset < 0) { Veketal.nekt(felt); return; }

        var url = new URL(window.location);
        url.searchParams.set("week", offset);
        window.location.href = url.toString();
    }

    Veketal.bind(felt, {
        gyldig: gyldig,
        ferdig: gaa,
        escape: function () { felt.value = ""; opna(false); knapp.focus(); }
    });
})();
