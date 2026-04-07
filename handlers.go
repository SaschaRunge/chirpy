package main

import (
	"net/http"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handlerJsonResponse(w http.ResponseWriter, r *http.Request) {
	expectedJSON, err := decodeJSON(r)
	if err != nil {
		respondWithError(w, 500, "Internal Server Error")
		return
	}
	if len(expectedJSON.Text) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	respondWithJSON(w, 200, map[string]any{
		"valid":        true,
		"cleaned_body": filterText(expectedJSON.Text)})
}
