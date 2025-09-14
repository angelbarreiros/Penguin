package middlewares

import (
	"net/http"
	"time"

	"github.com/angelbarreiros/Penguin/logger"
)

func WithLogging(hf http.HandlerFunc) http.HandlerFunc {
	return loggingMiddleware()(hf)
}

func loggingMiddleware() middlewareFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		var l = logger.GetConsoleLogger()
		return func(w http.ResponseWriter, r *http.Request) {
			var start time.Time = time.Now()
			hf(w, r)
			var duration time.Duration = time.Since(start)
			var method string = r.Method
			var path string = r.URL.Path
			var ip string = r.Header.Get("X-Real-IP")
			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}
			if ip == "" {
				ip = r.RemoteAddr
			}
			l.Info("Method: %s, Path: %s, IP: %s, Duration: %s", method, path, ip, duration)
		}
	}
}
