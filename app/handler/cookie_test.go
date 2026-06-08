package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAuthCookie_SameSiteAndHttpOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		setAuthCookie(c, "auth_token", "test-jwt-value", 3600)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Cari cookie yang di-set
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1, "harus ada 1 cookie")

	c := cookies[0]
	assert.Equal(t, "auth_token", c.Name)
	assert.Equal(t, "test-jwt-value", c.Value)
	assert.Equal(t, "/", c.Path)
	assert.Equal(t, 3600, c.MaxAge)
	assert.True(t, c.HttpOnly, "HttpOnly wajib true untuk JWT cookie")
	assert.Equal(t, http.SameSiteStrictMode, c.SameSite,
		"SameSite harus Strict untuk mitigasi CSRF")
	assert.False(t, c.Secure, "dev mode: Secure=false (production via reverse proxy)")
}

func TestSetAuthCookie_DeleteWithNegativeMaxAge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		setAuthCookie(c, "auth_token", "", -3600)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "", cookies[0].Value, "nilai cookie harus kosong untuk delete")
	// Go's http library normalizes negative MaxAge ke -1 untuk
	// Set-Cookie header. Yang penting: nilai < 0 (browser akan delete).
	assert.Less(t, cookies[0].MaxAge, 0, "MaxAge harus negatif untuk delete")
	assert.Equal(t, http.SameSiteStrictMode, cookies[0].SameSite,
		"SameSite juga harus strict untuk cookie delete (preventive)")
}
