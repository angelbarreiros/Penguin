package router

import (
	"angelotero/commonBackend/router/auth"
	"angelotero/commonBackend/router/cors"
	"context"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func WithAuthMiddleWare(auth auth.PlainAuthInterface, hf handleFunc) handleFunc {
	return authMiddleWareFunc(auth)(hf)
}

func authMiddleWareFunc(auth auth.PlainAuthInterface) middlewareFunc {
	return func(hf handleFunc) handleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if auth == nil {
				hf(w, r)
				return
			}
			if authorize, err := auth.Authorize(r); !authorize || err != nil {
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return

			}
			user, err := auth.GetUser(r)

			if err != nil {
				http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}
			var ctx, cancel = context.WithTimeout(context.Background(), auth.GetTimeout())
			defer cancel()
			ctx = context.WithValue(r.Context(), auth.GetContextKey(), user)
			r = r.WithContext(ctx)
			hf(w, r)
		}
	}
}
func WithCorsMiddleware(corrsConfig *cors.CORSConfig, hf handleFunc) handleFunc {
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

func WithAuthAndRBAC(authType auth.RBACAuthInterface, roles []string, hf handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if authorize, err := authType.Authorize(r); !authorize || err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		user, err := authType.GetUser(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), authType.GetTimeout())
		defer cancel()
		ctx = context.WithValue(ctx, authType.GetContextKey(), user)
		r = r.WithContext(ctx)
		if !authType.RBAC(roles) {
			http.Error(w, "Forbidden: You don't have the required role", http.StatusForbidden)
			return
		}

		hf(w, r)
	}
}
func WithQueryParametersObligation(queryParameters []string, hf handleFunc) handleFunc {
	return queryParametersObligation(queryParameters)(hf)
}
func queryParametersObligation(queryParameters []string) middlewareFunc {
	return func(hf handleFunc) handleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			for _, queryParameter := range queryParameters {
				if r.URL.Query().Get(queryParameter) == "" {
					http.Error(w, "Query parameter "+queryParameter+" is required", http.StatusBadRequest)
					return
				}
			}
			hf(w, r)
		}
	}
}
