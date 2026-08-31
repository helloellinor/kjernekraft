// Week numbers — the rules for a week field, written once.
//
// Three places let you type a week number: the week picker in the
// schedule, "from week" in the new-class form, and the field that picks
// one class out of a run. All three follow the same rules, and the rules
// are learnt rather than obvious:
//
//   * Digits only. Anything else falls away as you type.
//   * Two digits *is* finished. You do not have to press anything. One
//     digit is not: "3" is not a wish for week three, it is halfway to
//     "37".
//   * A week that has passed does not exist. Ask for week 2 while standing
//     in week 51 and you mean the *coming* week 2 — the studio does not
//     show weeks that are over, so there is no other one.
//   * A refusal is a colour, not a message. The field flashes the warning
//     colour; no dialog to read and close.
//
// This all lived in timeplan-veke.js and applied only there.
window.Veketal = (function () {
    "use strict";

    // Is the number a week at all? A year has 52 weeks, sometimes 53.
    function gyldig(v, vekerIAar) {
        return !isNaN(v) && v >= 1 && v <= vekerIAar;
    }

    // How many weeks ahead of `naa` is week `v`?
    //
    // The difference in week number, until the year turns. If the arithmetic
    // points below the floor we add a year: that is the rule about a past week
    // not existing, written as arithmetic.
    //
    // `golv` is how far back you *can* go. In the schedule it is minus the
    // week you have already paged to; in a form it is zero, because a new
    // class cannot start yesterday.
    function framover(v, naa, vekerIAar, golv) {
        golv = golv || 0;
        var skilnad = v - naa;
        if (golv + skilnad < 0) {
            skilnad += vekerIAar;
        }
        return skilnad;
    }

    // tolk les eit *sett* av vikor: «37», eit spenn «37-40», fleire
    // skilde med komma «37,39,41», og blandingar av dei.
    //
    // Eit spenn som gjeng den andre vegen gjeng yver nyttaar. «51-3» er
    // 51, 52, 1, 2, 3 og ikkje ein skrivefeil — det er den same regelen
    // som `framover`: ei vike som er gjengi finst ikkje, so det einaste
    // veke 3 som finst er den komande.
    //
    // Ho gjev null naar ho ikkje finn eit einaste tal. Tomt sett og
    // «ikkje eit tal» er tvo ulike svar: det fyrste tyder «ingen vikor
    // valde», det andre «eg skjøna deg ikkje».
    function tolk(tekst, vekerIAar) {
        var ut = [], sett = {}, fann = false;

        function legg(v) {
            if (!gyldig(v, vekerIAar) || sett[v]) return;
            sett[v] = true; ut.push(v);
        }

        String(tekst).split(",").forEach(function (bit) {
            bit = bit.trim();
            if (!bit) return;
            // Baade bindestrek og tankestrek: det eine er tasten, det
            // andre er det teksthandsamaren gjer av honom.
            var spenn = bit.split(/\s*[-–—]\s*/);
            if (spenn.length === 2) {
                var a = parseInt(spenn[0], 10), b = parseInt(spenn[1], 10);
                if (isNaN(a) || isNaN(b)) return;
                fann = true;
                var n = b >= a ? b - a : (b + vekerIAar) - a;
                // Eit spenn kann ikkje vera lengre enn aaret.
                if (n >= vekerIAar) n = vekerIAar - 1;
                for (var i = 0; i <= n; i++) {
                    var v = a + i;
                    while (v > vekerIAar) v -= vekerIAar;
                    legg(v);
                }
                return;
            }
            var einskild = parseInt(bit, 10);
            if (isNaN(einskild)) return;
            fann = true;
            legg(einskild);
        });

        if (!fann) return null;
        return ut;
    }

    // Neiet. Feltet vert tomt og fer ein augneblink i aatvaringsfargen.
    function nekt(felt) {
        felt.value = "";
        felt.classList.add("nekta");
        setTimeout(function () { felt.classList.remove("nekta"); }, 900);
    }

    // Tastaturet og siffervaska. `ferdig` vert kalla naar tale er
    // skrive ut: paa blur, paa Enter, og naar det ikkje kann verta
    // lengre. `tomt` naar feltet vert tømt, `escape` paa Escape.
    function bind(felt, val) {
        // Markøren, ikkje eit merke. Feltet er tomt naar det opnar, so
        // dette set berre markøren der han skal vera og hindrar at
        // nettlesaren merkjer noko av seg sjølv.
        felt.addEventListener("focus", function () {
            felt.setSelectionRange(felt.value.length, felt.value.length);
        });
        felt.addEventListener("blur", function () { val.ferdig(felt); });
        felt.addEventListener("keydown", function (e) {
            if (e.key === "Enter") { e.preventDefault(); felt.blur(); }
            if (e.key === "Escape" && val.escape) val.escape(felt);
        });
        felt.addEventListener("input", function () {
            felt.value = felt.value.replace(/[^0-9]/g, "");
            if (felt.value === "") { if (val.tomt) val.tomt(felt); return; }
            // Tvo siffer er ferdig skrive — men berre um dei er ei vike.
            // «99» ventar paa blur og fær neiet der, so talet ikkje
            // sprett burt medan fingeren enno er paa veg.
            if (felt.value.length === 2 &&
                (!val.gyldig || val.gyldig(parseInt(felt.value, 10)))) {
                val.ferdig(felt);
            }
        });
    }

    return { gyldig: gyldig, framover: framover, tolk: tolk, nekt: nekt, bind: bind };
})();
