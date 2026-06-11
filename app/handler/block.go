package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type BlockHandler struct {
	blockService services.BlockService
	walletRepo   repository.UserWalletRepository
}

func NewBlockHandler(blockService services.BlockService, walletRepo repository.UserWalletRepository) *BlockHandler {
	return &BlockHandler{
		blockService: blockService,
		walletRepo:   walletRepo,
	}
}

func (h *BlockHandler) GenerateBlock(c *gin.Context) {
	// Use retry logic to handle lock timeouts
	block, err := h.blockService.GenerateBlock()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"block": block})
}

func (h *BlockHandler) GetBlocks(c *gin.Context) {

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	if limitStr == "" {
		limitStr = "10"
	}

	if offsetStr == "" {
		offsetStr = "0"
	}

	var limit, offset int
	_, err := fmt.Sscan(limitStr, &limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	_, err = fmt.Sscan(offsetStr, &offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
		return
	}

	blocks, err := h.blockService.GetBlocks(limit, offset)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"blocks": blocks})
}

func (h *BlockHandler) GetBlockByID(c *gin.Context) {
	idParam := c.Param("id")

	var id int64
	_, err := fmt.Sscan(idParam, &id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid block ID"})
		return
	}

	block, err := h.blockService.GetBlockByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"block": block})
}

func (h *BlockHandler) CheckBlockchainIntegrity(c *gin.Context) {
	err := h.blockService.CheckBlockchainIntegrity()

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "invalid", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "valid"})
}

func (h *BlockHandler) GetBlockByBlockNumber(c *gin.Context) {
	numberParam := c.Param("number")

	var blockNumber int64
	_, err := fmt.Sscan(numberParam, &blockNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid block number"})
		return
	}

	block, err := h.blockService.GetDetailsByBlockNumber(blockNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"block": block})
}

func (h *BlockHandler) GetTransactionsByBlockNumber(c *gin.Context) {
	numberParam := c.Param("number")

	var blockNumber int64
	_, err := fmt.Sscan(numberParam, &blockNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("invalid block number"))
		return
	}

	ctx := c.Request.Context()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	txs, err := h.blockService.GetTransactionByBlockNumber(ctx, blockNumber)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to retrieve transactions for block"))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(map[string]interface{}{
		"transactions": txs,
	}))
}

func (h *BlockHandler) SearchBlocksByHash(c *gin.Context) {
	hash := c.Query("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("hash query parameter is required"))
		return
	}

	ctx := c.Request.Context()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	blocks, err := h.blockService.SearchBlocksByHash(ctx, hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to search blocks by hash"))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(blocks))
}

func (h *BlockHandler) GetBlocksInRange(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("from and to query parameters are required"))
		return
	}

	var from, to int64
	_, err := fmt.Sscan(fromStr, &from)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("invalid from block number"))
		return
	}

	_, err = fmt.Sscan(toStr, &to)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("invalid to block number"))
		return
	}

	if from > to {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("from block number must be less than or equal to to block number"))
		return
	}

	// Range size limit: tanpa limit, request ?from=1&to=999999999
	// → DB query unbounded → high latency + memory consumption.
	if to-from > 1000 {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("range terlalu besar, maksimum 1000 blocks"))
		return
	}

	ctx := c.Request.Context()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	blocks, err := h.blockService.GetBlocksInRange(ctx, from, to)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to retrieve blocks in range"))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(blocks))
}

func (h *BlockHandler) GetBlockStats(c *gin.Context) {
	ctx := c.Request.Context()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	stats, err := h.blockService.GetBlockStats(ctx)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to retrieve block stats"))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(stats))
}

func (h *BlockHandler) SearchBlocksByMinerAddress(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("address query parameter is required"))
		return
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	if limitStr == "" {
		limitStr = "10"
	}

	if offsetStr == "" {
		offsetStr = "0"
	}

	var limit, offset int
	_, err := fmt.Sscan(limitStr, &limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("invalid limit"))
		return
	}

	_, err = fmt.Sscan(offsetStr, &offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("invalid offset"))
		return
	}

	ctx := c.Request.Context()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	blocks, err := h.blockService.SearchBlocksByMinerAddress(ctx, address, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to search blocks by miner address"))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(blocks))
}

// GetRichList mengembalikan top holders berdasarkan YTE balance.
// Endpoint: GET /explorer/richlist?limit=100
func (h *BlockHandler) GetRichList(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	wallets, err := h.walletRepo.GetTopBalances(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string]("failed to get rich list"))
		return
	}

	// Format response: address + balance + rank
	type richEntry struct {
		Rank       int     `json:"rank"`
		Address    string  `json:"address"`
		YTEBalance float64 `json:"yte_balance"`
	}

	result := make([]richEntry, len(wallets))
	for i, w := range wallets {
		result[i] = richEntry{
			Rank:       i + 1,
			Address:    w.UserAddress,
			YTEBalance: w.YTEBalance,
		}
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(gin.H{
		"holders": result,
		"total":   len(result),
		"limit":   limit,
	}))
}
