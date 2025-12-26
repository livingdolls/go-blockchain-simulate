package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app/handler"
	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/app/websocket"
	"github.com/livingdolls/go-blockchain-simulate/app/worker"
	"github.com/livingdolls/go-blockchain-simulate/database"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

func main() {

	// initialize database
	db, err := database.NewDBConn()

	if err != nil {
		panic(err)
	}
	defer db.Close()

	jwt := security.NewJWTAdapter("yurinahirate-verysecret", 24*time.Hour)

	redisClient, err := redis.NewRedisMemory()
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	redisServices, err := redis.NewMemoryAdapter(redisClient, 1024)

	if err != nil {
		panic(err)
	}

	hub := websocket.NewHub()
	go hub.Run()

	hubHandler := websocket.GinHandler(hub, jwt)
	publisherWS := publisher.NewPublisherWS(hub)

	userRepo := repository.NewUserRepository(db.GetDB())
	userService := services.NewRegisterService(userRepo, jwt, redisServices)
	userHandler := handler.NewRegisterHandler(userService)

	txRepo := repository.NewTransactionRepository(db.GetDB())
	ledgerRepo := repository.NewLedgerRepository(db.GetDB())

	txVerify := services.NewVerifyTxService(redisServices)

	transactionService := services.NewTransactionService(userRepo, txRepo, ledgerRepo, redisServices, txVerify)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	balanceService := services.NewBalanceService(userRepo, txRepo)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	marketRepo := repository.NewMarketRepository(db.GetDB())
	marketService := services.NewMarketEngineService(marketRepo)

	blockRepo := repository.NewBlockRepository(db.GetDB())
	blockService := services.NewBlockService(blockRepo, txRepo, userRepo, ledgerRepo, marketService, publisherWS)
	blockHandler := handler.NewBlockHandler(blockService)

	rewardService := services.NewRewardHandler(blockRepo)
	rewardHandler := handler.NewRewardHandler(rewardService, blockService)

	profileService := services.NewProfileService(userRepo)
	profileHandler := handler.NewUserHandler(profileService, jwt)

	marketHandler := handler.NewMarketHandler(marketService)

	// start worker
	generateBlockWorker := worker.NewGenerateBlockWorker(blockService)
	generateBlockWorker.Start(10 * time.Second)

	r := gin.Default()

	allowedOrigins := map[string]bool{
		"http://192.168.88.178:3001": true,
		"http://localhost:3001":      true,
		"http://192.168.88.178:3000": true,
		"http://192.168.88.178:3002": true,
	}

	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.POST("/register", userHandler.Register)
	r.POST("/challenge/:address", userHandler.Challenge)
	r.POST("/challenge/verify", userHandler.Verify)
	r.POST("/send", transactionHandler.Send)
	r.GET("/generate-tx-nonce/:address", transactionHandler.GenerateNonce)
	r.GET("/transaction/:id", transactionHandler.GetTransaction)
	r.POST("/transaction/buy", transactionHandler.Buy)
	r.POST("/transaction/sell", transactionHandler.Sell)
	r.GET("/balance/:address", balanceHandler.GetBalance)
	r.GET("/wallet/:address", balanceHandler.GetWalletBalance)
	r.POST("/generate-block", blockHandler.GenerateBlock)
	r.GET("/blocks", blockHandler.GetBlocks)
	r.GET("/blocks/:id", blockHandler.GetBlockByID)
	r.GET("/blocks/detail/:number", blockHandler.GetBlockByBlockNumber)
	r.GET("/blocks/integrity", blockHandler.CheckBlockchainIntegrity)
	r.GET("/reward/schedule/:number", rewardHandler.GetRewardSchedule)
	r.GET("/reward/block/:number", rewardHandler.GetBlockReward)
	r.GET("/reward/info", rewardHandler.GetRewardInfo)
	r.GET("/ws/market", hubHandler)
	r.GET("/market", marketHandler.GetMarketEngineState)

	protected := r.Group("/profile")
	protected.Use(handler.JWTMiddleware(jwt))
	{
		protected.GET("", profileHandler.Me)
	}

	r.Run(":3010")
}
