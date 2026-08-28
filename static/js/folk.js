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

    // Veljarane paa timeplanfana deler datalista «laerarar». Tenaren
    // skriv henne, men ei rolla som vert slegi paa her skal vera valbar
    // der med ein gong — elles ser knappen ut som um han ikkje gjorde
    // noko fyrr sida vert lasta paa nytt.
    function settLaerarar(namn) {
        if (!namn) return;
        [].forEach.call(document.querySelectorAll("#laerarar"), function (dl) {
            dl.innerHTML = "";
            namn.forEach(function (n) {
                var o = document.createElement("option");
                o.value = n;
                dl.appendChild(o);
            });
        });
    }

    // Rollone. Merket skiftar med ein gong, og spring attende um tenaren
    // segjer nei — flata skal ikkje syna ei rolla basen ikkje hev.
    rot.addEventListener("click", function (e) {
        var b = e.target.closest(".loyvemerke");
        if (!b) return;

        var paa = b.getAttribute("aria-pressed") !== "true";
        b.setAttribute("aria-pressed", paa ? "true" : "false");
        b.disabled = true;

        fetch("/api/admin/loyve?brukar=" + encodeURIComponent(b.dataset.brukar)
            + "&loyve=" + encodeURIComponent(b.dataset.loyvemerke)
            + "&paa=" + (paa ? "1" : "0"), { method: "POST" })
            .then(function (svar) {
                if (!svar.ok) throw new Error(svar.status);
                return svar.json();
            })
            .then(function (d) { settLaerarar(d.laerarar); })
            .catch(function () {
                b.setAttribute("aria-pressed", paa ? "false" : "true");
                b.classList.add("nekta");
                setTimeout(function () { b.classList.remove("nekta"); }, 1600);
            })
            .then(function () { b.disabled = false; });
    });

    tel(radar.length);
})();
