package helpers

import (
	"encoding/json"
	"net/http"
	"strings"
)

func SendSuccessResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if data != nil {
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "` + err.Error() + `"}`))
			return
		}
		w.Write(jsonBytes)
	}
}
func SendValidationErrorResponse(w http.ResponseWriter, errors []string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	var validErrors []string
	for _, err := range errors {
		if strings.TrimSpace(err) != "" {
			validErrors = append(validErrors, err)
		}
	}

	response := map[string][]string{"error": validErrors}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "` + err.Error() + `"}`))
		return
	}
	w.Write(jsonBytes)
}
func SendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "` + err.Error() + `"}`))
		return
	}
	w.Write(jsonBytes)
}

func SendNoContentResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Versiones con traducción automática

func SendI18NValidationErrorResponse(w http.ResponseWriter, r *http.Request, errors map[string][]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	i18nInst := GetI18nInstance()
	var validErrors []string
	for errKey, args := range errors {
		if strings.TrimSpace(errKey) != "" {
			translatedErr := errKey
			if i18nInst != nil && r != nil {
				translatedErr = i18nInst.TranslateFromAcceptLanguageWithVars(r, errKey, args...)
			}
			validErrors = append(validErrors, translatedErr)
		}
	}

	response := map[string][]string{"error": validErrors}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "` + err.Error() + `"}`))
		return
	}
	w.Write(jsonBytes)
}

func SendI18NErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, message string, args ...any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	translatedMsg := message
	i18nInst := GetI18nInstance()
	if i18nInst != nil && r != nil {
		translatedMsg = i18nInst.TranslateFromAcceptLanguageWithVars(r, message, args...)
	}

	response := map[string]string{"error": translatedMsg}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "` + err.Error() + `"}`))
		return
	}
	w.Write(jsonBytes)
}
