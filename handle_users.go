package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ChernakovEgor/chirpy/internal/auth"
	"github.com/ChernakovEgor/chirpy/internal/database"
	"github.com/google/uuid"
)

type userResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (a *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var userRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&userRequest)
	if err != nil {
		log.Fatalf("could not decode body: %v", err)
	}

	email := userRequest.Email
	hashedPassword, err := auth.HashPassword(userRequest.Password)
	if err != nil {
		log.Fatalf("could not hash password: %v", err)
	}

	userParams := database.CreateUserParams{Email: email, HashedPassword: hashedPassword}
	user, err := a.dbQueries.CreateUser(context.Background(), userParams)
	if err != nil {
		log.Fatalf("could not create user: %v", err)
	}

	responseStruct := userResponse{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusCreated, responseStruct)
}

func (a *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		ExpiresIn int    `json:"expires_in_seconds,omitempty"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&loginRequest)
	if err != nil {
		log.Fatalf("could not decode request: %v", err)
	}

	var tokenDuration time.Duration
	if loginRequest.ExpiresIn == 0 || loginRequest.ExpiresIn > 3600 {
		tokenDuration = time.Hour
	} else {
		tokenDuration = time.Duration(loginRequest.ExpiresIn) * time.Second
	}

	user, err := a.dbQueries.GetUserByEmail(context.Background(), loginRequest.Email)
	if err := auth.CheckPasswordHash(loginRequest.Password, user.HashedPassword); err != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	loggedUser := userResponse{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
	}
	respondWithJSON(w, http.StatusOK, loggedUser)
}
