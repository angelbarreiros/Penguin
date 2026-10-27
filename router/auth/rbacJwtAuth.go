package auth

import (
	"crypto/ecdsa"
	"net/http"
	"slices"
	"sync"
)

type jwtRbacOptionsFunc func(*jwtRbacAuthOptions)

// RBACJwtAuth is safe for concurrent use: it keeps no per-request state.
type RBACJwtAuth struct {
	verifier *verifier
	options  *jwtRbacAuthOptions
}
type jwtRbacAuthOptions struct {
	ContextKey any
	Issuer     string
	Audience   string
}

var jwtRbacAuthInstance *RBACJwtAuth
var jwtRbacOnce sync.Once

func NewJwtAuthWithRbac(secret *ecdsa.PrivateKey, claimsType rBACClaimsInterface, options ...jwtRbacOptionsFunc) *RBACJwtAuth {
	return initJwtAuthRbac(publicKeyOf(secret), claimsType, options...)
}

// NewJwtAuthWithRbacAndPublicKey verifies tokens without needing the signing key.
func NewJwtAuthWithRbacAndPublicKey(publicKey *ecdsa.PublicKey, claimsType rBACClaimsInterface, options ...jwtRbacOptionsFunc) *RBACJwtAuth {
	return initJwtAuthRbac(publicKey, claimsType, options...)
}

func NewSingletonJwtAuthWithRbac(secret *ecdsa.PrivateKey, claimsType rBACClaimsInterface, options ...jwtRbacOptionsFunc) *RBACJwtAuth {
	jwtRbacOnce.Do(func() { jwtRbacAuthInstance = initJwtAuthRbac(publicKeyOf(secret), claimsType, options...) })
	return jwtRbacAuthInstance
}

// Authorize checks that a well-formed bearer token is present.
func (j *RBACJwtAuth) Authorize(r *http.Request) (bool, error) {
	if _, err := bearerToken(r); err != nil {
		return false, err
	}
	return true, nil
}

// GetUser verifies the request's token and returns a new claims value.
func (j *RBACJwtAuth) GetUser(r *http.Request) (any, error) {
	return j.verifier.parse(r)
}

// RBAC reports whether user (as returned by GetUser) has any of allowedRoles.
func (j *RBACJwtAuth) RBAC(user any, allowedRoles []string) bool {
	claims, ok := user.(rBACClaimsInterface)
	if !ok {
		return false
	}
	for _, role := range claims.GetRoles() {
		if slices.Contains(allowedRoles, role) {
			return true
		}
	}
	return false
}

func (j *RBACJwtAuth) GetContextKey() any {
	return j.options.ContextKey
}
func JwtAuthRbacWithCustomContextKey(key any) jwtRbacOptionsFunc {
	return func(o *jwtRbacAuthOptions) {
		o.ContextKey = key
	}
}
func JwtAuthRbacWithIssuer(issuer string) jwtRbacOptionsFunc {
	return func(o *jwtRbacAuthOptions) {
		o.Issuer = issuer
	}
}
func JwtAuthRbacWithAudience(audience string) jwtRbacOptionsFunc {
	return func(o *jwtRbacAuthOptions) {
		o.Audience = audience
	}
}

func initJwtAuthRbac(publicKey *ecdsa.PublicKey, claimsType rBACClaimsInterface, options ...jwtRbacOptionsFunc) *RBACJwtAuth {
	var opts *jwtRbacAuthOptions = &jwtRbacAuthOptions{
		ContextKey: DefaultContextKey,
	}
	for _, opt := range options {
		opt(opts)
	}
	return &RBACJwtAuth{
		verifier: newVerifier(publicKey, claimsType, opts.Issuer, opts.Audience),
		options:  opts,
	}
}
