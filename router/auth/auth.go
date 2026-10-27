package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

// UserContextKey is the default context key under which the auth middlewares
// store the authenticated user. Its unexported type means no other package
// can collide with it.
var UserContextKey any = contextKey("user")

const (
	// Deprecated: the default key is now UserContextKey. To keep reading
	// r.Context().Value("user"), pass WithCustomContextKey(DefaultContextKey)
	// (or JwtAuthRbacWithCustomContextKey) explicitly.
	DefaultContextKey string = "user"
)

var (
	ErrAuthorizationHeaderMissing = errors.New("authorization header is missing")
	ErrBearerTokenMalformed       = errors.New("bearer token is missing or malformed")
	ErrTokenInvalid               = errors.New("token is invalid")
	ErrTokenExpired               = errors.New("token has expired")
)

type PlainAuthInterface interface {
	Authorize(r *http.Request) (bool, error)
	GetUser(r *http.Request) (any, error)
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

// RBACAuthInterface checks roles against the user returned by GetUser for the
// same request. It must not keep any per-request state.
type RBACAuthInterface interface {
	Authorize(r *http.Request) (bool, error)
	GetUser(r *http.Request) (any, error)
	RBAC(user any, allowedRoles []string) bool
	GetContextKey() any
}

// LoadPrivateKeyFromPEM parses an EC private key in SEC1 ("EC PRIVATE KEY")
// or PKCS#8 ("PRIVATE KEY") PEM format.
func LoadPrivateKeyFromPEM(keyPem []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(keyPem)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	switch block.Type {
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		ecKey, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("PKCS#8 key is not an ECDSA key")
		}
		return ecKey, nil
	default:
		return nil, fmt.Errorf("unsupported PEM block type %q", block.Type)
	}
}

// Deprecated: the argument is PEM bytes, not a file path. Use LoadPrivateKeyFromPEM.
func LoadPrivateKeyFromFile(keyPem []byte) (*ecdsa.PrivateKey, error) {
	return LoadPrivateKeyFromPEM(keyPem)
}

// SigningMethodForKey returns the ES* method matching the key's curve.
func SigningMethodForKey(key *ecdsa.PublicKey) (*jwt.SigningMethodECDSA, error) {
	if key == nil {
		return nil, fmt.Errorf("ecdsa key is nil")
	}
	switch key.Curve {
	case elliptic.P256():
		return jwt.SigningMethodES256, nil
	case elliptic.P384():
		return jwt.SigningMethodES384, nil
	case elliptic.P521():
		return jwt.SigningMethodES512, nil
	default:
		return nil, fmt.Errorf("unsupported ecdsa curve")
	}
}

func bearerToken(r *http.Request) (string, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return "", ErrAuthorizationHeaderMissing
	}
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", ErrBearerTokenMalformed
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrBearerTokenMalformed
	}
	return token, nil
}

// verifier holds the immutable configuration shared by both auth types.
type verifier struct {
	publicKey  *ecdsa.PublicKey
	claimsType reflect.Type
	parser     *jwt.Parser
}

func newVerifier(publicKey *ecdsa.PublicKey, claimsTemplate jwt.Claims, issuer string, audience string) *verifier {
	claimsType := reflect.TypeOf(claimsTemplate)
	if claimsType == nil || claimsType.Kind() != reflect.Pointer {
		panic("auth: claims type must be a non-nil pointer, e.g. &auth.PlainClaims{}")
	}

	parserOptions := []jwt.ParserOption{jwt.WithExpirationRequired()}
	if publicKey != nil {
		method, err := SigningMethodForKey(publicKey)
		if err != nil {
			panic("auth: " + err.Error())
		}
		parserOptions = append(parserOptions, jwt.WithValidMethods([]string{method.Alg()}))
	}
	if issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(issuer))
	}
	if audience != "" {
		parserOptions = append(parserOptions, jwt.WithAudience(audience))
	}

	return &verifier{
		publicKey:  publicKey,
		claimsType: claimsType.Elem(),
		parser:     jwt.NewParser(parserOptions...),
	}
}

// parse returns a fresh claims value for every call so concurrent requests
// never share state.
func (v *verifier) parse(r *http.Request) (jwt.Claims, error) {
	tokenString, err := bearerToken(r)
	if err != nil {
		return nil, err
	}
	if v.publicKey == nil {
		return nil, fmt.Errorf("auth: no verification key configured")
	}

	claims := reflect.New(v.claimsType).Interface().(jwt.Claims)
	token, err := v.parser.ParseWithClaims(tokenString, claims, func(*jwt.Token) (any, error) {
		return v.publicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	if !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

func publicKeyOf(secret *ecdsa.PrivateKey) *ecdsa.PublicKey {
	if secret == nil {
		return nil
	}
	return &secret.PublicKey
}
