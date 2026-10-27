package middlewares

import (
	"encoding/json"
	"net/http"
)

type middlewareFunc func(http.HandlerFunc) http.HandlerFunc

func writeJSONError(w http.ResponseWriter, status int, message string) {
	body, _ := json.Marshal(map[string]string{"error": message})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}
