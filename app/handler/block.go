package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type BlockHandler struct {
	blockService services.BlockService
}

func NewBlockHandler(blockService services.BlockService) *BlockHandler {
	return &BlockHandler{
		blockService: blockService,
	}
}

func (h *BlockHandler) GenerateBlock(c *gin.Context) {
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
