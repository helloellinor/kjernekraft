// Rutenettet — eit utviklarverkty, ikkje ein del av sida.
//
// Natti til 30.8.2026 vart merket og radteksten skalerte i ni
// umgangar, kvar bit prøvd i sitt eige utsnitt, og heilskapen fall
// sund utan at nokon saag det skje. Tilhøve prøver ein paa heile
// sida — difor dette: eit taktnett lagt YVER den verkelege sida, so
// ein ser kva som ligg paa lina og kva som ikkje gjer det.
//
// Ctrl+G slær det av og paa (berre lokalt), og ?rutenett i adressa
// opnar med det paa — det er slik dei hovudlause skoti fær det fram.
// I drift gjer fila ingen ting utan at nokon bed um det i adressa.
(function () {
    var paa = new URLSearchParams(location.search).has("rutenett");
    var lokalt = /^(localhost|127\.)/.test(location.hostname);
    if (!paa && !lokalt) return;
    function slaa() { document.documentElement.classList.toggle("rutenett"); }
    if (paa) slaa();
    if (lokalt) {
        document.addEventListener("keydown", function (e) {
            if (e.ctrlKey && e.key === "g") { e.preventDefault(); slaa(); }
        });
    }
})();
