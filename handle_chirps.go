package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ChernakovEgor/chirpy/internal/database"
	"github.com/google/uuid"
)

type chirpEntry struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (a *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	var chirpRequest struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("could not read body: %v", err)
	}

	err = json.Unmarshal(b, &chirpRequest)
	if err != nil {
		log.Fatalf("could not unmarshal request: %v", err)
	}

	chirpParams := database.CreateChirpParams{Body: chirpRequest.Body, UserID: chirpRequest.UserId}
	chirp, err := a.dbQueries.CreateChirp(context.Background(), chirpParams)
	if err != nil {
		log.Fatalf("could not create chirp: %v", err)
	}

	chirpResponse := chirpEntry{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, chirpResponse)
}

func (a *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := a.dbQueries.GetAllChirps(context.Background())
	if err != nil {
		log.Fatalf("could not get chirps: %v", err)
	}

	var chirpResponse []chirpEntry
	for _, chirp := range chirps {
		chirpResponse = append(chirpResponse, chirpEntry{chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.Body, chirp.UserID})
	}

	respondWithJSON(w, http.StatusOK, chirpResponse)
}

func (a *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	log.Print(chirpID)
	log.Print(uuid.Validate(chirpID))
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, 404, "Incorrect UUID string")
		return
	}
	chirp, err := a.dbQueries.GetChirpByID(context.Background(), chirpUUID)
	if err != nil {
		respondWithError(w, 404, "Chirp not found")
		return
	}

	chirpResponse := chirpEntry{
		chirp.ID,
		chirp.CreatedAt,
		chirp.UpdatedAt,
		chirp.Body,
		chirp.UserID,
	}
	respondWithJSON(w, http.StatusOK, chirpResponse)
}
