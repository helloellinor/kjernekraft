package handsamarar

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/russross/blackfriday/v2"
)

// StudioEpost er adressa folk når studioet på. Ho stend her og ikkje
// spreidd i malarne, so ho kann bytast ein stad.
const StudioEpost = "post@kjernekraftoslo.no"

var (
	termsOnce sync.Once
	termsBody template.HTML
	termsErr  error
)

func lesVilkår() (template.HTML, error) {
	content, err := os.ReadFile("static/vilkår.md")
	if err != nil {
		return "", err
	}
	// Kjelda er ei fil me eig sjølve, ikkje noko ein brukar hev skrive,
	// so det er trygt aa lata henne stå som HTML.
	return template.HTML(blackfriday.Run(content)), nil
}

func vilkaarsinnhald() (template.HTML, error) {
	// Utvikling skal framleis syna markdown-endringar ved neste last.
	if IsDevelopment() {
		return lesVilkår()
	}
	termsOnce.Do(func() { termsBody, termsErr = lesVilkår() })
	return termsBody, termsErr
}

// Terms syner handelsvilkaari.
//
// Han skreiv ut naken blackfriday-HTML fyrr — utan <html>, utan
// teiknsett, utan stilark. Det var lesbart berre av di nettlesaren
// gissa. No gjeng han gjenom malen som alt anna; `?fragment=1` gjev
// framleis den nakne bolken, som er det skuffa på registreringssida
// hentar.
func (a *App) Terms(w http.ResponseWriter, r *http.Request) {
	body, err := vilkaarsinnhald()
	if err != nil {
		log.Printf("vilkaar: %v", err)
		http.Error(w, "Kunde ikkje lasta vilkaari", http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("fragment") == "1" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(body))
		return
	}

	lang := GetLanguageFromRequest(r)
	renderPage(w, r, "pages/vilkaar", sidedata(r, SidaVilkaar, t(lang, "terms.title"), map[string]any{
		"Terms": body,
	}))
}

// SignUpPage syner registreringi.
//
// Ho var ei rein fil på disken, servert med http.ServeFile, med Arial
// og sine eigne fargar og ingi umsetjing. No er ho ein mal som alle hine.
func (a *App) SignUpPage(w http.ResponseWriter, r *http.Request) {
	if IsLoggedIn(r) {
		http.Redirect(w, r, "/elev/hjem", http.StatusSeeOther)
		return
	}

	lang := GetLanguageFromRequest(r)
	renderPage(w, r, "pages/registrering", sidedata(r, SidaRegistrering, t(lang, "signup.title"), nil))
}

// GløymtPassord syner vegen attende til kontoen.
//
// Det finst ingi sjølvvend attendestilling enno — ho krev eingongsmerke
// med utlaupstid og ei e-postutsending, og ingen av delarne er sette
// upp. Sida segjer det rett ut og syner vegen som verkar.
func (a *App) GløymtPassord(w http.ResponseWriter, r *http.Request) {
	lang := GetLanguageFromRequest(r)
	renderPage(w, r, "pages/gloymt-passord", sidedata(r, SidaGloymtPassord, t(lang, "forgot.title"), map[string]any{
		"StudioEpost": StudioEpost,
	}))
}
