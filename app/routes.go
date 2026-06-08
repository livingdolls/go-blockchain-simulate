package app

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/handler"
	mw "github.com/livingdolls/go-blockchain-simulate/app/middleware"
	"github.com/livingdolls/go-blockchain-simulate/app/websocket"
)

// SetupRoutes configures all HTTP routes
func (a *AppConfig) SetupRoutes(r *gin.Engine) {
	// Health check (public, tidak butuh auth). Liveness untuk livenessProbe,
	// readiness untuk readinessProbe atau load balancer health check.
	r.GET("/healthz", a.HealthHandler.Liveness)
	r.GET("/readyz", a.HealthHandler.Readiness)

	// Rate limiter untuk endpoint sensitif. Redis-backed sehingga terdistribusi
	// aman di multiple instance. Identifier: client IP (atau X-Forwarded-For).
	authLimiter := mw.RateLimitMiddleware(a.deps.RedisClient,
		mw.RateLimiter{KeyPrefix: "ratelimit:auth", Limit: 10, Window: time.Minute},
		mw.RateLimiter{KeyPrefix: "ratelimit:tx", Limit: 30, Window: time.Minute},
	)

	// Auth routes
	authGroup := r.Group("")
	authGroup.Use(authLimiter)
	{
		authGroup.POST("/register", a.UserHandler.Register)
		authGroup.POST("/challenge/:address", a.UserHandler.Challenge)
		authGroup.POST("/challenge/verify", a.UserHandler.Verify)
	}

	// Transaction routes
	txGroup := r.Group("/transaction")
	txGroup.Use(authLimiter)
	// Per-address rate limit (defense in depth di samping IP-based).
	// Mengambil address dari body (from_address / address) untuk membatasi
	// request per address wallet, bukan per IP. Mitigasi DoS flood RMQ.
	txGroup.Use(mw.PerAddressRateLimiter(a.deps.RedisClient, "ratelimit:tx-address", 10, time.Minute))
	// Idempotency-Key untuk mencegah double-submit saat client retry.
	// Hanya POST /transaction/send dan /transaction/buy yang di-cache
	// (idempotency key hanya relevan untuk operasi non-idempotent).
	idemCfg := mw.IdempotencyConfig{
		Memory:    a.RedisServices,
		TTL:       24 * time.Hour,
		KeyPrefix: "idempotency:tx",
		RequiredScope: []string{
			"POST /transaction/send",
			"POST /transaction/buy",
			"POST /transaction/sell",
		},
	}
	txGroup.Use(mw.IdempotencyMiddleware(idemCfg))
	{
		txGroup.POST("/send", a.TransactionHandler.Send)
		txGroup.GET("/:id", a.TransactionHandler.GetTransaction)
		txGroup.POST("/buy", a.TransactionHandler.Buy)
		txGroup.POST("/sell", a.TransactionHandler.Sell)
	}

	// Admin Auth routes (public - no middleware required)
	adminAuthGroup := r.Group("/admin/auth")
	{
		adminAuthGroup.POST("/login", a.AdminLoginHandler.Login)
		adminAuthGroup.POST("/logout", a.AdminLoginHandler.Logout)
	}

	// Admin dashboard & management routes (protected with middleware)
	adminGroup := r.Group("/admin")
	adminGroup.Use(handler.AdminMiddleware(a.JWTAdmin, a.AdminRepo))
	{
		adminGroup.GET("/dashboard", a.AdminHandler.Dashboard)
		adminGroup.GET("/admins", a.AdminHandler.ListAdmins)
		adminGroup.POST("/admins", a.AdminHandler.CreateAdmin)
		adminGroup.PUT("/admins/:id/role", a.AdminHandler.UpdateAdminRole)
		adminGroup.PUT("/admins/:id/status", a.AdminHandler.UpdateAdminStatus)
		adminGroup.DELETE("/admins/:id", a.AdminHandler.DeleteAdmin)
		adminGroup.GET("/activity-logs", a.AdminHandler.GetActivityLogs)
		adminGroup.GET("/activity-logs/recent", a.AdminHandler.RecentActivityLogs)
	}

	// Nonce generation
	r.GET("/generate-tx-nonce/:address", a.TransactionHandler.GenerateNonce)

	// Balance routes. GET publik (read-only), topup wajib JWT
	// (self-only: address di-body harus match address di-claims).
	balanceGroup := r.Group("/balance")
	{
		balanceGroup.GET("/:address", a.BalanceHandler.GetUserWithUSDBalance)
		balanceGroup.POST("/topup", handler.JWTMiddleware(a.JWT), a.BalanceHandler.TopUpUSDBalance)
	}

	// Wallet routes
	r.GET("/wallet/:address", a.BalanceHandler.GetWalletBalance)

	// Block routes. Generate adalah operasi privileged (mining CPU-heavy
	// dan mempengaruhi integritas chain) sehingga hanya admin yang boleh
	// memicu manual. Worker periodik (GenerateBlockWorker) tetap men-trigger
	// otomatis setiap 10 detik tanpa lewat endpoint.
	blockGroup := r.Group("/blocks")
	{
		blockGroup.GET("", a.BlockHandler.GetBlocks)
		blockGroup.GET("/:id", a.BlockHandler.GetBlockByID)
		blockGroup.GET("/detail/:number", a.BlockHandler.GetBlockByBlockNumber)
		blockGroup.GET("/integrity", a.BlockHandler.CheckBlockchainIntegrity)
		blockGroup.GET("/transaction/:number", a.BlockHandler.GetTransactionsByBlockNumber)
		blockGroup.GET("/search", a.BlockHandler.SearchBlocksByHash)
		blockGroup.GET("/range", a.BlockHandler.GetBlocksInRange)
		blockGroup.GET("/stats", a.BlockHandler.GetBlockStats)
		blockGroup.GET("/search/miner/", a.BlockHandler.SearchBlocksByMinerAddress)
	}
	// /blocks/generate dipisah di luar group agar middleware admin bisa
	// dipasang spesifik hanya untuk endpoint ini.
	r.POST("/blocks/generate", handler.AdminMiddleware(a.JWTAdmin, a.AdminRepo), a.BlockHandler.GenerateBlock)

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
