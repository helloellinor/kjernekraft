// Ein dag løfta fram.
//
// Dagnamna over rutenettet er knappar, men dei filtrerer ikkje. Å
// filtrere ville teke frå rutenettet det einaste han kan som ei liste
// ikkje kan: syne heile vika på ein gong. Det er slik ein ser at Vinyasa
// går tysdag òg, og at torsdagen er full medan tysdagen ikkje er det.
//
// Eit trykk dempar dei andre dagane i staden. Ingenting går bort, og eit
// trykk til slepper dagen att.
(function () {
    "use strict";

    var hovud = document.querySelector(".timehovud");
    var plan = document.querySelector(".timeplan");
    if (!hovud || !plan) return;

    var vald = null;

    function rutor() {
        return [].slice.call(plan.querySelectorAll(".dagmerke"));
    }

    function vel(dag) {
        vald = dag;
        [].forEach.call(hovud.querySelectorAll(".dagnamn[data-dag]"), function (k) {
            k.setAttribute("aria-pressed", k.dataset.dag === dag ? "true" : "false");
        });
        rutor().forEach(function (r) {
            // Dagen står som `--dag` i stilen på ruta — det er den same
            // verdien som seier kva spalte ho høyrer til.
            var d = r.style.getPropertyValue("--dag").trim();
            r.classList.toggle("utanfor-fokus", dag !== null && d !== dag);
        });
        plan.classList.toggle("dagfokus", dag !== null);
    }

    hovud.addEventListener("click", function (e) {
        var knapp = e.target.closest(".dagnamn[data-dag]");
        if (!knapp) return;
        vel(knapp.dataset.dag === vald ? null : knapp.dataset.dag);
    });
})();
