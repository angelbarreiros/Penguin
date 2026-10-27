package helpers

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/angelbarreiros/Penguin/router/auth"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateJwtToken signs claims with the ES* method matching the key's curve
// (P-256 → ES256, P-384 → ES384, P-521 → ES512).
func GenerateJwtToken(claims jwt.Claims, secret *ecdsa.PrivateKey) (string, error) {
	if secret == nil {
		return "", fmt.Errorf("failed to sign token: private key is nil")
	}
	method, err := auth.SigningMethodForKey(&secret.PublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	token := jwt.NewWithClaims(method, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
