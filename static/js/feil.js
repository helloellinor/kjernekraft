// Ein feil fraa tenaren ber ei melding i kroppen (handlers/berging.go
// for panikk, svarFeil i handlers/feilsvar.go elles), men htmx byter
// ikkje inn feilsvar av seg sjølv. Utan dette stod lasteplassen og
// snurra i all æva og innsendingsknappen gjorde berre ingen ting —
// feilen fanst einast i nettverksfana. Meldingi skal standa der svaret
// elles skulde stade.
//
// >= 400, ikkje berre >= 500: eit avslag — 403 utan kjennemerke, 400
// for eit skjema som ikkje gjekk gjenom — er like usynlegt som eit
// krasj um ingen syner det fram.
(function () {
    document.body.addEventListener("htmx:beforeSwap", function (e) {
        if (e.detail.xhr.status >= 400) {
            e.detail.shouldSwap = true;
        }
    });
})();
