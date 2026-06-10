package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

type AdminLoginHandler struct {
	authService services.AdminAuthService
	jwtService  security.AdminJWTService
	memory      redis.MemoryAdapter
}

type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminLoginResponse struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Token    string `json:"token"`
}

func NewAdminLoginHandler(authService services.AdminAuthService, jwtService security.AdminJWTService, memory redis.MemoryAdapter) *AdminLoginHandler {
	return &AdminLoginHandler{
		authService: authService,
		jwtService:  jwtService,
		memory:      memory,
	}
}

func (h *AdminLoginHandler) Login(c *gin.Context) {
	var req AdminLoginRequest

	if !dto.BindJSON(c, &req) {
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	admin, err := h.authService.AuthenticateAdmin(ctx, req.Username, req.Password)
	if err != nil {
		// Generic message: jangan leak apakah username ada atau password
		// salah (mencegah username enumeration). Log error asli di server.
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Invalid username or password"))
		return
	}

	token, err := h.jwtService.GenerateAdminToken(admin.ID, admin.Username, admin.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("Failed to generate token: "+err.Error()))
		return
	}

	setAuthCookie(c, "admin_token", token, 24*3600)

	resp := AdminLoginResponse{
		ID:       admin.ID,
		UserID:   admin.UserID,
		Username: admin.Username,
		Role:     admin.Role,
		Token:    token,
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(resp))
}

func (h *AdminLoginHandler) Logout(c *gin.Context) {
	token, _ := c.Cookie("admin_token")
	blacklistToken(c, token, h.memory)
	setAuthCookie(c, "admin_token", "", -3600)
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Logged out successfully"))
}
