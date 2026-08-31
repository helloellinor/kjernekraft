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

    function start() {
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
    // vikenummer. Rekninga stend i `veketal.js`.
    //
    // `null` tyder «den vika er gjengi», og daa skal feltet nekta.
    //
    // Kvifor det trengst: `framover` legg til eit heilt aar naar tale
    // peikar bakyver, av di ei vike som er gjengi ikkje finst — og den
    // regelen er rett for eit *skjema*, der ein kann setja upp ein time i
    // veke 3 til aaret medan ein stend i veke 51. Her er han ei fella.
    // Stod du i veke 36 og skreiv 35 — vika som var — rekna han seg fram
    // til veke 35 *neste* aar, gjekk til `?week=51`, og der stend det
    // ingen timar i det heile. Sida vart tom, og tale i adressa hadde
    // ingen ting med tale du skreiv aa gjera.
    //
    // Difor eit tak: eit sprang paa meir enn eit halvt aar er ikkje eit
    // ynske um neste aar, det er ei vike som er gjengi. Under taket stend
    // aarsskiftet: skriv du 2 i veke 51, er det tri vikor fram, og det er
    // framleis rett.
    function offsetTil(v) {
        var steg = Veketal.framover(v, naa, vekerIAar, naaOffset);
        if (steg > vekerIAar / 2) return null;
        return naaOffset + steg;
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
        if (offset === null || offset < 0) { Veketal.nekt(felt); return; }

        var url = new URL(window.location);
        url.searchParams.set("week", offset);
        window.location.href = url.toString();
    }

    Veketal.bind(felt, {
        gyldig: gyldig,
        ferdig: gaa,
        escape: function () { felt.value = ""; opna(false); knapp.focus(); }
    });
    }

    start();
    // Vekeveljaren stend inne i <main>, so han er ny etter kvart
    // sidebyte. Sjaa sideskift.js.
    (window.__sideskift = window.__sideskift || []).push(start);
})();
