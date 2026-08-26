// Folkelista.
//
// Søket filtrerer det som alt ligg i sida. Studioet hev nokre hundrad
// medlemer; aa spørja tenaren per tastetrykk vilde vera tregare enn aa
// lesa det ein alt hev, og det vilde blinka.
(function () {
    "use strict";

    var rot = document.getElementById("folk");
    if (!rot) return;

    var felt = document.getElementById("folk-sok");
    var tal = document.getElementById("folk-tal");
    var radar = [].slice.call(rot.querySelectorAll(".folk-rad"));

    function tel(n) {
        tal.textContent = n === radar.length
            ? radar.length + " " + (tal.dataset.alle || "")
            : n + " " + (tal.dataset.av || "") + " " + radar.length;
    }

    function sok() {
        var q = felt.value.trim().toLowerCase();
        var n = 0;
        radar.forEach(function (r) {
            var traff = !q || r.dataset.sok.toLowerCase().indexOf(q) !== -1;
            r.hidden = !traff;
            if (traff) n++;
        });
        tel(n);
    }

    felt.addEventListener("input", sok);

    // Escape tømmer feltet i staden for aa gjeva det fraa seg. Ein er
    // midt i eit søk; ein vil byrja paa nytt, ikkje ut.
    felt.addEventListener("keydown", function (e) {
        if (e.key === "Escape" && felt.value) { e.preventDefault(); felt.value = ""; sok(); }
    });

    // Rada opnar seg der ho stend.
    rot.addEventListener("click", function (e) {
        var h = e.target.closest(".folk-hovud");
        if (!h) return;
        var meir = h.nextElementSibling;
        var open = h.getAttribute("aria-expanded") === "true";
        h.setAttribute("aria-expanded", open ? "false" : "true");
        meir.hidden = open;
    });

    tel(radar.length);
})();
