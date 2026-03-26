package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AdminClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

type AdminJWTService interface {
	GenerateAdminToken(userID int) (string, error)
	ValidateAdminToken(token string) (*AdminClaims, error)
}

type AdminJWTAdapter struct {
	secret string
	ttl    time.Duration
}

func NewAdminJWTAdapter(secret string, ttl time.Duration) AdminJWTService {
	return &AdminJWTAdapter{
		secret: secret,
		ttl:    ttl,
	}
}

// GenerateAdminToken implements [AdminJWTService].
func (a *AdminJWTAdapter) GenerateAdminToken(userID int) (string, error) {
	claims := &AdminClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(a.secret))

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ValidateAdminToken implements [AdminJWTService].
func (a *AdminJWTAdapter) ValidateAdminToken(token string) (*AdminClaims, error) {
	claims := &AdminClaims{}

	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenMalformed
		}

		return []byte(a.secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		return nil, err
	}

	claims, ok := t.Claims.(*AdminClaims)

	if !ok || !t.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
