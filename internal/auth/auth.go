package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return str, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	if !token.Valid {
		return uuid.UUID{}, errors.New("invalid token")
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	id, err := uuid.Parse(sub)
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	token := strings.Split(headers.Get("Authorization"), " ")
	if len(token) != 2 || len(token[1]) == 0 {
		return "", errors.New("bearer token is missing")
	}

	return token[1], nil
}

func MakeRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	key := strings.Split(headers.Get("Authorization"), " ")
	if len(key) != 2 || len(key) == 0 {
		return "", errors.New("api key is missing")
	}

	return key[1], nil
}
