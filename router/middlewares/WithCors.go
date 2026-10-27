package middlewares

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/angelbarreiros/Penguin/router/cors"
)

func WithCors(corsConfig *cors.CORSConfig, hf http.HandlerFunc) http.HandlerFunc {
	return corsMiddleware(corsConfig)(hf)
}

func corsMiddleware(corsConfig *cors.CORSConfig) middlewareFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if corsConfig == nil {
				hf(w, r)
				return
			}

			origin := r.Header.Get("Origin")
			header := w.Header()
			// The response depends on Origin whenever origins are echoed back,
			// so caches must key on it.
			if !corsConfig.AllowsAllOrigins() {
				header.Add("Vary", "Origin")
			}

			isPreflight := r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
			if isPreflight {
				header.Add("Vary", "Access-Control-Request-Method")
				header.Add("Vary", "Access-Control-Request-Headers")
			}

			// Empty or same origin: non-browser or same-origin request, no CORS
			// headers needed. Browsers also send Origin on same-origin POSTs.
			if origin == "" || isSameOrigin(origin, r) {
				if isPreflight && !corsConfig.OptionsPassthrough() {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				hf(w, r)
				return
			}

			if !corsConfig.IsOriginAllowed(origin) {
				writeJSONError(w, http.StatusForbidden, "Origin not allowed")
				return
			}

			if corsConfig.AllowsAllOrigins() {
				header.Set("Access-Control-Allow-Origin", cors.AllowAllOrigin)
			} else {
				header.Set("Access-Control-Allow-Origin", origin)
			}
			if corsConfig.AllowCredentials() {
				header.Set("Access-Control-Allow-Credentials", "true")
			}

			if isPreflight {
				header.Set("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders(), ", "))
				header.Set("Access-Control-Max-Age", strconv.Itoa(corsConfig.MaxAge()))
				if corsConfig.OptionsPassthrough() {
					hf(w, r)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if len(corsConfig.ExposedHeaders()) > 0 {
				header.Set("Access-Control-Expose-Headers", strings.Join(corsConfig.ExposedHeaders(), ", "))
			}
			hf(w, r)
		}
	}
}

// isSameOrigin reports whether origin names the host the request was sent to.
// Only the host is compared, since the scheme is unknown behind a TLS proxy.
func isSameOrigin(origin string, r *http.Request) bool {
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	return strings.EqualFold(u.Host, r.Host)
}
