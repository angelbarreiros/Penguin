package auth

import (
	"crypto/ecdsa"
	"net/http"
	"sync"
)

type jwtOptionsFunc func(*jwtAuthOptions)

// JwtAuth is safe for concurrent use: it keeps no per-request state.
type JwtAuth struct {
	verifier *verifier
	options  *jwtAuthOptions
}
type jwtAuthOptions struct {
	ContextKey any
	Issuer     string
	Audience   string
}

var jwtAuthInstance *JwtAuth
var jwtOnce sync.Once

func NewJwtAuth(secret *ecdsa.PrivateKey, claimsType plainClaimsInterface, options ...jwtOptionsFunc) *JwtAuth {
	return initJwtAuth(publicKeyOf(secret), claimsType, options...)
}

// NewJwtAuthWithPublicKey verifies tokens without needing the signing key.
func NewJwtAuthWithPublicKey(publicKey *ecdsa.PublicKey, claimsType plainClaimsInterface, options ...jwtOptionsFunc) *JwtAuth {
	return initJwtAuth(publicKey, claimsType, options...)
}

func NewSingletonJwtAuth(secret *ecdsa.PrivateKey, claimsType plainClaimsInterface, options ...jwtOptionsFunc) *JwtAuth {
	jwtOnce.Do(func() { jwtAuthInstance = initJwtAuth(publicKeyOf(secret), claimsType, options...) })
	return jwtAuthInstance
}

// Authorize checks that a well-formed bearer token is present.
func (j *JwtAuth) Authorize(r *http.Request) (bool, error) {
	if _, err := bearerToken(r); err != nil {
		return false, err
	}
	return true, nil
}

// GetUser verifies the request's token and returns a new claims value.
func (j *JwtAuth) GetUser(r *http.Request) (any, error) {
	return j.verifier.parse(r)
}

func (j *JwtAuth) GetContextKey() any {
	return j.options.ContextKey
}
func WithCustomContextKey(key any) jwtOptionsFunc {
	return func(o *jwtAuthOptions) {
		o.ContextKey = key
	}
}
func WithIssuer(issuer string) jwtOptionsFunc {
	return func(o *jwtAuthOptions) {
		o.Issuer = issuer
	}
}
func WithAudience(audience string) jwtOptionsFunc {
	return func(o *jwtAuthOptions) {
		o.Audience = audience
	}
}

func initJwtAuth(publicKey *ecdsa.PublicKey, claimsType plainClaimsInterface, options ...jwtOptionsFunc) *JwtAuth {
	var opts *jwtAuthOptions = &jwtAuthOptions{
		ContextKey: DefaultContextKey,
	}
	for _, opt := range options {
		opt(opts)
	}
	return &JwtAuth{
		verifier: newVerifier(publicKey, claimsType, opts.Issuer, opts.Audience),
		options:  opts,
	}
}
