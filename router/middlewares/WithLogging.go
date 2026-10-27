package middlewares

import (
	"net/http"
	"time"

	"github.com/angelbarreiros/Penguin/logger"
	"github.com/angelbarreiros/Penguin/router/internal/clientip"
)

func WithLogging(hf http.HandlerFunc) http.HandlerFunc {
	return loggingMiddleware(nil)(hf)
}

// WithLoggingTrustedProxies logs the forwarded client IP only when the peer
// is one of the given proxy IPs/CIDRs.
func WithLoggingTrustedProxies(trustedProxies []string, hf http.HandlerFunc) http.HandlerFunc {
	return loggingMiddleware(trustedProxies)(hf)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

func (s *statusRecorder) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}

func loggingMiddleware(trustedProxies []string) middlewareFunc {
	resolver := clientip.NewResolver(trustedProxies)
	return func(hf http.HandlerFunc) http.HandlerFunc {
		var l = logger.GetConsoleLogger()
		return func(w http.ResponseWriter, r *http.Request) {
			var start time.Time = time.Now()
			rec := &statusRecorder{ResponseWriter: w}
			hf(rec, r)
			if rec.status == 0 {
				rec.status = http.StatusOK
			}
			// %q escapes control characters so request data cannot forge log lines.
			l.Info("Method: %q, Path: %q, Status: %d, IP: %q, Duration: %s",
				r.Method, r.URL.Path, rec.status, resolver.ClientIP(r), time.Since(start))
		}
	}
}
