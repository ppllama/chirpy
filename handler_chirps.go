package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ppllama/chirpy/internal/auth"
	"github.com/ppllama/chirpy/internal/database"
)

type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time	`json:"created_at"`
		UpdatedAt time.Time	`json:"updated_at"`
		Body      string	`json:"body"`
		UserID    uuid.UUID	`json:"user_id"`
	}

func(cfg *apiConfig) handlerPostChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type responseCleaned struct {
		Cleaned_body string `json:"cleaned_body"`
	}

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

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleaned := badWordReplace(params.Body)

	chirpParams := database.CreateChirpParams{
		Body: cleaned,
		UserID: UserID,
	}

	newChirp, err := cfg.db.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not create Chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID: newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body: newChirp.Body,
		UserID: newChirp.UserID,
	})
}

func(cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {

	authorIDStr := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")

	var chirps []database.Chirp
	var err error

	if authorIDStr == "" {
		chirps, err = cfg.db.GetAllChirps(r.Context())
	} else {
		authorID, err := uuid.Parse(authorIDStr)
		if err != nil {
			chirps, err = cfg.db.GetAllChirps(r.Context())
		}
		chirps, err = cfg.db.GetAllChirpsByUser(r.Context(), authorID)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting all chirps", err)
		return
	}

	if sortOrder == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt)})
	} else {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)})
	}
	

	allChirps := []Chirp{}
	for _, chirp := range(chirps) {
		allChirps = append(allChirps, Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, allChirps)
}

func(cfg *apiConfig) handlerChirp(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("chirp_id")
	if requestID == "" {
		respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
		return
	}
	id, err := uuid.Parse(requestID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}
	responseChirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not get chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID: responseChirp.ID,
		CreatedAt: responseChirp.CreatedAt,
		UpdatedAt: responseChirp.UpdatedAt,
		Body: responseChirp.Body,
		UserID: responseChirp.UserID,
	})
}

func(cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {

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

	requestID := r.PathValue("chirp_id")
	if requestID == "" {
		respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
		return
	}
	id, err := uuid.Parse(requestID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	responseChirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Could not get chirp", err)
		return
	}

	if responseChirp.UserID != UserID {
		respondWithError(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), responseChirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting chirp", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func badWordReplace(input string) string {

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert": {},
		"fornax": {},
	}

	splitInput := strings.Split(input, " ")
	
	for i, word := range(splitInput) {
		if _, ok := badWords[strings.ToLower(word)]; ok {
			splitInput[i] = "****"
		}
	}


	return strings.Join(splitInput, " ")
}