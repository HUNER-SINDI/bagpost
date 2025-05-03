package store

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

// CustomClaims defines the structure for the JWT claims
type CustomClaims struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
	jwt.RegisteredClaims
}

// CustomRefreshClaims defines the structure for the refresh token claims
type CustomRefreshClaims struct {
	ID int `json:"id"`
	jwt.RegisteredClaims
}

var secretKey = []byte(os.Getenv("Stores_JWT_SECRET"))
var refreshSecretKey = []byte(os.Getenv("Stores_REFRESH_SECRET"))

// GenerateToken generates a JWT token for a warehouse user
func GenerateToken(id int, email string, isActive bool) (string, error) {
	claims := CustomClaims{
		ID:       id,
		Email:    email,
		IsActive: isActive,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ParseToken validates the JWT token and returns the claims
func ParseToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("could not parse claims")
	}

	// Return claims as map[string]interface{} for flexibility in middleware
	return map[string]interface{}{
		"id":        claims.ID,
		"email":     claims.Email,
		"is_active": claims.IsActive,
	}, nil
}

// GenerateRefreshToken generates a refresh token with a longer expiry
func GenerateRefreshToken(id int) (string, error) {
	claims := CustomRefreshClaims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecretKey)
}

// ParseRefreshToken validates a refresh token and returns the warehouse ID
func ParseRefreshToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomRefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		return refreshSecretKey, nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*CustomRefreshClaims)
	if !ok {
		return 0, errors.New("could not parse refresh token claims")
	}

	return claims.ID, nil
}
