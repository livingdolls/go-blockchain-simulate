package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type StakingHandler struct {
	stakingService services.StakingService
}

func NewStakingHandler(stakingService services.StakingService) *StakingHandler {
	return &StakingHandler{stakingService: stakingService}
}

// StakeRequest untuk POST /staking/stake
type StakeRequest struct {
	Address     string  `json:"address" binding:"required,eth_addr"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	LockDays    int     `json:"lock_days" binding:"required,gte=1"`
}

// Stake handler: POST /staking/stake
func (h *StakingHandler) Stake(c *gin.Context) {
	var req StakeRequest
	if !dto.BindJSON(c, &req) {
		return
	}

	// Convert days to seconds
	lockDuration := int64(req.LockDays) * 86400

	record, err := h.stakingService.Stake(
		c.Request.Context(),
		req.Address,
		req.Amount,
		lockDuration,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse(gin.H{
		"stake_id":    record.ID,
		"amount":      record.Amount,
		"lock_until":  record.LockUntil,
		"status":      record.Status,
		"days_locked": req.LockDays,
	}))
}

// Unstake handler: POST /staking/unstake
func (h *StakingHandler) Unstake(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required,eth_addr"`
		StakeID int64  `json:"stake_id" binding:"required,gt=0"`
	}

	if !dto.BindJSON(c, &req) {
		return
	}

	if err := h.stakingService.Unstake(c.Request.Context(), req.Address, req.StakeID); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(gin.H{
		"message":  "Unstake berhasil, balance telah dikembalikan",
		"stake_id": req.StakeID,
	}))
}

// GetStatus handler: GET /staking/status?address=0x...
func (h *StakingHandler) GetStatus(c *gin.Context) {
	address := strings.TrimSpace(c.Query("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("address is required"))
		return
	}

	status, err := h.stakingService.GetStakingStatus(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(status))
}

// GetInfo handler: GET /staking/info
func (h *StakingHandler) GetInfo(c *gin.Context) {
	info, err := h.stakingService.GetStakingInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(info))
}

// Helper unused suppressor
var _ = fmt.Sprintf
var _ = strconv.Itoa
