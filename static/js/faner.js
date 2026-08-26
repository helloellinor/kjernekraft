// Fanone.
//
// Klikk og piltastar byter bolk. Bolkarne vert ikkje bytte ut — dei
// stend alle i dokumentet, og det er `data-bolk` paa rommet som avgjer
// kva som er synleg. Difor hoppar ikkje sida naar ein byter fana.
//
// Valet stend i adressa. Lastar ein sida paa nytt, kjem ein attende til
// den fana ein stod paa, og ei lenkja til ein bolk peikar dit han er.
(function () {
    "use strict";

    function vel(ark, bolk, skrivAdresse) {
        var rom = ark.querySelector(".fanerom");
        var faner = ark.querySelectorAll(".fane[data-bolk]");
        var finst = false;

        for (var i = 0; i < faner.length; i++) {
            if (faner[i].getAttribute("data-bolk") === bolk) { finst = true; break; }
        }
        if (!finst) return;

        if (rom) rom.setAttribute("data-bolk", bolk);
        for (var j = 0; j < faner.length; j++) {
            var vald = faner[j].getAttribute("data-bolk") === bolk;
            faner[j].setAttribute("aria-selected", vald ? "true" : "false");
            // Berre den valde fana er i tabbrekkja; piltastarne fører
            // millom dei hine. Det er slik ei fanerekkja skal te seg.
            faner[j].tabIndex = vald ? 0 : -1;
        }

        // Eit felt med `autofocus` i ein bolk som var gøymd naar sida
        // lasta fekk aldri fokus — nettlesaren hev alt gjenge forbi det.
        // Difor set me det naar bolken vert synleg.
        if (rom) {
            var f = rom.querySelector('[data-bolk="' + bolk + '"] [autofocus]');
            if (f && document.activeElement !== f) {
                try { f.focus({ preventScroll: true }); } catch (e) { f.focus(); }
            }
        }

        if (skrivAdresse && window.history && window.history.replaceState) {
            window.history.replaceState(null, "", "#" + bolk);
        }
    }

    function start(ark) {
        var faner = [].slice.call(ark.querySelectorAll(".fane[data-bolk]"));
        if (!faner.length) return;

        ark.addEventListener("click", function (e) {
            var f = e.target.closest(".fane[data-bolk]");
            if (f) vel(ark, f.getAttribute("data-bolk"), true);
        });

        ark.addEventListener("keydown", function (e) {
            var f = e.target.closest(".fane[data-bolk]");
            if (!f) return;
            var i = faner.indexOf(f);
            var neste = null;
            if (e.key === "ArrowRight" || e.key === "ArrowDown") neste = faner[(i + 1) % faner.length];
            else if (e.key === "ArrowLeft" || e.key === "ArrowUp") neste = faner[(i - 1 + faner.length) % faner.length];
            else if (e.key === "Home") neste = faner[0];
            else if (e.key === "End") neste = faner[faner.length - 1];
            if (!neste) return;
            e.preventDefault();
            vel(ark, neste.getAttribute("data-bolk"), true);
            neste.focus();
        });

        function fraaAdressa() {
            return window.location.hash.replace(/^#/, "");
        }

        // Eit hopp som berre skiftar hash lastar ikkje dokumentet paa
        // nytt. Utan denne lyttaren peika ei lenkja til ei fana rett
        // nok i adressa, men synte kva som helst.
        window.addEventListener("hashchange", function () {
            vel(ark, fraaAdressa(), false);
        });

        vel(ark, fraaAdressa() || faner[0].getAttribute("data-bolk"), false);
    }

    function startAlle() {
        [].forEach.call(document.querySelectorAll(".faneark"), start);
    }

    if (document.readyState === "loading") {
        document.addEventListener("DOMContentLoaded", startAlle);
    } else {
        startAlle();
    }
})();
