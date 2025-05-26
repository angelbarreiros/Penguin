package helpers

import (
	"fmt"

	"github.com/angelbarreiros/Penguin/router/auth"
	"github.com/golang-jwt/jwt/v5"
)

func GeneratePlainToken(claims auth.PlainClaims, privateKeyData []byte) (string, error) {

	privateKey, err := auth.LoadPrivateKeyFromFile(privateKeyData)
	if err != nil {
		return "", fmt.Errorf("failed to load private key: %w", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func GenerateRBACToken(claims auth.RBACClaims, privateKeyData []byte) (string, error) {

	privateKey, err := auth.LoadPrivateKeyFromFile(privateKeyData)
	if err != nil {
		return "", fmt.Errorf("failed to load private key: %w", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
