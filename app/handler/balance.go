package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type BalanceHandler struct {
	service services.BalanceService
}

func NewBalanceHandler(service services.BalanceService) *BalanceHandler {
	return &BalanceHandler{service: service}
}

func (h *BalanceHandler) GetBalance(c *gin.Context) {
	address := c.Param("address")

	if address == "" {
		c.JSON(400, gin.H{"error": "address is required"})
		return
	}

	user, err := h.service.GetBalance(address)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": user.Address,
		"balance": user.Balance,
	})
}

func (h *BalanceHandler) GetWalletBalance(c *gin.Context) {
	address := c.Param("address")

	if address == "" {
		c.JSON(400, gin.H{"error": "address is required"})
		return
	}

	walletResponse, err := h.service.GetWalletBalance(address)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wallet": walletResponse,
	})
}
