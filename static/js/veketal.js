// Veketal — reglane for eit vekefelt, skrivne ein gong.
//
// Tri stader i huset let ein skriva eit vekenummer: vikeveljaren i
// timeplanen, «fraa veke» i skjemaet for ein ny time, og feltet som
// plukkar ut éin time or ei timerekkje. Alle tri fylgjer dei same
// reglane, og reglane er ikkje sjølvsagde — dei er lærde:
//
//   * Berre siffer. Alt anna fell burt medan du skriv.
//   * Tvo siffer *er* ferdig skrive. Daa treng ein ikkje trykkja noko.
//     Eitt siffer er det ikkje: «3» er ikkje eit ynske um veke tri, det
//     er halve vegen til «37».
//   * Ei vike som er gjengi finst ikkje. Bed du um veke 2 medan du
//     stend i veke 51, er det den *komande* veke 2 du meiner — studioet
//     syner ikkje vikor som er ute, so det finst ikkje noka onnor.
//   * Eit nei er ein farge, ikkje ei melding. Feltet fer ein augneblink
//     i aatvaringsfargen; ingi rute som lyt lesast og lukkast.
//
// Fyrr laag alt dette i `timeplan-veke.js` og galdt berre der. Det er
// her no, og den fila les det som dei hine.
window.Veketal = (function () {
    "use strict";

    // Er talet ei vike i det heile? Aaret hev 52 vikor, stundom 53.
    function gyldig(v, vekerIAar) {
        return !isNaN(v) && v >= 1 && v <= vekerIAar;
    }

    // Kor mange vikor fram fraa `naa` ligg vike `v`?
    //
    // Skilnaden i vekenummer, heilt til aaret skiftar. Peikar
    // reknestykket bakum golvet, legg me eit aar til: det er regelen um
    // at ei gjengi vike ikkje finst, skrivi som aritmetikk.
    //
    // `golv` er kor langt attende ein *kann* koma. I timeplanen er det
    // minus den vika du alt hev bladd fram til; i eit skjema er det
    // null, av di ein ny time ikkje kann byrja i gaar.
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
