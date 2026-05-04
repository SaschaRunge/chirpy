package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
)

// TODO: untested
func GetAPIKey(headers http.Header) (string, error) {
	apiKey := headers.Get("Authorization")
	apiKey, found := strings.CutPrefix(apiKey, "ApiKey")
	if !found || apiKey == "" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	apiKey = strings.Trim(apiKey, " ")

	return apiKey, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	bearer := headers.Get("Authorization")
	bearer, found := strings.CutPrefix(bearer, "Bearer")
	if !found || bearer == "" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	bearer = strings.Trim(bearer, " ")

	return bearer, nil
}

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
