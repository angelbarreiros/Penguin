package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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
func GetI18nInstance(...i18nFunctions) *i18n {
	if instancei18n == nil {
		oncei18n.Do(func() {
			instancei18n = &i18n{
				DefaultLocale:    "en",
				DefaultDirectory: "i18n",
			}
			for _, fn := range []i18nFunctions{} {
				fn(instancei18n)
			}
		})
	}
	return instancei18n

}
func (i i18n) Translate(key string, r *http.Request) string {
	var sb strings.Builder

	sb.WriteString(i.GetLanguage(r))
	sb.WriteString(".json")
	var i18nFilePath string = filepath.Join(i.DefaultDirectory, sb.String())

	if _, err := os.Stat(i18nFilePath); os.IsNotExist(err) {
		return defaultTranslate(key, i.DefaultLocale)
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

	return key
}
func defaultTranslate(key, locale string) string {
	var sb strings.Builder

	sb.WriteString(locale)
	sb.WriteString(".json")
	var defaultFilePath string = filepath.Join(instancei18n.DefaultDirectory, sb.String())

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
