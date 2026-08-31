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
    // Berre `aria-current`. Selektoren spurde etter `.nav-item.active`
    // fyrst, og `querySelector` gjev fyrste treffet i *dokumentrekkje-
    // fylgd* og ikkje det fyrste i lista — so ei gamal `active` som stod
    // lenger framme i leidingi vann yver den rette `aria-current`. Stripa
    // vart difor liggjande under den sida du kom fraa, og berre naar du
    // gjekk attende i rekkja hamna ho rett.
    var valt = hovud.querySelector('.nav-item a[aria-current="page"]')
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
  // Resize fyrer titals gonger i sekundet medan ein dreg i glaskanten,
  // og `still` *les* layout (getBoundingClientRect) og *skriv* han
  // (style.setProperty) i same slengen. Ei lesing etter ei skriving
  // tvingar nettlesaren til å rekna ut layouten på nytt der og då — ein
  // gong per hending. Med rAF gjeng det ein gong per bilete i staden.
  var ventar = false;
  window.addEventListener("resize", function () {
    if (ventar) return;
    ventar = true;
    requestAnimationFrame(function () { ventar = false; still(); });
  });
  window.addEventListener("pageshow", still);

  // Etter eit sidebyte lyt målingi venta eit bilete.
  //
  // `still` bergar seg med `if (!lr.width) return` når lenkja ikkje hev
  // noko breidd å måla enno — og då stend bandet att der det låg, under
  // den sida du kom frå. Ved eit boost-byte er det nett det som hender:
  // `htmx:afterSwap` kjem fyre nettlesaren hev lagt ut det nye
  // innhaldet, og `aria-current` er sett same augneblinken (sideskift.js).
  //
  // Difor: mål i neste bilete, når layouten er sett. `afterSettle` kjem
  // etter `afterSwap` og er den htmx sjølv peikar på for måling; me tek
  // baae, og rAF-en gjer at det ikkje kostar tvo målingar.
  function stillSnart() {
    if (stillVentar) return;
    stillVentar = true;
    requestAnimationFrame(function () { stillVentar = false; still(); });
  }
  var stillVentar = false;
  document.addEventListener("htmx:afterSwap", stillSnart);
  document.addEventListener("htmx:afterSettle", stillSnart);
  still();
})();
