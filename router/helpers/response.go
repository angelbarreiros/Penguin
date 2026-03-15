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

func SendI18NSuccessResponse(w http.ResponseWriter, data any, r *http.Request) {
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

func SendI18NValidationErrorResponse(w http.ResponseWriter, errors []string, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	i18nInst := GetI18nInstance()
	var validErrors []string
	for _, err := range errors {
		if strings.TrimSpace(err) != "" {
			translatedErr := err
			if i18nInst != nil && r != nil {
				translatedErr = i18nInst.TranslateFromAcceptLanguage(err, r)
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

func SendI18NErrorResponse(w http.ResponseWriter, statusCode int, message string, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	translatedMsg := message
	i18nInst := GetI18nInstance()
	if i18nInst != nil && r != nil {
		translatedMsg = i18nInst.TranslateFromAcceptLanguage(message, r)
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
