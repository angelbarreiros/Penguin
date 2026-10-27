package middlewares

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/angelbarreiros/Penguin/router/auth"
	"github.com/angelbarreiros/Penguin/router/cors"
	"github.com/angelbarreiros/Penguin/router/helpers"
	"github.com/golang-jwt/jwt/v5"
)

func signToken(t *testing.T, key *ecdsa.PrivateKey, subject string, roles []string) string {
	t.Helper()
	claims := &auth.RBACClaims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := helpers.GenerateJwtToken(claims, key)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestRBACConcurrentRequestsDoNotShareIdentity(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	rbac := auth.NewJwtAuthWithRbac(key, &auth.RBACClaims{})
	adminToken := signToken(t, key, "admin", []string{"admin"})
	guestToken := signToken(t, key, "guest", []string{"user"})

	var mismatches, guestsAllowed atomic.Int32
	handler := WithAuthAndRBAC(rbac, []string{"admin"}, func(_ http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(auth.DefaultContextKey).(*auth.RBACClaims)
		if claims.Subject != r.Header.Get("X-Expect") {
			mismatches.Add(1)
		}
		if r.Header.Get("X-Expect") == "guest" {
			guestsAllowed.Add(1)
		}
	})

	var wg sync.WaitGroup
	for i := range 2000 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if i%2 == 0 {
				r.Header.Set("Authorization", "Bearer "+adminToken)
				r.Header.Set("X-Expect", "admin")
			} else {
				r.Header.Set("Authorization", "Bearer "+guestToken)
				r.Header.Set("X-Expect", "guest")
			}
			handler(httptest.NewRecorder(), r)
		}()
	}
	wg.Wait()

	if mismatches.Load() != 0 || guestsAllowed.Load() != 0 {
		t.Fatalf("identity mismatches=%d, guest requests allowed=%d", mismatches.Load(), guestsAllowed.Load())
	}
}

func TestPlainJwtRejectsWrongAlgorithmAndExpired(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	plain := auth.NewJwtAuth(key, &auth.PlainClaims{})
	handler := WithAuthMiddleware(plain, func(http.ResponseWriter, *http.Request) {})

	expired, _ := helpers.GenerateJwtToken(&auth.PlainClaims{RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
	}}, key)
	noExp, _ := helpers.GenerateJwtToken(&auth.PlainClaims{}, key)
	hmac, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte("secret"))

	for name, token := range map[string]string{"expired": expired, "no exp": noExp, "hs256": hmac, "garbage": `a"b.c.d`} {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler(w, r)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want 401", name, w.Code)
		}
		var body map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Errorf("%s: body is not valid JSON: %q", name, w.Body.String())
		}
	}
}

type falseNilAuth struct{}

func (falseNilAuth) Authorize(*http.Request) (bool, error) { return false, nil }
func (falseNilAuth) GetUser(*http.Request) (any, error)    { return nil, nil }
func (falseNilAuth) GetContextKey() any                    { return "user" }

func TestAuthorizeFalseWithNilErrorDoesNotPanic(t *testing.T) {
	w := httptest.NewRecorder()
	WithAuthMiddleware(falseNilAuth{}, func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler must not run")
	})(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestNilAuthPanicsAtConstruction(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil auth")
		}
	}()
	WithAuthMiddleware(nil, func(http.ResponseWriter, *http.Request) {})
}

func rateLimitedRequest(h http.HandlerFunc, remoteAddr, forwardedFor string) int {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		r.Header.Set("X-Forwarded-For", forwardedFor)
	}
	w := httptest.NewRecorder()
	h(w, r)
	return w.Code
}

func TestRateLimitIgnoresSpoofedForwardedFor(t *testing.T) {
	h := WithRateLimiting(func(http.ResponseWriter, *http.Request) {}, RateLimitOptStartingLimit(1), RateLimitOptLimitPerSecond(0.001))
	allowed := 0
	for i := range 20 {
		if rateLimitedRequest(h, "203.0.113.5:1000", strings.Repeat("1", i+1)) == http.StatusOK {
			allowed++
		}
	}
	if allowed != 1 {
		t.Fatalf("allowed = %d, want 1", allowed)
	}
}

func TestRateLimitTrustedProxyUsesForwardedClient(t *testing.T) {
	h := WithRateLimiting(func(http.ResponseWriter, *http.Request) {},
		RateLimitOptStartingLimit(1), RateLimitOptLimitPerSecond(0.001), RateLimitOptTrustedProxies([]string{"10.0.0.0/8"}))
	if code := rateLimitedRequest(h, "10.0.0.1:1", "198.51.100.1"); code != http.StatusOK {
		t.Fatalf("first client status = %d", code)
	}
	if code := rateLimitedRequest(h, "10.0.0.1:1", "198.51.100.2"); code != http.StatusOK {
		t.Fatalf("second client status = %d", code)
	}
	if code := rateLimitedRequest(h, "10.0.0.1:1", "198.51.100.1"); code != http.StatusTooManyRequests {
		t.Fatalf("repeat client status = %d, want 429", code)
	}
}

func TestRateLimitKeepsFractionalTokens(t *testing.T) {
	limiter := &rateLimiter{
		config:    rateLimitConfig{startingLimit: 2, limitPerSec: 5, maxEntries: 10},
		buckets:   map[string]*tokenBucket{},
		lastSweep: time.Now(),
		idleAfter: time.Hour,
	}
	now := time.Now()
	allowed := 0
	for i := range 10 {
		if ok, _ := limiter.take("k", now.Add(time.Duration(i)*150*time.Millisecond)); ok {
			allowed++
		}
	}
	// Capacity 2, +0.75 tokens per 150ms gap: requests 5 and 9 are the only
	// rejections. Flooring the refill (the old bug) allowed just 2.
	if allowed != 8 {
		t.Fatalf("allowed = %d, want 8", allowed)
	}
}

func TestRateLimitEvictsIdleBucketsAndCapsEntries(t *testing.T) {
	limiter := &rateLimiter{
		config:    rateLimitConfig{startingLimit: 1, limitPerSec: 1, maxEntries: 2},
		buckets:   map[string]*tokenBucket{},
		lastSweep: time.Now(),
		idleAfter: time.Second,
	}
	now := time.Now()
	limiter.take("a", now)
	limiter.take("b", now)
	if ok, _ := limiter.take("c", now); ok {
		t.Fatal("new client admitted while table is full")
	}
	if ok, _ := limiter.take("c", now.Add(2*time.Second)); !ok {
		t.Fatal("idle buckets were not evicted")
	}
}

func TestCorsVaryAndDisallowedPreflight(t *testing.T) {
	config := cors.NewCORSConfig(cors.WithAllowedOrigins([]string{"https://ok.example"}))
	h := WithCors(config, func(http.ResponseWriter, *http.Request) {})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://ok.example")
	w := httptest.NewRecorder()
	h(w, r)
	if w.Header().Get("Access-Control-Allow-Origin") != "https://ok.example" || !strings.Contains(strings.Join(w.Header().Values("Vary"), ","), "Origin") {
		t.Fatalf("allowed origin headers = %v", w.Header())
	}

	r = httptest.NewRequest(http.MethodOptions, "/", nil)
	r.Header.Set("Origin", "https://evil.example")
	r.Header.Set("Access-Control-Request-Method", "POST")
	w = httptest.NewRecorder()
	h(w, r)
	if w.Code != http.StatusForbidden || w.Header().Get("Access-Control-Allow-Origin") != "" || w.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatalf("disallowed preflight: code=%d headers=%v", w.Code, w.Header())
	}
}

func TestCorsWildcardWithCredentialsPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	cors.NewCORSConfig(cors.WithAllowedOrigins([]string{cors.AllowAllOrigin}), cors.WithAllowCredentials(true))
}

func TestRecoveryAfterPartialWriteDoesNotAppendBody(t *testing.T) {
	w := httptest.NewRecorder()
	func() {
		defer func() {
			if p := recover(); p != http.ErrAbortHandler {
				t.Fatalf("recover = %v, want ErrAbortHandler", p)
			}
		}()
		WithRecovery(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("partial"))
			panic("boom")
		})(w, httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	if w.Body.String() != "partial" {
		t.Fatalf("body = %q", w.Body.String())
	}
}
