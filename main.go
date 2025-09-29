package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	mux := http.NewServeMux()
	config := apiConfig{}

	mux.Handle("/app/", http.StripPrefix("/app/", config.middlewareMetricsInc(http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", getServerReadiness)
	mux.HandleFunc("GET /api/metrics", config.getRequestCount)
	mux.HandleFunc("POST /api/reset", config.resetRequestCount)

	server := &http.Server{Addr: ":8080", Handler: mux}
	log.Fatal(server.ListenAndServe())
}

func getServerReadiness(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (cfg *apiConfig) getRequestCount(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Appendf(nil, "Hits: %d", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) resetRequestCount(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.Swap(0)
	w.WriteHeader(http.StatusOK)
}
