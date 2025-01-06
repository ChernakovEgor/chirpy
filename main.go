package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ChernakovEgor/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      database.Queries
	platform       string
}

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits.Add(1)
		log.Println(a.fileserverHits.Load())
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}

func (a *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	metricsPage := `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`
	io.WriteString(w, fmt.Sprintf(metricsPage, a.fileserverHits.Load()))
}

func (a *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if a.platform != "dev" {
		respondWithError(w, 403, "")
	}
	err := a.dbQueries.ResetUsers(context.Background())
	if err != nil {
		log.Fatalf("could not reset users: %v", err)
	}
	w.WriteHeader(200)
	a.fileserverHits.Store(0)
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("could not connect to db: %v", err)
	}

	dbQueries := database.New(db)
	mux := http.NewServeMux()
	apiCfg := apiConfig{dbQueries: *dbQueries, platform: platform}
	fileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileserverHandler))
	mux.HandleFunc("GET /api/healthz", handleHealthz)
	mux.HandleFunc("POST /api/validate_chirp", handleValidate)
	mux.HandleFunc("POST /api/users", apiCfg.handleUsers)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	server := http.Server{Addr: ":8080", Handler: mux}

	log.Fatalln(server.ListenAndServe())
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
