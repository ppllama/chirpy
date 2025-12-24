package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ppllama/chirpy/internal/auth"
	"github.com/ppllama/chirpy/internal/database"
)

type parameters struct {
		Email 				string	`json:"email"`
		Password 			string 	`json:"password"`
	}

type User struct {
		ID        		uuid.UUID 	`json:"id"`
		CreatedAt 		time.Time 	`json:"created_at"`
		UpdatedAt 		time.Time 	`json:"updated_at"`
		Email     		string    	`json:"email"`
		Token			string		`json:"token"`
		RefreshToken 	string		`json:"refresh_token"`
		IsChirpyRed		bool		`json:"is_chirpy_red"`
	}


func(cfg *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	
	params, err := getEmailPassword(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	createUserParams := database.CreateUserParams{
		Email: params.Email,
		HashedPassword: hashedPassword,
	}

	newUser, err := cfg.db.CreateUser(r.Context(), createUserParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, User{
		ID: newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email: newUser.Email,
		IsChirpyRed: newUser.IsChirpyRed,
	})
}

func(cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	
	params, err := getEmailPassword(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user", err)
		return
	}

	ok, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error verifying user", err)
	}

	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret)
	refreshTokenCandidate, err := auth.MakeRefreshToken()

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token: refreshTokenCandidate,
		UserID: user.ID,
	}

	refreshToken, err := cfg.db.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		Token: token,
		RefreshToken: refreshToken.Token,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func(cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorised", err)
		return
	}

	refreshToken, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorised", err)
		return
	}

	newAccessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating new access token", err)
		return
	}

	type AccessToken struct{
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, AccessToken{
		Token: newAccessToken,
	})
}

func getEmailPassword(r *http.Request) (parameters, error) {

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		return parameters{}, err
	}

	return params, nil
}

func(cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorised", err)
		return
	}

	err = cfg.db.UpdateRevoke(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error revoking token", err)
		return
	}
	respondWithJSON(w, 204, nil)
}

func(cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorised", err)
		return
	}

	UserID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorised", err)
		return
	}
	
	params, err := getEmailPassword(r)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	updateUserParams := database.UpdateUserParams{
		ID: UserID,
		Email: params.Email,
		HashedPassword: hashedPassword,
	}

	editedUser, err := cfg.db.UpdateUser(r.Context(), updateUserParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}
	
	respondWithJSON(w, http.StatusOK, User{
		ID: editedUser.ID,
		CreatedAt: editedUser.CreatedAt,
		UpdatedAt: editedUser.UpdatedAt,
		Email: editedUser.Email,
		IsChirpyRed: editedUser.IsChirpyRed,
	})
}