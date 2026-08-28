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
    var inkje = document.getElementById("folk-inkje");
    var radar = [].slice.call(rot.querySelectorAll(".folk-rad"));
    var fanor = [].slice.call(rot.querySelectorAll(".rollefaner .fane[data-rolla]"));

    // Bunken ein stend i. Tom streng er «alle», og det er utgangsstoda.
    var bunke = "";

    function tel(n) {
        tal.textContent = n === radar.length
            ? radar.length + " " + (tal.dataset.alle || "")
            : n + " " + (tal.dataset.av || "") + " " + radar.length;
    }

    // Rolla vert lesi av merki i rada, ikkje av eit `data-` paa henne.
    //
    // Merki *er* rolla: dei er det Ida trykkjer paa naar ho gjev nokon
    // ei, og `aria-pressed` er den einaste staden svaret bur i sida. Eit
    // `data-rolla` skrive av tenaren hadde vore det same svaret ein gong
    // til, og det andre svaret hadde vorte gale i det sekundet nokon
    // trykte paa merket (§7: eitt spursmaal, eitt svar).
    function har(rad, loyve) {
        var m = rad.querySelector('.loyvemerke[data-loyve="' + loyve + '"]');
        return !!m && m.getAttribute("aria-pressed") === "true";
    }

    // Ein elev er den som ikkje hev noko løyve.
    //
    // Ikkje «alle», endaa ein lærar ogso trenar her: daa hadde «Elevar»
    // vore det same knappetrykket som «Alle», og ein bunke som er heile
    // lista er ingen bunke. Ein som er baade lærar og sjef stend i baae
    // dei tvo — eit løyve er noko ein *hev*, og ein kann ha tvo.
    function iBunken(rad) {
        if (!bunke) return true;
        if (bunke === "laerar") return har(rad, "teacher");
        if (bunke === "sjef") return har(rad, "admin");
        return !har(rad, "teacher") && !har(rad, "admin");
    }

    function sok() {
        var q = felt.value.trim().toLowerCase();
        var n = 0;
        radar.forEach(function (r) {
            var traff = iBunken(r)
                && (!q || r.dataset.sok.toLowerCase().indexOf(q) !== -1);
            r.hidden = !traff;
            if (traff) n++;
        });
        tel(n);
        // Eit filter som gjev ingen ting lyt segja det. Berre naar det
        // *er* folk i lista — stend ho tom fraa tenaren, hev ho alt sagt
        // det med sine eigne ord, og tvo setningar hadde stade der daa.
        if (inkje) inkje.hidden = !(radar.length && n === 0);
    }

    felt.addEventListener("input", sok);

    // Bunken. Ingi adressa og ingen bolk som vert bytt: rada stend der
    // ho stend, og det er `hidden` som avgjer kva ein ser (§3).
    if (fanor.length) {
        rot.addEventListener("click", function (e) {
            var k = e.target.closest(".rollefaner .fane[data-rolla]");
            if (!k) return;
            bunke = k.dataset.rolla || "";
            fanor.forEach(function (f) {
                f.setAttribute("aria-selected",
                    (f.dataset.rolla || "") === bunke ? "true" : "false");
            });
            sok();
        });
    }

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
            + "&loyve=" + encodeURIComponent(b.dataset.loyve)
            + "&paa=" + (paa ? "1" : "0"), { method: "POST" })
            .then(function (svar) {
                if (!svar.ok) throw new Error(svar.status);
                return svar.json();
            })
            // Lista vert *ikkje* sikta paa nytt her. Rada Ida nett gav ei
            // rolla høyrer straks til ein annan bunke, men ho skal ikkje
            // kverva under handi hennar — same avgjerdi som i
            // rabattkravlista, der svaret stend att i rada (§3). Talet
            // held seg sant med det: det tel det som stend paa skjermen.
            // Neste tastetrykk eller neste bunke sorterer henne dit ho
            // høyrer heime.
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
