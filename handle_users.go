package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ChernakovEgor/chirpy/internal/auth"
	"github.com/ChernakovEgor/chirpy/internal/database"
	"github.com/google/uuid"
)

type userResponse struct {
	Id           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	JWTtoken     string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
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

	user, err := a.dbQueries.GetUserByEmail(context.Background(), loginRequest.Email)
	if err := auth.CheckPasswordHash(loginRequest.Password, user.HashedPassword); err != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	jwtToken, err := a.registerJWT(user.ID, loginRequest.ExpiresIn)
	if err != nil {
		log.Fatalf("could not create JWT token: %v", err)
	}

	refreshToken, err := a.registerRefreshToken(user.ID)
	if err != nil {
		log.Fatalf("could not create refresh token: %v", err)
	}

	loggedUser := userResponse{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
		jwtToken,
		refreshToken,
	}
	respondWithJSON(w, http.StatusOK, loggedUser)
}

func (a *apiConfig) registerJWT(userID uuid.UUID, expiresIn int) (string, error) {
	var tokenDuration time.Duration
	if expiresIn == 0 || expiresIn > 3600 {
		tokenDuration = time.Hour
	} else {
		tokenDuration = time.Duration(expiresIn) * time.Second
	}

	token, err := auth.MakeJWT(userID, a.jwtSecret, tokenDuration)
	if err != nil {
		return "", fmt.Errorf("could not create token: %v", err)
	}

	return token, nil
}

func (a *apiConfig) registerRefreshToken(userID uuid.UUID) (string, error) {
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		return "", fmt.Errorf("could not create refresh token: %v", err)
	}

	createRefreshTokenParams := database.CreateRefreshTokenParams{Token: refreshToken, UserID: userID}
	_, err = a.dbQueries.CreateRefreshToken(context.Background(), createRefreshTokenParams)
	if err != nil {
		return "", fmt.Errorf("could not insert refresh token: %v", err)
	}

	return refreshToken, nil
}

func (a *apiConfig) handleUpdateCredentials(w http.ResponseWriter, r *http.Request) {
	jwtToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "jwt token expired")
		return
	}

	userID, err := auth.ValidateJWT(jwtToken, a.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid jwt token")
		return
	}

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&body)
	if err != nil {
		log.Fatalf("could not decode body: %v", err)
	}

	hashedPass, err := auth.HashPassword(body.Password)
	if err != nil {
		log.Fatalf("could not hash password: %v", err)
	}

	updateUserParams := database.UpdateEmailAndPasswordParams{ID: userID, Email: body.Email, HashedPassword: hashedPass}
	user, err := a.dbQueries.UpdateEmailAndPassword(context.Background(), updateUserParams)
	if err != nil {
		log.Fatalf("could not update user: %v", err)
	}

	userRes := userResponse{
		Id:        userID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, http.StatusOK, userRes)
}
