package utils

import (
	"errors"
	"time"

	"github.com/campus-project-hub/api/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID    uuid.UUID `json:"userId"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TokenType TokenType `json:"tokenType"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

func GenerateTokenPair(userID uuid.UUID, email, role string) (*TokenPair, error) {
	cfg := config.GetConfig()

	// Access Token
	accessExpiry := time.Now().Add(time.Duration(cfg.JWT.ExpiryHours) * time.Hour)
	accessClaims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.App.Name,
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return nil, err
	}

	// Refresh Token
	refreshExpiry := time.Now().Add(time.Duration(cfg.JWT.RefreshExpiryHours) * time.Hour)
	refreshClaims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.App.Name,
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(cfg.JWT.ExpiryHours * 3600),
	}, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.GetConfig()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}
