// Dokka.
//
// Ho held det du ser paa i ro medan resten rullar. Du vel ein dag i
// timeplanen, og timen stend i dokka til du vel ein annan — so du kann
// samanlikna tysdag og torsdag utan at noko flyttar seg, og knappen er
// alltid paa den same staden.
//
// Høgdi hennar vert *maald* og skrivi attende til --dokkhogd. Sida fær
// luft under seg av det talet. Sette me høgdi og lufti av same fasttalet,
// vilde eit litt lengre namn skuva innhaldet under dokka.
(function () {
    "use strict";

    var dokk = document.getElementById("timedokk");
    if (!dokk) return;

    var felt = {
        tittel: document.getElementById("dokk-tittel"),
        laerar: document.getElementById("dokk-laerar"),
        rom: document.getElementById("dokk-rom"),
        dag: document.getElementById("dokk-dag"),
        klokke: document.getElementById("dokk-klokke"),
        maalar: document.getElementById("dokk-maalar"),
        teke: document.getElementById("dokk-teke"),
        av: document.getElementById("dokk-av"),
        knapp: document.getElementById("dokk-knapp")
    };

    var valdMerke = null;

    function maalDokka() {
        document.documentElement.style.setProperty(
            "--dokkhogd", dokk.hidden ? "0px" : dokk.offsetHeight + "px");
    }

    function teiknMaalar(teke, plassar) {
        // Éi rute per plass, som i klippekortmaalaren. Her er ei tend
        // rute ein teken plass, so ein full time er heilt tend.
        felt.maalar.textContent = "";
        for (var i = 0; i < plassar; i++) {
            var r = document.createElement("span");
            r.className = i < teke ? "plass teken" : "plass";
            felt.maalar.appendChild(r);
        }
    }

    function opna(merke) {
        var d = merke.dataset;

        felt.tittel.textContent = d.tittel;
        felt.laerar.textContent = d.laerar;
        felt.rom.textContent = d.rom || "";
        felt.dag.textContent = d.dag;
        felt.klokke.textContent = d.klokke;

        var teke = parseInt(d.teke, 10) || 0;
        var plassar = parseInt(d.plassar, 10) || 0;
        teiknMaalar(teke, plassar);
        felt.teke.textContent = teke;
        felt.av.textContent = "/" + plassar;

        var full = d.full === "true";
        felt.knapp.textContent = full ? felt.knapp.dataset.venteliste : felt.knapp.dataset.meldPaa;
        felt.knapp.className = full ? "btn-danger" : "btn-primary";
        felt.knapp.dataset.event = d.event;

        if (valdMerke) valdMerke.setAttribute("aria-pressed", "false");
        merke.setAttribute("aria-pressed", "true");
        valdMerke = merke;

        dokk.hidden = false;
        maalDokka();
    }

    function lukk() {
        dokk.hidden = true;
        if (valdMerke) valdMerke.setAttribute("aria-pressed", "false");
        valdMerke = null;
        maalDokka();
    }

    document.addEventListener("click", function (e) {
        var merke = e.target.closest(".dagmerke:not([disabled])");
        if (merke) { opna(merke); return; }
        if (e.target.closest("#dokk-lukk")) lukk();
    });

    // Escape lukkar. Ei dokk som ikkje lèt seg lukka med tastaturet er
    // ei dokk som stend i vegen.
    document.addEventListener("keydown", function (e) {
        if (e.key === "Escape" && !dokk.hidden) lukk();
    });

    window.addEventListener("resize", maalDokka);
    maalDokka();
})();
