package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type i18n struct {
	DefaultLocale    string
	DefaultDirectory string

	// files caches decoded locale files by path; a nil map value means the
	// file is missing or invalid.
	files sync.Map
}

var (
	instancei18n *i18n
	oncei18n     sync.Once
)

type i18nFunctions func(*i18n)

func SetDefaultLocale(locale string) i18nFunctions {
	return func(i *i18n) {
		i.DefaultLocale = locale
	}
}
func SetDirectory(directory string) i18nFunctions {
	return func(i *i18n) {
		i.DefaultDirectory = directory
	}
}

// GetI18nInstance returns the shared instance. Options only apply on the
// first call; later calls ignore them.
func GetI18nInstance(fns ...i18nFunctions) *i18n {
	oncei18n.Do(func() {
		instancei18n = &i18n{
			DefaultLocale:    "en",
			DefaultDirectory: "i18n",
		}
		for _, fn := range fns {
			fn(instancei18n)
		}
		instancei18n.DefaultLocale = normalizeLanguage(instancei18n.DefaultLocale, "en")
	})
	return instancei18n
}

// Reload drops cached locale files so they are read again on next use.
func (i *i18n) Reload() {
	i.files.Clear()
}

// TranslateFromAcceptLanguage traduce una clave basada en el header Accept-Language
func (i *i18n) TranslateFromAcceptLanguage(r *http.Request, key string) string {
	language := i.GetLanguage(r)
	return i.translateWithLanguage(key, language)
}

// TranslateFromAcceptLanguageWithVars traduce una clave con variables {{0}}, {{1}}, etc.
func (i *i18n) TranslateFromAcceptLanguageWithVars(r *http.Request, key string, args ...any) string {
	language := i.GetLanguage(r)
	translation := i.translateWithLanguage(key, language)
	return i.replaceVariables(translation, args...)
}

// TranslateWithLanguage traduce una clave usando solo key + language, sin request.
func (i *i18n) TranslateWithLanguage(key string, language string) string {
	return i.translateWithLanguage(key, normalizeLanguage(language, i.DefaultLocale))
}

// TranslateWithLanguageWithVars traduce una clave usando key + language, sin request,
// reemplazando variables {{0}}, {{1}}, etc.
func (i *i18n) TranslateWithLanguageWithVars(key string, language string, args ...any) string {
	translation := i.TranslateWithLanguage(key, language)
	return i.replaceVariables(translation, args...)
}

func (i *i18n) loadTranslations(language string) map[string]any {
	path := filepath.Join(i.DefaultDirectory, language+".json")
	if cached, ok := i.files.Load(path); ok {
		return cached.(map[string]any)
	}

	var translations map[string]any
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &translations); err != nil {
			translations = nil
		}
	}
	actual, _ := i.files.LoadOrStore(path, translations)
	return actual.(map[string]any)
}

// translateWithLanguage realiza la lógica común de traducción para un lenguaje específico
func (i *i18n) translateWithLanguage(key string, language string) string {
	if value, exists := lookupTranslation(i.loadTranslations(language), key); exists {
		return value
	}
	if language != i.DefaultLocale {
		if value, exists := lookupTranslation(i.loadTranslations(i.DefaultLocale), key); exists {
			return value
		}
	}
	return key
}

var placeholderRegex = regexp.MustCompile(`\{\{(\d+)\}\}`)

// replaceVariables reemplaza {{0}}, {{1}}, etc. con los valores proporcionados.
// It substitutes in a single pass, so placeholders inside an argument are
// left as they are.
func (i *i18n) replaceVariables(text string, args ...any) string {
	if len(args) == 0 {
		return text
	}
	return placeholderRegex.ReplaceAllStringFunc(text, func(placeholder string) string {
		idx, err := strconv.Atoi(placeholder[2 : len(placeholder)-2])
		if err != nil || idx >= len(args) {
			return placeholder
		}
		return fmt.Sprint(args[idx])
	})
}

// lookupTranslation supports both flat keys and dot-separated nested object paths.
func lookupTranslation(translations map[string]any, key string) (string, bool) {
	if translations == nil {
		return "", false
	}
	if directValue, exists := translations[key]; exists {
		if stringValue, ok := directValue.(string); ok {
			return stringValue, true
		}
	}

	if !strings.Contains(key, ".") {
		return "", false
	}

	var current any = translations
	for _, part := range strings.Split(key, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return "", false
		}
		next, exists := object[part]
		if !exists {
			return "", false
		}
		current = next
	}

	stringValue, ok := current.(string)
	return stringValue, ok
}

// normalizeLanguage reduces a tag like "es-ES" to its lowercase primary
// subtag ("es"). Anything that is not 2-3 ASCII letters yields fallback.
func normalizeLanguage(tag string, fallback string) string {
	primary, _, _ := strings.Cut(strings.TrimSpace(tag), "-")
	primary, _, _ = strings.Cut(primary, "_")
	primary = strings.ToLower(primary)
	if len(primary) < 2 || len(primary) > 3 {
		return fallback
	}
	for _, c := range primary {
		if c < 'a' || c > 'z' {
			return fallback
		}
	}
	return primary
}

// GetLanguage picks the highest-weighted language from Accept-Language,
// honouring q-values. Wildcards and malformed entries are skipped.
func (i *i18n) GetLanguage(r *http.Request) string {
	type candidate struct {
		lang  string
		q     float64
		order int
	}
	var candidates []candidate
	for order, part := range strings.Split(r.Header.Get("Accept-Language"), ",") {
		tag, params, _ := strings.Cut(part, ";")
		lang := normalizeLanguage(tag, "")
		if lang == "" {
			continue
		}
		q := 1.0
		if params = strings.TrimSpace(params); strings.HasPrefix(params, "q=") {
			parsed, err := strconv.ParseFloat(strings.TrimPrefix(params, "q="), 64)
			if err != nil {
				continue
			}
			q = parsed
		}
		if q <= 0 {
			continue
		}
		candidates = append(candidates, candidate{lang: lang, q: q, order: order})
	}
	if len(candidates) == 0 {
		return i.DefaultLocale
	}
	sort.SliceStable(candidates, func(a, b int) bool { return candidates[a].q > candidates[b].q })
	return candidates[0].lang
}
