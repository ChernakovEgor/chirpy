package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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
	w.WriteHeader(200)
	a.fileserverHits.Store(0)
}

func main() {
	mux := http.NewServeMux()
	apiCfg := apiConfig{}
	fileserverHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileserverHandler))
	mux.HandleFunc("GET /api/healthz", handleHealthz)
	mux.HandleFunc("POST /api/validate_chirp", handleValidate)
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

func handleValidate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type validateRequest struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	var vRequest validateRequest
	err := decoder.Decode(&vRequest)
	if err != nil {
		log.Printf("could not decode request: %v", err)
	}

	if len(vRequest.Body) > 140 {
		rError := struct {
			Error string `json:"error"`
		}{Error: "Chirp is too long"}
		b, err := json.Marshal(rError)
		if err != nil {
			log.Printf("could not encode error: %v", err)
		}
		w.WriteHeader(400)
		_, err = w.Write(b)
		if err != nil {
			log.Printf("could not write response: %v", err)
		}
	} else {
		rValid := struct {
			Valid bool `json:"valid"`
		}{Valid: true}
		b, err := json.Marshal(rValid)
		if err != nil {
			log.Printf("could not encode valid: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(b)
		if err != nil {
			log.Printf("could not write response: %v", err)
		}
	}
}
