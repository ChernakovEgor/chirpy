package main

import (
	"context"
	"net/http"

	"github.com/ChernakovEgor/chirpy/internal/auth"
)

func (a *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect token")
		return
	}

	_, err = a.dbQueries.RevokeToken(context.Background(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token expired")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
