package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/ppllama/chirpy/internal/auth"
)

type PolkaWebhook struct {
	Event 	string	`json:"event"`
	Data 	struct{
		UserID	string 	`json:"user_id"`
	}
}

func(cfg *apiConfig) handlerUpgradeUser(w http.ResponseWriter, r *http.Request) {

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting api key", err)
		return
	}

	if apiKey != cfg.polka_key {
		respondWithError(w, http.StatusUnauthorized, "Incorrect ApiKey", err)
	}

	decoder := json.NewDecoder(r.Body)

	requestData := PolkaWebhook{}
	if err := decoder.Decode(&requestData); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	if requestData.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	userID, err := uuid.Parse(requestData.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing user id", err)
		return
	}

	if err := cfg.db.UpgradeUser(r.Context(), userID); err != nil {
		if err.Error() == "sql: no rows in result set" {
			respondWithError(w, http.StatusNotFound, "User not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error upgrading user", err)
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}