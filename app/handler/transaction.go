package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/app/worker"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
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
	rmqClient          *rabbitmq.Client
}

func NewTransactionHandler(transactionService services.TransactionService, rmqClient *rabbitmq.Client) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		rmqClient:          rmqClient,
	}
}

func (h *TransactionHandler) Send(c *gin.Context) {
	var req SendTransactionWithSignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	msg := worker.TransactionMessage{
		Type:      "SEND",
		Address:   req.FromAddress,
		ToAddress: req.ToAddress,
		Amount:    req.Amount,
		Nonce:     req.Nonce,
		Signature: req.Signature,
	}

	body, _ := json.Marshal(msg)

	if err := h.rmqClient.Publish(
		c.Request.Context(),
		rabbitmq.TransactionExchange,
		rabbitmq.TransactionSubmittedKey,
		body,
	); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to publish transaction message"))
		return
	}

	c.JSON(202, dto.NewSuccessResponse(map[string]interface{}{
		"message": "Transaction submitted successfully and is being processed",
	}))
}

func (h *TransactionHandler) Buy(c *gin.Context) {
	var req BuySellTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	msg := worker.TransactionMessage{
		Type:      "BUY",
		Address:   req.Address,
		Amount:    req.Amount,
		Nonce:     req.Nonce,
		Signature: req.Signature,
	}

	body, _ := json.Marshal(msg)

	if err := h.rmqClient.Publish(
		c.Request.Context(),
		rabbitmq.TransactionExchange,
		rabbitmq.TransactionSubmittedKey,
		body,
	); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to publish buy transaction message"))
		return
	}

	c.JSON(202, dto.NewSuccessResponse(map[string]interface{}{
		"message": "Buy transaction submitted successfully and is being processed",
	}))
}

func (h *TransactionHandler) Sell(c *gin.Context) {
	var req BuySellTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	msg := worker.TransactionMessage{
		Type:      "SELL",
		Address:   req.Address,
		Amount:    req.Amount,
		Nonce:     req.Nonce,
		Signature: req.Signature,
	}

	body, _ := json.Marshal(msg)

	if err := h.rmqClient.Publish(
		c.Request.Context(),
		rabbitmq.TransactionExchange,
		rabbitmq.TransactionSubmittedKey,
		body,
	); err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to publish sell transaction message"))
		return
	}

	c.JSON(202, dto.NewSuccessResponse(map[string]interface{}{
		"message": "Sell transaction submitted successfully and is being processed",
	}))
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
