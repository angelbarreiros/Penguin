package handlers

import (
	"net/http"

	"github.com/bytedance/sonic"
)

func SendSuccessResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if data != nil {
		sonicBytes, err := sonic.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(sonicBytes)
	}
}
func SendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	sonicBytes, err := sonic.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(sonicBytes)
}

func SendNoContentResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
