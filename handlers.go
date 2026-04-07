package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handlerJsonResponse(w http.ResponseWriter, r *http.Request) {
	type expectedJSON struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	expJSON := expectedJSON{}
	err := decoder.Decode(&expJSON)
	if err != nil {
		respondWithError(w, 500, "Internal Server Error")
		return
	}
	if len(expJSON.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	respondWithJSON(w, 200, map[string]any{
		"valid":        true,
		"cleaned_body": filterText(expJSON.Body)})
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
