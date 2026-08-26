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
        melding: document.getElementById("dokk-melding"),
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

        teiknKnapp(merke);

        if (valdMerke) valdMerke.setAttribute("aria-pressed", "false");
        merke.setAttribute("aria-pressed", "true");
        valdMerke = merke;

        syn("");
        dokk.hidden = false;
        maalDokka();
    }

    // Knappen fylgjer stoda paa merket, so han og merket aldri kann
    // segja kvar sitt.
    function teiknKnapp(merke) {
        var d = merke.dataset;
        var paameld = d.paameld === "true";
        var full = d.full === "true";

        felt.knapp.dataset.event = d.event;
        felt.knapp.dataset.handling = paameld ? "meld-av" : "meld-paa";
        felt.knapp.disabled = false;

        if (paameld) {
            felt.knapp.textContent = felt.knapp.dataset.meldAv;
            felt.knapp.className = "btn-danger";
        } else if (full) {
            felt.knapp.textContent = felt.knapp.dataset.venteliste;
            felt.knapp.className = "btn";
        } else {
            felt.knapp.textContent = felt.knapp.dataset.meldPaa;
            felt.knapp.className = "btn-primary";
        }
    }

    function syn(bod, gale) {
        felt.melding.textContent = bod || "";
        felt.melding.className = gale ? "dokkmelding gale" : "dokkmelding";
    }

    // Paamelding gjeng til tenaren. Fyrr flytte skriptet berre kort
    // kring i dokumentet: du saag deg sjølv paameld, tenaren visste
    // ingen ting, og neste lasting sa noko anna.
    async function send(handling, eventID) {
        var url = handling === "meld-av" ? "/api/events/cancel-signup" : "/api/events/signup";
        var svar = await fetch(url, {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: "event_id=" + encodeURIComponent(eventID)
        });
        var tekst = await svar.text();
        if (!svar.ok) throw new Error(tekst);
        return tekst;
    }

    function oppdaterMerke(merke, delta) {
        var teke = (parseInt(merke.dataset.teke, 10) || 0) + delta;
        var plassar = parseInt(merke.dataset.plassar, 10) || 0;
        if (teke < 0) teke = 0;

        merke.dataset.teke = teke;
        merke.dataset.paameld = delta > 0 ? "true" : "false";
        merke.dataset.full = teke >= plassar ? "true" : "false";
        merke.classList.toggle("full", teke >= plassar);

        var tal = merke.querySelector(".dagtal");
        if (tal) tal.innerHTML = teke + '<span class="av">/' + plassar + "</span>";
        var fyll = merke.querySelector(".dagfyll");
        if (fyll && plassar) fyll.style.setProperty("--fyll", Math.round(teke * 100 / plassar) + "%");

        teiknMaalar(teke, plassar);
        felt.teke.textContent = teke;
        teiknKnapp(merke);
    }

    function lukk() {
        dokk.hidden = true;
        if (valdMerke) valdMerke.setAttribute("aria-pressed", "false");
        valdMerke = null;
        maalDokka();
    }

    document.addEventListener("click", async function (e) {
        // .daud er hol og ikkje knappar. [disabled] fangar deim ikkje:
        // dei er <span>, og attributtet gjeld berre skjemaelement.
        var merke = e.target.closest(".dagmerke:not(.daud):not([disabled])");
        if (merke) { opna(merke); return; }
        if (e.target.closest("#dokk-lukk")) { lukk(); return; }

        if (e.target.closest("#dokk-knapp") && valdMerke) {
            var knapp = felt.knapp;
            var handling = knapp.dataset.handling;
            knapp.disabled = true;
            syn("");
            try {
                await send(handling, knapp.dataset.event);
                oppdaterMerke(valdMerke, handling === "meld-av" ? -1 : 1);
            } catch (feil) {
                // Tenaren segjer kvifor — for tidleg, fullt, alt paameld.
                // Bodskapen hans er meir nyttig enn ein me finn paa.
                syn(feil.message, true);
                knapp.disabled = false;
            }
        }
    });

    // Escape lukkar. Ei dokk som ikkje lèt seg lukka med tastaturet er
    // ei dokk som stend i vegen.
    document.addEventListener("keydown", function (e) {
        if (e.key === "Escape" && !dokk.hidden) lukk();
    });

    window.addEventListener("resize", maalDokka);
    maalDokka();
})();
