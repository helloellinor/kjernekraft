// Ein 5xx fraa tenaren ber no ei melding i kroppen (handlers/berging.go),
// men htmx byter ikkje inn feilsvar av seg sjølv. Utan dette stod
// lasteplassen og snurra i all æva og innsendingsknappen gjorde berre
// ingen ting — feilen fanst einast i nettverksfana. Meldingi skal
// standa der svaret elles skulde stade.
(function () {
    document.body.addEventListener("htmx:beforeSwap", function (e) {
        if (e.detail.xhr.status >= 500) {
            e.detail.shouldSwap = true;
        }
    });
})();
