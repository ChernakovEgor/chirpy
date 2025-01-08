package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/ChernakovEgor/chirpy/internal/auth"
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
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Fatalf("could not get token from headers: %v", err)
	}

	userID, err := auth.ValidateJWT(token, a.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid JWT token")
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("could not read body: %v", err)
	}

	err = json.Unmarshal(b, &chirpRequest)
	if err != nil {
		log.Fatalf("could not unmarshal request: %v", err)
	}

	chirpParams := database.CreateChirpParams{Body: chirpRequest.Body, UserID: userID}
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
	authorIdString := r.URL.Query().Get("author_id")
	var chirps []database.Chirp
	var err error

	if authorIdString == "" {
		chirps, err = a.dbQueries.GetAllChirps(context.Background())
		if err != nil {
			log.Fatalf("could not get chirps: %v", err)
		}
	} else {
		authorID, err := uuid.Parse(authorIdString)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "invalid id")
			return
		}
		chirps, err = a.dbQueries.GetChirpByAuthor(context.Background(), authorID)
	}

	sortOrder := r.URL.Query().Get("sort")
	if sortOrder == "desc" {
		slices.Reverse(chirps)
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

func (a *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token invalid")
		return
	}

	userID, err := auth.ValidateJWT(token, a.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token invalid")
		return
	}

	chirpIdString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIdString)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "incorrect chirp id")
		return
	}

	// get chirp info
	chirp, err := a.dbQueries.GetChirpByID(context.Background(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "incorrect chirp id")
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "incorrect chirp id")
		return
	}

	// delete chirp
	deleteChirpParams := database.DeleteChirpParams{UserID: userID, ID: chirpID}
	_, err = a.dbQueries.DeleteChirp(context.Background(), deleteChirpParams)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}

	w.WriteHeader(204)
}
