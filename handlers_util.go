package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type requestContent struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`

	Email    string `json:"email"`
	Password string `json:"password"`
}

// TODO: make generic
func decodeJSON(r *http.Request) (requestContent, error) {
	decoder := json.NewDecoder(r.Body)
	expectedJSON := requestContent{}
	err := decoder.Decode(&expectedJSON)
	if err != nil {
		return requestContent{}, err
	}
	return expectedJSON, nil
}

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
