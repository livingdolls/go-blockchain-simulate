package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	Address string `json:"address"`
	jwt.RegisteredClaims
}

type JWTService interface {
	GenerateToken(address string) (string, error)
	ValidateToken(token string) (*JWTClaims, error)
}

type JWTAdapter struct {
	secret string
	ttl    time.Duration
}

func NewJWTAdapter(secret string, ttl time.Duration) JWTService {
	return &JWTAdapter{
		secret: secret,
		ttl:    ttl,
	}
}

// GenerateToken implements services.JWTService.
func (j *JWTAdapter) GenerateToken(address string) (string, error) {
	claims := &JWTClaims{
		Address: address,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ValidateToken implements services.JWTService.
func (j *JWTAdapter) ValidateToken(token string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenMalformed
		}
		return []byte(j.secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, err
	}

	claims, ok := t.Claims.(*JWTClaims)

	if !ok || !t.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
