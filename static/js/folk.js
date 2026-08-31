// The people list.
//
// The search filters what is already in the page. The studio has a few
// hundred members; asking the server per keystroke would be slower than
// reading what you already have, and it would flicker.
(function () {
    "use strict";

    function start() {
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

    // The role is read from the marks in the row, not from a data- attribute
    // on it.
    //
    // The marks *are* the role: they are what you press to grant one, and
    // aria-pressed is the only place the answer lives in the page. A data-rolla
    // written by the server would be the same answer a second time, and the
    // second answer would be wrong the moment somebody pressed the mark (§7:
    // one question, one answer).
    function har(rad, loyve) {
        var m = rad.querySelector('.loyvemerke[data-loyve="' + loyve + '"]');
        return !!m && m.getAttribute("aria-pressed") === "true";
    }

    // A student is someone with no permission.
    //
    // Not "everyone", even though a teacher trains here too: then "Students"
    // would be the same press as "All", and a filter that is the whole list is
    // no filter. Someone who is both teacher and manager stands in both — a
    // permission is something you *have*, and you can have two.
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
        // A filter that gives nothing has to say so. Only when there *are* people
        // in the list — if it comes empty from the server it has already said so
        // in its own words, and two sentences would stand there.
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

    // Escape empties the field rather than giving it up. You are mid-search;
    // you want to start over, not leave.
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

    // The selects on the schedule tab share the "laerarar" datalist. The
    // server writes it, but a role switched on here should be selectable there
    // at once — otherwise the button looks like it did nothing until the page
    // reloads.
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

    // The roles. The mark changes immediately and springs back if the server
    // says no — the surface must not show a role the database does not have.
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

    // Fokus i søkjefeltet, utan aa rulla dit.
    //
    // `autofocus` i malen gjorde det same, men han rullar: nettlesaren
    // dreg feltet inn i glaset, og paa ei laag skjerm hamna heile sida
    // nedrulla med fanerekkja utanfor. `preventScroll` er heile
    // skilnaden — det same grepet faner.js gjorde daa bolkarne var
    // gøymde i staden for teikna av tenaren.
    //
    // Berre naar ingen ting anna hev fokus: kjem ein hit ved aa trykkja
    // paa ei fana, skal fokus standa der trykket var.
    if (document.activeElement === document.body) {
        try { felt.focus({ preventScroll: true }); } catch (e) { /* eldre nettlesarar rullar; lat dei */ }
    }
    }

    start();
    (window.__sideskift = window.__sideskift || []).push(start);
})();
