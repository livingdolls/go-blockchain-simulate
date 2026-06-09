package utils

import (
	"github.com/gin-gonic/gin"
)

// SetupSSEHeaders sets the necessary headers for Server-Sent Events (SSE).
// Access-Control-Allow-Origin TIDAK di-set di sini; rely pada global
// CORSMiddleware yang sudah menangani origin validation dengan benar.
// Sebelumnya ada 'Access-Control-Allow-Origin: *' yang bypass CORS
// middleware → website manapun bisa baca SSE stream.
func SetupSSEHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
}
