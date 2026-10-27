package middlewares

import (
	"net/http"
	"strconv"
)

// WithFeatureEnabledByHeader gates a handler on a client-supplied header.
//
// This is NOT an access control: any client can send the header. Use it only
// for opt-in behaviour (betas, previews); gate privileged features on
// server-side configuration (WithFeatureEnable) or on auth claims instead.
func WithFeatureEnabledByHeader(header string, hf http.HandlerFunc) http.HandlerFunc {
	return featureEnabledByHeader(header)(hf)
}
func featureEnabledByHeader(header string) middlewareFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			isEnabled, err := strconv.ParseBool(r.Header.Get(header))
			if err != nil || !isEnabled {
				writeJSONError(w, http.StatusForbidden, "Feature not enabled")
				return
			}
			hf(w, r)
		}
	}
}
