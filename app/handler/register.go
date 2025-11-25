package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type RegisterRequest struct {
	Name string `json:"name" binding:"required"`
}

type RegisterHandler struct {
	service services.RegisterService
}

func NewRegisterHandler(service services.RegisterService) *RegisterHandler {
	return &RegisterHandler{service: service}
}

func (h *RegisterHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Register(req.Name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	resp := &models.UserRegisterResponse{
		Address:    user.Address,
		PublicKey:  user.PublicKey,
		PrivateKey: user.PrivateKey,
		Mnemonic:   user.Mnemonic,
		Username:   user.Username,
	}

	c.JSON(200, resp)
}
