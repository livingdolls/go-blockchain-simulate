package app

import (
	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/config"
)

// DefaultAllowedOrigins adalah fallback origin yang dipakai saat CORSOrigins
// pada konfigurasi kosong. Tetap disediakan untuk development lokal.
var DefaultAllowedOrigins = map[string]bool{
	"http://localhost:3000": true,
	"http://localhost:3001": true,
	"http://localhost:3002": true,
}

// CORSMiddleware mengatur header CORS untuk semua request.
// Daftar origin yang diizinkan dibaca dari konfigurasi server.
// Jika konfigurasi kosong, fallback ke DefaultAllowedOrigins.
func CORSMiddleware(cfg *config.ServerConfig) gin.HandlerFunc {
	allowed := make(map[string]bool, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		allowed[o] = true
	}
	if len(allowed) == 0 {
		allowed = DefaultAllowedOrigins
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
