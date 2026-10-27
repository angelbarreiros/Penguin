package helpers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

const internalErrorBody = `{"error":"Internal Server Error"}`

// writeJSON marshals before writing headers so a marshal failure can still
// produce a proper 500 instead of a 200 with an error body.
func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	jsonBytes, err := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(internalErrorBody))
		return
	}
	w.WriteHeader(statusCode)
	w.Write(jsonBytes)
}

func SendSuccessResponse(w http.ResponseWriter, data any) {
	if data == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}
	writeJSON(w, http.StatusOK, data)
}
func SendValidationErrorResponse(w http.ResponseWriter, errors []string) {
	var validErrors []string
	for _, err := range errors {
		if strings.TrimSpace(err) != "" {
			validErrors = append(validErrors, err)
		}
	}
	writeJSON(w, http.StatusBadRequest, map[string][]string{"error": validErrors})
}
func SendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

func SendNoContentResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Versiones con traducción automática

func SendI18NValidationErrorResponse(w http.ResponseWriter, r *http.Request, errors map[string][]any) {
	// Sort keys so the error order is deterministic.
	keys := make([]string, 0, len(errors))
	for errKey := range errors {
		if strings.TrimSpace(errKey) != "" {
			keys = append(keys, errKey)
		}
	}
	sort.Strings(keys)

	i18nInst := GetI18nInstance()
	var validErrors []string
	for _, errKey := range keys {
		translatedErr := errKey
		if r != nil {
			translatedErr = i18nInst.TranslateFromAcceptLanguageWithVars(r, errKey, errors[errKey]...)
		}
		validErrors = append(validErrors, translatedErr)
	}
	writeJSON(w, http.StatusBadRequest, map[string][]string{"error": validErrors})
}

func SendI18NErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, message string, args ...any) {
	translatedMsg := message
	if r != nil {
		translatedMsg = GetI18nInstance().TranslateFromAcceptLanguageWithVars(r, message, args...)
	}
	writeJSON(w, statusCode, map[string]string{"error": translatedMsg})
}
