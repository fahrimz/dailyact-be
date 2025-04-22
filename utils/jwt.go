package utils

import (
	"dailyact/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID uint        `json:"user_id"`
	Email  string      `json:"email"`
	Role   models.Role `json:"role"`
	jwt.StandardClaims
}

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(user models.User) (string, error) {
	// Set expiration time (24 hours from now)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "dailyact-api",
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates and parses a JWT token
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
