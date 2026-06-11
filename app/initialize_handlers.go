package app

import (
	"github.com/livingdolls/go-blockchain-simulate/app/handler"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

// InitializeHandlers menginisialisasi semua HTTP handler. Setiap handler
// membungkus service yang sesuai; tidak ada logika bisnis di sini.
func (a *AppConfig) InitializeHandlers() {
	a.UserHandler = handler.NewRegisterHandler(a.UserService, a.RedisServices)
	a.TransactionHandler = handler.NewTransactionHandler(a.TransactionService, a.RMQClient, a.TxRepo)
	a.BalanceHandler = handler.NewBalanceHandler(a.BalanceService)
	a.BlockHandler = handler.NewBlockHandler(a.BlockService, a.WalletRepo, a.TxRepo)
	a.RewardHandler = handler.NewRewardHandler(a.RewardService, a.BlockService)
	a.ProfileHandler = handler.NewUserHandler(a.ProfileService, a.JWT)
	a.MarketHandler = handler.NewMarketHandler(a.MarketService)
	a.CandleHandler = handler.NewCandleHandler(a.CandleService)
	// Candle stream butuh candle stream service terpisah
	a.CandleStreamHandler = handler.NewCandleStreamHandler(services.NewCandleStreamService(a.RedisServices), a.CandleService)
	a.AdminLoginHandler = handler.NewAdminLoginHandler(a.AdminAuthService, a.JWTAdmin, a.RedisServices)
	a.AdminHandler = handler.NewAdminHandler(a.AdminService, a.UserRepo)
	a.HealthHandler = handler.NewHealthHandler(a.DBConn, a.deps.RedisClient, a.deps.Config.App.Name, a.deps.Config.App.Version)
	a.SwaggerHandler = handler.NewSwaggerHandler("docs/openapi.yaml")
	a.NotificationHandler = handler.NewNotificationHandler(a.NotificationRepo)
	logger.LogInfo("All handlers initialized successfully")
}
