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

// Slett ein rad.
//
// Han vert ikkje borte med det same. Han vert *merkt*, og det er dokka
// som utfører det når du lagrar — same veg som alle andre endringar på
// sida. Angre-knappen tek han attende, og du ser kva du har gjort før
// det er gjort.
(function () {
    "use strict";

    var skjema = document.getElementById("prisar");
    if (!skjema) return;

    skjema.addEventListener("click", function (e) {
        var knapp = e.target.closest("[data-slett]");
        if (!knapp) return;

        var rad = knapp.closest(".prislapp");
        var id = rad.getAttribute("data-medlemskap");
        if (!id) {                       // ein ny rad er berre borte
            rad.remove();
            skjema.dispatchEvent(new CustomEvent("endringar:nye"));
            return;
        }

        var merke = rad.querySelector('input[name="slett-' + id + '"]');
        if (merke) {
            merke.remove();
            rad.classList.remove("slettast");
        } else {
            merke = document.createElement("input");
            merke.type = "hidden";
            merke.name = "slett-" + id;
            merke.value = "1";
            // Løynde felt vert ikkje talde. Dette skal teljast, so det er
            // eit felt som *er* der og som er ulikt det lagra.
            merke.dataset.tel = "1";
            rad.appendChild(merke);
            rad.classList.add("slettast");
        }
        skjema.dispatchEvent(new CustomEvent("endringar:nye"));
    });
})();
