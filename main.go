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
}

func main() {
	const filePathRoot = "."
	const port = ":8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	dbQueries := database.New(dbConn)
	cfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db: dbQueries,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.getMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))

	server := &http.Server{
		Addr: port,
		Handler: mux,
	}
	
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

