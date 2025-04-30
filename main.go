package main

import (
	"angelotero/commonBackend/router"
	"angelotero/commonBackend/router/auth"
	"net/http"
	"os"
)

func handlerUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
	panic("Error")

}

type customClaims struct {
	Id uint64 `json:"id"`
	auth.PlainClaims
}

func main() {
	file, _ := os.Open("private_key.pem")
	r := router.Router()
	var claims = new(customClaims)
	var auth = auth.NewSingletonJwtAuth(file, claims)
	r.NewRoute(router.Route{
		Path:   "/",
		Method: "GET",
		Handler: router.WithAuthMiddleWare(auth,
			router.WithRateLimiting(
				router.WithLoggingMiddleware(
					router.WithRecovery(handlerUser)))),
		AditionalMethods: []router.HTTPMethod{router.HEAD, router.OPTIONS},
	})

	r.StartServer(":8080")

}
