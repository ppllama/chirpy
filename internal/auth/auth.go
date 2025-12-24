package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	claim := jwt.RegisteredClaims{
		Issuer: "chirpy",
		Subject: userID.String(),
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(3600) * time.Second)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("token invalid or expired")
	}

	if claim, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		id, err := uuid.Parse(claim.Subject)
		if err != nil {
			return uuid.Nil, err
		}
		return id, nil
	}
	
	return uuid.Nil, fmt.Errorf("Unknown claims type")
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeaders := headers.Values("Authorization")
	if len(authHeaders) < 1 {
		return "", fmt.Errorf("authorization header not found")
	}

	authHeader := authHeaders[0]
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", fmt.Errorf("authorization header does not contain Bearer token")
	}

	token := strings.TrimSpace(authHeader[len(prefix):])
	if token == "" {
		return "", fmt.Errorf("bearer token is empty")
	}

	return token, nil
}

func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeaders := headers.Values("Authorization")
	if len(authHeaders) < 1 {
		return "", fmt.Errorf("authorization header not found")
	}

	authHeader := authHeaders[0]
	const prefix = "ApiKey "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", fmt.Errorf("authorization header does not contain ApiKey")
	}

	apiKey := strings.TrimSpace(authHeader[len(prefix):])
	if apiKey == "" {
		return "", fmt.Errorf("ApiKey is empty")
	}

	return apiKey, nil
}