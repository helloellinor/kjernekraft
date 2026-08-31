// Bolkveljaren.
//
// Dette er det som er att av faner.js. Dei 139 linone der gjorde tri
// ting: dei gøymde bolkar, dei skreiv ein hash i adressa, og dei batt
// seg upp att for kvart sidebyte. Dei tvo siste er tenaren sitt arbeid
// no — `?fane=` i adressa, éin bolk teikna — og fanone i huset er
// vanlege lenkjor. Sjaa handsamarar/faner.go og common/faner.html.
//
// Att stend éin bruk som ikkje kann vera ei lenkja: brytaren i
// timestyringi som segjer um du held paa heile rekkja eller den eine
// timen. Baae bolkarne lyt standa i dokumentet — det du hev skrive i
// den eine skal vera der endaa naar du kjem attende — og ei lenkja
// hadde gjenge til tenaren og kome att med tomme felt.
//
// Alt her heng paa `document`, ikkje paa elementi. Difor treng det ingi
// upp-att-binding naar htmx byter innhaldet, og difor stend det ikkje i
// lista til sideskift.js slik faner.js gjorde.
(function () {
    "use strict";

    // Fanerekkjor kann liggja inni kvarandre, so alt lyt vera avgrensa
    // til *næraste* ark: utan dette fann det ytre arket fanone til det
    // indre og skreiv aria-selected paa deim med.
    function eigne(ark, veljar) {
        return [].slice.call(ark.querySelectorAll(veljar)).filter(function (el) {
            return el.closest(".faneark") === ark;
        });
    }

    function vel(ark, bolk) {
        var rom = eigne(ark, ".fanerom")[0];
        var faner = eigne(ark, ".fane[data-bolk]");

        if (rom) {
            // Bolkarne vert gøymde her og ikkje i stilarket. Fyrr stod
            // namni deira i ein regel i CSS-en — ei ny rekkja kravde ei
            // ny lina der, og ho vart gløymd.
            var bolkar = rom.children;
            for (var b = 0; b < bolkar.length; b++) {
                var namn = bolkar[b].getAttribute("data-bolk");
                if (!namn) continue;
                if (namn === bolk) bolkar[b].removeAttribute("hidden");
                else bolkar[b].setAttribute("hidden", "");
            }
        }
        for (var j = 0; j < faner.length; j++) {
            var vald = faner[j].getAttribute("data-bolk") === bolk;
            faner[j].setAttribute("aria-selected", vald ? "true" : "false");
            // Berre den valde fana er i tabbrekkja; piltastarne fører
            // millom dei hine. Det er slik ei fanerekkja skal te seg.
            faner[j].tabIndex = vald ? 0 : -1;
        }
    }

    document.addEventListener("click", function (e) {
        var f = e.target.closest(".fane[data-bolk]");
        if (!f) return;
        vel(f.closest(".faneark"), f.getAttribute("data-bolk"));
    });

    document.addEventListener("keydown", function (e) {
        var f = e.target.closest(".fane[data-bolk]");
        if (!f) return;
        var ark = f.closest(".faneark");
        var faner = eigne(ark, ".fane[data-bolk]");
        var i = faner.indexOf(f);
        var neste = null;
        if (e.key === "ArrowRight" || e.key === "ArrowDown") neste = faner[(i + 1) % faner.length];
        else if (e.key === "ArrowLeft" || e.key === "ArrowUp") neste = faner[(i - 1 + faner.length) % faner.length];
        else if (e.key === "Home") neste = faner[0];
        else if (e.key === "End") neste = faner[faner.length - 1];
        if (!neste) return;
        e.preventDefault();
        vel(ark, neste.getAttribute("data-bolk"));
        neste.focus();
    });
})();
