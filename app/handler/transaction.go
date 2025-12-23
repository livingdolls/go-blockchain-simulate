package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type SendTransactionRequest struct {
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Amount      float64 `json:"amount"`
	PrivateKey  string  `json:"private_key"`
}

type SendTransactionWithSignatureRequest struct {
	FromAddress string  `json:"from_address"`
	ToAddress   string  `json:"to_address"`
	Amount      float64 `json:"amount"`
	Nonce       string  `json:"nonce"`
	Signature   string  `json:"signature"`
}

type BuySellTransactionRequest struct {
	Address   string  `json:"address"`
	Amount    float64 `json:"amount"`
	Nonce     string  `json:"nonce"`
	Signature string  `json:"signature"`
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
	var req SendTransactionWithSignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.transactionService.SendWithSignature(c.Request.Context(), req.FromAddress, req.ToAddress, req.Amount, req.Nonce, req.Signature)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":     "Transaction created successfully",
		"transaction": tx,
		"breakdown": gin.H{
			"amount":             tx.Amount,
			"fee":                tx.Fee,
			"total_cost":         tx.Amount + tx.Fee,
			"recipient_receives": tx.Amount,
		},
		"status": "PENDING",
		"note":   "Transaction will be confirmed when included a block",
	})
}

func (h *TransactionHandler) Buy(c *gin.Context) {
	var req BuySellTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.transactionService.Buy(c.Request.Context(), req.Address, req.Nonce, req.Signature, req.Amount)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":     "Buy transaction created successfully",
		"transaction": tx,
		"breakdown": gin.H{
			"amount":             tx.Amount,
			"fee":                tx.Fee,
			"total_cost":         tx.Amount + tx.Fee,
			"recipient_receives": tx.Amount,
		},
		"status": "PENDING",
		"note":   "Transaction will be confirmed when included a block",
	})
}

func (h *TransactionHandler) Sell(c *gin.Context) {
	var req BuySellTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.transactionService.Sell(c.Request.Context(), req.Address, req.Nonce, req.Signature, req.Amount)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":     "Sell transaction created successfully",
		"transaction": tx,
		"breakdown": gin.H{
			"amount":             tx.Amount,
			"fee":                tx.Fee,
			"total_cost":         tx.Amount + tx.Fee,
			"recipient_receives": tx.Amount,
		},
		"status": "PENDING",
		"note":   "Transaction will be confirmed when included a block",
	})
}

func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	idStr := c.Param("id")

	var id int64
	_, err := fmt.Sscan(idStr, &id)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid transaction ID"})
		return
	}

	tx, err := h.transactionService.GetTransactionByID(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"transaction": tx})
}

func (h *TransactionHandler) GenerateNonce(c *gin.Context) {
	address := c.Param("address")

	if address == "" {
		c.JSON(400, gin.H{"error": "address is required"})
		return
	}

	nonce := h.transactionService.GenerateTransactionNonce(c.Request.Context(), address)

	c.JSON(200, gin.H{
		"nonce": nonce,
	})
}
