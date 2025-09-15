package middlewares

import (
	"net/http"
	"strconv"
)

func WithFeatureEnabledByHeader(header string, hf http.HandlerFunc) http.HandlerFunc {
	return featureEnabledByHeader(header)(hf)
}
func featureEnabledByHeader(header string) middlewareFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var headerValue string = r.Header.Get(header)
			if headerValue == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error": "Feature not enabled"}`))
				return
			}
			var isEnabled, err = strconv.ParseBool(headerValue)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error": "Feature not enabled"}`))
				return
			}
			if !isEnabled {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error": "Feature not enabled"}`))
				return
			}
			hf(w, r)
		}
	}
}
