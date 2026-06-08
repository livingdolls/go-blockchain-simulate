package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxBodySize_AllowsSmallBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(MaxBodySizeMiddleware(100)) // 100 byte limit
	r.POST("/test", func(c *gin.Context) {
		// Use raw read, not ShouldBindJSON, untuk test murni middleware
		buf := make([]byte, 200)
		n, err := c.Request.Body.Read(buf)
		if err != nil {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"read": n})
	})

	body := bytes.Repeat([]byte("a"), 50) // 50 byte, di bawah limit
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "body 50 byte harus lolos limit 100 byte")
}

func TestMaxBodySize_RejectsOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(MaxBodySizeMiddleware(100)) // 100 byte limit
	r.POST("/test", func(c *gin.Context) {
		buf := make([]byte, 200)
		_, err := c.Request.Body.Read(buf)
		if err != nil {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	body := bytes.Repeat([]byte("a"), 200) // 200 byte, di atas limit 100
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// MaxBytesReader akan return error dari Read; handler kami return 413.
	// Yang penting: handler TIDAK boleh memproses 200 byte (yang berbahaya).
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code,
		"body 200 byte harus ditolak oleh limit 100 byte")
}

func TestMaxBodySize_ExactlyAtLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(MaxBodySizeMiddleware(100))
	r.POST("/test", func(c *gin.Context) {
		buf := make([]byte, 200)
		n, err := c.Request.Body.Read(buf)
		if err != nil {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"read": n})
	})

	body := bytes.Repeat([]byte("a"), 100) // exactly 100
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// MaxBytesReader allow exactly maxBytes; Read return 100 + nil error.
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMaxBodySize_NoBody(t *testing.T) {
	// Request tanpa body harus tetap lolos (GET, atau POST tanpa body).
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(MaxBodySizeMiddleware(100))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
