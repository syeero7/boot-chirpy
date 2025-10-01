package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("password1")
	if err != nil || len(hash) == 0 {
		t.Error("failed to hash password")
	}
}

func TestCheckHash(t *testing.T) {
	password := "password1"
	hash, _ := HashPassword(password)
	match, err := CheckPasswordHash(password, hash)
	if err != nil || !match {
		t.Errorf("password mismatch: expected '%s', got '%s'", password, hash)
	}
}

func TestMakeJWT(t *testing.T) {
	token, err := MakeJWT(uuid.New(), "secret", 5*time.Minute)
	if err != nil || len(token) == 0 {
		t.Error("failed to make jwt token")
	}
}

func TestValidateJWT(t *testing.T) {
	secret := "secret"
	userID := uuid.New()
	token, _ := MakeJWT(userID, secret, 5*time.Minute)
	id, err := ValidateJWT(token, secret)
	if err != nil || id.String() != userID.String() {
		t.Error("failed to validate jwt")
	}

	if _, err := ValidateJWT(token, "wrong"); err == nil {
		t.Error("wrong secret should error")
	}

	expired, _ := MakeJWT(userID, secret, 4*time.Second)
	time.Sleep(5 * time.Second)
	if _, err := ValidateJWT(expired, secret); err == nil {
		t.Error("expired token should error")
	}
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer my_token")
	str, err := GetBearerToken(headers)
	if err != nil || str != "my_token" {
		t.Error("failed to retrieve bearer token")
	}
}
