// Sideskift: det som lyt hende når <main> vert bytt og resten stend.
//
// `hx-boost` byter berre innhaldet. Det er heile vinsten — stilarket,
// skriftene, merke_defs og skripti stend — men tvo ting fylgjer ikkje
// med av seg sjølv, og dei er det denne fila gjer.
(function () {
    "use strict";

    // ---- 1. Leidingi veit ikkje at du hev flytt deg ----
    //
    // `aria-current="page"` vart sett av malen då sida vart teikna. Med
    // eit boost-byte kjem det aldri ei ny leiding, so merket stod att på
    // den sida du kom frå. Adressa er fasiten, so me les henne.
    function merkSida() {
        var no = location.pathname.replace(/\/$/, "");
        var lenkjer = document.querySelectorAll(".header a[href^='/']");
        for (var i = 0; i < lenkjer.length; i++) {
            var l = lenkjer[i];
            var sti = l.getAttribute("href").replace(/\/$/, "");
            if (sti === no) {
                l.setAttribute("aria-current", "page");
            } else {
                l.removeAttribute("aria-current");
            }
        }

        // Namneknappen er ingi lenkja og hev ingen `href` aa samanlikna
        // med. Han fylgjer profil-lenkja i menyen sin, som nettupp fekk
        // svaret sitt yver.
        var namn = document.querySelector(".user-name");
        if (namn) {
            if (document.querySelector('.namnemeny a[aria-current="page"]')) {
                namn.setAttribute("aria-current", "page");
            } else {
                namn.removeAttribute("aria-current");
            }
        }
    }

    // ---- 2. Skript som batt seg til element inne i <main> ----
    //
    // Element utanfor <main> stend, so lyttarane deira stend med. Men det
    // som batt seg til noko *inni* mista taket då innhaldet vart bytt.
    //
    // Skripti sjølve vert ikkje køyrde um att — dei ligg utanfor <main>,
    // og det er med vilje. Difor lyt dei bindast på nytt her, og kvart
    // skript som treng det legg seg i lista under.
    function bindPåNytt() {
        // veketal.js, timeplan-veke.js og folk.js legg kvar sin
        // startfunksjon her når dei vert lasta. Finst han ikkje, er
        // skriptet ikkje med på denne sida, og det er greitt.
        //
        // faner.js stod her ogso. Han gjer det ikkje lenger: fanone er
        // lenkjor og tenaren teiknar bolken, og det som er att av
        // honom — bolkveljar.js — heng på `document` og treng inkje.
        var att = window.__sideskift || [];
        for (var i = 0; i < att.length; i++) {
            try {
                att[i]();
            } catch (e) {
                // Eit skript som fell skal ikkje taka dei hine med seg.
                console.error("sideskift: ", e);
            }
        }
    }

    document.body.addEventListener("htmx:afterSwap", function (e) {
        // Berre heile sidebyte. Eit stykke som byter seg sjølv — lista
        // yver påmeldingar, helsingi — skal ikkje setja leidingi på nytt.
        if (e.detail && e.detail.boosted === false) return;
        merkSida();
        bindPåNytt();
    });

    // ---- 3. Bufferet skal ikkje halda paa ei innlogga sida ----
    //
    // htmx tek vare paa heile <body> for kvar sida han hev vore paa, i
    // sessionStorage. Etter ei utlogging ligg dei der endaa, og eit
    // «attende» teiknar den førre brukaren si sida upp att — namn,
    // medlemskap, folkelista — utan aa spyrja tenaren um lov.
    //
    // Er hovudet burte, er ingen innlogga (base.html teiknar det berre
    // for ein brukar), og daa skal bufferet vera tomt.
    if (!document.querySelector(".header")) {
        try { sessionStorage.removeItem("htmx-history-cache"); } catch (e) { /* privat modus */ }
    }

    // htmx melder frå når historikken vert nytta (fram/attende).
    document.body.addEventListener("htmx:historyRestore", function () {
        merkSida();
        bindPåNytt();
    });
})();
