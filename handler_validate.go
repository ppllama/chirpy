package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type responseCleaned struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
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

	respondWithJSON(w, http.StatusOK, responseCleaned{
		Cleaned_body: cleaned,
	})
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