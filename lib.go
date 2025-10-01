package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorRes struct {
		Error string `json:"error"`
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	returnErr := errorRes{Error: msg}
	if data, err := json.Marshal(&returnErr); err == nil {
		w.Write(data)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func filterProfanity(s string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Fields(s)

	for i, word := range words {
		if slices.Contains(badWords, strings.ToLower(word)) {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

func isLessThanHour(seconds int) (time.Duration, bool) {
	duration := time.Duration(seconds) * time.Second
	return duration, duration.Hours() < 1.00
}
