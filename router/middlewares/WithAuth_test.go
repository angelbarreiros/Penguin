package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type plainAuthStub struct {
	user             any
	contextKey       any
	authorizeContext context.Context
	getUserContext   context.Context
}

func (a *plainAuthStub) Authorize(r *http.Request) (bool, error) {
	a.authorizeContext = r.Context()
	return true, nil
}

func (a *plainAuthStub) GetUser(r *http.Request) (any, error) {
	a.getUserContext = r.Context()
	return a.user, nil
}

func (a *plainAuthStub) GetContextKey() any {
	return a.contextKey
}

type rbacAuthStub struct {
	plainAuthStub
	allowed bool
}

func (a *rbacAuthStub) RBAC(_ []string) bool {
	return a.allowed
}

func TestWithAuthMiddlewareWithoutTimeoutInheritsRequestContext(t *testing.T) {
	t.Parallel()

	type contextKey string
	const userKey contextKey = "user"
	user := struct{ Name string }{Name: "Ada"}
	auth := &plainAuthStub{user: user, contextKey: userKey}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	var handlerContext context.Context
	handler := WithAuthMiddleware(auth, func(_ http.ResponseWriter, r *http.Request) {
		handlerContext = r.Context()
		if got := r.Context().Value(userKey); got != user {
			t.Fatalf("context user = %#v, want %#v", got, user)
		}
	})

	handler(response, req)

	if _, ok := auth.authorizeContext.Deadline(); ok {
		t.Fatal("Authorize context unexpectedly has a deadline")
	}
	if _, ok := auth.getUserContext.Deadline(); ok {
		t.Fatal("GetUser context unexpectedly has a deadline")
	}
	if _, ok := handlerContext.Deadline(); ok {
		t.Fatal("handler context unexpectedly has a deadline")
	}
	if err := handlerContext.Err(); err != nil {
		t.Fatalf("handler context error = %v, want nil", err)
	}
}

func TestWithAuthMiddlewareTimeoutCoversAuthenticationAndHandler(t *testing.T) {
	t.Parallel()

	const timeout = time.Minute
	const userKey = "user"
	const user = "authenticated user"
	auth := &plainAuthStub{user: user, contextKey: userKey}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	startedAt := time.Now()

	var handlerContext context.Context
	handler := WithAuthMiddleware(auth, func(_ http.ResponseWriter, r *http.Request) {
		handlerContext = r.Context()
		if got := r.Context().Value(userKey); got != user {
			t.Fatalf("context user = %#v, want %#v", got, user)
		}
	}, WithTimeout(timeout))

	handler(response, req)

	assertDeadlineNear(t, auth.authorizeContext, startedAt.Add(timeout))
	assertDeadlineNear(t, auth.getUserContext, startedAt.Add(timeout))
	assertDeadlineNear(t, handlerContext, startedAt.Add(timeout))
	if err := handlerContext.Err(); err != context.Canceled {
		t.Fatalf("handler context error after return = %v, want %v", err, context.Canceled)
	}
}

func TestWithAuthMiddlewareDoesNotExtendParentDeadline(t *testing.T) {
	t.Parallel()

	auth := &plainAuthStub{user: "user", contextKey: "user"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	parentContext, cancel := context.WithTimeout(req.Context(), time.Minute)
	defer cancel()
	parentDeadline, _ := parentContext.Deadline()
	req = req.WithContext(parentContext)

	var handlerDeadline time.Time
	handler := WithAuthMiddleware(auth, func(_ http.ResponseWriter, r *http.Request) {
		handlerDeadline, _ = r.Context().Deadline()
	}, WithTimeout(2*time.Minute))

	handler(httptest.NewRecorder(), req)

	if !handlerDeadline.Equal(parentDeadline) {
		t.Fatalf("handler deadline = %v, want parent deadline %v", handlerDeadline, parentDeadline)
	}
}

func TestWithAuthAndRBACTimeoutAndContextValue(t *testing.T) {
	t.Parallel()

	const timeout = time.Minute
	const userKey = "user"
	const user = "administrator"
	auth := &rbacAuthStub{
		plainAuthStub: plainAuthStub{user: user, contextKey: userKey},
		allowed:       true,
	}
	startedAt := time.Now()

	var handlerContext context.Context
	handler := WithAuthAndRBAC(auth, []string{"admin"}, func(_ http.ResponseWriter, r *http.Request) {
		handlerContext = r.Context()
		if got := r.Context().Value(userKey); got != user {
			t.Fatalf("context user = %#v, want %#v", got, user)
		}
	}, WithTimeout(timeout))

	handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assertDeadlineNear(t, auth.authorizeContext, startedAt.Add(timeout))
	assertDeadlineNear(t, auth.getUserContext, startedAt.Add(timeout))
	assertDeadlineNear(t, handlerContext, startedAt.Add(timeout))
}

func TestAuthMiddlewareRejectsNonPositiveTimeout(t *testing.T) {
	t.Parallel()

	auth := &plainAuthStub{}
	rbacAuth := &rbacAuthStub{}
	handler := func(http.ResponseWriter, *http.Request) {}

	tests := []struct {
		name  string
		build func()
	}{
		{
			name:  "plain zero",
			build: func() { WithAuthMiddleware(auth, handler, WithTimeout(0)) },
		},
		{
			name:  "plain negative",
			build: func() { WithAuthMiddleware(auth, handler, WithTimeout(-time.Second)) },
		},
		{
			name:  "RBAC zero",
			build: func() { WithAuthAndRBAC(rbacAuth, nil, handler, WithTimeout(0)) },
		},
		{
			name:  "RBAC negative",
			build: func() { WithAuthAndRBAC(rbacAuth, nil, handler, WithTimeout(-time.Second)) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if recover() == nil {
					t.Fatal("middleware construction did not panic")
				}
			}()
			tt.build()
		})
	}
}

func TestAuthMiddlewareUsesLastTimeoutOption(t *testing.T) {
	t.Parallel()

	auth := &plainAuthStub{user: "user", contextKey: "user"}
	startedAt := time.Now()
	var handlerContext context.Context
	handler := WithAuthMiddleware(auth, func(_ http.ResponseWriter, r *http.Request) {
		handlerContext = r.Context()
	}, WithTimeout(time.Second), WithTimeout(time.Minute))

	handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assertDeadlineNear(t, handlerContext, startedAt.Add(time.Minute))
}

func assertDeadlineNear(t *testing.T, ctx context.Context, want time.Time) {
	t.Helper()
	if ctx == nil {
		t.Fatal("context was not captured")
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("context does not have a deadline")
	}
	const tolerance = time.Second
	if difference := deadline.Sub(want); difference < -tolerance || difference > tolerance {
		t.Fatalf("deadline = %v, want within %v of %v", deadline, tolerance, want)
	}
}
