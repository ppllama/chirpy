package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/ppllama/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
	secret string
	polka_key string
}

func main() {
	const filePathRoot = "."
	const port = ":8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	JWT_Secret := os.Getenv("JWT_SECRET")
	POLKA_KEY := os.Getenv("POLKA_KEY")

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	dbQueries := database.New(dbConn)
	cfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db: dbQueries,
		platform: platform,
		secret: JWT_Secret,
		polka_key: POLKA_KEY,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)
	mux.HandleFunc("POST /api/chirps", cfg.handlerPostChirps)
	mux.HandleFunc("GET /api/chirps", cfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirp_id}", cfg.handlerChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirp_id}", cfg.handlerDeleteChirp)
	mux.HandleFunc("POST /api/users", cfg.handlerUsers)
	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handlerRevoke)
	mux.HandleFunc("PUT /api/users", cfg.handlerUpdateUser)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.handlerUpgradeUser)
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))

	server := &http.Server{
		Addr: port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

