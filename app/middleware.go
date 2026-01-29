package app

import (
	"github.com/gin-gonic/gin"
)

var allowedOrigins = map[string]bool{
	"http://192.168.88.178:3001": true,
	"http://localhost:3001":      true,
	"http://192.168.88.178:3000": true,
	"http://192.168.88.178:3002": true,
}

// CORSMiddleware handles CORS headers for all requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowedOrigins[origin] {
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
