package middlewares

import (
	"context"
	"net/http"

	"github.com/angelbarreiros/Penguin/router/auth"
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
