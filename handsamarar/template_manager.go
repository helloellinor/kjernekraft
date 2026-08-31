package handsamarar

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"kjernekraft/handsamarar/config"
	"kjernekraft/models"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TemplateManager handles centralized template loading and parsing
type TemplateManager struct {
	mu        sync.RWMutex
	templates map[string]*template.Template
	basePath  string

	// Fingerprint of the template directory as it looked when last read.
	//
	// In development *every* template was reparsed on every request — and
	// since each page carries the layout, all components and all modules, that
	// is hundreds of file opens per view: 53 ms per request against 0.4 ms in
	// production.
	//
	// Now the files are stat-ed instead. If nothing changed, the templates we
	// have are right. A change to an .html file still shows on the next reload
	// — that was the whole point of rereading — but the cost drops from
	// parsing to a handful of stat calls.
	fingeravtrykk string
}

var templateManager *TemplateManager
var tmplOnce sync.Once

// GetTemplateManager returns the singleton instance of TemplateManager
func GetTemplateManager() *TemplateManager {
	tmplOnce.Do(func() {
		wd, _ := os.Getwd()
		templateManager = &TemplateManager{
			templates: make(map[string]*template.Template),
			basePath:  filepath.Join(wd, "handsamarar", "templates"),
		}
		templateManager.loadTemplates()
	})
	return templateManager
}

// Template function map with commonly used functions
func getTemplateFuncs() template.FuncMap {
	settings := config.GetInstance()
	return template.FuncMap{
		"asset": staticURL,
		// Klassetypen som ein CSS-krok. Verdet kjem raatt frå
		// `events.class_type`, der administrasjonen skriv fritt, so det
		// vert vaska: smaa bokstavar og ingen ting utanum a-z. Er det
		// tomt eller ukjent, fell `.slag-*` burt og vengen tek
		// standardfargen sin.
		//
		// The list lives in slag.go; the activity board uses the same one.
		"slagklasse": SlagKlasse,
		// Ruta utan time. Ho er den same figuren for alle, so han vert
		// rekna ein gong og ikkje per rute.
		"markDead":     DeadSilhouette,
		"markIndeks":   SkiveIndeks,
		"markSilhuett": Silhuett,
		"markDagfane":  Dagfane,
		"markGlas": func() map[string]float64 {
			x, y, r := GlasMidt()
			return map[string]float64{"X": x, "Y": y, "R": r}
		},
		"markSkivetal": SkiveTal,
		"markViewBox": func() string {
			return fmt.Sprintf("%.2f 0 %.2f %.2f", MarkLeft, MarkWidth, MarkHeight)
		},
		"markWidth":  func() float64 { return MarkWidth },
		"markHeight": func() float64 { return MarkHeight },
		"sub": func(a, b int) int {
			return a - b
		},
		// Pengar er lagra i øre og skal lesast i kronor. Tabellen synte
		// «104000 øre» — på den skjermen Ida redigerer prisar på.
		// Tusenskiljet er eit hardt mellomrom, so talet ikkje bryt
		// midt i lina.
		//
		// Han tek imot båe heiltal og flyttal, av di prisen er `int`
		// på Membership og `float64` på FreezeRequest. Det er ein
		// lyte i modellane — pengar bør vera eitt slag heile vegen —
		// men malen skal ikkje falla på det.
		"kroner": func(v interface{}) string {
			var oere int64
			switch n := v.(type) {
			case int:
				oere = int64(n)
			case int32:
				oere = int64(n)
			case int64:
				oere = n
			case float32:
				oere = int64(n)
			case float64:
				oere = int64(n)
			default:
				return ""
			}
			return skrivKronor(oere / 100)
		},
		// Kronor som eit reint tal, til eit talfelt.
		"kronerTal": func(v interface{}) int64 {
			switch n := v.(type) {
			case int:
				return int64(n) / 100
			case int64:
				return n / 100
			case float64:
				return int64(n) / 100
			}
			return 0
		},
		"divf": func(a, b interface{}) float64 {
			var aFloat, bFloat float64
			switch v := a.(type) {
			case int:
				aFloat = float64(v)
			case float64:
				aFloat = v
			default:
				return 0
			}
			switch v := b.(type) {
			case int:
				bFloat = float64(v)
			case float64:
				bFloat = v
			default:
				return 0
			}
			if bFloat == 0 {
				return 0
			}
			return aFloat / bFloat
		},
		// Go's Format gives English day and month names and there is no way to
		// ask it for anything else, so the names live here. "Thursday 27. August"
		// stood in the dock on a Norwegian page.
		//
		// The day as three letters for the grid header, and the date as a string
		// so the template can compare with "today".
		"dagKort3": func(namn string) string {
			r := []rune(namn)
			if len(r) > 3 {
				r = r[:3]
			}
			return string(r)
		},
		"dagStreng": func(t time.Time) string { return t.Format("2006-01-02") },
		// "three weeks ago" instead of a date. A date has to be worked out;
		// "three weeks ago" answers immediately what you are actually asking — is
		// this someone who is here, or someone
		"sidan": func(lang string, t0 *time.Time) string {
			if t0 == nil {
				return ""
			}
			d := time.Since(*t0)
			switch dagar := int(d.Hours() / 24); {
			case dagar <= 0:
				return t(lang, "admin.ago_today")
			case dagar == 1:
				return t(lang, "admin.ago_yesterday")
			case dagar < 14:
				return fmt.Sprintf(t(lang, "admin.ago_days"), dagar)
			case dagar < 60:
				return fmt.Sprintf(t(lang, "admin.ago_weeks"), dagar/7)
			default:
				return fmt.Sprintf(t(lang, "admin.ago_months"), dagar/30)
			}
		},
		// printf translation: the string itself says where the number goes, so
		// word order can differ between languages.
		"tf": func(lang, key string, a ...interface{}) string {
			return fmt.Sprintf(t(lang, key), a...)
		},
		"norskDato": func(t time.Time) string {
			return fmt.Sprintf("%s %d. %s", norskeDagar[t.Weekday()], t.Day(), norskeMaanader[t.Month()])
		},
		// With the year. "fredag 26. mars" is enough for a class this week, but a
		// renewal date without a year is not a date — it could be next month or in
		// two years.
		//
		// The expiry on a membership. The rules live in the model because they are
		// rules and not drawing — four branches and a clock that stops while the
		// membership is frozen — and because there they can be tested. nil means
		// "no expiry", and the card draws ∞.
		"gjeldTil": func(m models.MembershipWithDetails) *time.Time {
			return m.GjeldTil(time.Now())
		},
		"norskDatoÅr": func(t time.Time) string {
			return fmt.Sprintf("%d. %s %d", t.Day(), norskeMaanader[t.Month()], t.Year())
		},
		"formatDateTime": func(t time.Time) string {
			return t.In(settings.GetLocation()).Format("2006-01-02 15:04")
		},
		"currentTime": func() time.Time {
			return settings.GetCurrentTime()
		},
		"currentTimezone": func() string {
			return settings.GetTimezone()
		},
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"t": func(lang, key string) string {
			loc := GetLocalization()
			return loc.T(lang, key)
		},
		// One or several. "1 gonger" stood in the rule row because the number and
		// the word were written separately and the word never saw the number. The
		// house already had the key pair; there was nowhere to make the choice
		// once.
		//
		// The weekday a date falls on, as a translation key. The rule has
		// VekedagNykel itself, but a single day can be moved away from the rule's
		// day, and then it is the date that knows.
		"vekedagnykel": func(t time.Time) string {
			return [...]string{
				"timeplan.sunday", "timeplan.monday", "timeplan.tuesday",
				"timeplan.wednesday", "timeplan.thursday", "timeplan.friday",
				"timeplan.saturday",
			}[t.Weekday()]
		},
		// Maanaden ein dato fell i, som umsetjingsnykel. Stutt form: han
		// stend under vekenummeret på dagrada som ei merkelapp, og ei
		// merkelapp skal ikkje taka meir rom enn ho er verd (§5).
		"maanadnykel": func(t time.Time) string {
			return [...]string{
				"", "timeplan.jan", "timeplan.feb", "timeplan.mar", "timeplan.apr",
				"timeplan.mai", "timeplan.jun", "timeplan.jul", "timeplan.aug",
				"timeplan.sep", "timeplan.okt", "timeplan.nov", "timeplan.des",
			}[int(t.Month())]
		},
		// The week number a date falls in, ISO-8601 — the same number the week
		// picker counts in, so "week 37" means the same in both places. A series
		// repeats weekly, so the number is the one thing separating one class in
		// the run from the others.
		"veketal": func(t time.Time) int {
			_, v := t.ISOWeek()
			return v
		},
		// `eq` in a template requires the same type. The group's id is int64 from
		// the database and the class's is int, and comparing the two is a runtime
		// error in the middle of rendering.
		"int64": func(n interface{}) int64 {
			switch v := n.(type) {
			case int:
				return int64(v)
			case int64:
				return v
			}
			return 0
		},
		// Namni i ei setning: «Leon», «Leon og Vikar», «Leon, Anna og
		// Vikar». Rekkja er lærar fyrst, so vikarane — sjå
		// `laerarane` i timeserie.go — so setningi les seg som «Leon,
		// og so ein vikar ei vika».
		//
		// Ordet imillom kjem or umsetjingane som alt anna (§7). Han
		// stend i malen og ikkje i ei Go-lista av di det er *skrifti* og
		// ikkje rekkja som skil maali.
		"namneliste": func(lang string, namn []string) string {
			switch len(namn) {
			case 0:
				return ""
			case 1:
				return namn[0]
			}
			siste := namn[len(namn)-1]
			return strings.Join(namn[:len(namn)-1], ", ") +
				" " + t(lang, "admin.stats_and") + " " + siste
		},
		"gonger": func(lang string, n int) string {
			if n == 1 {
				return t(lang, "admin.times_one")
			}
			return t(lang, "admin.times")
		},
		"toJS": func(s string) template.JS {
			// Escape string for JavaScript use
			escaped := strings.ReplaceAll(s, "\\", "\\\\")
			escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
			escaped = strings.ReplaceAll(escaped, "'", "\\'")
			escaped = strings.ReplaceAll(escaped, "\n", "\\n")
			escaped = strings.ReplaceAll(escaped, "\r", "\\r")
			escaped = strings.ReplaceAll(escaped, "\t", "\\t")
			return template.JS("\"" + escaped + "\"")
		},
		"seq": func(n int) []int {
			var result []int
			for i := 0; i < n; i++ {
				result = append(result, i)
			}
			return result
		},
		// Ei setning med eit val i seg står som «Medlemer %s oppgradere
		// …» i umsetjingsfila — heile setninga på éin nykel, so han som
		// omset ser samanhengen. `deling` deler henne ved %s, so malen
		// kan setje veljaren midt inne i henne.
		"deling": func(s string) []string {
			i := strings.Index(s, "%s")
			if i < 0 {
				return []string{s, ""}
			}
			return []string{s[:i], s[i+2:]}
		},
		"add": func(a, b int) int { return a + b },
		// Ti hòl, alltid. Talet bur i `models.HolPerKort` av di rekninga
		// og malen lyt telja likt: stod det ein tiar i malen, kunde
		// konstanten skifta utan at rekkja fylgde med.
		"holPerKort": func() int { return models.HolPerKort },
		// Kor mange rader i vika som er heilt gjengne. Malen treng talet
		// til knappen som hentar dei fram att.
		"pastRowCount": PastRowCount,
		// `list` saman med `dict` er det som skal til for å gje ein
		// komponent ei liste med ting frå ein mal — fanerekkja tek ei.
		"list": func(values ...interface{}) []interface{} { return values },
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				return nil // Must have even number of arguments (key-value pairs)
			}
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					continue // Skip non-string keys
				}
				dict[key] = values[i+1]
			}
			return dict
		},
	}
}

// malfingeravtrykk gjeng gjenom malmappa og lagar ein streng av
// filnamn, storleik og endringstid. Skil han seg frå den fyrre, er
// noko endra og malarne lyt lesast på nytt.
func (tm *TemplateManager) malfingeravtrykk() string {
	var b strings.Builder
	filepath.WalkDir(tm.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		fmt.Fprintf(&b, "%s|%d|%d\n", path, info.Size(), info.ModTime().UnixNano())
		return nil
	})
	return b.String()
}

// lesPåNyttOmEndra rereads the templates only when a file actually changed.
func (tm *TemplateManager) lesPåNyttOmEndra() {
	nytt := tm.malfingeravtrykk()

	tm.mu.RLock()
	uendra := nytt == tm.fingeravtrykk
	tm.mu.RUnlock()
	if uendra {
		return
	}

	tm.loadTemplates()

	tm.mu.Lock()
	tm.fingeravtrykk = nytt
	tm.mu.Unlock()
}

// loadTemplates loads all templates from the templates directory
func (tm *TemplateManager) loadTemplates() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Walk through the templates directory
	filepath.WalkDir(tm.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		// Get relative path for template name
		relPath, _ := filepath.Rel(tm.basePath, path)
		name := strings.TrimSuffix(relPath, ".html")
		name = strings.ReplaceAll(name, string(os.PathSeparator), "/")

		// Load template with base layouts if it's a page template
		if strings.HasPrefix(relPath, "pages/") {
			tm.loadPageTemplate(name, path)
		} else {
			tm.loadComponentTemplate(name, path)
		}

		return nil
	})
}

// loadPageTemplate loads a page template with base layout
func (tm *TemplateManager) loadPageTemplate(name, path string) {
	t := template.New(name).Funcs(getTemplateFuncs())

	// Try to load base layout first
	baseLayoutPath := filepath.Join(tm.basePath, "layouts", "base.html")
	if _, err := os.Stat(baseLayoutPath); err == nil {
		var parseErr error
		t, parseErr = t.ParseFiles(baseLayoutPath)
		if parseErr != nil {
			log.Printf("Error parsing base layout %s: %v", baseLayoutPath, parseErr)
			return
		}
	}

	// Load all components
	componentsPath := filepath.Join(tm.basePath, "components")
	if _, err := os.Stat(componentsPath); err == nil {
		filepath.WalkDir(componentsPath, func(compPath string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(compPath, ".html") {
				parsed, parseErr := t.ParseFiles(compPath)
				if parseErr != nil {
					log.Printf("Error parsing component %s: %v", compPath, parseErr)
					return nil
				}
				t = parsed
			}
			return nil
		})
	}

	// Load all modules
	modulesPath := filepath.Join(tm.basePath, "modules")
	if _, err := os.Stat(modulesPath); err == nil {
		filepath.WalkDir(modulesPath, func(modPath string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(modPath, ".html") {
				parsed, parseErr := t.ParseFiles(modPath)
				if parseErr != nil {
					log.Printf("Error parsing module %s: %v", modPath, parseErr)
					return nil
				}
				t = parsed
			}
			return nil
		})
	}

	// Finally parse the page template
	finalTemplate, err := t.ParseFiles(path)
	if err != nil {
		log.Printf("Error parsing page template %s: %v", path, err)
		return
	}
	tm.templates[name] = finalTemplate
}

// loadComponentTemplate loads a standalone component or module template.
//
// It gets the shared pieces under components/ with it. Without them a
// module rendered on its own — an htmx fragment, say — could not use a
// common component: its page worked, but the fragment answered "no such
// template". It is the same template and should behave the same wherever
// it is drawn.
//
// The error was also swallowed here: the template simply fell out of the
// set, and the first sign of it was a missing section.
func (tm *TemplateManager) loadComponentTemplate(name, path string) {
	t := template.New(name).Funcs(getTemplateFuncs())

	komponentar := filepath.Join(tm.basePath, "components")
	if _, err := os.Stat(komponentar); err == nil {
		filepath.WalkDir(komponentar, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".html") || p == path {
				return nil
			}
			parsed, parseErr := t.ParseFiles(p)
			if parseErr != nil {
				log.Printf("Error parsing shared component %s: %v", p, parseErr)
				return nil
			}
			t = parsed
			return nil
		})
	}

	parsed, err := t.ParseFiles(path)
	if err != nil {
		log.Printf("Error parsing component template %s: %v", path, err)
		return
	}
	tm.templates[name] = parsed
}

// GetTemplate returns a template by name.
//
// In development templates are reread per request, so a change to an .html
// file shows on the next reload — without a restart, and without wondering
// whether you are seeing the new one or the old.
func (tm *TemplateManager) GetTemplate(name string) (*template.Template, bool) {
	if IsDevelopment() {
		tm.lesPåNyttOmEndra()
	}
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	tmpl, exists := tm.templates[name]
	return tmpl, exists
}

// ReloadTemplates reloads all templates (useful for development)
func (tm *TemplateManager) ReloadTemplates() {
	tm.loadTemplates()
}

// GetAvailableTemplates returns a list of available template names (for debugging)
func (tm *TemplateManager) GetAvailableTemplates() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var names []string
	for name := range tm.templates {
		names = append(names, name)
	}
	return names
}

// ParseTemplate creates a template with components loaded
func (tm *TemplateManager) ParseTemplate(content string, name string) (*template.Template, error) {
	t := template.New(name).Funcs(getTemplateFuncs())

	// Load all components first
	componentsPath := filepath.Join(tm.basePath, "components")
	if _, err := os.Stat(componentsPath); err == nil {
		filepath.WalkDir(componentsPath, func(compPath string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(compPath, ".html") {
				t, _ = t.ParseFiles(compPath)
			}
			return nil
		})
	}

	// Parse the main template content
	return t.Parse(content)
}

// ExecuteTemplate executes a template by name with the given data
func (tm *TemplateManager) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	tmpl, exists := tm.GetTemplate(name)
	if !exists {
		return errors.New("template not found: " + name)
	}
	return tmpl.ExecuteTemplate(w, name, data)
}

// skrivKronor set hardt mellomrom som tusenskilje, so talet ikkje bryt
// midt i lina.
func skrivKronor(kr int64) string {
	s := strconv.FormatInt(kr, 10)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteRune('\u00a0')
		}
		b.WriteRune(c)
	}
	b.WriteString("\u00a0kr")
	return b.String()
}

// Day and month names in Norwegian. Go only knows English, and an
// interface that says "Thursday" to someone reading Nynorsk is not
// translated — it is only partly written.
var norskeDagar = map[time.Weekday]string{
	time.Monday: "måndag", time.Tuesday: "tysdag", time.Wednesday: "onsdag",
	time.Thursday: "torsdag", time.Friday: "fredag", time.Saturday: "laurdag",
	time.Sunday: "sundag",
}

var norskeMaanader = map[time.Month]string{
	time.January: "januar", time.February: "februar", time.March: "mars",
	time.April: "april", time.May: "mai", time.June: "juni",
	time.July: "juli", time.August: "august", time.September: "september",
	time.October: "oktober", time.November: "november", time.December: "desember",
}
