package middlewares

import (
	"net/http"
	"os"
	"strconv"
)

// Deprecated: use WithFeatureEnabled.
func WithFeatureEnable(env string, hf http.HandlerFunc) http.HandlerFunc {
	return featureEnabled(env)(hf)
}
func featureEnabled(env string) middlewareFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			isEnabled, err := strconv.ParseBool(os.Getenv(env))
			if err != nil || !isEnabled {
				writeJSONError(w, http.StatusForbidden, "Feature not enabled")
				return
			}
			hf(w, r)
		}
	}
}

// WithFeatureEnabled is the correctly spelled name of WithFeatureEnable.
func WithFeatureEnabled(env string, hf http.HandlerFunc) http.HandlerFunc {
	return featureEnabled(env)(hf)
}
