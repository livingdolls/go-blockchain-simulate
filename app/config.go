package app

import (
	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/handler"
	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/app/websocket"
	"github.com/livingdolls/go-blockchain-simulate/app/worker"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

// AppConfig holds all application dependencies
type AppConfig struct {
	// Database
	DB *sqlx.DB

	// Cache
	RedisServices redis.MemoryAdapter

	// Message Queue
	RMQClient *rabbitmq.Client

	// Auth
	JWT security.JWTService

	// WebSocket
	Hub         *websocket.Hub
	PublisherWS *publisher.PublisherWS

	// Repositories
	UserRepo        repository.UserRepository
	WalletRepo      repository.UserWalletRepository
	BalanceRepo     repository.UserBalanceRepository
	TxRepo          repository.TransactionRepository
	LedgerRepo      repository.LedgerRepository
	MarketRepo      repository.MarketRepository
	BlockRepo       repository.BlockRepository
	CandleRepo      repository.CandlesRepository
	DiscrepancyRepo repository.DiscrepancyRepository

	// Publishers
	PricingPublisher services.MarketPricingPublisher
	LedgerPublisher  services.LedgerPublisher
	RewardPublisher  services.RewardPublisher

	// Services
	UserService        services.RegisterService
	TransactionService services.TransactionService
	BalanceService     services.BalanceService
	MarketService      services.MarketEngineService
	CandleService      services.CandleService
	BlockService       services.BlockService
	RewardService      services.RewardService
	ProfileService     services.ProfileService

	// Handlers
	UserHandler         *handler.RegisterHandler
	TransactionHandler  *handler.TransactionHandler
	BalanceHandler      *handler.BalanceHandler
	BlockHandler        *handler.BlockHandler
	RewardHandler       *handler.RewardHandler
	ProfileHandler      *handler.UserHandler
	MarketHandler       *handler.MarketHandler
	CandleHandler       *handler.CandleHandler
	CandleStreamHandler *handler.CandleStreamHandler

	// Workers
	BlockWorker  *worker.GenerateBlockWorker
	CandleWorker *worker.GenerateCandleWorker

	// Consumers
	TransactionConsumer        *worker.TransactionConsumer
	PricingConsumer            *worker.MarketPricingConsumer
	VolumeConsumer             *worker.MarketVolumeConsumer
	LedgerPersistenceConsumer  *worker.LedgerPersistenceConsumer
	AuditConsumer              *worker.LedgerAuditConsumer
	ReconcileConsumer          *worker.LedgerReconcileConsumer
	RewardCalculationConsumer  *worker.RewardCalculationConsumer
	RewardDistributionConsumer *worker.RewardDistributionConsumer
}
