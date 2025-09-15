package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/angelbarreiros/Penguin/logger"
)

func WithRecovery(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer handlePanic(w)
		hf(w, r)
	}
}

func handlePanic(w http.ResponseWriter) {
	if err := recover(); err != nil {
		logger.GetConsoleLogger().Error("Panic recovered: %v\nStack: %s", err, debug.Stack())
		logger.GetFileLogger().Error("Panic recovered: %v\nStack: %s", err, debug.Stack())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal Server Error"}`))
	}
}
