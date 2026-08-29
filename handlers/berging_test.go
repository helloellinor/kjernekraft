package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Ein panikk midt i skrivingi skal verta ein heil 500, ikkje ei halv
// sida med 200 paa. Det var nettupp dette chi sin Recoverer ikkje
// greidde: handsamarane strøymer malarne, so fyrste byten var alt ute.
func TestRecovererPanikkMidtISkrivingi(t *testing.T) {
	h := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html>halvskrive"))
		panic("malen fall")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, venta 500", rec.Code)
	}
	kropp := rec.Body.String()
	if strings.Contains(kropp, "halvskrive") {
		t.Errorf("det halvskrivne kom ut: %q", kropp)
	}
	if !strings.Contains(kropp, "<!doctype html") {
		t.Errorf("ei full sida venta, fekk: %q", kropp)
	}
}

// htmx fær eit stykke og ikkje ei heil sida, med same klasse som
// skjemafeilarne elles ber, so meldingi kann standa der svaret skulde.
func TestRecovererHtmxFaerEitStykke(t *testing.T) {
	h := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("noko fall")
	}))

	req := httptest.NewRequest("POST", "/signup", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, venta 500", rec.Code)
	}
	kropp := rec.Body.String()
	if !strings.Contains(kropp, `class="error"`) || !strings.Contains(kropp, `role="alert"`) {
		t.Errorf("stykket ber ikkje feilklassa: %q", kropp)
	}
	if strings.Contains(kropp, "<!doctype") {
		t.Errorf("htmx fekk ei heil sida: %q", kropp)
	}
}

// Gjeng alt vel, skal bufringi vera usynleg: status, hovudlinor og
// kropp som handsamaren skreiv deim.
func TestRecovererSlepperGjenomDetHeile(t *testing.T) {
	h := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("alt vel"))
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, venta 418", rec.Code)
	}
	if rec.Body.String() != "alt vel" {
		t.Errorf("kropp = %q", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain" {
		t.Errorf("Content-Type = %q", ct)
	}
}

// http.ErrAbortHandler er net/http sin eigen stoggmaate og skal upp,
// ikkje verta 500.
func TestRecovererSlepperAbortVidare(t *testing.T) {
	h := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	defer func() {
		if p := recover(); p != http.ErrAbortHandler {
			t.Errorf("panikken vart %v, venta ErrAbortHandler", p)
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
}
