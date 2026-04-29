package auth

import (
	_ "crypto/rand"
	_ "crypto/rsa"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
)

const (
	totallyPrivateString string = "myPrivateKey"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	// TODO this is a mockup
	privateKey := []byte(totallyPrivateString) //
	// privateKey, err :=rsa.GenerateKey(rand.Reader, 1024)
	//if err != nil {
	//return "", fmt.Errorf("failed to generate key: %w", err)
	//}

	signedString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign jwt: %w", err)
	}

	return signedString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(totallyPrivateString), nil
	})

	//TODO differentiate errors
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("unexpected error parsing jwt: %w", err)
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("unable to retrieve uuid from token: %w", err)
	}
	idAsUuid, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("unable to parse uuid from token: %w", err)
	}

	return idAsUuid, nil
}
