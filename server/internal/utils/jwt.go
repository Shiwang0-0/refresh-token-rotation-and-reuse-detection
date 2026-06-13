package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	accessSecret  = []byte(os.Getenv("JWT_ACCESS_SECRET"))
	refreshSecret = []byte(os.Getenv("JWT_REFRESH_SECRET"))

	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")

	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

func GenerateAccessToken(userID int) (string, error) {
	expiry := time.Now().Add(AccessTokenTTL)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expiry.Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(accessSecret)
}

func GenerateRefreshToken(userID int) (string, error) {
	expiry := time.Now().Add(RefreshTokenTTL)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expiry.Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(refreshSecret)
}

// public functions just pass the right secret
func ValidateAccessToken(tokenString string) (int, error) {
	return validateToken(tokenString, accessSecret)
}

func ValidateRefreshToken(tokenString string) (int, error) {
	return validateToken(tokenString, refreshSecret)
}

func validateToken(tokenString string, secret []byte) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrTokenExpired
		}
		return 0, ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, ErrTokenInvalid
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, ErrTokenInvalid
	}

	return int(userID), nil
}
