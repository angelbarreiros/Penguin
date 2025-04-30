package auth

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtOptionsFunc func(*JwtAuth)
type JwtAuth struct {
	authKey     *ecdsa.PrivateKey
	claimsType  plainClaimsInterface
	tokenString string
	options     *jwtAuthOptions
}
type jwtAuthOptions struct {
	Timeout    time.Duration
	ContextKey any
}

var jwtAuthInstance *JwtAuth
var jwtOnce sync.Once

func NewJwtAuth(secret *os.File, claimsType plainClaimsInterface, options ...jwtOptionsFunc) *JwtAuth {
	return initJwtAuth(secret, claimsType, options...)
}

func NewSingletonJwtAuth(secret *os.File, claimsType plainClaimsInterface, options ...jwtOptionsFunc) *JwtAuth {
	jwtOnce.Do(func() { initJwtAuthInstance(secret, claimsType, options...) })
	return jwtAuthInstance
}

func (j *JwtAuth) Authorize(r *http.Request) (bool, error) {
	var jwtTokenString string = r.Header.Get("Authorization")
	if strings.TrimSpace(jwtTokenString) == "" {
		return false, fmt.Errorf("authorization header is missing")
	}
	if !strings.HasPrefix(jwtTokenString, "Bearer ") {
		return false, fmt.Errorf("authorization header must start with 'bearer '")
	}

	var tokenParts []string = strings.SplitN(jwtTokenString, "Bearer ", 2)
	if len(tokenParts) != 2 || strings.TrimSpace(tokenParts[1]) == "" {
		return false, fmt.Errorf("bearer token is missing or malformed")
	}

	var tokenString string = strings.TrimSpace(tokenParts[1])
	j.tokenString = tokenString
	return true, nil
}
func (j *JwtAuth) GetUser(r *http.Request) (any, error) {
	var jwtToken *jwt.Token
	var err error

	jwtToken, err = jwt.ParseWithClaims(j.tokenString, j.claimsType, func(t *jwt.Token) (any, error) {
		return &j.authKey.PublicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if !jwtToken.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	if jwtToken == nil {
		return nil, fmt.Errorf("no token available")
	}
	var expirationTime *jwt.NumericDate
	expirationTime, err = jwtToken.Claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get expiration time: %v", err)
	}
	if expirationTime == nil {
		return nil, fmt.Errorf("expiration time is missing")

	}
	if expirationTime.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	return jwtToken.Claims, nil
}

func (j *JwtAuth) GetTimeout() time.Duration {
	return j.options.Timeout
}
func (j *JwtAuth) GetContextKey() any {
	return j.options.ContextKey
}
func WithCustomTimeout(timeout time.Duration) jwtRbacOptionsFunc {
	return func(ja *RBACJwtAuth) {
		ja.options.Timeout = timeout
	}
}
func WithCustomContextKey(key any) jwtRbacOptionsFunc {
	return func(ja *RBACJwtAuth) {
		ja.options.ContextKey = key
	}
}

func initJwtAuthInstance(secret *os.File, claimsType plainClaimsInterface, options ...jwtOptionsFunc) {
	jwtAuthInstance = initJwtAuth(secret, claimsType, options...)
}
func initJwtAuth(secret *os.File, claimsType plainClaimsInterface, options ...jwtOptionsFunc) *JwtAuth {
	var bytes []byte
	var err error
	bytes, err = io.ReadAll(secret)
	if err != nil {
		panic("cannot load jwt secret file")
	}

	var key *ecdsa.PrivateKey
	key, err = loadPrivateKeyFromFile(bytes)
	if err != nil {
		panic("cannot load jwt secret file" + err.Error())
	}

	var jwtAuth *JwtAuth = &JwtAuth{
		authKey:    key,
		claimsType: claimsType,
		options: &jwtAuthOptions{
			Timeout:    time.Duration(DefaultContextTimeout),
			ContextKey: DefaultContextKey,
		},
	}

	for _, opt := range options {
		opt(jwtAuth)
	}
	return jwtAuth
}
