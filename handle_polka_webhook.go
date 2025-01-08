package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ChernakovEgor/chirpy/internal/auth"
	"github.com/google/uuid"
)

type polkaEvent struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (a *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect key")
		return
	}

	if key != a.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "incorrect key")
		return
	}

	var polkaRequest polkaEvent
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&polkaRequest)
	if err != nil {
		log.Fatalf("could not decode body: %v", err)
	}

	eventType := polkaRequest.Event
	if eventType != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// userID, err := uuid.Parse(polkaRequest.Data.UserID)
	_, err = a.dbQueries.UpgradeToRed(context.Background(), polkaRequest.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "user not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
