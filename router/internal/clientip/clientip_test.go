package clientip

import (
	"net/http/httptest"
	"testing"
)

func TestRealIPHeaderIgnoresForwardedFor(t *testing.T) {
	res := NewResolver([]string{"10.0.0.0/8"}, "X-Real-IP")
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:1234"
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	r.Header.Set("X-Real-IP", "198.51.100.9")
	if got := res.ClientIP(r); got != "198.51.100.9" {
		t.Fatalf("ClientIP = %q", got)
	}
}

func TestForwardedForHeaderIgnoresRealIP(t *testing.T) {
	res := NewResolver([]string{"10.0.0.0/8"}, "X-Forwarded-For")
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:1234"
	r.Header.Set("X-Real-IP", "1.2.3.4")
	if got := res.ClientIP(r); got != "10.0.0.1" {
		t.Fatalf("ClientIP = %q, want the proxy address", got)
	}
}

func TestUnsupportedHeaderPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	NewResolver(nil, "X-Client-IP")
}
