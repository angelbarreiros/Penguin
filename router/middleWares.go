package router

import (
	"angelotero/commonBackend/router/auth"
	"context"
	"net/http"
)

type AuthType = auth.AuthType

func WithAuthMiddleWare(auth AuthType, hf handleFunc) handleFunc {
	return authMiddleWareFunc(auth)(hf)
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
