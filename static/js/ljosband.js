// ljosband.js — LED-stripa i lina under hovudet.
//
// Lina under hovudet er tvo strekar med ei renne imellom (.header::after).
// Glødi ligg i renna (.header::before) og skal standa nett under den sida
// du er på. Kvar det er, veit berre nettlesaren: lenkjene er ulike breie og
// rada bryt seg på smale glas. So me måler den valde lenkja og skriv
// midten og breiddi hennar til hovudet som --glodmidt og --glodvidd.
//
// Står dei ikkje, er vidda null og glødi usynleg — betre enn eit ljos på
// feil stad medan sida lastar. Same grepet som i ordboka.
(function () {
  "use strict";

  function still() {
    var hovud = document.querySelector(".header");
    if (!hovud) return;
    // Heime står ingi lenkje i leidingi — det er merkeordet sjølv som er
    // sida du er på (det ber aria-current="page"). Stripa skal syne kvar du
    // er, so ho lyser under merket då.
    var valt = hovud.querySelector('.nav-item.active a, .nav-item a[aria-current="page"]')
            || hovud.querySelector('.merkeord[aria-current="page"]');
    if (!valt) {
      hovud.style.removeProperty("--glodmidt");
      hovud.style.removeProperty("--glodvidd");
      return;
    }
    var hr = hovud.getBoundingClientRect(), lr = valt.getBoundingClientRect();
    if (!lr.width) return;
    hovud.style.setProperty("--glodmidt",
      (lr.left - hr.left + lr.width / 2).toFixed(1) + "px");
    // Ljoset rekk ei grand ut yver bokstavarne, som i ordboka.
    hovud.style.setProperty("--glodvidd",
      Math.max(lr.width * 0.62, 48).toFixed(1) + "px");
  }

  document.addEventListener("DOMContentLoaded", still);
  // Skrifti kjem etter fyrste målingi og gjer lenkjene breiare.
  if (document.fonts && document.fonts.ready) document.fonts.ready.then(still);
  window.addEventListener("resize", still);
  window.addEventListener("pageshow", still);
  document.addEventListener("htmx:afterSwap", still);
  still();
})();
