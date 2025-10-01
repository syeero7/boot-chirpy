package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/syeero7/boot-chirpy/internal/database"
)

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	config := apiConfig{
		db:        database.New(db),
		platform:  platform,
		jwtSecret: jwtSecret,
	}

	mux.Handle("/app/", http.StripPrefix("/app/", config.middlewareMetricsInc(http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", getServerReadiness)
	mux.HandleFunc("GET /admin/metrics", config.getRequestCount)
	mux.HandleFunc("POST /admin/reset", config.resetServer)
	mux.HandleFunc("POST /api/users", config.createUser)
	mux.HandleFunc("POST /api/chirps", config.createChirp)
	mux.HandleFunc("GET /api/chirps", config.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirp_id}", config.getChirpByID)
	mux.HandleFunc("POST /api/login", config.loginUser)

	server := &http.Server{Addr: ":8080", Handler: mux}
	log.Fatal(server.ListenAndServe())
}
