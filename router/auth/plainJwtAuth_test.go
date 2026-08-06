package auth

import "testing"

func TestWithCustomContextKeyConfiguresPlainJwtAuth(t *testing.T) {
	t.Parallel()

	type contextKey string
	const key contextKey = "authenticated-user"

	jwtAuth := NewJwtAuth(nil, &PlainClaims{}, WithCustomContextKey(key))

	if got := jwtAuth.GetContextKey(); got != key {
		t.Fatalf("context key = %#v, want %#v", got, key)
	}
}
