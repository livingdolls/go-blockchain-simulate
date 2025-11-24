package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type RewardHandler struct {
	svr services.RewardService
	svb services.BlockService
}

func NewRewardHandler(svr services.RewardService, svb services.BlockService) *RewardHandler {
	return &RewardHandler{
		svr: svr,
		svb: svb,
	}
}

// get reward info
func (h *RewardHandler) GetRewardInfo(c *gin.Context) {
	info, err := h.svr.RewardInfo()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"reward_info": info})
}

// get next block reward info
func (h *RewardHandler) GetBlockReward(c *gin.Context) {
	blockNumberStr := c.Param("number")
	blockNumber, err := strconv.ParseInt(blockNumberStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid block number"})
		return
	}

	lastBlock, err := h.svb.GetBlockByBlockNumber(blockNumber)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	response := models.BlockRewardResponse{
		BlockNumber:  int64(lastBlock.BlockNumber),
		MinerAddress: lastBlock.MinerAddress,
		BlockReward:  lastBlock.BlockReward,
		TotalFees:    lastBlock.TotalFees,
		TotalEarned:  lastBlock.BlockReward + lastBlock.TotalFees,
		Timestamp:    lastBlock.Timestamp,
	}

	c.JSON(http.StatusOK, response)
}

func (h *RewardHandler) GetRewardSchedule(c *gin.Context) {
	blockStr := c.Query("blocks")
	blockCount, err := strconv.Atoi(blockStr)
	if err != nil || blockCount <= 0 {
		blockCount = 10 // default to 10 blocks
	}

	schedule, err := h.svr.GetRewardSchedule(blockCount)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"data": schedule})
}
