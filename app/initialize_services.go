package app

import (
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

// InitializeServices menginisialisasi semua service. Urutan penting karena
// ada dependency antar service (mis. txVerify dibutuhkan TransactionService).
func (a *AppConfig) InitializeServices() {
	// User
	a.UserService = services.NewRegisterService(a.UserRepo, a.WalletRepo, a.BalanceRepo, a.JWT, a.RedisServices)

	// Transaction
	txVerify := services.NewVerifyTxService(a.RedisServices)
	a.TransactionService = services.NewTransactionService(a.UserRepo, a.WalletRepo, a.BalanceRepo, a.TxRepo, a.LedgerRepo, a.RedisServices, txVerify)

	// Balance
	a.BalanceService = services.NewBalanceService(a.UserRepo, a.TxRepo, a.BalanceRepo, a.PublisherWS)

	// Market
	a.MarketService = services.NewMarketEngineService(a.MarketRepo)

	// Candle (butuh candle stream service terpisah)
	candleStream := services.NewCandleStreamService(a.RedisServices)
	a.CandleService = services.NewCandleService(a.CandleRepo, candleStream)

	// Block (butuh banyak dependency)
	a.BlockService = services.NewBlockService(
		a.BlockRepo, a.WalletRepo, a.BalanceRepo, a.TxRepo, a.UserRepo,
		a.CandleService, a.MarketService, a.PublisherWS, a.PricingPublisher, a.LedgerPublisher, a.RewardPublisher,
	)

	// Reward
	a.RewardService = services.NewRewardHandler(a.BlockRepo)

	// Profile
	a.ProfileService = services.NewProfileService(a.UserRepo)

	// Admin
	a.AdminService = services.NewAdminService(a.AdminRepo)
	a.AdminAuthService = services.NewAdminAuthService(a.UserRepo, a.AdminRepo)

	logger.LogInfo("All services initialized successfully")
}
