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

    // Yverskrifta er ein contenteditable <h1> og ikkje eit <input>, av
    // di ho skal brjota som all annan tekst paa sida. Ho speglar seg inn
    // i eit løynt felt, so skjemaet sender det same som fyrr.
    var speglar = [].slice.call(skjema.querySelectorAll("[data-speil]"));
    speglar.forEach(function (el) {
        var maal = document.getElementById(el.dataset.speil);
        if (!maal) return;
        felt.push(maal);

        function spegla() {
            // Eitt namn er éi linja. Lima ein inn fleire, vert dei til
            // mellomrom — teksti gjeng ikkje tapt, ho vert berre ei linja.
            var t = el.textContent.replace(/\s+/g, " ").trim();
            maal.value = t;
            maal.dispatchEvent(new Event("input", { bubbles: true }));
        }

        el.addEventListener("input", spegla);
        el.addEventListener("blur", function () {
            var t = el.textContent.replace(/\s+/g, " ").trim();
            if (el.textContent !== t) el.textContent = t;
            spegla();
        });
        // Linjeskift i eit namn er ikkje ei linja til, det er slutten.
        el.addEventListener("keydown", function (e) {
            if (e.key === "Enter") { e.preventDefault(); el.blur(); }
        });
        // Der plaintext-only ikkje finst, lyt me sjølve halda det reint.
        el.addEventListener("paste", function (e) {
            if (el.getAttribute("contenteditable") === "plaintext-only") return;
            e.preventDefault();
            var t = (e.clipboardData || window.clipboardData).getData("text");
            document.execCommand("insertText", false, t.replace(/\s+/g, " "));
        });
    });

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
            // Eit løynt felt kann ikkje syna at det er endra; merket
            // gjeng paa det som er synleg i staden.
            var synleg = f.dataset.visast ? document.getElementById(f.dataset.visast) : f;
            if (synleg) synleg.classList.toggle("endra", e);
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
            if (f.dataset.visast) {
                var synleg = document.getElementById(f.dataset.visast);
                if (synleg) synleg.textContent = f.dataset.lagra;
            }
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
