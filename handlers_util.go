package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

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
