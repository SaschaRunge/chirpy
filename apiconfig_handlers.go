package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/SaschaRunge/chirpy/internal/auth"
	"github.com/SaschaRunge/chirpy/internal/database"

	"github.com/google/uuid"
)

const (
	accessTokenExpirationTime  = time.Hour
	refreshTokenExpirationTime = time.Hour * 24 * 60 // 60 days

)

type user struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func userFrom(u database.User) user {
	return user{
		ID:          u.ID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Email:       u.Email,
		IsChirpyRed: u.IsChirpyRed,
	}
}

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func chirpFrom(c database.Chirp) chirp {
	return chirp{
		ID:        c.ID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body:      c.Body,
		UserID:    c.UserID,
	}
}

type apiConfig struct {
	dbQueries      *database.Queries
	fileServerHits atomic.Int32
	platform       string
	tokenSecret    string
	polkaKey       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerChangePassword(w http.ResponseWriter, r *http.Request) {
	type changePasswordRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	body, err := decodeJSON[changePasswordRequest](r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	id, err := validateUser(r, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	user, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:       id,
		Email:    body.Email,
		Password: hash,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	respondWithJSON(w, http.StatusOK, userFrom(user))
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type createChirpRequest struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	expectedJSON, err := decodeJSON[createChirpRequest](r)
	if err != nil {
		respondWithError(w, 500, "Internal Server Error")
		return
	}
	if len(expectedJSON.Body) > 140 {
		respondWithError(w, 400, "chirp is too long")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	createChirpParams := database.CreateChirpParams{
		Body:   expectedJSON.Body,
		UserID: id,
	}

	newChirp, err := cfg.dbQueries.CreateChirp(r.Context(), createChirpParams)
	if err != nil {
		respondWithError(w, 500, "unable to create chirp")
		return
	}
	respondWithJSON(w, http.StatusCreated, chirpFrom(newChirp))
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type createUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	expectedJSON, err := decodeJSON[createUserRequest](r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		fmt.Printf("Internal Server Error: %s", err)
		return
	}

	hash, err := auth.HashPassword(expectedJSON.Password)
	dbUser, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:    expectedJSON.Email,
		Password: hash,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error: Unable to create user")
		fmt.Printf("Internal Server Error: Unable to create user: %s", err)
		return
	}

	taggedUser := userFrom(dbUser)
	respondWithJSON(w, 201, taggedUser)
}

func (cfg *apiConfig) handlerDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	userID, err := validateUser(r, cfg.tokenSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirp_id"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error: Unable to parse UUID")
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Not Found")
		return
	}
	if userID != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}

	if err = cfg.dbQueries.DeleteChirpByID(r.Context(), chirpID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error: Unable to delete chirp.")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirp_id"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error: Unable to parse UUID")
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp Not Found")
		return
	}

	respondWithJSON(w, 200, chirpFrom(chirp))
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	//r.URL.Query() .

	jsonReadableChirps := []chirp{}
	for _, c := range chirps {
		jsonReadableChirps = append(jsonReadableChirps, chirpFrom(c))
	}
	respondWithJSON(w, http.StatusOK, jsonReadableChirps)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	expectedJSON, err := decodeJSON[loginRequest](r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error: Unable to decode JSON request.")
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), expectedJSON.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	hash := user.Password
	authorized, err := auth.CheckPasswordHash(expectedJSON.Password, hash)
	if !authorized || err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userResponse := userFrom(user)
	userResponse.Token, err = auth.MakeJWT(user.ID, cfg.tokenSecret, accessTokenExpirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error: Unable to generate JWT.")
		return
	}
	userResponse.RefreshToken = auth.MakeRefreshToken()

	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     userResponse.RefreshToken,
		UserID:    userResponse.ID,
		ExpiresAt: time.Now().Add(refreshTokenExpirationTime),
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error: Unable to store refresh token.")
		return
	}

	respondWithJSON(w, http.StatusOK, userResponse)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	errs := []error{}
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errs = append(errs, err)
	}

	tokenFromDB, err := cfg.dbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		errs = append(errs, err)
	}

	if tokenFromDB.ExpiresAt.Before(time.Now()) || tokenFromDB.RevokedAt.Valid || len(errs) > 0 {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	accessToken, err := auth.MakeJWT(tokenFromDB.UserID, cfg.tokenSecret, accessTokenExpirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error: Unable to generate JWT.")
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Bad Request: No valid refresh token in header.")
		return
	}
	err = cfg.dbQueries.RevokeToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Not Found: No matching refresh token.")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
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

func (cfg *apiConfig) handlerUpgradeUser(w http.ResponseWriter, r *http.Request) {
	type requestJSON struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	polkaKey, err := auth.GetAPIKey(r.Header)

	if err != nil || polkaKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	body, err := decodeJSON[requestJSON](r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error: Unable to generate JWT.")
		return
	}

	if body.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	if err = cfg.dbQueries.UpgradeUser(r.Context(), body.Data.UserID); err != nil {
		respondWithError(w, http.StatusNotFound, "Not Found")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
