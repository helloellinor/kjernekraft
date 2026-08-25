// CSRF-kjennemerket ligg i ei kaka sida sjølv kann lesa. Alt som
// endrar noko lyt bera det attende i ei hovudlina; tenaren
// samanliknar med det som stend i økti.
//
// Det ligg her og ikkje i kvart einskilt kall, av di eit kall som
// gløymer det fær 403 — og daa gløymer ein det ein gong og finn det
// att seks maanader seinare.
(function () {
    var mutating = { POST: 1, PUT: 1, PATCH: 1, DELETE: 1 };

    function token() {
        var m = document.cookie.match(/(?:^|;\s*)kjernekraft-csrf=([^;]*)/);
        return m ? decodeURIComponent(m[1]) : "";
    }

    // htmx
    document.body.addEventListener("htmx:configRequest", function (e) {
        if (mutating[(e.detail.verb || "").toUpperCase()]) {
            e.detail.headers["X-CSRF-Token"] = token();
        }
    });

    // fetch — berre mot vaar eigen tenar. Eit kjennemerke sendt til
    // ein framand vert eit kjennemerke den framande hev.
    var original = window.fetch;
    window.fetch = function (input, init) {
        init = init || {};
        var method = (init.method || (input && input.method) || "GET").toUpperCase();
        var url = typeof input === "string" ? input : (input && input.url) || "";
        var sameOrigin = url === "" || url.charAt(0) === "/" ||
            url.indexOf(window.location.origin) === 0;

        if (mutating[method] && sameOrigin) {
            var headers = new Headers(init.headers || (input && input.headers) || {});
            headers.set("X-CSRF-Token", token());
            init.headers = headers;
        }
        return original.call(this, input, init);
    };
})();

