package auth

import (
	"net/http"
	"time"
)

const (
	DefaultContextTimeout int    = 5
	DefaultContextKey     string = "user"
)

type AuthType interface {
	Authorize(r *http.Request) (bool, error)
	GetUser(r *http.Request) (any, error)
	GetTimeout() time.Duration
	GetContextKey() any
}
