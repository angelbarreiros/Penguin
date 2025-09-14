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
