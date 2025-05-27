package helpers

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJwtToken(claims jwt.Claims, secret *ecdsa.PrivateKey) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
