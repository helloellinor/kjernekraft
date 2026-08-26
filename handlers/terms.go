package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/russross/blackfriday/v2"
)

// StudioEpost er adressa folk naar studioet paa. Ho stend her og ikkje
// spreidd i malarne, so ho kann bytast ein stad.
const StudioEpost = "post@kjernekraftoslo.no"

// TermsHandler syner handelsvilkaari.
//
// Han skreiv ut naken blackfriday-HTML fyrr — utan <html>, utan
// teiknsett, utan stilark. Det var lesbart berre av di nettlesaren
// gissa. No gjeng han gjenom malen som alt anna; `?fragment=1` gjev
// framleis den nakne bolken, som er det skuffa paa registreringssida
// hentar.
func TermsHandler(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("static/vilkår.md")
	if err != nil {
		log.Printf("vilkaar: %v", err)
		http.Error(w, "Kunde ikkje lasta vilkaari", http.StatusInternalServerError)
		return
	}

	// Kjelda er ei fil me eig sjølve, ikkje noko ein brukar hev skrive,
	// so det er trygt aa lata henne staa som HTML.
	body := template.HTML(blackfriday.Run(content))

	if r.URL.Query().Get("fragment") == "1" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(body))
		return
	}

	lang := GetLanguageFromRequest(r)
	renderPage(w, r, "pages/vilkaar", map[string]interface{}{
		"Title":       t(lang, "terms.title"),
		"CurrentPage": "vilkaar",
		"Lang":        lang,
		"CSRFToken":   CSRFToken(r),
		"IsAdmin":     sessionIsAdmin(r),
		"UserName":    sessionUserName(r),
		"Terms":       body,
	})
}

// SignUpPageHandler syner registreringi.
//
// Ho var ei rein fil paa disken, servert med http.ServeFile, med Arial
// og sine eigne fargar og ingi umsetjing. No er ho ein mal som alle hine.
func SignUpPageHandler(w http.ResponseWriter, r *http.Request) {
	if IsLoggedIn(r) {
		http.Redirect(w, r, "/elev/hjem", http.StatusSeeOther)
		return
	}

	lang := GetLanguageFromRequest(r)
	renderPage(w, r, "pages/registrering", map[string]interface{}{
		"Title":       t(lang, "signup.title"),
		"CurrentPage": "registrering",
		"Lang":        lang,
		"CSRFToken":   CSRFToken(r),
		"IsAdmin":     sessionIsAdmin(r),
		"UserName":    sessionUserName(r),
	})
}

// GloymtPassordHandler syner vegen attende til kontoen.
//
// Det finst ingi sjølvvend attendestilling enno — ho krev eingongsmerke
// med utlaupstid og ei e-postutsending, og ingen av delarne er sette
// upp. Sida segjer det rett ut og syner vegen som verkar.
func GloymtPassordHandler(w http.ResponseWriter, r *http.Request) {
	lang := GetLanguageFromRequest(r)
	renderPage(w, r, "pages/gloymt-passord", map[string]interface{}{
		"Title":       t(lang, "forgot.title"),
		"CurrentPage": "gloymt-passord",
		"Lang":        lang,
		"CSRFToken":   CSRFToken(r),
		"IsAdmin":     sessionIsAdmin(r),
		"UserName":    sessionUserName(r),
		"StudioEpost": StudioEpost,
	})
}
