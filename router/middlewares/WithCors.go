package middlewares

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/angelbarreiros/Penguin/router/cors"
)

func WithCors(corrsConfig *cors.CORSConfig, hf http.HandlerFunc) http.HandlerFunc {
	return corsMiddleware(corrsConfig)(hf)
}

func corsMiddleware(corrsConfig *cors.CORSConfig) middlewareFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if corrsConfig == nil {
				hf(w, r)
				return
			}

			origin := r.Header.Get("Origin")
			allowAllOrigins := slices.Contains(corrsConfig.AllowedOrigins(), cors.AllowAllOrigin)

			if r.Method == http.MethodOptions {
				// For OPTIONS: echo back the actual origin if allowed, or "*" if all origins allowed
				if allowAllOrigins {
					w.Header().Set("Access-Control-Allow-Origin", cors.AllowAllOrigin)
				} else if origin != "" && slices.Contains(corrsConfig.AllowedOrigins(), origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}

				w.Header().Set("Access-Control-Allow-Headers", strings.Join(corrsConfig.AllowedHeaders(), ","))
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(corrsConfig.MaxAge()))
				if corrsConfig.AllowCredentials() {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				if corrsConfig.ExposedHeaders() != nil {
					w.Header().Set("Access-Control-Expose-Headers", strings.Join(corrsConfig.ExposedHeaders(), ","))
				}
				if corrsConfig.OptionsPassthrough() {
					hf(w, r)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// For regular requests: empty origin is allowed (same-origin or non-browser tools)
			// Only reject if origin is non-empty AND not in allowed list
			if !allowAllOrigins && origin != "" && !slices.Contains(corrsConfig.AllowedOrigins(), origin) {
				// Add CORS headers to error response so browser can read it
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(corrsConfig.AllowedHeaders(), ","))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error": "Origin not allowed"}`))
				return
			}

			// Set CORS headers for allowed requests
			if allowAllOrigins {
				w.Header().Set("Access-Control-Allow-Origin", cors.AllowAllOrigin)
			} else if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Headers", strings.Join(corrsConfig.AllowedHeaders(), ","))
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(corrsConfig.MaxAge()))
			if corrsConfig.AllowCredentials() {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if corrsConfig.ExposedHeaders() != nil {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(corrsConfig.ExposedHeaders(), ","))
			}
			hf(w, r)
		}
	}
}
