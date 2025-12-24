package auth

import (
	"testing"
	"time"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	secret := "fufufuufufuff9"
	userID1, _ := uuid.NewUUID()
	userID2, _ := uuid.NewUUID()

	// Create tokens
	JWT1, err := MakeJWT(userID1, secret)
	if err != nil {
		t.Fatalf("failed to make JWT1: %v", err)
	}

	JWT2, err := MakeJWT(userID2, secret)
	if err != nil {
		t.Fatalf("failed to make JWT2: %v", err)
	}


	tests := []struct {
		name        string
		token       string
		secret      string
		wantErr     bool
		expectedID  uuid.UUID
		sleepBefore time.Duration
	}{
		{
			name:       "Valid token",
			token:      JWT1,
			secret:     secret,
			wantErr:    false,
			expectedID: userID1,
		},
		{
			name:        "Expired token",
			token:       JWT2,
			secret:      secret,
			wantErr:     true,
			expectedID:  uuid.Nil,
			sleepBefore: 10 * time.Hour, // wait for token to expire
		},
		{
			name:       "Invalid token string",
			token:      "invalid.token.string",
			secret:     secret,
			wantErr:    true,
			expectedID: uuid.Nil,
		},
		{
			name:       "Wrong secret",
			token:      JWT1,
			secret:     "wrongsecret",
			wantErr:    true,
			expectedID: uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.sleepBefore > 0 {
				time.Sleep(tt.sleepBefore)
			}

			id, err := ValidateJWT(tt.token, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("[%s]ValidateJWT() error = %v, wantErr %v",tt.name, err, tt.wantErr)
			}
			if !tt.wantErr && id != tt.expectedID {
				t.Errorf("[%s]ValidateJWT() returned id = %v, expected %v",tt.name, id, tt.expectedID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		headers     http.Header
		wantToken   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "Valid Bearer token",
			headers:   http.Header{"Authorization": {"Bearer abc123"}},
			wantToken: "abc123",
			wantErr:   false,
		},
		{
			name:        "Missing Authorization header",
			headers:     http.Header{},
			wantToken:   "",
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "Authorization header without Bearer",
			headers:     http.Header{"Authorization": {"Basic abc123"}},
			wantToken:   "",
			wantErr:     true,
			errContains: "does not contain Bearer",
		},
		{
			name:        "Empty Bearer token",
			headers:     http.Header{"Authorization": {"Bearer "}},
			wantToken:   "",
			wantErr:     true,
			errContains: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.headers)

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error = %v, got err = %v", tt.wantErr, err)
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error to contain %q, got %v", tt.errContains, err)
			}

			if token != tt.wantToken {
				t.Errorf("expected token = %q, got %q", tt.wantToken, token)
			}
		})
	}
}
