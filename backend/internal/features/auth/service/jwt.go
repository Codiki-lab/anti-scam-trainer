package service

import (
	"anti-scam-trainer/backend/internal/core/domain"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TokenLifetime = 7 * 24 * time.Hour

type JWTManager struct{ secret []byte }

func NewJWTManager(secret string) (*JWTManager, error) {
	if secret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	return &JWTManager{secret: []byte(secret)}, nil
}

func (m *JWTManager) Issue(user domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":         strconv.Itoa(user.ID),
		"access_role": string(user.AccessRole),
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(TokenLifetime).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *JWTManager) Parse(raw string) (Identity, error) {
	token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return Identity{}, errors.New("invalid access token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Identity{}, errors.New("invalid access token claims")
	}
	if expiresAt, err := claims.GetExpirationTime(); err != nil || expiresAt == nil {
		return Identity{}, errors.New("invalid access token expiration")
	}
	if issuedAt, err := claims.GetIssuedAt(); err != nil || issuedAt == nil {
		return Identity{}, errors.New("invalid access token issue time")
	}
	subject, err := claims.GetSubject()
	if err != nil {
		return Identity{}, errors.New("invalid access token subject")
	}
	userID, err := strconv.Atoi(subject)
	if err != nil || userID < 1 {
		return Identity{}, errors.New("invalid access token subject")
	}
	role, ok := claims["access_role"].(string)
	if !ok || (domain.AccessRole(role) != domain.AccessRoleUser && domain.AccessRole(role) != domain.AccessRoleAdmin) {
		return Identity{}, fmt.Errorf("invalid access token role")
	}
	return Identity{UserID: userID, AccessRole: domain.AccessRole(role)}, nil
}
