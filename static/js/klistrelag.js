// klistrelag.js — kor høgt dei klistra linone stend.
//
// Toppen av sida er fleire linor som kvar held seg sjølv oppe: hovudlina
// med leidingi, og på timeplanen tittellina med vikeveljaren under henne.
// Ei klistra lina veit ikkje om ei onnor klistra lina — kvar av deim
// stend på det `top` ho fær — so den andre lyt vita kor høg den fyrste
// er, elles legg dei seg oppå kvarandre.
//
// Talet kann ikkje standa i CSS-en. Hovudet er 56 px på eit breidt glas
// og 96 på eit smalt, av di leidingi bryt til ei lina til under merket;
// tittellina er 80 px når vikeveljaren får plass attmed tittelen og 56
// når han ikkje gjer det. Skreiv me eit tal, ville det vera rett på éi
// breidd og gøyma ei lina bak ei onnor på alle dei hine.
//
// So me måler, og skriv høgdi til rota som --sidehovud-hogd og
// --tittellinje-hogd. Same grepet som ljosband.js: nettlesaren veit det,
// og CSS-en kann ikkje spyrja.
(function () {
  "use strict";

  var lag = [
    { vel: ".header", nykel: "--sidehovud-hogd" },
    { vel: ".timehovudlinje", nykel: "--tittellinje-hogd" },
  ];

  function maal() {
    var rot = document.documentElement;
    lag.forEach(function (l) {
      var el = document.querySelector(l.vel);
      // Fell laget burt — sida hev ingi tittellina, eller ein er ikkje
      // innlogga so hovudet ikkje stend — skal talet vera null og ikkje
      // det siste me målte. Ei lina som ikkje finst tek ingi høgd.
      if (!el) {
        rot.style.setProperty(l.nykel, "0px");
        return;
      }
      rot.style.setProperty(l.nykel, el.getBoundingClientRect().height + "px");
    });
  }

  // Breidd, zoom, ei lenkje som bryt: alt som endrar høgdi på ei lina
  // skal skriva talet på nytt. `ResizeObserver` ser det `resize` ikkje
  // ser — hovudet kann verta høgare utan at glaset endrar seg.
  var sjaaar = window.ResizeObserver ? new ResizeObserver(maal) : null;

  // Eit sidebyte lyt maalast um att, og linone lyt sjaaast um att.
  //
  // `.timehovudlinje` bur inni <main>, og `hx-boost` byter <main>. Den
  // gamle lina var burte og den nye vart aldri sedd, so `maal()` gjekk
  // aldri um att: kom du til timeplanen fraa ei onnor sida, stod
  // `--tittellinje-hogd` att paa 0px — verdet fraa sida utan tittellina —
  // og daghovudet klistra seg 134 piksel for høgt, rett attum
  // vikeveljaren. Eit friskt sidekall paa den same adressa gjorde det
  // ikkje. Difor «det kjem an paa kvar eg kom fraa».
  function start() {
    maal();
    if (!sjaaar) return;
    sjaaar.disconnect();
    lag.forEach(function (l) {
      var el = document.querySelector(l.vel);
      if (el) sjaaar.observe(el);
    });
  }

  start();
  document.addEventListener("DOMContentLoaded", start);
  // Skrifti kjem etter fyrste målingi og endrar høgdi.
  if (document.fonts && document.fonts.ready) document.fonts.ready.then(maal);
  (window.__sideskift = window.__sideskift || []).push(start);

  if (!window.ResizeObserver) {
    // Reserveløysingi for nettlesarar utan ResizeObserver. Han er
    // strupa som dei hine — ResizeObserver samlar sjølv, denne ikkje.
    var ventar = false;
    window.addEventListener("resize", function () {
      if (ventar) return;
      ventar = true;
      requestAnimationFrame(function () { ventar = false; maal(); });
    });
  }
})();
