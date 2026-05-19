package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	minSecretKeySize = 32
)

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTMaker{secretKey: secretKey}, nil
}

func (m *JWTMaker) CreateToken(username string, role string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, role, duration)
	if err != nil {
		return "", payload, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload) // payload is claims
	token, err := jwtToken.SignedString([]byte(m.secretKey))
	return token, payload, err
}

func (m *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// keyFunc
	keyFunc := func(tk *jwt.Token) (any, error) {
		if _, ok := tk.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(m.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrExpiredToken
		case errors.Is(err, jwt.ErrTokenSignatureInvalid), errors.Is(err, jwt.ErrTokenMalformed):
			return nil, ErrInvalidToken
		default:
			return nil, ErrInvalidToken
		}
	}

	if !jwtToken.Valid {
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
