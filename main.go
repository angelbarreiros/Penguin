package main

import (
	"angelotero/commonBackend/router"
	"angelotero/commonBackend/router/auth"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

func handlerUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
	panic("Error")

}

type customClaims struct {
	Id uint64 `json:"id"`
	auth.RBACClaims
	jwt.RegisteredClaims
}

func main() {
	file, _ := os.Open("private_key.pem")
	r := router.Router()
	var claims = new(customClaims)
	var auth = auth.NewSingletonJwtAuthWithRbac(file, claims)
	r.NewRoute(router.Route{
		Path:   "/",
		Method: "GET",
		Handler: router.WithAuthAndRBAC(auth, []string{"admin"},
			router.WithRateLimiting(
				router.WithLoggingMiddleware(
					router.WithRecovery(handlerUser)))),
		AditionalMethods: []router.HTTPMethod{router.HEAD, router.OPTIONS},
	})

	r.StartServer(":8080")

}
