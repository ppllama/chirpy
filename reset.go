package main

import (
	"fmt"
	"net/http"
)

func(cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {

	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Request is not using dev platform", fmt.Errorf("Request is not using dev platform"))
	}
	cfg.fileserverHits.Store(0)

	if err := cfg.db.DeleteAllUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "error resetting users database", err)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Successfully reset"))
}