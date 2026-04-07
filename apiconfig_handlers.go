package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/SaschaRunge/chirpy/internal/database"

	"github.com/google/uuid"
)

type user struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func userFrom(u database.User) user {
	return user{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
	}
}

type apiConfig struct {
	dbQueries      *database.Queries
	fileServerHits atomic.Int32
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	expectedJSON, err := decodeJSON(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		fmt.Printf("Internal Server Error: %s", err)
		return
	}

	dbUser, err := cfg.dbQueries.CreateUser(r.Context(), expectedJSON.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error: Unable to create user")
		fmt.Printf("Internal Server Error: Unable to create user: %s", err)
		return
	}

	taggedUser := userFrom(dbUser)
	respondWithJSON(w, 201, taggedUser)
}

func (cfg *apiConfig) handlerReturnFileServerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	s := fmt.Sprintf(
		`<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`,
		cfg.fileServerHits.Load())
	//s := fmt.Sprintf("Hits: %d", cfg.fileServerHits.Load())
	w.Write([]byte(s))
}

func (cfg *apiConfig) handlerResetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
	cfg.dbQueries.ResetUsers(r.Context())
	w.Write([]byte("Reset users db."))
}
