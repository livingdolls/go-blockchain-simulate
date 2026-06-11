package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/app/worker"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type SendTransactionWithSignatureRequest struct {
	FromAddress string  `json:"from_address" binding:"required,eth_addr"`
	ToAddress   string  `json:"to_address" binding:"required,eth_addr"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Nonce       string  `json:"nonce" binding:"required"`
	Signature   string  `json:"signature" binding:"required,len=132"` // 0x + 130 hex
}

type BuySellTransactionRequest struct {
	Address   string  `json:"address" binding:"required,eth_addr"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Nonce     string  `json:"nonce" binding:"required"`
	Signature string  `json:"signature" binding:"required,len=132"`
}

type TransactionHandler struct {
	transactionService services.TransactionService
	rmqClient          *rabbitmq.Client
	txRepo             repository.TransactionRepository
}

func NewTransactionHandler(transactionService services.TransactionService, rmqClient *rabbitmq.Client, txRepo repository.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		rmqClient:          rmqClient,
		txRepo:             txRepo,
	}
}

func (h *TransactionHandler) Send(c *gin.Context) {
	var req SendTransactionWithSignatureRequest
	if !dto.BindJSON(c, &req) {
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

	body, err := json.Marshal(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to serialize transaction"))
		return
	}

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
	if !dto.BindJSON(c, &req) {
		return
	}

	msg := worker.TransactionMessage{
		Type:      "BUY",
		Address:   req.Address,
		Amount:    req.Amount,
		Nonce:     req.Nonce,
		Signature: req.Signature,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to serialize buy transaction"))
		return
	}

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
	if !dto.BindJSON(c, &req) {
		return
	}

	msg := worker.TransactionMessage{
		Type:      "SELL",
		Address:   req.Address,
		Amount:    req.Amount,
		Nonce:     req.Nonce,
		Signature: req.Signature,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to serialize sell transaction"))
		return
	}

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
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("invalid transaction ID"))
		return
	}

	tx, err := h.transactionService.GetTransactionByID(id)
	if err != nil {
		// Database error atau not found: log error asli di server,
		// return generic message ke client (jangan leak internal error).
		c.JSON(http.StatusNotFound, dto.NewErrorResponse[string]("transaction not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": tx})
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

// EstimateFee menghitung estimasi fee berdasarkan:
// - Base fee (tier amount-based)
// - Network congestion (jumlah pending tx)
// - Priority level (low/medium/high)
//
// Endpoint: GET /transaction/fee/estimate?amount=100&priority=medium
func (h *TransactionHandler) EstimateFee(c *gin.Context) {
	amountStr := c.Query("amount")
	priority := c.DefaultQuery("priority", "low")

	var amount float64
	if _, err := fmt.Sscan(amountStr, &amount); err != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("invalid amount"))
		return
	}

	// Hitung jumlah pending tx untuk congestion multiplier
	pendingCount := 0
	if h.txRepo != nil {
		txs, err := h.txRepo.GetPendingTransactions(1000)
		if err == nil {
			pendingCount = len(txs)
		}
	}

	priorityMultiplier := utils.EstimatePriorityMultiplier(priority)
	estimatedFee := utils.EstimateFee(amount, pendingCount, priorityMultiplier)
	congestionLevel := utils.CongestionLevel(pendingCount)
	congestionPct := utils.CongestionPercentage(pendingCount)
	congestionMultiplier := utils.CalculateCongestionMultiplier(pendingCount)
	baseFee := utils.CalculateTransactionFee(amount)

	c.JSON(http.StatusOK, dto.NewSuccessResponse(gin.H{
		"base_fee":             baseFee,
		"congestion_multiplier": congestionMultiplier,
		"priority_multiplier":  priorityMultiplier,
		"estimated_fee":        estimatedFee,
		"pending_count":        pendingCount,
		"congestion_level":     congestionLevel,
		"congestion_percent":   congestionPct,
	}))
}

// GetPendingTransactions mengembalikan daftar transaksi pending (mempool).
// Endpoint: GET /transactions/pending?limit=50
//
// Data ini menunjukkan transaksi yang sudah dikirim oleh user tapi belum
// dikonfirmasi di block. Berguna untuk:
// - Monitoring congestion (berapa banyak tx menunggu)
// - Fee estimation (tx dengan fee tinggi diproses duluan)
// - Debugging (tx stuck karena nonce salah, dll)
func (h *TransactionHandler) GetPendingTransactions(c *gin.Context) {
	limit := 50
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 200 {
		limit = l
	}

	txs, err := h.txRepo.GetPendingTransactions(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to get pending transactions"))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(gin.H{
		"transactions": txs,
		"count":        len(txs),
		"limit":        limit,
	}))
}
