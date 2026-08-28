// Stadfestingi.
//
// Det som ikkje kann gjerast um att krev tvo trykk, ikkje eitt. Ikkje av
// di brukaren treng ei aatvaring — ho stend alt paa sida, attmed
// knappen, der ARKET §7 vil ha henne — men av di ein finger som glid er
// ein finger som glid, og «Sei upp» er eit ord ein ikkje fær attende.
//
// Ingen dialog. §7 forbyd ei rute som spring upp *etter* trykket og
// fortel deg noko du skulde visst fyrr. Han forbyd ikkje at knappen bed
// um trykk nummer tvo. Det er tvo ulike ting, og det er det fyrste som
// er gale.
//
// Knappen skiftar ord der han stend, og han skiftar ikkje storleik:
// plassen til det lengste av dei tvo ordi vert teken av med ein gong,
// fyrr nokon hev trykt. Ei maaling *ved* trykket hadde ikkje halde —
// `min-width` er eit golv, og «Ja, sei opp» er lengre enn «Sei opp», so
// knappen hadde vakse likevel og skuva alt attmed seg (§3).
(function () {
    "use strict";

    var opa = null;

    // Breiddi til det breidaste ordet, maald ein gong. Alt her hender i
    // same synkrone bolken, so nettlesaren teiknar aldri mellomsteget.
    function tak(knapp) {
        if (knapp.dataset.stadfestbreidd) return;

        var fyrr = knapp.textContent;
        var smal = knapp.offsetWidth;
        knapp.textContent = knapp.dataset.stadfest;
        var brei = knapp.offsetWidth;
        knapp.textContent = fyrr;

        knapp.style.minWidth = Math.max(smal, brei) + "px";
        knapp.dataset.stadfestbreidd = "1";
    }

    function takAlle(rot) {
        var alle = (rot || document).querySelectorAll("[data-stadfest]");
        for (var i = 0; i < alle.length; i++) tak(alle[i]);
    }

    function slepp(knapp) {
        if (!knapp) return;
        if (knapp.dataset.stadfestFyrr !== undefined) {
            knapp.textContent = knapp.dataset.stadfestFyrr;
            delete knapp.dataset.stadfestFyrr;
        }
        // Breiddi stend. Ho er teki av for det lengste ordet, og skal
        // ikkje gjevast fraa seg att.
        knapp.removeAttribute("data-stadfestar");
        if (opa === knapp) opa = null;
    }

    // Fangstfasen. Handlingane bur i sine eigne skript og lyder paa
    // dokumentet i boblefasen; stoggar me hendingi her, naar ho deim
    // aldri, og dei treng ikkje vita at det finst ei stadfesting.
    document.addEventListener("click", function (e) {
        var knapp = e.target.closest("[data-stadfest]");

        if (!knapp) { slepp(opa); return; }

        // Trykk nummer tvo. Han slepp gjenom til den som skal ha honom.
        if (knapp.hasAttribute("data-stadfestar")) {
            slepp(knapp);
            return;
        }

        // Trykk nummer ein. Han kjem ikkje lenger enn hit.
        e.preventDefault();
        e.stopPropagation();

        slepp(opa);
        tak(knapp);
        knapp.dataset.stadfestFyrr = knapp.textContent;
        knapp.textContent = knapp.dataset.stadfest;
        knapp.setAttribute("data-stadfestar", "");
        opa = knapp;
    }, true);

    // Escape slepper honom, som alt anna i huset som kann opnast.
    document.addEventListener("keydown", function (e) {
        if (e.key === "Escape") slepp(opa);
    });

    // Korti paa betalingssida vert teikna paa nytt etter kvar handling,
    // og timeplanen kjem inn med htmx. Nye knappar lyt taka plassen sin
    // dei og.
    document.addEventListener("DOMContentLoaded", function () { takAlle(); });
    document.body.addEventListener("htmx:afterSwap", function (e) { takAlle(e.target); });
    takAlle();
})();
