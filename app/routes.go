package app

import (
	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/handler"
	"github.com/livingdolls/go-blockchain-simulate/app/websocket"
)

// SetupRoutes configures all HTTP routes
func (a *AppConfig) SetupRoutes(r *gin.Engine) {
	// Auth routes
	authGroup := r.Group("")
	{
		authGroup.POST("/register", a.UserHandler.Register)
		authGroup.POST("/challenge/:address", a.UserHandler.Challenge)
		authGroup.POST("/challenge/verify", a.UserHandler.Verify)
	}

	// Transaction routes
	txGroup := r.Group("/transaction")
	{
		txGroup.POST("/send", a.TransactionHandler.Send)
		txGroup.GET("/:id", a.TransactionHandler.GetTransaction)
		txGroup.POST("/buy", a.TransactionHandler.Buy)
		txGroup.POST("/sell", a.TransactionHandler.Sell)
	}

	// Nonce generation
	r.GET("/generate-tx-nonce/:address", a.TransactionHandler.GenerateNonce)

	// Balance routes
	balanceGroup := r.Group("/balance")
	{
		balanceGroup.GET("/:address", a.BalanceHandler.GetUserWithUSDBalance)
		balanceGroup.POST("/topup", a.BalanceHandler.TopUpUSDBalance)
	}

	// Wallet routes
	r.GET("/wallet/:address", a.BalanceHandler.GetWalletBalance)

	// Block routes
	blockGroup := r.Group("/blocks")
	{
		blockGroup.POST("/generate", a.BlockHandler.GenerateBlock)
		blockGroup.GET("", a.BlockHandler.GetBlocks)
		blockGroup.GET("/:id", a.BlockHandler.GetBlockByID)
		blockGroup.GET("/detail/:number", a.BlockHandler.GetBlockByBlockNumber)
		blockGroup.GET("/integrity", a.BlockHandler.CheckBlockchainIntegrity)
	}

	// Reward routes
	rewardGroup := r.Group("/reward")
	{
		rewardGroup.GET("/schedule/:number", a.RewardHandler.GetRewardSchedule)
		rewardGroup.GET("/block/:number", a.RewardHandler.GetBlockReward)
		rewardGroup.GET("/info", a.RewardHandler.GetRewardInfo)
	}

	// Market routes
	r.GET("/market", a.MarketHandler.GetMarketEngineState)

	// Candle routes
	candleGroup := r.Group("/candles")
	{
		candleGroup.GET("", a.CandleHandler.GetCandle)
		candleGroup.GET("/range", a.CandleHandler.GetCandleFrom)
	}

	// Streaming routes
	r.GET("/sse/candles", a.CandleStreamHandler.StreamCandles)
	r.GET("/sse/ping", a.CandleStreamHandler.Ping)

	// WebSocket routes
	r.GET("/ws/market", a.setupWebSocketHandler())

	// Protected routes
	protected := r.Group("/profile")
	protected.Use(handler.JWTMiddleware(a.JWT))
	{
		protected.GET("", a.ProfileHandler.Me)
	}
}

// setupWebSocketHandler creates WebSocket handler
func (a *AppConfig) setupWebSocketHandler() gin.HandlerFunc {
	return websocket.GinHandler(a.Hub, a.JWT)
}
