package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"presupuesto-rapido/backend/internal/domain"
)

type TokenClaims struct {
	UserID string      `json:"uid"`
	Email  string      `json:"email"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

func CreateAccessToken(user domain.SessionUser, secret string, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("JWT_SECRET is required")
	}
	now := time.Now().UTC()
	claims := TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ParseAccessToken(tokenString, secret string) (domain.SessionUser, error) {
	if secret == "" {
		return domain.SessionUser{}, errors.New("JWT_SECRET is required")
	}
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return domain.SessionUser{}, err
	}
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return domain.SessionUser{}, errors.New("invalid token")
	}
	return domain.SessionUser{ID: claims.UserID, Email: claims.Email, Role: claims.Role}, nil
}

func NewRefreshToken() (plain string, hash []byte, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, err
	}
	plain = base64.RawURLEncoding.EncodeToString(raw)
	h := sha256.Sum256([]byte(plain))
	return plain, h[:], nil
}

func HashRefreshToken(plain string) []byte {
	h := sha256.Sum256([]byte(plain))
	return h[:]
}
