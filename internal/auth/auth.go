package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", fmt.Errorf("could not hash password: %v", err)
	}
	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	issued := time.Now()
	expires := issued.Add(expiresIn)
	claims := jwt.RegisteredClaims{Issuer: "chirpy",
		IssuedAt:  &jwt.NumericDate{issued},
		ExpiresAt: &jwt.NumericDate{expires},
		Subject:   string(userID.String())}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	res, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return res, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var claims struct {
		jwt.RegisteredClaims
	}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not parse token: %v", err)
	}

	idString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not get subject from token: %v", err)
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not get issuer from token: %v", err)
	}
	if issuer != "chirpy" {
		return uuid.Nil, fmt.Errorf("invalid issuer: %v", issuer)
	}

	userID, err := uuid.Parse(idString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not parse uuid: %v", err)
	}
	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	value := headers.Get("Authorization")
	if value == "" {
		return "", fmt.Errorf("no header 'Authorization'")
	}

	token := strings.TrimPrefix(value, "Bearer ")
	return token, nil
}

func MakeRefreshToken() (string, error) {
	data := make([]byte, 32)
	_, err := rand.Read(data)
	if err != nil {
		return "", fmt.Errorf("could not generate random string data: %v", err)
	}

	token := hex.EncodeToString(data)
	return token, nil
}
