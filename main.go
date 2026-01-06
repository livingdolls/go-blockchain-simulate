package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	userBalanceRepository := repository.NewUserBalanceRepository(db.GetDB())
	walletRepo := repository.NewUserWalletRepository(db.GetDB())

	userRepo := repository.NewUserRepository(db.GetDB())
	userService := services.NewRegisterService(userRepo, walletRepo, userBalanceRepository, jwt, redisServices)
	userHandler := handler.NewRegisterHandler(userService)

	txRepo := repository.NewTransactionRepository(db.GetDB())
	ledgerRepo := repository.NewLedgerRepository(db.GetDB())

	txVerify := services.NewVerifyTxService(redisServices)

	transactionService := services.NewTransactionService(userRepo, walletRepo, txRepo, ledgerRepo, redisServices, txVerify)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	balanceService := services.NewBalanceService(userRepo, txRepo, userBalanceRepository, publisherWS)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	marketRepo := repository.NewMarketRepository(db.GetDB())
	marketService := services.NewMarketEngineService(marketRepo)

	candleStreamServices := services.NewCandleStreamService(redisServices)
	candleRepo := repository.NewCandleRepository(db.GetDB())
	candleService := services.NewCandleService(candleRepo, candleStreamServices)
	candleHandler := handler.NewCandleHandler(candleService)
	candleStreamHandler := handler.NewCandleStreamHandler(candleStreamServices, candleService)

	blockRepo := repository.NewBlockRepository(db.GetDB())
	blockService := services.NewBlockService(blockRepo, walletRepo, txRepo, userRepo, ledgerRepo, candleService, marketService, publisherWS)
	blockHandler := handler.NewBlockHandler(blockService)

	rewardService := services.NewRewardHandler(blockRepo)
	rewardHandler := handler.NewRewardHandler(rewardService, blockService)

	profileService := services.NewProfileService(userRepo)
	profileHandler := handler.NewUserHandler(profileService, jwt)

	marketHandler := handler.NewMarketHandler(marketService)

	// start worker
	generateBlockWorker := worker.NewGenerateBlockWorker(blockService)
	generateBlockWorker.Start(10 * time.Second)

	candleWorker := worker.NewGenerateCandlesWorker(candleService, 4)
	candleWorker.SetJobTimeout(45 * time.Second)
	candleWorker.Start(1 * time.Second)

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
	r.POST("/balance/topup", balanceHandler.TopUpUSDBalance)
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
	r.GET("/sse/candles", candleStreamHandler.StreamCandles)
	r.GET("/sse/ping", candleStreamHandler.Ping)

	candleGroup := r.Group("/candles")
	{
		candleGroup.GET("", candleHandler.GetCandle)
		candleGroup.GET("/range", candleHandler.GetCandleFrom)
	}

	protected := r.Group("/profile")
	protected.Use(handler.JWTMiddleware(jwt))
	{
		protected.GET("", profileHandler.Me)
	}

	// setup gracefull shutdown
	go func() {
		log.Println("Server starting on port :3010")

		if err := r.Run(":3010"); err != nil && err.Error() != "http: Server closed" {
			log.Printf("Server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// wait signal
	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down...\n", sig)

	// gracefull shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// stop workers

	stopWorkers(ctx, generateBlockWorker, candleWorker)

	// close websocket hub
	closeHub(hub, 15*time.Second)

	log.Println("Server gracefully stopped")
	os.Exit(0)

}

func stopWorkers(ctx context.Context, workers ...interface{}) {
	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	wg.Add(len(workers))

	go func() {
		for _, w := range workers {
			go func(workerInstance interface{}) {
				defer wg.Done()
				switch v := workerInstance.(type) {
				case *worker.GenerateBlockWorker:
					v.Stop()
					log.Println("block worker stopped")
				case *worker.GenerateCandleWorker:
					v.Stop()
					log.Println("candle worker stopped")
				}
			}(w)
		}
		wg.Wait()
		close(stopChan)
	}()

	select {
	case <-stopChan:
		log.Println("All workers stopped")
	case <-ctx.Done():
		log.Println("Timeout while stopping workers")
	}
}

func closeHub(hub *websocket.Hub, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// close client connections
	done := make(chan struct{})

	go func() {
		hub.Close()
		close(done)
	}()

	select {
	case <-done:
		log.Println("WebSocket hub closed all connections")
	case <-ctx.Done():
		log.Println("Timeout while closing WebSocket hub connections")
	}
}
