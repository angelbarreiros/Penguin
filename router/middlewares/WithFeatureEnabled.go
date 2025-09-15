package middlewares

import (
	"net/http"
	"os"
	"strconv"
)

func WithFeatureEnable(env string, hf http.HandlerFunc) http.HandlerFunc {
	return featureEnabled(env)(hf)
}
func featureEnabled(env string) middlewareFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var osEnvEnabled string = os.Getenv(env)
			var isEnabled, err = strconv.ParseBool(osEnvEnabled)
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
