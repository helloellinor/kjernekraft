// Profilen.
//
// Kvart felt merkjer seg sjølv naar det er ulikt det som er lagra, og
// dokka nedst ber handlingi. Ingen tekst segjer «du hev ulagra
// endringar» — feltet *ser* endra ut, og det er nok.
(function () {
    "use strict";

    var skjema = document.getElementById("profil");
    var dokk = document.getElementById("lagre-dokk");
    if (!skjema || !dokk) return;

    var tal = document.getElementById("endringstal");
    var angra = document.getElementById("angra");
    var felt = [].slice.call(skjema.querySelectorAll("input:not([type=hidden])"));

    // Det som stod der daa sida vart lasta. Det er *dette* eit felt vert
    // samanlikna med — ikkje med seg sjølv for eit augneblink sidan.
    felt.forEach(function (f) {
        f.dataset.lagra = f.type === "checkbox" ? String(f.checked) : f.value;
    });

    function endra(f) {
        return (f.type === "checkbox" ? String(f.checked) : f.value) !== f.dataset.lagra;
    }

    function maal() {
        var n = 0;
        felt.forEach(function (f) {
            var e = endra(f);
            f.classList.toggle("endra", e);
            if (e) n++;
        });

        dokk.hidden = n === 0;
        if (n) {
            tal.textContent = n === 1
                ? tal.dataset.ein
                : (tal.dataset.fleire || "%d").replace("%d", n);
        }
        document.documentElement.style.setProperty(
            "--dokkhogd", dokk.hidden ? "0px" : dokk.offsetHeight + "px");
    }

    skjema.addEventListener("input", maal);
    skjema.addEventListener("change", maal);

    angra.addEventListener("click", function () {
        felt.forEach(function (f) {
            if (f.type === "checkbox") f.checked = f.dataset.lagra === "true";
            else f.value = f.dataset.lagra;
        });
        maal();
    });

    // Ei aatvaring fraa nettlesaren om ein gjeng fraa sida med ulagra
    // endringar. Feltet og dokka syner det medan du er her; dette er for
    // den som gjeng sin veg utan aa sjaa etter.
    window.addEventListener("beforeunload", function (e) {
        if (!dokk.hidden) { e.preventDefault(); e.returnValue = ""; }
    });
    skjema.addEventListener("submit", function () {
        dokk.hidden = true;   // slepp aatvaringi naar ein *lagrar*
    });

    maal();
})();
