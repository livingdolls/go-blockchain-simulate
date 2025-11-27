package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

func JWTMiddleware(jwtService security.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("auth_token")
		if err != nil || strings.TrimSpace(token) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    http.StatusUnauthorized,
				"error":   "Unauthorized: missing or invalid token",
			})
			return
		}

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    http.StatusUnauthorized,
				"error":   err.Error(),
			})
			return
		}

		// Set user info ke context
		c.Set("user", claims)
		c.Next()
	}
}

func GetUserClaims(c *gin.Context) (*security.JWTClaims, bool) {
	claims, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*security.JWTClaims)
	if !ok {
		return nil, false
	}

	return userClaims, true
}
