package main

import (
	"angelotero/commonBackend/router"
	"angelotero/commonBackend/router/auth"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func handleRequests(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
	log.Println("Request received")

}
func main() {
	var file, err = os.Open("private_key.pem")
	if err != nil {
		panic(err)
	}
	type UserClaims struct {
		Expiration time.Time `json:"expiration"`
		Id         uint      `json:"id"`
		Role       string    `json:"role"`
		jwt.RegisteredClaims
	}
	var r = router.NewRouter()
	var jwtAuth = auth.NewSingletonJwtAuth(file, UserClaims{}, auth.WithCustomContextKey("custom_key"))

	r.NewRoute("/hello",
		handleRequests,
		router.WithAuthMiddleWate(jwtAuth),
		router.WithMethodMiddleWare(router.POST))

	r.StartServer(":8080")

}
