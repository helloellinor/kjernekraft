// Ein dag vald.
//
// Dagnamna over rutenettet dempa dei andre fyrr og filtrerte aldri. Å
// filtrere ville teke frå rutenettet det einaste han kan som ei liste
// ikkje kan: syne heile vika på ein gong. Det er slik ein ser at Vinyasa
// går tysdag òg, og at torsdagen er full medan tysdagen ikkje er det.
//
// Grunnen held framleis — men han held mot eit filter som er
// *utgangsstoda*. «Alle dagar» står fyrst og er vald når sida kjem, so
// heile vika er det du får utan å be om noko. Vel du ein dag, er det du
// som tok bort dei andre, og eitt trykk til gjev deg vika att.
//
// Valet står i adressa, som på fanone elles i huset: `#dag-3`. Ei lenkje
// til ein dag peikar dit han er, og ei omlasting kjem attende dit du var.
(function () {
    "use strict";

    // Netet vert slege upp kvar gong, ikkje ein gong.
    //
    // Her stod `var plan = document.querySelector(".timeplan")` heilt
    // ytst, med eit `if (!plan) return` under. Det gjorde tvo ting gale,
    // og baae saag ut som «dagbytet er ustøtt»:
    //
    //   1. Lasta du ei sida utan timeplan — heimesida — gav uppslaget
    //      null, skriptet gav seg med ein gong, og det kom aldri
    //      attende. Gjekk du so til timeplanen gjenom leidingi (som
    //      byter <main> og ikkje dokumentet), var dagbytet daudt heile
    //      økti. Lasta du /elev/timeplan beint, verka det.
    //
    //   2. Naar vikebladingi vart eit byte i staden for ei sidelasting,
    //      vart `.timeplan` bytt ut under føtene paa variabelen. Han
    //      peika daa paa eit element som ikkje stod i dokumentet lenger:
    //      `setAttribute` verka, men paa noko ingen kunde sjaa.
    //
    // Difor: ingen fanga element. Kvart kall spør dokumentet paa nytt,
    // og `start()` ligg i lista til sideskift.js som dei hine.
    function planen() { return document.querySelector(".timeplan"); }

    // Skyvaren i vikehovudet. Han er den same `.faner` som fanone paa
    // dei hine sidone; berre kva han styrer er annleis — her er «bolken»
    // det same nettet, sikta til ein dag.
    function knappar() {
        return [].slice.call(document.querySelectorAll(".dagfaner .fane[data-dag]"));
    }

    function merke(plan) {
        return [].slice.call(plan.querySelectorAll(".dagmerke"));
    }

    function dagenTil(el) {
        // Dagen står som `--dag` i stilen — den same verdien som seier kva
        // spalte ruta høyrer til.
        return el.style.getPropertyValue("--dag").trim();
    }

    function vel(dag, skrivAdresse) {
        var plan = planen();
        if (!plan) return;
        var ingen = !dag;

        if (ingen) {
            plan.removeAttribute("data-dag");
        } else {
            plan.setAttribute("data-dag", dag);
        }

        knappar().forEach(function (k) {
            var eigen = k.dataset.dag || "";
            k.setAttribute("aria-selected", eigen === (dag || "") ? "true" : "false");
        });

        merke(plan).forEach(function (r) {
            r.classList.toggle("utanfor-dagen", !ingen && dagenTil(r) !== dag);
        });

        // Ei rad utan noko den dagen er ikkje eit svar på spørsmålet, og
        // ho skal ikkje stå att som ei tom line. Den daude ruta tel ikkje
        // — ho *er* «her går det ingen ting».
        [].forEach.call(plan.querySelectorAll(".timerad"), function (rad) {
            if (ingen) {
                rad.classList.remove("rad-utanfor");
                return;
            }
            var noko = [].slice.call(rad.querySelectorAll(".dagmerke")).some(function (r) {
                return dagenTil(r) === dag && !r.classList.contains("daud");
            });
            rad.classList.toggle("rad-utanfor", !noko);
        });

        // Eit filter som gjev ingen ting lyt segja det. Elles stend
        // berre eit tomt nett att, og det ser ut som om sida rauk.
        var tom = plan.querySelector(".dagtom");
        if (tom) {
            var att = [].slice.call(plan.querySelectorAll(".timerad")).some(function (rad) {
                return !rad.classList.contains("rad-utanfor");
            });
            tom.hidden = ingen || att;
        }

        if (skrivAdresse && window.history && window.history.replaceState) {
            window.history.replaceState(null, "", ingen ? window.location.pathname + window.location.search : "#dag-" + dag);
        }
    }

    function fraaAdressa() {
        var m = /^#dag-([1-7])$/.exec(window.location.hash);
        return m ? m[1] : "";
    }

    // Lyttarane heng paa `document` og `window`, so dei vert bundne éin
    // gong og stend gjenom kvart byte. Det er berre *stoda* som lyt
    // setjast paa nytt naar netet er eit anna — og det er det `start`
    // gjer.
    document.addEventListener("click", function (e) {
        var k = e.target.closest(".dagfaner .fane[data-dag]");
        if (!k) return;
        var plan = planen();
        if (!plan) return;
        var eigen = k.dataset.dag || "";
        // Eit trykk på dagen du står på slepper han att.
        vel(eigen && eigen === plan.getAttribute("data-dag") ? "" : eigen, true);
    });

    window.addEventListener("hashchange", function () { vel(fraaAdressa(), false); });

    function start() {
        // Ingi dagrekkja: sida hev ingen timeplan, og det er greitt.
        if (document.querySelector(".dagfaner")) vel(fraaAdressa(), false);
    }

    start();
    // Timeplanen kjem inn med eit sidebyte, og netet vert bytt ut naar
    // ein bladar ei vike. Sjaa sideskift.js.
    (window.__sideskift = window.__sideskift || []).push(start);
})();
