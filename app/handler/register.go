package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type RegisterHandler struct {
	service services.RegisterService
}

func NewRegisterHandler(service services.RegisterService) *RegisterHandler {
	return &RegisterHandler{service: service}
}

func (h *RegisterHandler) Register(c *gin.Context) {
	var req models.UserRegister
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("invalid request body"))
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
		return
	}

	resp := &models.UserRegisterResponse{
		Address:  user.Address,
		Username: user.Username,
	}

	c.SetCookie("auth_token", user.Token, int(24*time.Hour.Seconds()), "/", "", false, true)

	c.JSON(200, dto.NewSuccessResponse(resp))
}

func (h *RegisterHandler) Challenge(c *gin.Context) {
	address := c.Param("address")

	challenge, err := h.service.Challenge(c.Request.Context(), address)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"challenge": challenge})
}

func (h *RegisterHandler) Verify(c *gin.Context) {
	var req struct {
		Address   string `json:"address"`
		Signature string `json:"signature"`
		Nonce     string `json:"nonce"`
		Username  string `json:"username"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	valid, err := h.service.Verify(c.Request.Context(), req.Address, req.Nonce, req.Signature, req.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("auth_token", valid, int(24*time.Hour.Seconds()), "/", "", false, true)

	c.JSON(200, gin.H{"valid": true})
}
