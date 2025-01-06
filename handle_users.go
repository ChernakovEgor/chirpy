package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (a *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var emailStruct struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&emailStruct)
	if err != nil {
		log.Fatalf("could not decode body: %v", err)
	}

	email := emailStruct.Email
	user, err := a.dbQueries.CreateUser(context.Background(), email)
	if err != nil {
		log.Fatalf("could not create user: %v", err)
	}

	responseStruct := struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusCreated, responseStruct)
}
