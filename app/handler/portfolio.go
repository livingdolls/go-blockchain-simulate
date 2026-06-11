package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type PortfolioHandler struct {
	userRepo    repository.UserRepository
	walletRepo  repository.UserWalletRepository
	balanceRepo repository.UserBalanceRepository
	marketSvc   services.MarketEngineService
}

func NewPortfolioHandler(
	userRepo repository.UserRepository,
	walletRepo repository.UserWalletRepository,
	balanceRepo repository.UserBalanceRepository,
	marketSvc services.MarketEngineService,
) *PortfolioHandler {
	return &PortfolioHandler{
		userRepo:    userRepo,
		walletRepo:  walletRepo,
		balanceRepo: balanceRepo,
		marketSvc:   marketSvc,
	}
}

// GetPortfolio mengembalikan portfolio analytics untuk user tertentu.
//
// Endpoint: GET /portfolio/:address
//
// Response mencakup:
// - YTE + USD balance
// - Total value dalam USD (yte_balance * price + usd_balance)
// - P&L (realized vs unrealized)
// - Asset allocation
func (h *PortfolioHandler) GetPortfolio(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse[string]("address is required"))
		return
	}

	// Get wallet (YTE balance)
	wallet, err := h.walletRepo.GetByAddress(address)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse[string]("wallet not found"))
		return
	}

	// Get balance (USD + trading stats)
	balance, err := h.balanceRepo.GetByAddress(address)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.NewErrorResponse[string]("balance not found"))
		return
	}

	// Get market price
	ytePrice := 1.0
	if h.marketSvc != nil {
		state, err := h.marketSvc.GetState()
		if err == nil && state.Price > 0 {
			ytePrice = state.Price
		}
	}

	// Portfolio calculations
	yteBalance := wallet.YTEBalance
	usdBalance := balance.USDBalance
	totalValueUSD := yteBalance*ytePrice + usdBalance

	totalDeposited := balance.TotalDeposited
	totalWithdrawn := balance.TotalWithdrawn
	totalTraded := balance.TotalTraded

	// P&L calculations
	costBasis := totalDeposited
	realizedPnL := totalTraded - costBasis
	unrealizedPnL := totalValueUSD - costBasis

	var pnlPercent float64
	if costBasis > 0 {
		pnlPercent = ((totalValueUSD - costBasis) / costBasis) * 100
	}

	// Asset allocation
	var ytePercent, usdPercent float64
	if totalValueUSD > 0 {
		ytePercent = (yteBalance * ytePrice / totalValueUSD) * 100
		usdPercent = (usdBalance / totalValueUSD) * 100
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(gin.H{
		"address":           address,
		"yte_balance":       yteBalance,
		"usd_balance":       usdBalance,
		"yte_price":         ytePrice,
		"total_value_usd":   totalValueUSD,
		"total_deposited":   totalDeposited,
		"total_withdrawn":   totalWithdrawn,
		"total_traded":      totalTraded,
		"realized_pnl":      realizedPnL,
		"unrealized_pnl":    unrealizedPnL,
		"pnl_percent":       pnlPercent,
		"allocation": gin.H{
			"yte_percent": ytePercent,
			"usd_percent": usdPercent,
		},
	}))
}
