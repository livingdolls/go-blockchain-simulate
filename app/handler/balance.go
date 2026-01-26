package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
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
	})
}

func (h *BalanceHandler) GetWalletBalance(c *gin.Context) {
	address := c.Param("address")

	if address == "" {
		c.JSON(400, gin.H{"error": "address is required"})
		return
	}

	filter := models.TransactionFilter{
		Address: address,
		Type:    c.DefaultQuery("type", "all"),
		Status:  c.DefaultQuery("status", "all"),
		SortBy:  c.DefaultQuery("sort_by", "id"),
		Order:   c.DefaultQuery("order", "desc"),
	}

	//parse page
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	filter.Page = page

	//parse limit
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	filter.Limit = limit

	walletResponse, err := h.service.GetWalletBalance(filter)

	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, walletResponse)
}

func (h *BalanceHandler) TopUpUSDBalance(c *gin.Context) {
	var req dto.TopUpRequestDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.NewErrorResponse[string]("invalid request body"))
		return
	}

	res, err := h.service.TopUpUSDBalance(req.Address, req.Amount, req.ReferenceID, req.Description)

	if err != nil {
		status := http.StatusInternalServerError

		switch {
		case errors.Is(err, entity.ErrAddressNotFound),
			errors.Is(err, entity.ErrUserBalanceNotFound),
			errors.Is(err, entity.ErrAmountMustBePositive):
			status = http.StatusBadRequest
		}
		c.JSON(status, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *BalanceHandler) GetUserWithUSDBalance(c *gin.Context) {
	address := c.Param("address")

	if address == "" {
		c.JSON(400, dto.NewErrorResponse[string]("address is required"))
		return
	}

	userWithBalance, err := h.service.GetUserWithUSDBalance(address)

	if err != nil {
		c.JSON(404, dto.NewErrorResponse[string](err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(userWithBalance))
}
