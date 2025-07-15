package auth

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JWT struct {
	UserID uint      `json:"uid" validate:"required"`
	Email  string    `json:"email" validate:"required"`
	Exp    time.Time `json:"exp" validate:"required"`
	AppID  uint      `json:"app_id" validate:"required"`
}

type Claims struct {
	Email string `json:"email"`
	AppID uint   `json:"app_id"`
	jwt.RegisteredClaims
}

func ParseJWT(tokenStr string, pubKey ed25519.PublicKey) (*JWT, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodEdDSA.Alg() {
			return nil, fmt.Errorf("unexpected signing method=%v", token.Header["alg"])
		}
		return pubKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	uid, err := parseSubjectToUint(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("invalid subject: %w", err)
	}

	if claims.ExpiresAt == nil {
		return nil, errors.New("missing exp")
	}

	return &JWT{
		UserID: uid,
		Email:  claims.Email,
		Exp:    claims.ExpiresAt.Time,
		AppID:  claims.AppID,
	}, nil

}

func ParseJWTUnverified(tokenStr string) (*JWT, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodEdDSA.Alg() {
			return nil, fmt.Errorf("unexpected signing method=%v", token.Header["alg"])
		}
		return nil, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	uid, err := parseSubjectToUint(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("invalid subject: %w", err)
	}

	if claims.ExpiresAt == nil {
		return nil, errors.New("missing exp")
	}

	return &JWT{
		UserID: uid,
		Email:  claims.Email,
		Exp:    claims.ExpiresAt.Time,
		AppID:  claims.AppID,
	}, nil
}

func parseSubjectToUint(sub string) (uint, error) {
	var uid uint64
	_, err := fmt.Sscanf(sub, "%d", &uid)
	if err != nil {
		return 0, err
	}
	return uint(uid), nil
}
