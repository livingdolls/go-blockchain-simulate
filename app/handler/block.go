package handler

import (
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
