package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

type AdminLoginHandler struct {
	authService services.AdminAuthService
	jwtService  security.AdminJWTService
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

func NewAdminLoginHandler(authService services.AdminAuthService, jwtService security.AdminJWTService) *AdminLoginHandler {
	return &AdminLoginHandler{
		authService: authService,
		jwtService:  jwtService,
	}
}

func (h *AdminLoginHandler) Login(c *gin.Context) {
	var req AdminLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("Invalid request payload"))
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	admin, err := h.authService.AuthenticateAdmin(ctx, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse[string]("Authentication failed: "+err.Error()))
		return
	}

	token, err := h.jwtService.GenerateAdminToken(admin.ID, admin.Username, admin.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("Failed to generate token: "+err.Error()))
		return
	}

	c.SetCookie("admin_token", token, int(24*time.Hour.Seconds()), "/", "", false, true)

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
	c.SetCookie("admin_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Logged out successfully"))
}
