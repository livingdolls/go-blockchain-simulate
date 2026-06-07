package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	// Tanpa X-Request-ID header, middleware harus generate UUID baru.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/test", func(c *gin.Context) {
		id, ok := c.Get(GinKeyRequestID)
		assert.True(t, ok, "request_id harus diset di gin.Context")
		c.String(http.StatusOK, id.(string))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// UUID v4 panjang 36 karakter
	assert.Len(t, rec.Body.String(), 36, "request_id harus UUID (36 char)")
	assert.NotEmpty(t, rec.Header().Get(HeaderRequestID), "X-Request-ID harus di response header")
	assert.Equal(t, rec.Body.String(), rec.Header().Get(HeaderRequestID),
		"response body (id) dan response header harus sama")
}

func TestRequestIDMiddleware_AcceptsIncoming(t *testing.T) {
	// Client mengirim X-Request-ID harus dipakai apa adanya (untuk tracing).
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get(GinKeyRequestID)
		c.String(http.StatusOK, id.(string))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(HeaderRequestID, "client-supplied-id-12345")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, "client-supplied-id-12345", rec.Body.String())
	assert.Equal(t, "client-supplied-id-12345", rec.Header().Get(HeaderRequestID))
}

func TestRequestIDMiddleware_PropagatesToRequestContext(t *testing.T) {
	// request_id harus bisa diakses dari context.Context (bukan hanya gin.Context)
	// agar service-layer logger.FromContext bisa mengambilnya.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/test", func(c *gin.Context) {
		id, ok := c.Request.Context().Value(RequestIDKey).(string)
		assert.True(t, ok, "request_id harus ada di context.Context")
		assert.NotEmpty(t, id)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(HeaderRequestID, "test-id-abc")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
