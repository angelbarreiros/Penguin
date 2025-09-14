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
				http.Error(w, "Feature not enabled", http.StatusForbidden)
				return
			}
			var isEnabled, err = strconv.ParseBool(headerValue)
			if err != nil {
				http.Error(w, "Feature not enabled", http.StatusForbidden)
				return
			}
			if !isEnabled {
				http.Error(w, "Feature not enabled", http.StatusForbidden)
				return
			}
			hf(w, r)
		}
	}
}
