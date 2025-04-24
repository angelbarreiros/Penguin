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

	"slices"

	"github.com/golang-jwt/jwt/v5"
)

type jwtRbacOptionsFunc func(*RBACJwtAuth)
type RBACJwtAuth struct {
	authKey     *ecdsa.PrivateKey
	claimsType  RBACClaims
	tokenString string
	options     *jwtRbacAuthOptions
}
type jwtRbacAuthOptions struct {
	Timeout    time.Duration
	ContextKey any
}

var jwtRbacAuthInstance *RBACJwtAuth
var jwtRbacOnce sync.Once

func NewJwtAuthWithRbac(secret *os.File, claimsType jwt.Claims, options ...jwtRbacOptionsFunc) *RBACJwtAuth {
	return initJwtAuthRbac(secret, claimsType, options...)
}

func NewSingletonJwtAuthWithRbac(secret *os.File, claimsType jwt.Claims, options ...jwtRbacOptionsFunc) *RBACJwtAuth {
	jwtRbacOnce.Do(func() { initJwtAuthRbacInstance(secret, claimsType, options...) })
	return jwtRbacAuthInstance
}

func (j *RBACJwtAuth) Authorize(r *http.Request) (bool, error) {
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
func (j *RBACJwtAuth) GetUser(r *http.Request) (any, error) {
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
func (j *RBACJwtAuth) RBAC(allowedRoles []string) bool {

	var roles []string = j.claimsType.GetRoles()
	if len(roles) == 0 {
		return false
	}
	for _, role := range roles {
		if slices.Contains(allowedRoles, role) {
			return true
		}
	}
	return false
}

func (j *RBACJwtAuth) GetTimeout() time.Duration {
	return j.options.Timeout
}
func (j *RBACJwtAuth) GetContextKey() any {
	return j.options.ContextKey
}
func JwtAuthRbacWithCustomTimeout(timeout time.Duration) jwtRbacOptionsFunc {
	return func(ja *RBACJwtAuth) {
		ja.options.Timeout = timeout
	}
}
func JwtAuthRbacWithCustomContextKey(key any) jwtRbacOptionsFunc {
	return func(ja *RBACJwtAuth) {
		ja.options.ContextKey = key
	}
}

func initJwtAuthRbacInstance(secret *os.File, claimsType jwt.Claims, options ...jwtRbacOptionsFunc) {
	jwtRbacAuthInstance = initJwtAuthRbac(secret, claimsType, options...)
}
func initJwtAuthRbac(secret *os.File, claimsType jwt.Claims, options ...jwtRbacOptionsFunc) *RBACJwtAuth {
	var rbaclaims, ok = claimsType.(RBACClaims)
	if !ok {
		panic("claimsType must implement RBACClaims interface")
	}
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

	var jwtAuth *RBACJwtAuth = &RBACJwtAuth{
		authKey:    key,
		claimsType: rbaclaims,
		options: &jwtRbacAuthOptions{
			Timeout:    time.Duration(DefaultContextTimeout),
			ContextKey: DefaultContextKey,
		},
	}

	for _, opt := range options {
		opt(jwtAuth)
	}
	return jwtAuth
}
