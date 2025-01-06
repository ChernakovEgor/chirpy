package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateJWT(t *testing.T) {
	want, _ := uuid.NewRandom()
	t.Logf("Generated id is: %v", want)
	token, err := MakeJWT(want, "secret", time.Hour)
	if err != nil {
		t.Errorf("could not make JWT: %v", err)
	}
	t.Logf("Token: %v", token)

	got, err := ValidateJWT(token, "secret")
	if err != nil {
		t.Errorf("could not validate token: %v", err)
	}

	t.Logf("Got id: %v", got)
	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}
