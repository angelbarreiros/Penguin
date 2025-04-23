package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type optionsFunc func(*JwtAuth)
type JwtAuth struct {
	authKey    *ecdsa.PrivateKey
	claimsType jwt.Claims
	tokeString string
	options    *jwtAuthOptions
}
type jwtAuthOptions struct {
	Timeout    time.Duration
	ContextKey any
}

var jwtAuthInstance *JwtAuth
var jwtOnce sync.Once

func NewJwtAuth(secret *os.File, claimsType jwt.Claims, options ...optionsFunc) *JwtAuth {
	return initJwtAuth(secret, claimsType, options...)
}

func NewSingletonJwtAuth(secret *os.File, claimsType jwt.Claims, options ...optionsFunc) *JwtAuth {
	jwtOnce.Do(func() { initJwtAuthInstance(secret, claimsType, options...) })
	return jwtAuthInstance
}

func (j *JwtAuth) Authorize(r *http.Request) (bool, error) {
	var jwtTokenString string = r.Header.Get("Authorization")
	if strings.TrimSpace(jwtTokenString) == "" {
		return false, fmt.Errorf("Authorization header is missing")
	}
	if !strings.HasPrefix(jwtTokenString, "Bearer ") {
		return false, fmt.Errorf("Authorization header must start with 'Bearer '")
	}

	var tokenParts []string = strings.SplitN(jwtTokenString, "Bearer ", 2)
	if len(tokenParts) != 2 || strings.TrimSpace(tokenParts[1]) == "" {
		return false, fmt.Errorf("Bearer token is missing or malformed")
	}

	var tokenString string = strings.TrimSpace(tokenParts[1])
	j.tokeString = tokenString
	return true, nil
}
func (j *JwtAuth) GetUser(r *http.Request) (any, error) {
	var jwtToken *jwt.Token
	var err error
	jwtToken, err = jwt.ParseWithClaims(j.tokeString, j.claimsType, func(t *jwt.Token) (any, error) {
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
		return nil, fmt.Errorf("token is expired")
	}
	return jwtToken.Claims, nil
}
func (j *JwtAuth) GetTimeout() time.Duration {
	return j.options.Timeout
}
func (j *JwtAuth) GetContextKey() any {
	return j.options.ContextKey
}
func WithCustomTimeout(timeout time.Duration) optionsFunc {
	return func(ja *JwtAuth) {
		ja.options.Timeout = timeout
	}
}
func WithCustomContextKey(key any) optionsFunc {
	return func(ja *JwtAuth) {
		ja.options.ContextKey = key
	}
}

func initJwtAuthInstance(secret *os.File, claimsType jwt.Claims, options ...optionsFunc) {
	jwtAuthInstance = initJwtAuth(secret, claimsType, options...)
}
func initJwtAuth(secret *os.File, claimsType jwt.Claims, options ...optionsFunc) *JwtAuth {
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
	var jwtAuth *JwtAuth
	jwtAuth = &JwtAuth{
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

func loadPrivateKeyFromFile(keyPem []byte) (*ecdsa.PrivateKey, error) {

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
