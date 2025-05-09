package middlewares

import (
	"angelotero/commonBackend/logger"
	"net/http"
	"runtime/debug"
)

func WithRecovery(hf handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer handlePanic(w)
		hf(w, r)
	}
}

func handlePanic(w http.ResponseWriter) {
	if err := recover(); err != nil {
		logger.GetConsoleLogger().Error("Panic recovered: %v\nStack: %s", err, debug.Stack())
		logger.GetFileLogger().Error("Panic recovered: %v\nStack: %s", err, debug.Stack())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
