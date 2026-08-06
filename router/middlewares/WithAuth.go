package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/angelbarreiros/Penguin/router/auth"
)

type authMiddlewareOption func(*authMiddlewareOptions)

type authMiddlewareOptions struct {
	timeout    time.Duration
	timeoutSet bool
}

func WithTimeout(timeout time.Duration) authMiddlewareOption {
	return func(options *authMiddlewareOptions) {
		options.timeout = timeout
		options.timeoutSet = true
	}
}

func WithAuthMiddleware(auth auth.PlainAuthInterface, hf http.HandlerFunc, opts ...authMiddlewareOption) http.HandlerFunc {
	return authMiddlewareFunc(auth, buildAuthMiddlewareOptions(opts...))(hf)
}

func authMiddlewareFunc(auth auth.PlainAuthInterface, options authMiddlewareOptions) middlewareFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r, cancel := withOptionalTimeout(r, options)
			defer cancel()

			if auth == nil {
				hf(w, r)
				return
			}
			if authorize, err := auth.Authorize(r); !authorize || err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Unauthorized: ` + err.Error() + `"}`))
				return

			}
			user, err := auth.GetUser(r)

			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Unauthorized: ` + err.Error() + `"}`))
				return
			}
			ctx := context.WithValue(r.Context(), auth.GetContextKey(), user)
			r = r.WithContext(ctx)
			hf(w, r)
		}
	}
}

func WithAuthAndRBAC(authType auth.RBACAuthInterface, roles []string, hf http.HandlerFunc, opts ...authMiddlewareOption) http.HandlerFunc {
	options := buildAuthMiddlewareOptions(opts...)

	return func(w http.ResponseWriter, r *http.Request) {
		r, cancel := withOptionalTimeout(r, options)
		defer cancel()

		if authorize, err := authType.Authorize(r); !authorize || err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Unauthorized: ` + err.Error() + `"}`))
			return
		}

		user, err := authType.GetUser(r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Unauthorized: ` + err.Error() + `"}`))
			return
		}

		ctx := context.WithValue(r.Context(), authType.GetContextKey(), user)
		r = r.WithContext(ctx)
		if !authType.RBAC(roles) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "Forbidden: You don't have the required role"}`))
			return
		}

		hf(w, r)
	}
}

func buildAuthMiddlewareOptions(opts ...authMiddlewareOption) authMiddlewareOptions {
	var options authMiddlewareOptions
	for _, opt := range opts {
		opt(&options)
	}
	if options.timeoutSet && options.timeout <= 0 {
		panic("middlewares: timeout must be greater than zero")
	}
	return options
}

func withOptionalTimeout(r *http.Request, options authMiddlewareOptions) (*http.Request, context.CancelFunc) {
	if !options.timeoutSet {
		return r, func() {}
	}
	ctx, cancel := context.WithTimeout(r.Context(), options.timeout)
	return r.WithContext(ctx), cancel
}
