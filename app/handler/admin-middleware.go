package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

func AdminMiddleware(jwtService security.AdminJWTService, adminRepo repository.AdminRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// extract token from cookie
		token, err := c.Cookie("admin_token")

		if err != nil || token == strings.TrimSpace(token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: missing token"))
			return
		}

		// validate token
		claims, err := jwtService.ValidateAdminToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: invalid token"))
			return
		}

		ctx := c.Request.Context()

		// check if user is admin
		admin, err := adminRepo.GetAdminByUserID(ctx, claims.UserID)
		if err != nil {
			if err.Error() == "admin not found" {
				c.AbortWithStatusJSON(http.StatusForbidden, dto.NewErrorResponse[string]("Forbidden: not an admin"))
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("Internal Server Error"))
			}

			return
		}

		// check admin status
		if admin.Status != "active" {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.NewErrorResponse[string]("Forbidden: admin account is not active"))
			return
		}

		// set admin info to context
		c.Set("admin", admin)
		c.Set("user", claims)

		// update last login time
		_ = adminRepo.UpdateLastLogin(ctx, admin.ID)

		c.Next()
	}
}

// AdminWithPermissionMiddleware checks if admin has required permission
func AdminWithPermissionMiddleware(jwtService security.AdminJWTService, adminRepo repository.AdminRepository, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// extract token from cookie
		token := getTokenFromCookie(c)

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: missing token"))
			return
		}

		// validate token
		claims, err := jwtService.ValidateAdminToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Unauthorized: invalid token"))
			return
		}

		// check if user is admin
		ctx := c.Request.Context()
		admin, err := adminRepo.GetAdminByUserID(ctx, claims.UserID)
		if err != nil {
			if err.Error() == "admin not found" {
				c.AbortWithStatusJSON(http.StatusForbidden, dto.NewErrorResponse[string]("Forbidden: not an admin"))
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("Internal Server Error"))
			}
			return
		}

		// check admin status
		if admin.Status != "active" {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.NewErrorResponse[string]("Forbidden: admin account is not active"))
			return
		}

		// check permission
		if !hasPermission(admin, requiredPermission) {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.NewErrorResponse[string]("Forbidden: insufficient permissions"))
			return
		}

		// set admin info to context
		c.Set("admin", admin)
		c.Set("user", claims)

		// update last login time
		_ = adminRepo.UpdateLastLogin(ctx, admin.ID)

		c.Next()
	}
}

func hasPermission(admin *models.AdminWithUser, requiredPermission string) bool {
	// admin with * role has all permissions

	if admin.Role == "admin" {
		return true
	}

	// parse permissions from JSON
	if !admin.Permissions.Valid {
		return false
	}

	var permissions []string
	err := json.Unmarshal([]byte(admin.Permissions.String), &permissions)
	if err != nil {
		return false
	}

	// check if required permission is in the list
	for _, perm := range permissions {
		if perm == "*" || perm == requiredPermission {
			return true
		}
	}

	return false
}

// GetAdminFromContext extracts admin info from gin context
func GetAdminFromContext(c *gin.Context) (*models.AdminWithUser, error) {
	adminVal, exists := c.Get("admin")
	if !exists {
		return nil, fmt.Errorf("admin not found in context")
	}

	admin, ok := adminVal.(*models.AdminWithUser)
	if !ok {
		return nil, fmt.Errorf("invalid admin context")
	}

	return admin, nil
}

// getTokenFromCookie extracts admin token from cookie
func getTokenFromCookie(c *gin.Context) string {
	token, err := c.Cookie("admin_token")
	if err == nil && strings.TrimSpace(token) != "" {
		return token
	}
	return ""
}
