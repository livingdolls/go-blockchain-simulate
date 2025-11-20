package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type SendTransactionRequest struct {
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Amount      float64 `json:"amount"`
	PrivateKey  string  `json:"private_key"`
}

type TransactionHandler struct {
	transactionService services.TransactionService
}

func NewTransactionHandler(transactionService services.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

func (h *TransactionHandler) Send(c *gin.Context) {
	var req SendTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.transactionService.Send(req.FromAddress, req.ToAddress, req.PrivateKey, req.Amount)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, tx)
}
