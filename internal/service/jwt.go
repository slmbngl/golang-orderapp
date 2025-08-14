package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret-key")

// define token durations
const (
	AccessTokenDuration  = 15 * time.Minute   // 15 dakika
	RefreshTokenDuration = 7 * 24 * time.Hour // 7 gün
)

// make Access Token (15 minutes)
func GenerateAccessToken(userID int, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"type":    "access", // Token type
		"exp":     time.Now().Add(AccessTokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// make Refresh Token (7 days)
func GenerateRefreshToken() (string, error) {
	// create 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateJWT(userID int, role string) (string, error) {
	return GenerateAccessToken(userID, role)
}

// parse JWT token
func ParseJWT(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
}

// Access token's validation
func ValidateAccessToken(tokenString string) (int, string, error) {
	token, err := ParseJWT(tokenString)
	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check token type
		if tokenType, exists := claims["type"]; !exists || tokenType != "access" {
			return 0, "", errors.New("invalid token type")
		}

		userID := int(claims["user_id"].(float64))
		role := claims["role"].(string)
		return userID, role, nil
	}

	return 0, "", errors.New("invalid token")
}

func ValidateJWT(tokenString string) (uint, error) {
	userID, _, err := ValidateAccessToken(tokenString)
	if err != nil {
		return 0, err
	}
	return uint(userID), nil
}

func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
