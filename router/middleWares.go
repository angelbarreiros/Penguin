package router

import (
	"angelotero/commonBackend/router/auth"
	"context"
	"net/http"
)

type AuthType = auth.AuthType

func methodMiddlewareFunc(method HTTPMethod) middlewareFunc {
	return func(next handleFunc) handleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != string(method) {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			next(w, r)
		}
	}
}

func methodMiddleware(method HTTPMethod) *middleware {
	return &middleware{
		priority:       1,
		middlewareFunc: methodMiddlewareFunc(method),
	}

}
func WithMethodMiddleWare(method HTTPMethod) routeOptions {
	return func(r *route) {
		if _, exists := r.usedMiddlewareMap[MethodMiddleware]; exists {
			panic("Method middleware already used")

		}
		r.middlewares = append(r.middlewares, methodMiddleware(method))
		r.usedMiddlewareMap[MethodMiddleware] = true
	}
}

func WithAuthMiddleWate(auth AuthType) routeOptions {
	return func(r *route) {
		if _, exists := r.usedMiddlewareMap[AuthMiddleware]; exists {
			panic("Method middleware already used")

		}
		r.middlewares = append(r.middlewares, authMiddleWare(auth))
		r.usedMiddlewareMap[AuthMiddleware] = true
	}
}

func authMiddleWare(auth AuthType) *middleware {
	return &middleware{
		priority:       0,
		middlewareFunc: authMiddleWareFunc(auth),
	}
}
func authMiddleWareFunc(auth AuthType) middlewareFunc {
	return func(hf handleFunc) handleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
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
