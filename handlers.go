package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/syeero7/boot-chirpy/internal/auth"
	"github.com/syeero7/boot-chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
	polkaKey       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (cfg *apiConfig) getRequestCount(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Appendf(nil, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) resetServer(w http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	if err := cfg.db.DeleteUsers(req.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	cfg.fileserverHits.Swap(0)
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	params := reqParams{}

	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	userData := database.CreateUserParams{Email: params.Email, HashedPassword: hash}
	user, err := cfg.db.CreateUser(req.Context(), userData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJSON(w, http.StatusCreated, &user)
}

func (cfg *apiConfig) loginUser(w http.ResponseWriter, req *http.Request) {
	type reqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	params := reqParams{}

	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	refreshTokenData := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour * 60),
	}

	if err := cfg.db.CreateRefreshToken(req.Context(), refreshTokenData); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	type ResData struct {
		ID           uuid.UUID `json:"id"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}

	data := ResData{
		ID:           user.ID,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		IsChirpyRed:  user.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusOK, &data)
}

func (cfg *apiConfig) updateUserData(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	type reqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(req.Body)
	params := reqParams{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	userData := database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hash,
		UpdatedAt:      time.Now(),
	}

	user, err := cfg.db.UpdateUser(req.Context(), userData)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	type resData struct {
		ID          uuid.UUID `json:"id"`
		Email       string    `json:"email"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	data := resData{
		ID:          user.ID,
		Email:       user.Email,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		IsChirpyRed: user.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusOK, &data)
}

func (cfg *apiConfig) createRefreshToken(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	user, err := cfg.db.GetUserByRefreshToken(req.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	str, err := auth.MakeJWT(user.ID, cfg.jwtSecret, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	type resData struct {
		Token string `json:"token"`
	}

	data := resData{
		Token: str,
	}

	respondWithJSON(w, http.StatusOK, &data)
}

func (cfg *apiConfig) revokeRefreshToken(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	revokeTokenData := database.RevokeRefreshTokenParams{
		Token:     token,
		RevokedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: time.Now(),
	}

	if err := cfg.db.RevokeRefreshToken(req.Context(), revokeTokenData); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	type reqParams struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := reqParams{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if limit := 140; len(params.Body) > limit {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	chirpData := database.CreateChirpParams{UserID: userID, Body: filterProfanity(params.Body)}
	chirp, err := cfg.db.CreateChirp(req.Context(), chirpData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJSON(w, http.StatusCreated, &chirp)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.db.GetChirps(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJSON(w, http.StatusOK, &chirps)
}

func (cfg *apiConfig) getChirpByID(w http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirp_id"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	chirp, err := cfg.db.GetChirpByID(req.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	respondWithJSON(w, http.StatusOK, &chirp)
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	chirpID, err := uuid.Parse(req.PathValue("chirp_id"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	chirp, err := cfg.db.GetChirpByID(req.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	if err := cfg.db.DeleteChirp(req.Context(), chirpID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) upgradeChirpyMembership(w http.ResponseWriter, req *http.Request) {
	if key, err := auth.GetAPIKey(req.Header); err != nil || key != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	type reqParams struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(req.Body)
	params := reqParams{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	chirpyData := database.SetUserChirpyRedParams{
		ID:          userID,
		IsChirpyRed: true,
	}

	if err := cfg.db.SetUserChirpyRed(req.Context(), chirpyData); err != nil {
		respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getServerReadiness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
