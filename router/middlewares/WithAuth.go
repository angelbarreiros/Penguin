package middlewares

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/angelbarreiros/Penguin/router/auth"
)

type authMiddlewareOption func(*authMiddlewareOptions)

type authMiddlewareOptions struct {
	timeout    time.Duration
	timeoutSet bool
}

// WithTimeout sets a deadline on the request context for authentication and
// the handler. Handlers must observe r.Context() for it to have any effect.
func WithTimeout(timeout time.Duration) authMiddlewareOption {
	return func(options *authMiddlewareOptions) {
		options.timeout = timeout
		options.timeoutSet = true
	}
}

// WithAuthMiddleware panics if auth is nil: a missing authenticator must never
// silently disable authentication.
func WithAuthMiddleware(auth auth.PlainAuthInterface, hf http.HandlerFunc, opts ...authMiddlewareOption) http.HandlerFunc {
	if auth == nil {
		panic("middlewares: WithAuthMiddleware requires a non-nil auth")
	}
	return authMiddlewareFunc(auth, buildAuthMiddlewareOptions(opts...))(hf)
}

func authMiddlewareFunc(auth auth.PlainAuthInterface, options authMiddlewareOptions) middlewareFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r, cancel := withOptionalTimeout(r, options)
			defer cancel()

			user, ok := authenticate(w, r, auth)
			if !ok {
				return
			}
			ctx := context.WithValue(r.Context(), auth.GetContextKey(), user)
			hf(w, r.WithContext(ctx))
		}
	}
}

// WithAuthAndRBAC panics if authType is nil.
func WithAuthAndRBAC(authType auth.RBACAuthInterface, roles []string, hf http.HandlerFunc, opts ...authMiddlewareOption) http.HandlerFunc {
	if authType == nil {
		panic("middlewares: WithAuthAndRBAC requires a non-nil auth")
	}
	options := buildAuthMiddlewareOptions(opts...)

	return func(w http.ResponseWriter, r *http.Request) {
		r, cancel := withOptionalTimeout(r, options)
		defer cancel()

		user, ok := authenticate(w, r, authType)
		if !ok {
			return
		}
		if !authType.RBAC(user, roles) {
			writeJSONError(w, http.StatusForbidden, "Forbidden: You don't have the required role")
			return
		}

		ctx := context.WithValue(r.Context(), authType.GetContextKey(), user)
		hf(w, r.WithContext(ctx))
	}
}

func authenticate(w http.ResponseWriter, r *http.Request, a auth.PlainAuthInterface) (any, bool) {
	if authorized, err := a.Authorize(r); !authorized || err != nil {
		writeJSONError(w, http.StatusUnauthorized, unauthorizedMessage(err))
		return nil, false
	}
	user, err := a.GetUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, unauthorizedMessage(err))
		return nil, false
	}
	return user, true
}

// unauthorizedMessage only exposes the library's own sentinel errors; anything
// else (custom implementations, parser internals) becomes a generic message.
func unauthorizedMessage(err error) string {
	for _, known := range []error{
		auth.ErrAuthorizationHeaderMissing,
		auth.ErrBearerTokenMalformed,
		auth.ErrTokenInvalid,
		auth.ErrTokenExpired,
	} {
		if errors.Is(err, known) {
			return "Unauthorized: " + known.Error()
		}
	}
	return "Unauthorized"
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
