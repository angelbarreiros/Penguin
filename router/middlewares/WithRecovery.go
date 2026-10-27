package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/angelbarreiros/Penguin/logger"
)

func WithRecovery(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &headerTracker{ResponseWriter: w}
		defer handlePanic(rec)
		hf(rec, r)
	}
}

type headerTracker struct {
	http.ResponseWriter
	wroteHeader bool
}

func (h *headerTracker) WriteHeader(code int) {
	h.wroteHeader = true
	h.ResponseWriter.WriteHeader(code)
}

func (h *headerTracker) Write(b []byte) (int, error) {
	h.wroteHeader = true
	return h.ResponseWriter.Write(b)
}

func (h *headerTracker) Unwrap() http.ResponseWriter {
	return h.ResponseWriter
}

func handlePanic(w *headerTracker) {
	err := recover()
	if err == nil {
		return
	}
	// net/http uses ErrAbortHandler to abort a response on purpose.
	if err == http.ErrAbortHandler {
		panic(err)
	}
	stack := debug.Stack()
	logger.GetConsoleLogger().Error("Panic recovered: %v\nStack: %s", err, stack)
	logger.GetFileLogger().Error("Panic recovered: %v\nStack: %s", err, stack)
	if w.wroteHeader {
		// The response is already partially sent; abort the connection rather
		// than appending an error body to it.
		panic(http.ErrAbortHandler)
	}
	writeJSONError(w, http.StatusInternalServerError, "Internal Server Error")
}
