package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"kjernekraft/handlers/config"
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
}

var templateManager *TemplateManager
var tmplOnce sync.Once

// GetTemplateManager returns the singleton instance of TemplateManager
func GetTemplateManager() *TemplateManager {
	tmplOnce.Do(func() {
		wd, _ := os.Getwd()
		templateManager = &TemplateManager{
			templates: make(map[string]*template.Template),
			basePath:  filepath.Join(wd, "handlers", "templates"),
		}
		templateManager.loadTemplates()
	})
	return templateManager
}

// Template function map with commonly used functions
func getTemplateFuncs() template.FuncMap {
	settings := config.GetInstance()
	return template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"substr": func(s string, start int, length int) string {
			if start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		// Pengar er lagra i øre og skal lesast i kronor. Tabellen synte
		// «104000 øre» — paa den skjermen Ida redigerer prisar paa.
		// Tusenskiljet er eit hardt mellomrom, so talet ikkje bryt
		// midt i lina.
		//
		// Han tek imot baade heiltal og flyttal, av di prisen er `int`
		// paa Membership og `float64` paa FreezeRequest. Det er ein
		// lyte i modellane — pengar bør vera eitt slag heile vegen —
		// men malen skal ikkje falla paa det.
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
		"formatTime": func(t time.Time, format string) string {
			return t.In(settings.GetLocation()).Format(format)
		},
		// Go sitt Format gjev engelske dag- og maanadsnamn, og det finst
		// ingen maate aa be honom um noko anna. Difor stend namni her.
		// «Thursday 27. August» stod i dokka paa ei norsk sida.
		"norskDag": func(t time.Time) string {
			return norskeDagar[t.Weekday()]
		},
		"norskDato": func(t time.Time) string {
			return fmt.Sprintf("%s %d. %s", norskeDagar[t.Weekday()], t.Day(), norskeMaanader[t.Month()])
		},
		"formatTimeShort": func(t time.Time) string {
			return t.In(settings.GetLocation()).Format("15:04")
		},
		"formatDateShort": func(t time.Time) string {
			return t.In(settings.GetLocation()).Format("02.01")
		},
		"formatDateTime": func(t time.Time) string {
			return t.In(settings.GetLocation()).Format("2006-01-02 15:04")
		},
		"formatDateTimeLocal": func(t time.Time) string {
			return t.In(settings.GetLocation()).Format("2006-01-02T15:04")
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
		"translate": func(lang, key string) string {
			loc := GetLocalization()
			return loc.T(lang, key)
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

// loadComponentTemplate loads a standalone component template.
//
// Feilen vart slukt her fyrr — malen datt berre ut or samlingi, og den
// fyrste ein visste um det var at ei sida mangla ein bolk.
func (tm *TemplateManager) loadComponentTemplate(name, path string) {
	t := template.New(name).Funcs(getTemplateFuncs())
	parsed, err := t.ParseFiles(path)
	if err != nil {
		log.Printf("Error parsing component template %s: %v", path, err)
		return
	}
	tm.templates[name] = parsed
}

// GetTemplate returns a template by name
func (tm *TemplateManager) GetTemplate(name string) (*template.Template, bool) {
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

// Dag- og maanadsnamn paa norsk. Go kann berre engelsk, og eit
// grensesnitt som segjer «Thursday» til ein som les nynorsk er ikkje
// omsett — det er berre delvis skrive.
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
