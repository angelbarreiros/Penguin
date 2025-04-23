package main

import (
	"angelotero/commonBackend/router"
	"angelotero/commonBackend/router/auth"
	"angelotero/commonBackend/router/cors"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims implementa la interfaz RBACClaims
type CustomClaims struct {
	ID    uint     `json:"id"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func (c CustomClaims) GetRoles() []string {
	return c.Roles
}

func publicHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, Public Route!"))
	log.Println("Public request received")
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, Admin Route!"))
	log.Println("Admin request received")
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	if claims, ok := context.Value(auth.DefaultContextKey).(*CustomClaims); ok {
		log.Println("User ID:", claims.ID)
	}

	w.Write([]byte("Hello, User Route!"))
	log.Println("User request received")
}

func main() {
	var file, err = os.Open("private_key.pem")
	defer file.Close()
	if err != nil {
		log.Fatal("Error opening private key:", err)
	}
	var file2, err2 = os.Open("private_key2.pem")
	if err2 != nil {
		log.Fatal("Error opening private key:", err2)
	}
	defer file.Close()

	r := router.Router()

	// Crear instancia de CustomClaims para el auth
	var claims = new(CustomClaims)

	// Configurar auth y cors
	var rbacConfig = auth.NewSingletonJwtAuthWithRbac(file, claims)
	var _ = auth.NewSingletonJwtAuth(file2, claims)
	var corsConfig = cors.NewCORSConfig(
		cors.WithAllowedOrigins([]string{"*"}),
		cors.WithAllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "OPTIONS", "HEAD"}),
		cors.WithAllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Ruta pública
	r.NewRoute(router.Route{
		Path:    "/public",
		Method:  router.GET,
		Handler: router.WithCorsMiddleware(corsConfig, publicHandler),
	})

	// Ruta protegida solo para admins
	r.NewRoute(router.Route{
		Path:   "/admin",
		Method: router.GET,
		Handler: router.WithCorsMiddleware(corsConfig,
			router.WithAuthAndRBAC(rbacConfig, []string{"admin"}, adminHandler)),
	})

	// Ruta protegida para usuarios normales
	r.NewRoute(router.Route{
		Path:   "/user",
		Method: router.GET,
		Handler: router.WithCorsMiddleware(corsConfig,
			router.WithAuthAndRBAC(rbacConfig, []string{"user"}, userHandler)),
	})
	r.NewRoute(router.Route{
		Path:   "/prueba",
		Method: router.GET,
		Handler: router.WithCorsMiddleware(corsConfig,
			router.WithAuthMiddleWare(rbacConfig, userHandler)),
	})

	log.Println("Server starting on :8080...")
	r.StartServer(":8080")

}

// func loadPrivateKeyFromFile(keyPem []byte) (*ecdsa.PrivateKey, error) {

// 	block, _ := pem.Decode(keyPem)
// 	if block == nil || block.Type != "EC PRIVATE KEY" {
// 		return nil, fmt.Errorf("failed to decode PEM block containing private key")
// 	}

// 	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return privateKey, nil
// }
// func createJWTTokenFromFile(file *os.File, claims *CustomClaims) (string, error) {
// 	// Read the private key from the file
// 	keyPem, err := os.ReadFile(file.Name())
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read private key file: %w", err)
// 	}

// 	// Load the private key
// 	privateKey, err := loadPrivateKeyFromFile(keyPem)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to load private key: %w", err)
// 	}

// 	// Create a new token with the claims
// 	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)

// 	// Sign the token with the private key
// 	signedToken, err := token.SignedString(privateKey)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to sign token: %w", err)
// 	}

// 	return signedToken, nil
// }
