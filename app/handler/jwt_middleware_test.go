package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/livingdolls/go-blockchain-simulate/security"
	"github.com/stretchr/testify/assert"
)

// fakeJWTService mengimplement security.JWTService untuk test.
// GenerateToken/ValidateToken deterministik dengan secret statis.
type fakeJWTService struct {
	secret []byte
}

func (f *fakeJWTService) GenerateToken(address string) (string, error) {
	claims := &security.JWTClaims{
		Address: address,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(f.secret)
}

func (f *fakeJWTService) ValidateToken(tokenStr string) (*security.JWTClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &security.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return f.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*security.JWTClaims)
	if !ok || !parsed.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

var _ security.JWTService = (*fakeJWTService)(nil)

func newJWTMiddlewareTestRouter(jwtSvc security.JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTMiddleware(jwtSvc, nil)) // nil memory = skip blacklist check
	r.GET("/profile", func(c *gin.Context) {
		claims, _ := GetUserClaims(c)
		c.JSON(http.StatusOK, gin.H{"address": claims.Address})
	})
	return r
}

func TestJWTMiddleware_HappyPath(t *testing.T) {
	jwtSvc := &fakeJWTService{secret: []byte("test-secret-32-chars-aaaaaaaaaaa")}
	r := newJWTMiddlewareTestRouter(jwtSvc)

	token, _ := jwtSvc.GenerateToken("0xABC123")
	req := httptest.NewRequest("GET", "/profile", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	jwtSvc := &fakeJWTService{secret: []byte("test-secret-32-chars-aaaaaaaaaaa")}
	r := newJWTMiddlewareTestRouter(jwtSvc)

	req := httptest.NewRequest("GET", "/profile", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	jwtSvc := &fakeJWTService{secret: []byte("test-secret-32-chars-aaaaaaaaaaa")}
	r := newJWTMiddlewareTestRouter(jwtSvc)

	req := httptest.NewRequest("GET", "/profile", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "garbage.token.value"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTMiddleware_TokenWithEmptyAddress(t *testing.T) {
	// Token valid (signed benar) tapi claims.Address kosong. Middleware
	// harus reject karena endpoint self-only butuh address valid.
	jwtSvc := &fakeJWTService{secret: []byte("test-secret-32-chars-aaaaaaaaaaa")}
	r := newJWTMiddlewareTestRouter(jwtSvc)

	token, _ := jwtSvc.GenerateToken("") // empty address
	req := httptest.NewRequest("GET", "/profile", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code,
		"token dengan address kosong harus ditolak")
}

func TestJWTMiddleware_EmptyTokenString(t *testing.T) {
	// Cookie ada tapi value whitespace saja.
	jwtSvc := &fakeJWTService{secret: []byte("test-secret-32-chars-aaaaaaaaaaa")}
	r := newJWTMiddlewareTestRouter(jwtSvc)

	req := httptest.NewRequest("GET", "/profile", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "   "})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
