package app

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/handler"
	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/app/websocket"
	"github.com/livingdolls/go-blockchain-simulate/app/worker"
	"github.com/livingdolls/go-blockchain-simulate/database"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

// InitializeInfrastructure initializes database, cache, message queue, and auth
func (a *AppConfig) InitializeInfrastructure() error {
	// Database
	db, err := database.NewDBConn()
	if err != nil {
		return err
	}
	a.DB = db.GetDB()

	// Redis
	redisClient, err := redis.NewRedisMemory()
	if err != nil {
		return err
	}
	redisServices, err := redis.NewMemoryAdapter(redisClient, 1024)
	if err != nil {
		return err
	}
	a.RedisServices = redisServices

	// RabbitMQ
	rmqClient, err := rabbitmq.NewClient("amqp://guest:guest@localhost:5672/", 10)
	if err != nil {
		return err
	}
	a.RMQClient = rmqClient

	// JWT
	a.JWT = security.NewJWTAdapter("yurinahirate-verysecret", 24*time.Hour)

	log.Println("[BOOTSTRAP] Infrastructure initialized successfully")
	return nil
}

// SetupRabbitMQTopology sets up queues, exchanges, and bindings
func (a *AppConfig) SetupRabbitMQTopology() error {
	queues := getQueueDefinitions()
	exchanges := getExchangeDefinitions()
	binds := getBindingDefinitions()

	for _, q := range queues {
		if err := a.RMQClient.DeclareQueue(q); err != nil {
			log.Printf("Warning: Failed to declare queue %s: %v\n", q.Name, err)
		}
	}

	for _, e := range exchanges {
		if err := a.RMQClient.DeclareExchange(e); err != nil {
			log.Printf("Warning: Failed to declare exchange %s: %v\n", e.Name, err)
		}
	}

	for _, b := range binds {
		if err := a.RMQClient.Bind(b); err != nil {
			log.Printf("Warning: Failed to bind queue %s: %v\n", b.Queue, err)
		}
	}

	log.Println("[RABBITMQ] Topology initialized successfully")
	return nil
}

// InitializeWebSocket initializes WebSocket hub and publisher
func (a *AppConfig) InitializeWebSocket() {
	a.Hub = websocket.NewHub()
	go a.Hub.Run()
	a.PublisherWS = publisher.NewPublisherWS(a.Hub)
	log.Println("[WEBSOCKET] Hub initialized successfully")
}

// InitializeRepositories initializes all data repositories
func (a *AppConfig) InitializeRepositories() {
	a.UserRepo = repository.NewUserRepository(a.DB)
	a.WalletRepo = repository.NewUserWalletRepository(a.DB)
	a.BalanceRepo = repository.NewUserBalanceRepository(a.DB)
	a.TxRepo = repository.NewTransactionRepository(a.DB)
	a.LedgerRepo = repository.NewLedgerRepository(a.DB)
	a.MarketRepo = repository.NewMarketRepository(a.DB)
	a.BlockRepo = repository.NewBlockRepository(a.DB)
	a.CandleRepo = repository.NewCandleRepository(a.DB)
	log.Println("[REPOSITORIES] All repositories initialized successfully")
}

// InitializePublishers initializes message publishers
func (a *AppConfig) InitializePublishers() {
	a.PricingPublisher = services.NewMarketPricingPublisher(a.RMQClient)
	a.LedgerPublisher = services.NewLedgerPublisher(a.RMQClient)
	log.Println("[PUBLISHERS] All publishers initialized successfully")
}

// InitializeServices initializes all business logic services
func (a *AppConfig) InitializeServices() {
	// User service
	a.UserService = services.NewRegisterService(a.UserRepo, a.WalletRepo, a.BalanceRepo, a.JWT, a.RedisServices)

	// Transaction service
	txVerify := services.NewVerifyTxService(a.RedisServices)
	a.TransactionService = services.NewTransactionService(a.UserRepo, a.WalletRepo, a.BalanceRepo, a.TxRepo, a.LedgerRepo, a.RedisServices, txVerify)

	// Balance service
	a.BalanceService = services.NewBalanceService(a.UserRepo, a.TxRepo, a.BalanceRepo, a.PublisherWS)

	// Market service
	a.MarketService = services.NewMarketEngineService(a.MarketRepo)

	// Candle service
	candleStream := services.NewCandleStreamService(a.RedisServices)
	a.CandleService = services.NewCandleService(a.CandleRepo, candleStream)

	// Block service
	a.BlockService = services.NewBlockService(
		a.BlockRepo, a.WalletRepo, a.BalanceRepo, a.TxRepo, a.UserRepo,
		a.CandleService, a.MarketService, a.PublisherWS, a.PricingPublisher, a.LedgerPublisher,
	)

	// Reward service
	a.RewardService = services.NewRewardHandler(a.BlockRepo)

	// Profile service
	a.ProfileService = services.NewProfileService(a.UserRepo)

	log.Println("[SERVICES] All services initialized successfully")
}

// InitializeHandlers initializes all HTTP request handlers
func (a *AppConfig) InitializeHandlers() {
	a.UserHandler = handler.NewRegisterHandler(a.UserService)
	a.TransactionHandler = handler.NewTransactionHandler(a.TransactionService, a.RMQClient)
	a.BalanceHandler = handler.NewBalanceHandler(a.BalanceService)
	a.BlockHandler = handler.NewBlockHandler(a.BlockService)
	a.RewardHandler = handler.NewRewardHandler(a.RewardService, a.BlockService)
	a.ProfileHandler = handler.NewUserHandler(a.ProfileService, a.JWT)
	a.MarketHandler = handler.NewMarketHandler(a.MarketService)
	a.CandleHandler = handler.NewCandleHandler(a.CandleService)
	a.CandleStreamHandler = handler.NewCandleStreamHandler(services.NewCandleStreamService(a.RedisServices), a.CandleService)
	log.Println("[HANDLERS] All handlers initialized successfully")
}

// InitializeWorkers initializes background workers
func (a *AppConfig) InitializeWorkers() {
	// Block generation worker
	a.BlockWorker = worker.NewGenerateBlockWorker(a.BlockService)
	a.BlockWorker.Start(10 * time.Second)

	// Candle generation worker
	a.CandleWorker = worker.NewGenerateCandlesWorker(a.CandleService, 4)
	a.CandleWorker.SetJobTimeout(45 * time.Second)
	a.CandleWorker.Start(1 * time.Second)

	log.Println("[WORKERS] All background workers initialized successfully")
}

// InitializeConsumers initializes all message consumers
func (a *AppConfig) InitializeConsumers() {
	a.TransactionConsumer = worker.NewTransactionConsumer(a.RMQClient, a.TransactionService, 5)
	a.PricingConsumer = worker.NewMarketPricingConsumer(a.RMQClient, a.MarketRepo, a.PublisherWS, 3)
	a.VolumeConsumer = worker.NewMarketVolumeConsumer(a.RMQClient, a.MarketRepo, 2)
	a.AuditConsumer = worker.NewLedgerAuditConsumer(a.RMQClient, 3)

	reconcileConfig := worker.RecoilConfig{
		WorkerCount:       5,
		ReconcileWorkers:  3,
		ProcessingTimeout: 30 * time.Second,
		MaxDiscrepancies:  100000,
	}
	a.ReconcileConsumer = worker.NewLedgerReconcileConsumer(a.RMQClient, a.WalletRepo, a.LedgerRepo, reconcileConfig)

	log.Println("[CONSUMERS] All message consumers initialized successfully")
}

// StartConsumers starts all message consumers asynchronously
func (a *AppConfig) StartConsumers() {
	go func() {
		if err := a.TransactionConsumer.Start(context.Background()); err != nil {
			log.Println("[TRANSACTION_CONSUMER] Error starting:", err)
		}
	}()

	go func() {
		if err := a.PricingConsumer.Start(); err != nil {
			log.Println("[PRICING_CONSUMER] Error starting:", err)
		}
	}()

	go func() {
		if err := a.VolumeConsumer.Start(); err != nil {
			log.Println("[VOLUME_CONSUMER] Error starting:", err)
		}
	}()

	go func() {
		if err := a.AuditConsumer.Start(); err != nil {
			log.Println("[AUDIT_CONSUMER] Error starting:", err)
		}
	}()

	go func() {
		if err := a.ReconcileConsumer.Start(); err != nil {
			log.Println("[RECONCILE_CONSUMER] Error starting:", err)
		}
	}()

	log.Println("[CONSUMERS] All message consumers started successfully")
}

// Shutdown gracefully shuts down all components
func (a *AppConfig) Shutdown() {
	log.Println("[BOOTSTRAP] Starting graceful shutdown...")

	stopWorkers(
		a.BlockWorker,
		a.CandleWorker,
		a.TransactionConsumer,
		a.PricingConsumer,
		a.VolumeConsumer,
		a.AuditConsumer,
		a.ReconcileConsumer,
	)

	if a.Hub != nil {
		closeHub(a.Hub, 15*time.Second)
	}

	if a.RMQClient != nil {
		a.RMQClient.Close()
	}

	log.Println("[BOOTSTRAP] Shutdown complete")
}

// stopWorkers stops all workers gracefully
func stopWorkers(workers ...interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
					log.Println("[WORKER] block worker stopped")
				case *worker.GenerateCandleWorker:
					v.Stop()
					log.Println("[WORKER] candle worker stopped")
				case *worker.TransactionConsumer:
					v.Stop()
					log.Println("[WORKER] transaction consumer stopped")
				case *worker.MarketPricingConsumer:
					v.Stop()
					log.Println("[WORKER] market pricing consumer stopped")
				case *worker.MarketVolumeConsumer:
					v.Stop()
					log.Println("[WORKER] market volume consumer stopped")
				case *worker.LedgerAuditConsumer:
					v.Stop()
					log.Println("[WORKER] ledger audit consumer stopped")
				case *worker.LedgerReconcileConsumer:
					v.Stop()
					log.Println("[WORKER] ledger reconcile consumer stopped")
				}
			}(w)
		}
		wg.Wait()
		close(stopChan)
	}()

	select {
	case <-stopChan:
		log.Println("[WORKER] All workers stopped")
	case <-ctx.Done():
		log.Println("[WORKER] Timeout while stopping workers")
	}
}

// closeHub closes WebSocket hub connections with timeout
func closeHub(hub *websocket.Hub, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		hub.Close()
		close(done)
	}()

	select {
	case <-done:
		log.Println("[WEBSOCKET] WebSocket hub closed all connections")
	case <-ctx.Done():
		log.Println("[WEBSOCKET] Timeout while closing WebSocket hub connections")
	}
}
