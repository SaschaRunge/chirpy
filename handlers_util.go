package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SaschaRunge/chirpy/internal/auth"
	"github.com/google/uuid"
)

func validateUser(r *http.Request, tokenSecret string) (uuid.UUID, error) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return uuid.UUID{}, err
	}

	id, err := auth.ValidateJWT(token, tokenSecret)
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

func decodeJSON[T any](r *http.Request) (T, error) {
	decoder := json.NewDecoder(r.Body)
	var expectedJSON T
	err := decoder.Decode(&expectedJSON)
	if err != nil {
		var zero T
		return zero, err
	}
	return expectedJSON, nil
}

// TODO: make generic

func filterText(text string) string {
	const censored = "****"
	badWords := [3]string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(text, " ")
	for i, word := range words {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				words[i] = censored
			}
		}
	}
	return strings.Join(words, " ")
}
