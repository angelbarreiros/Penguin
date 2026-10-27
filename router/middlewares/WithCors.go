package middlewares

import (
	"net/http"
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

			// Empty origin: same-origin or non-browser request, no CORS headers needed.
			if origin == "" {
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
