package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

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
		respondWithError(w, 400, "Chirp is too long")
		return
	} else {
		msg := filterProfane(vRequest.Body)
		filtered := struct {
			CleanedBody string `json:"cleaned_body"`
		}{CleanedBody: msg}
		respondWithJSON(w, http.StatusOK, filtered)
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	body := struct {
		Error string `json:"error"`
	}{Error: msg}
	w.WriteHeader(code)
	response, err := json.Marshal(body)
	if err != nil {
		log.Fatalf("could not marshal error: %v", err)
	}

	_, err = w.Write(response)
	if err != nil {
		log.Fatalf("could not write error: %v", err)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("could not marshal response: %v", err)
	}

	w.WriteHeader(code)

	_, err = w.Write(response)
	if err != nil {
		log.Fatalf("could not write response: %v", err)
	}
}

func filterProfane(msg string) string {
	banned := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Fields(msg)
	for i, word := range words {
		for _, bannedWord := range banned {
			if strings.ToLower(word) == bannedWord {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}
