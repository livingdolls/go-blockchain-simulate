package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// HeaderRequestID adalah nama header HTTP yang digunakan untuk
	// request ID. Disimpan di request dan response header.
	HeaderRequestID = "X-Request-ID"
	// GinKeyRequestID adalah key yang dipakai untuk menyimpan request ID
	// di dalam gin.Context (bukan context.Context) untuk akses handler-side.
	GinKeyRequestID = "request_id"
)

// RequestIDMiddleware membuat atau membaca request ID dari header X-Request-ID,
// lalu menyebarkannya ke:
//   - context.Context request (via stdlib http.Request)
//   - gin.Context (untuk akses handler-side)
//   - response header X-Request-ID (untuk tracing client-side)
//
// Middleware ini TIDAK melakukan logging; gunakan RequestLogMiddleware
// jika juga ingin log request start/finish.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(HeaderRequestID)
		if reqID == "" {
			reqID = uuid.NewString()
		}

		// store in gin context for handler access
		c.Set(GinKeyRequestID, reqID)

		// store in request context (context.Context) for service/logger access
		ctx := ContextWithRequestID(c.Request.Context(), reqID)
		c.Request = c.Request.WithContext(ctx)

		// expose to client for tracing
		c.Writer.Header().Set(HeaderRequestID, reqID)

		c.Next()
	}
}

// RequestLogMiddleware mencatat log untuk setiap request HTTP yang masuk
// dan keluar, lengkap dengan method, path, status, latency, dan request ID.
// Harus dipasang SETELAH RequestIDMiddleware agar request_id tersedia
// di context.
func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// Log setelah handler selesai; FromContext otomatis menyertakan request_id.
		FromContext(c.Request.Context()).Info("http_request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}
