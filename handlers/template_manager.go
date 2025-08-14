package handlers

import (
	"html/template"
	"io/fs"
	"kjernekraft/handlers/config"
	"os"
	"path/filepath"
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
		t, _ = t.ParseFiles(baseLayoutPath)
	}

	// Load all components
	componentsPath := filepath.Join(tm.basePath, "components")
	if _, err := os.Stat(componentsPath); err == nil {
		filepath.WalkDir(componentsPath, func(compPath string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(compPath, ".html") {
				t, _ = t.ParseFiles(compPath)
			}
			return nil
		})
	}

	// Finally parse the page template
	t, err := t.ParseFiles(path)
	if err == nil {
		tm.templates[name] = t
	}
}

// loadComponentTemplate loads a standalone component template
func (tm *TemplateManager) loadComponentTemplate(name, path string) {
	t := template.New(name).Funcs(getTemplateFuncs())
	t, err := t.ParseFiles(path)
	if err == nil {
		tm.templates[name] = t
	}
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