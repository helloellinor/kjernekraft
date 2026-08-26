// Ein ny prisrad.
//
// Sida er ei side der ein endrar, og då skal ein kunne leggje til her
// òg — ikkje på eit eige skjema ein annan stad som gjer det same ein
// gong til. Raden kjem tom nedst i lista, og han vert lagra av den same
// dokka som resten: du fyller han ut, dokka seier «1 endring», du
// trykkjer Lagre.
(function () {
    "use strict";

    var knapp = document.getElementById("ny-prislapp");
    var lista = document.getElementById("prisliste");
    var mal = document.getElementById("prislapp-mal");
    var skjema = document.getElementById("prisar");
    if (!knapp || !lista || !mal || !skjema) return;

    var teljar = 0;

    knapp.addEventListener("click", function () {
        teljar++;
        var rad = mal.content.cloneNode(true);
        // Feltnamna ber id-en. Ein ny rad har ingen, so han får eit namn
        // tenaren kjenner att som ny: «ny1», «ny2» …
        [].forEach.call(rad.querySelectorAll("[name]"), function (f) {
            f.name = f.name.replace("NY", "ny" + teljar);
        });
        lista.appendChild(rad);

        var lagd = lista.lastElementChild;
        // Dokka skal telje felta i den nye raden med. Ho tok berre dei
        // som stod der då sida vart lasta.
        skjema.dispatchEvent(new CustomEvent("endringar:nye", { bubbles: false }));

        var fyrste = lagd.querySelector("input");
        if (fyrste) fyrste.focus();
    });
})();
