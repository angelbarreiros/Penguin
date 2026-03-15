package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type i18n struct {
	DefaultLocale    string
	DefaultDirectory string
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
func GetI18nInstance(fns ...i18nFunctions) *i18n {
	if instancei18n == nil {
		oncei18n.Do(func() {
			instancei18n = &i18n{
				DefaultLocale:    "en",
				DefaultDirectory: "i18n",
			}
			for _, fn := range fns {
				fn(instancei18n)
			}
		})
	}
	return instancei18n
}

// TranslateFromAcceptLanguage traduce una clave basada en el header Accept-Language
func (i i18n) TranslateFromAcceptLanguage(r *http.Request, key string) string {
	language := i.GetLanguage(r)
	return i.translateWithLanguage(key, language)
}

// TranslateFromAcceptLanguageWithVars traduce una clave con variables {{0}}, {{1}}, etc.
func (i i18n) TranslateFromAcceptLanguageWithVars(r *http.Request, key string, args ...any) string {
	language := i.GetLanguage(r)
	translation := i.translateWithLanguage(key, language)
	return i.replaceVariables(translation, args...)
}

// Translate mantiene compatibilidad con versiones anteriores
func (i i18n) Translate(r *http.Request, key string) string {
	return i.TranslateFromAcceptLanguage(r, key)
}

// translateWithLanguage realiza la lógica común de traducción para un lenguaje específico
func (i i18n) translateWithLanguage(key string, language string) string {
	var sb strings.Builder

	sb.WriteString(language)
	sb.WriteString(".json")
	var i18nFilePath string = filepath.Join(i.DefaultDirectory, sb.String())

	if _, err := os.Stat(i18nFilePath); os.IsNotExist(err) {
		return i.translateFromDefault(key)
	}
	file, err := os.Open(i18nFilePath)
	if err != nil {
		return key
	}
	defer file.Close()

	var translations map[string]string
	if err := json.NewDecoder(file).Decode(&translations); err != nil {
		return key
	}

	if value, exists := translations[key]; exists {
		return value
	}

	return i.translateFromDefault(key)
}

// translateFromDefault busca la traducción en el archivo del idioma por defecto
func (i i18n) translateFromDefault(key string) string {
	var sb strings.Builder

	sb.WriteString(i.DefaultLocale)
	sb.WriteString(".json")
	var defaultFilePath string = filepath.Join(i.DefaultDirectory, sb.String())

	if _, err := os.Stat(defaultFilePath); os.IsNotExist(err) {
		return key
	}
	file, err := os.Open(defaultFilePath)
	if err != nil {
		return key
	}
	defer file.Close()

	var translations map[string]string
	if err := json.NewDecoder(file).Decode(&translations); err != nil {
		return key
	}

	if value, exists := translations[key]; exists {
		return value
	}

	return key
}

// replaceVariables reemplaza {{0}}, {{1}}, etc. con los valores proporcionados
func (i i18n) replaceVariables(text string, args ...any) string {
	result := text
	for idx, arg := range args {
		placeholder := "{{" + strconv.Itoa(idx) + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprint(arg))
	}
	return result
}
func (i i18n) GetLanguage(r *http.Request) string {
	al := r.Header.Get("Accept-Language")
	if al == "" {
		return i.DefaultLocale
	}
	if len(al) < 2 {
		return i.DefaultLocale
	}
	return strings.ToLower(al[:2])
}
