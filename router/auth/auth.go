package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	DefaultContextTimeout int    = 5
	DefaultContextKey     string = "user"
)

type PlainAuthInterface interface {
	Authorize(r *http.Request) (bool, error)
	GetUser(r *http.Request) (any, error)
	GetTimeout() time.Duration
	GetContextKey() any
}
type rBACClaimsInterface interface {
	jwt.Claims
	GetRoles() []string
}
type plainClaimsInterface interface {
	jwt.Claims
}
type PlainClaims struct {
	jwt.RegisteredClaims
}
type RBACClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func (c *RBACClaims) GetRoles() []string {
	return c.Roles
}

type RBACAuthInterface interface {
	Authorize(r *http.Request) (bool, error)
	GetUser(r *http.Request) (any, error)
	RBAC(allowedRoles []string) bool
	GetTimeout() time.Duration
	GetContextKey() any
}

func LoadPrivateKeyFromFile(keyPem []byte) (*ecdsa.PrivateKey, error) {

	block, _ := pem.Decode(keyPem)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
