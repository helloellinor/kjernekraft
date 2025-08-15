package handlers

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

// Localization holds the translation data
type Localization struct {
	mu        sync.RWMutex
	languages map[string]map[string]interface{}
	basePath  string
}

var localization *Localization
var locOnce sync.Once

// GetLocalization returns the singleton instance of Localization
func GetLocalization() *Localization {
	locOnce.Do(func() {
		localization = &Localization{
			languages: make(map[string]map[string]interface{}),
			basePath:  "locales",
		}
		localization.loadTranslations()
	})
	return localization
}

// loadTranslations loads all translation files
func (l *Localization) loadTranslations() {
	l.mu.Lock()
	defer l.mu.Unlock()

	languages := []string{"nb", "nn", "en"}
	for _, lang := range languages {
		langMap := make(map[string]interface{})
		
		// Load common.json for each language
		commonPath := filepath.Join(l.basePath, lang, "common.json")
		if data, err := ioutil.ReadFile(commonPath); err == nil {
			var translations map[string]interface{}
			if err := json.Unmarshal(data, &translations); err == nil {
				langMap = translations
			}
		}
		
		l.languages[lang] = langMap
	}
}

// T returns a translation for the given key and language
func (l *Localization) T(lang, key string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Default to Norwegian bokmål if language not found
	if _, exists := l.languages[lang]; !exists {
		lang = "nb"
	}

	langMap := l.languages[lang]
	return l.getNestedValue(langMap, key)
}

// getNestedValue retrieves a value from nested map using dot notation
func (l *Localization) getNestedValue(data map[string]interface{}, key string) string {
	keys := strings.Split(key, ".")
	current := data

	for i, k := range keys {
		if i == len(keys)-1 {
			// Last key, return the value
			if val, ok := current[k]; ok {
				if str, ok := val.(string); ok {
					return str
				}
			}
		} else {
			// Navigate deeper
			if val, ok := current[k]; ok {
				if nextMap, ok := val.(map[string]interface{}); ok {
					current = nextMap
				} else {
					return key // Return key if navigation fails
				}
			} else {
				return key // Return key if key not found
			}
		}
	}

	return key // Return key if not found
}

// GetSupportedLanguages returns list of supported languages
func (l *Localization) GetSupportedLanguages() []string {
	return []string{"nb", "nn", "en"}
}

// GetLanguageName returns the display name for a language code
func (l *Localization) GetLanguageName(code string) string {
	names := map[string]string{
		"nb": "Norsk bokmål",
		"nn": "Norsk nynorsk", 
		"en": "English",
	}
	if name, exists := names[code]; exists {
		return name
	}
	return code
}