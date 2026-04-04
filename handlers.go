package main

import (
	"encoding/json"
	"net/http"
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
	respondWithJSON(w, 200, map[string]bool{"valid": true})
}
