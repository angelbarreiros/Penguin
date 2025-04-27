package main

import (
	"angelotero/commonBackend/router"
	"net/http"

	_ "github.com/lib/pq"
)

func handlerUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
	panic("Error")

}
func main() {
	r := router.Router()
	r.NewRoute(router.Route{
		Path:             "/",
		Method:           "GET",
		Handler:          router.WithRateLimiting(router.WithLoggingMiddleware(router.WithRecovery(handlerUser))),
		AditionalMethods: []router.HTTPMethod{router.HEAD, router.OPTIONS},
	})

	r.StartServer(":8080")

}
