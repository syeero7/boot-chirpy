package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/syeero7/boot-chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
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
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := reqParams{}

	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	respondWithJSON(w, http.StatusCreated, &user)
}

func getServerReadiness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func validateChirp(w http.ResponseWriter, req *http.Request) {
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

	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	body := returnVals{CleanedBody: replaceProfane(params.Body)}
	respondWithJSON(w, http.StatusOK, &body)
}
