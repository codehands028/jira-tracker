package utils

import (
	"errors"
	"time"

	"jira-tracker-backend/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    uint
	Phone     string
	Role      string
	SessionID string
	IP        string
	UserAgent string
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, phone, role, sessionID, ip, userAgent string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Phone:     phone,
		Role:      role,
		SessionID: sessionID,
		IP:        ip,
		UserAgent: userAgent,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.GlobalConfig.JWT.Expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GlobalConfig.JWT.Secret))
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(config.GlobalConfig.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
