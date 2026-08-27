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
    // Tenaren tel vikor fram fraa den ein stend i, so det er
    // skilnaden i vikenummer — heilt til aaret skiftar. Veke 2 sedd
    // fraa veke 51 er tri vikor *fram*, ikkje ni og førti attende, og
    // fyrr rekna dette ut eit stort negativt tal som vart avvist i
    // stilla. Nær aarsskiftet legg me difor aaret til.
    function offsetTil(v) {
        var skilnad = v - naa;
        if (skilnad < 0 && naa > vekerIAar - 12 && v < 13) {
            skilnad += vekerIAar;
        }
        return naaOffset + skilnad;
    }

    // Ei vike som er gjengi finst ikkje i timeplanen. Fyrr sette feltet
    // seg berre attende og sa ingen ting, og daa ser det ut som um
    // feltet ikkje verkar. Bobla fer ein augneblink i aatvaringsfargen:
    // ingen ring, og ingi melding som lyt lesast og lukkast.
    function nekt() {
        felt.value = "";
        felt.classList.add("nekta");
        setTimeout(function () { felt.classList.remove("nekta"); }, 900);
    }

    function gyldig(v) {
        return !isNaN(v) && v >= 1 && v <= 53;
    }

    function gaa() {
        var v = parseInt(felt.value, 10);
        if (!gyldig(v) || v === naa) {
            felt.value = "";    // ikkje eit tal, eller det same — attende til framlegget
            return;
        }
        var offset = offsetTil(v);
        if (offset < 0) { nekt(); return; }

        var url = new URL(window.location);
        url.searchParams.set("week", offset);
        window.location.href = url.toString();
    }

    // Markøren, ikkje eit merke. Feltet er tomt naar det opnar, so dette
    // set berre markøren der han skal vera og hindrar at nettlesaren
    // merkjer noko av seg sjølv.
    felt.addEventListener("focus", function () {
        felt.setSelectionRange(felt.value.length, felt.value.length);
    });
    felt.addEventListener("blur", gaa);
    felt.addEventListener("keydown", function (e) {
        if (e.key === "Enter") { e.preventDefault(); felt.blur(); }
        if (e.key === "Escape") { felt.value = ""; opna(false); knapp.focus(); }
    });

    // Tvo siffer *er* ferdig skrive; daa treng ein ikkje trykkja noko.
    felt.addEventListener("input", function () {
        felt.value = felt.value.replace(/[^0-9]/g, "");
        if (felt.value.length === 2 && gyldig(parseInt(felt.value, 10))) gaa();
    });
})();
