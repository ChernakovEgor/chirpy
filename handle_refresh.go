package main

import (
	"context"
	"log"
	"net/http"

	"github.com/ChernakovEgor/chirpy/internal/auth"
)

func (a *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token does not exist")
		return
	}

	userID, err := a.dbQueries.GetUserByToken(context.Background(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token expired")
		return
	}

	jwtToken, err := a.registerJWT(userID, 3600)
	if err != nil {
		log.Fatalf("could not create JWT token: %v", err)
	}

	response := struct {
		Token string `json:"token"`
	}{Token: jwtToken}

	respondWithJSON(w, http.StatusOK, response)
}
