// namnemeny.js — namnet i hovudet, og dei tvo utgangane attum det.
//
// Same grepet som vikeveljaren: ordet er knappen. Profilen og
// utloggingi tek ingi plass fyre ein spør etter deim.
//
// Han bur i si eigi fil, og det er ikkje ei smakssak. Han stod i
// `faner.js` fyrr, av di han er ein veljar og fanone er veljarar — men
// `faner.js` vart berre send til dei sidone som *hadde* ei fanerekkje
// (klippekort, medlemskap, betaling, admin, arket). Hovudet stend paa
// kvar side. Paa heimesida og timeplanen fanst difor knappen utan at
// noko lydde paa honom: du trykte, og ingen ting hende — ikkje seint,
// men aldri.
//
// Regelen som fylgjer av det: eit skript som høyrer til *hovudet* eller
// *botnlina* lyt sendast like breidt som dei, og kann ikkje liggja i ein
// bunt som er knytt til innhaldet paa sida. Sjaa base.html, der dei
// staar samla med ljosband.js og klistrelag.js.
(function () {
    "use strict";
    var knapp = document.getElementById("namn-knapp");
    var meny = document.getElementById("namn-meny");
    if (!knapp || !meny) return;

    function set(open) {
        meny.hidden = !open;
        knapp.setAttribute("aria-expanded", open ? "true" : "false");
    }

    knapp.addEventListener("click", function () { set(meny.hidden); });
    document.addEventListener("click", function (e) {
        if (!meny.hidden && !meny.contains(e.target) && e.target !== knapp) set(false);
    });
    document.addEventListener("keydown", function (e) {
        if (e.key === "Escape" && !meny.hidden) { set(false); knapp.focus(); }
    });
})();
