package middlewares

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/angelbarreiros/Penguin/router/cors"
)

func WithCors(corrsConfig *cors.CORSConfig, hf handleFunc) handleFunc {
	return corsMiddleware(corrsConfig)(hf)
}

func corsMiddleware(corrsConfig *cors.CORSConfig) middlewareFunc {
	return func(hf handleFunc) handleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if corrsConfig == nil {
				hf(w, r)
				return
			}
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Origin", strings.Join(corrsConfig.AllowedOrigins(), ","))
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

			var origin string = r.Header.Get("Origin")
			var allAllowedOrigin bool = slices.Contains(corrsConfig.AllowedOrigins(), cors.AllowAllOrigin)
			if !allAllowedOrigin {
				if strings.TrimSpace(origin) == "" || !slices.Contains(corrsConfig.AllowedOrigins(), origin) {
					http.Error(w, "Origin not allowed", http.StatusForbidden)
					return
				}

			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
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
