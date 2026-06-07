package app

import (
	"context"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/config"
	"github.com/livingdolls/go-blockchain-simulate/logger"

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

// AppDependencies berisi seluruh dependensi yang dibutuhkan aplikasi.
// Dibuat dari konfigurasi terpusat, lalu diteruskan ke AppConfig.
// Definisi struct ada di app/config.go.

// NewAppDependencies membangun seluruh dependensi infrastruktur dari konfigurasi.
// Dipanggil dari main sebelum inisialisasi handler/service/worker.
func NewAppDependencies(cfg *config.Config) (*AppDependencies, error) {
	deps := &AppDependencies{Config: cfg}

	if err := deps.initializeInfrastructure(); err != nil {
		return nil, err
	}

	return deps, nil
}

// initializeInfrastructure membuka koneksi database, redis, dan rabbitmq
// serta membuat adapter JWT berdasarkan konfigurasi.
func (d *AppDependencies) initializeInfrastructure() error {
	// Database
	db, err := database.NewDBConn(database.Config{
		Driver:          d.Config.Database.Driver,
		DSN:             d.Config.Database.DSN,
		MaxOpenConns:    d.Config.Database.MaxOpenConns,
		MaxIdleConns:    d.Config.Database.MaxIdleConns,
		ConnMaxLifetime: d.Config.Database.ConnMaxLifetime,
	})
	if err != nil {
		return err
	}
	d.DBConn = db

	// Redis
	redisClient, err := redis.NewRedisClient(redis.Config{
		Addr:         d.Config.Redis.Addr,
		Password:     d.Config.Redis.Password,
		DB:           d.Config.Redis.DB,
		PoolSize:     d.Config.Redis.PoolSize,
		MinIdleConns: d.Config.Redis.MinIdleConns,
	})
	if err != nil {
		_ = db.Close()
		return err
	}
	d.RedisClient = redisClient

	redisServices, err := redis.NewMemoryAdapter(redisClient, 1024)
	if err != nil {
		_ = redisClient.Close()
		_ = db.Close()
		return err
	}
	d.RedisServices = redisServices

	// RabbitMQ
	rmqClient, err := rabbitmq.NewClient(d.Config.RabbitMQ.URL, d.Config.RabbitMQ.PoolSize)
	if err != nil {
		_ = redisClient.Close()
		_ = db.Close()
		return err
	}
	d.RMQClient = rmqClient

	// JWT
	d.JWT = security.NewJWTAdapter(d.Config.JWT.UserSecret, d.Config.JWT.UserTTL)
	d.JWTAdmin = security.NewAdminJWTAdapter(d.Config.JWT.AdminSecret, d.Config.JWT.AdminTTL)

	logger.LogInfo("infrastruktur berhasil diinisialisasi")
	return nil
}

// InitializeInfrastructure tetap dipertahankan untuk backward compatibility.
// Memindahkan nilai dari AppDependencies ke AppConfig.
// Disarankan migrasi ke NewAppDependencies di caller.
func (a *AppConfig) InitializeInfrastructure() error {
	if a.deps == nil {
		return entity.ErrDependenciesNotInitialized
	}
	a.DBConn = a.deps.DBConn
	a.DB = a.deps.DBConn.GetDB()
	a.RedisServices = a.deps.RedisServices
	a.RMQClient = a.deps.RMQClient
	a.JWT = a.deps.JWT
	a.JWTAdmin = a.deps.JWTAdmin
	return nil
}

// SetDeps menempelkan AppDependencies ke AppConfig.
// Cara yang disarankan di main, menggantikan pemanggilan NewAppDependencies
// terpisah + InitializeInfrastructure terpisah.
func (a *AppConfig) SetDeps(deps *AppDependencies) {
	a.deps = deps
	a.DBConn = deps.DBConn
	a.DB = deps.DBConn.GetDB()
	a.RedisServices = deps.RedisServices
	a.RMQClient = deps.RMQClient
	a.JWT = deps.JWT
	a.JWTAdmin = deps.JWTAdmin
}

// SetupRabbitMQTopology sets up queues, exchanges, and bindings
func (a *AppConfig) SetupRabbitMQTopology() error {
	queues := getQueueDefinitions()
	exchanges := getExchangeDefinitions()
	binds := getBindingDefinitions()

	for _, q := range queues {
		if err := a.RMQClient.DeclareQueue(q); err != nil {
			logger.LogError("Failed to declare queue", err)
		}
	}

	for _, e := range exchanges {
		if err := a.RMQClient.DeclareExchange(e); err != nil {
			logger.LogError("Failed to declare exchange", err)
		}
	}

	for _, b := range binds {
		if err := a.RMQClient.Bind(b); err != nil {
			logger.LogError("Failed to bind queue", err)
		}
	}

	logger.LogInfo("RabbitMQ topology initialized successfully")
	return nil
}

// InitializeWebSocket initializes WebSocket hub and publisher
func (a *AppConfig) InitializeWebSocket() {
	a.Hub = websocket.NewHub()
	go a.Hub.Run()
	a.PublisherWS = publisher.NewPublisherWS(a.Hub)
	logger.LogInfo("WebSocket hub initialized successfully")
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
	a.DiscrepancyRepo = repository.NewDiscrepancyRepository(a.DB)
	a.AdminRepo = repository.NewAdminRepository(a.DB)
	logger.LogInfo("All repositories initialized successfully")
}

// InitializePublishers initializes message publishers
func (a *AppConfig) InitializePublishers() {
	a.PricingPublisher = services.NewMarketPricingPublisher(a.RMQClient)
	a.LedgerPublisher = services.NewLedgerPublisher(a.RMQClient)
	a.RewardPublisher = services.NewRewardPublisher(a.RMQClient)
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
		a.CandleService, a.MarketService, a.PublisherWS, a.PricingPublisher, a.LedgerPublisher, a.RewardPublisher,
	)

	// Reward service
	a.RewardService = services.NewRewardHandler(a.BlockRepo)

	// Profile service
	a.ProfileService = services.NewProfileService(a.UserRepo)

	a.RewardPublisher = services.NewRewardPublisher(a.RMQClient)

	// Admin services
	a.AdminService = services.NewAdminService(a.AdminRepo)
	a.AdminAuthService = services.NewAdminAuthService(a.UserRepo, a.AdminRepo)

	logger.LogInfo("All services initialized successfully")
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
	a.AdminLoginHandler = handler.NewAdminLoginHandler(a.AdminAuthService, a.JWTAdmin)
	a.AdminHandler = handler.NewAdminHandler(a.AdminService)
	a.HealthHandler = handler.NewHealthHandler(a.DBConn, a.deps.RedisClient, a.deps.Config.App.Name, a.deps.Config.App.Version)
	logger.LogInfo("All handlers initialized successfully")
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

	logger.LogInfo("All background workers initialized successfully")
}

// InitializeConsumers initializes all message consumers
func (a *AppConfig) InitializeConsumers() {
	a.TransactionConsumer = worker.NewTransactionConsumer(a.RMQClient, a.TransactionService, 5)
	a.PricingConsumer = worker.NewMarketPricingConsumer(a.RMQClient, a.MarketRepo, a.PublisherWS, 3)
	a.VolumeConsumer = worker.NewMarketVolumeConsumer(a.RMQClient, a.MarketRepo, 2)
	a.AuditConsumer = worker.NewLedgerAuditConsumer(a.RMQClient, 3)
	a.LedgerPersistenceConsumer = worker.NewLedgerPersistenceConsumer(a.RMQClient, a.LedgerRepo, 5)

	reconcileConfig := worker.RecoilConfig{
		WorkerCount:       5,
		ReconcileWorkers:  3,
		ProcessingTimeout: 30 * time.Second,
		MaxDiscrepancies:  100000,
	}
	a.ReconcileConsumer = worker.NewLedgerReconcileConsumer(a.RMQClient, a.WalletRepo, a.LedgerRepo, a.DiscrepancyRepo, reconcileConfig)

	// Reward Calculation Consumer
	rewardCalcConfig := worker.RewardEngineConfig{
		ConsumerWorkers:   3,
		CalcWorkers:       5,
		ProcessingTimeout: 30 * time.Second,
		QueueSize:         1000,
	}
	a.RewardCalculationConsumer = worker.NewRewardCalculationConsumer(a.RMQClient, a.RewardPublisher, rewardCalcConfig)

	// reward distribution consumer
	rewardDistConfig := worker.RewardDistConfig{
		ConsumerWorkers:   3,
		DistWorkers:       5,
		ProcessingTimeout: 30 * time.Second,
		QueueSize:         1000,
		CleanupInterval:   5 * time.Minute,
		TTLDuration:       24 * time.Hour,
	}

	a.RewardDistributionConsumer = worker.NewRewardDistributionConsumer(a.RMQClient, a.WalletRepo, rewardDistConfig)

	logger.LogInfo("All message consumers initialized successfully")
}

// StartConsumers starts all message consumers asynchronously
func (a *AppConfig) StartConsumers() {
	go func() {
		if err := a.TransactionConsumer.Start(context.Background()); err != nil {
			logger.LogError("Error starting transaction consumer", err)
		}
	}()

	go func() {
		if err := a.PricingConsumer.Start(); err != nil {
			logger.LogError("Error starting pricing consumer", err)
		}
	}()

	go func() {
		if err := a.VolumeConsumer.Start(); err != nil {
			logger.LogError("Error starting volume consumer", err)
		}
	}()

	go func() {
		if err := a.LedgerPersistenceConsumer.Start(); err != nil {
			logger.LogError("Error starting ledger persistence consumer", err)
		}
	}()

	go func() {
		if err := a.AuditConsumer.Start(); err != nil {
			logger.LogError("Error starting audit consumer", err)
		}
	}()

	go func() {
		if err := a.ReconcileConsumer.Start(); err != nil {
			logger.LogError("Error starting reconcile consumer", err)
		}
	}()

	go func() {
		if err := a.RewardCalculationConsumer.Start(); err != nil {
			logger.LogError("Error starting reward calculation consumer", err)
		}
	}()

	go func() {
		if err := a.RewardDistributionConsumer.Start(); err != nil {
			logger.LogError("Error starting reward distribution consumer", err)
		}
	}()

	logger.LogInfo("All message consumers started successfully")
}

// Shutdown gracefully shuts down all components
func (a *AppConfig) Shutdown() {
	logger.LogInfo("Starting graceful shutdown...")

	stopWorkers(
		a.BlockWorker,
		a.CandleWorker,
		a.TransactionConsumer,
		a.PricingConsumer,
		a.VolumeConsumer,
		a.AuditConsumer,
		a.ReconcileConsumer,
		a.RewardCalculationConsumer,
		a.RewardDistributionConsumer,
	)

	if a.Hub != nil {
		closeHub(a.Hub, 15*time.Second)
	}

	if a.RMQClient != nil {
		a.RMQClient.Close()
	}

	logger.LogInfo("Shutdown complete")
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
					logger.LogInfo("Block worker stopped")
				case *worker.GenerateCandleWorker:
					v.Stop()
					logger.LogInfo("Candle worker stopped")
				case *worker.TransactionConsumer:
					v.Stop()
					logger.LogInfo("Transaction consumer stopped")
				case *worker.MarketPricingConsumer:
					v.Stop()
					logger.LogInfo("Market pricing consumer stopped")
				case *worker.MarketVolumeConsumer:
					v.Stop()
					logger.LogInfo("Market volume consumer stopped")
				case *worker.LedgerAuditConsumer:
					v.Stop()
					logger.LogInfo("Ledger audit consumer stopped")
				case *worker.LedgerReconcileConsumer:
					v.Stop()
					logger.LogInfo("Ledger reconcile consumer stopped")
				case *worker.RewardCalculationConsumer:
					v.Stop()
					logger.LogInfo("Reward calculation consumer stopped")
				case *worker.RewardDistributionConsumer:
					v.Stop()
					logger.LogInfo("Reward distribution consumer stopped")
				}
			}(w)
		}
		wg.Wait()
		close(stopChan)
	}()

	select {
	case <-stopChan:
		logger.LogInfo("All workers stopped")
	case <-ctx.Done():
		logger.LogInfo("Timeout while stopping workers")
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
		logger.LogInfo("WebSocket hub closed all connections")
	case <-ctx.Done():
		logger.LogInfo("Timeout while closing WebSocket hub connections")
	}
}
