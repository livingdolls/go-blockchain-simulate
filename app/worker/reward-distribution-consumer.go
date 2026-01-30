package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/logger"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type RewardStats struct {
	MinerAddress    string
	TotalRewards    float64
	TotalUSDValue   float64
	RewardCount     int
	LastRewardAt    int64
	LastBlockNumber int
}

type ProcessedReward struct {
	BlockNumber int
	Timestamp   int64
	ExpiresAt   int64
}

type RewardDistConfig struct {
	ConsumerWorkers   int
	DistWorkers       int
	ProcessingTimeout time.Duration
	QueueSize         int
	CleanupInterval   time.Duration // TTL cleanup interval
	TTLDuration       time.Duration // TTL for processed rewards (default 24h)
}

type RewardDistributionConsumer struct {
	client     *rabbitmq.Client
	walletRepo repository.UserWalletRepository
	cfg        RewardDistConfig
	mu         sync.Mutex
	isRunning  bool
	stopCtx    context.Context
	stopCancel context.CancelFunc

	distQueue chan dto.RewardDistributionEvent
	statsMu   sync.RWMutex
	stats     map[string]RewardStats

	// Idempotency tracking
	processedMu   sync.RWMutex
	processed     map[string]ProcessedReward // Key: blockNumber_minerAddress
	retryQueue    chan dto.RewardDistributionEvent
	retryTickerC  <-chan time.Time
	cleanupTicker <-chan time.Time
}

func NewRewardDistributionConsumer(client *rabbitmq.Client, walletRepo repository.UserWalletRepository, cfg RewardDistConfig) *RewardDistributionConsumer {
	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = 5 * time.Minute
	}
	if cfg.TTLDuration == 0 {
		cfg.TTLDuration = 24 * time.Hour
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &RewardDistributionConsumer{
		client:     client,
		walletRepo: walletRepo,
		cfg:        cfg,
		stopCtx:    ctx,
		stopCancel: cancel,
		distQueue:  make(chan dto.RewardDistributionEvent, cfg.QueueSize),
		retryQueue: make(chan dto.RewardDistributionEvent, cfg.QueueSize/2),
		stats:      make(map[string]RewardStats),
		processed:  make(map[string]ProcessedReward),
	}
}

func (rdc *RewardDistributionConsumer) Start() error {
	rdc.mu.Lock()

	if rdc.isRunning {
		rdc.mu.Unlock()
		return nil
	}

	rdc.isRunning = true
	rdc.mu.Unlock()

	logger.LogInfo("Starting reward distribution consumer")

	// Start worker pool before consuming messages
	for i := 0; i < rdc.cfg.DistWorkers; i++ {
		go rdc.distWorker(i)
	}

	// Start cleanup worker
	go rdc.cleanupWorker()

	// Start retry worker
	go rdc.retryPendingRewardsWorker()

	return rdc.client.Consume(
		rabbitmq.RewardDistributionQueue,
		rdc.cfg.ConsumerWorkers,
		rdc.handleMessage,
	)
}

func (rdc *RewardDistributionConsumer) handleMessage(msg amqp091.Delivery) {
	var event dto.RewardDistributionEvent

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logger.LogError("Failed to unmarshal message", err)
		if err := msg.Nack(false, false); err != nil {
			logger.LogError("Failed to nack message", err)
		}
		return
	}

	// Proper queue full handling with nack for retry
	select {
	case rdc.distQueue <- event:
		logger.LogInfo(fmt.Sprintf("Enqueued reward distribution event for block %d, miner %s", event.BlockNumber, event.MinerAddress))
		if err := msg.Ack(false); err != nil {
			logger.LogError("Failed to ack message", err)
		}
	default:
		logger.LogInfo(fmt.Sprintf("Distribution queue full for block %d, nacking for retry", event.BlockNumber))
		// Nack without requeue to send to dead letter queue, will be retried by retry worker
		if err := msg.Nack(false, true); err != nil {
			logger.LogError("Failed to nack message", err)
		}
	}
}

// distWorker processes rewards from the distribution queue
func (rdc *RewardDistributionConsumer) distWorker(id int) {
	logger.LogInfo(fmt.Sprintf("Distribution worker %d started", id))

	for {
		select {
		case <-rdc.stopCtx.Done():
			logger.LogInfo(fmt.Sprintf("Distribution worker %d stopping", id))
			return
		case event, ok := <-rdc.distQueue:
			if !ok {
				logger.LogInfo(fmt.Sprintf("Distribution worker %d stopping - queue closed", id))
				return
			}
			// ✅ FIX #3: Use stopCtx instead of Background
			ctx, cancel := context.WithTimeout(rdc.stopCtx, rdc.cfg.ProcessingTimeout)
			rdc.process(ctx, event)
			cancel()
		}
	}
}

// process handles the actual reward distribution with transactional persistence
func (rdc *RewardDistributionConsumer) process(ctx context.Context, event dto.RewardDistributionEvent) {
	// Check context timeout before processing
	select {
	case <-ctx.Done():
		logger.LogInfo(fmt.Sprintf("Context cancelled for block %d", event.BlockNumber))
		return
	default:
	}

	// Check idempotency - prevent duplicate distribution
	if rdc.isProcessed(event.BlockNumber, event.MinerAddress) {
		logger.LogInfo(fmt.Sprintf("Reward already distributed for block %d, miner %s (skipping)", event.BlockNumber, event.MinerAddress))
		return
	}

	// Implement transactional wallet update with persistence
	tx, err := rdc.walletRepo.BeginTx()
	if err != nil {
		logger.LogError("Failed to begin transaction, retrying", err)
		rdc.enqueueForRetry(event)
		return
	}

	// Lock wallet for update to prevent race conditions
	wallets, err := rdc.walletRepo.GetMultipleByAddressWithTx(tx, []string{event.MinerAddress})
	if err != nil {
		tx.Rollback()
		logger.LogError("Failed to get miner wallet, retrying", err)
		rdc.enqueueForRetry(event)
		return
	}

	var minerWallet *models.UserWallet

	if len(wallets) > 0 {
		minerWallet = &wallets[0]
	} else {
		// Create wallet if not exists
		if err := rdc.walletRepo.UpsertEmptyIfNotExistsWithTx(tx, event.MinerAddress); err != nil {
			tx.Rollback()
			logger.LogError("Failed to create wallet for miner", err)
			rdc.enqueueForRetry(event)
			return
		}
		minerWallet = &models.UserWallet{
			UserAddress: event.MinerAddress,
			YTEBalance:  0,
		}
	}

	prev := minerWallet.YTEBalance
	newBalance := minerWallet.YTEBalance + event.MinerReward

	// Update wallet balance in database
	if err := rdc.walletRepo.UpdateWalletWithTx(tx, event.MinerAddress, newBalance); err != nil {
		tx.Rollback()
		logger.LogError("Failed to update wallet balance, retrying", err)
		rdc.enqueueForRetry(event)
		return
	}

	// Record wallet history for audit
	history := models.WalletHistory{
		UserAddress:   event.MinerAddress,
		ChangeType:    "MINING",
		Amount:        event.MinerReward,
		BalanceBefore: prev,
		BalanceAfter:  newBalance,
		Description:   stringPtr(fmt.Sprintf("Block reward for block #%d", event.BlockNumber)),
		ReferenceID:   stringPtr(fmt.Sprintf("BLOCK_%d", event.BlockNumber)),
	}

	if err := rdc.walletRepo.InsertHistoryWithTx(tx, history); err != nil {
		tx.Rollback()
		logger.LogError("Failed to record wallet history, retrying", err)
		rdc.enqueueForRetry(event)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		logger.LogError("Failed to commit transaction, retrying", err)
		rdc.enqueueForRetry(event)
		return
	}

	// Mark as processed (idempotency)
	rdc.markProcessed(event.BlockNumber, event.MinerAddress)

	// Update stats
	rdc.updateStats(event)

	// Log result
	rdc.logResult(event, prev, newBalance)
}

// enqueueForRetry adds failed events back to retry queue
func (rdc *RewardDistributionConsumer) enqueueForRetry(event dto.RewardDistributionEvent) {
	select {
	case rdc.retryQueue <- event:
		logger.LogInfo(fmt.Sprintf("Queued for retry: block %d, miner %s", event.BlockNumber, event.MinerAddress))
	default:
		logger.LogInfo(fmt.Sprintf("Retry queue full, dropping reward for block %d (will be lost)", event.BlockNumber))
	}
}

// retryPendingRewardsWorker periodically retries failed reward distributions
func (rdc *RewardDistributionConsumer) retryPendingRewardsWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	retryAttempts := make(map[string]int)
	maxRetries := 5

	logger.LogInfo("Retry worker started")

	for {
		select {
		case <-rdc.stopCtx.Done():
			logger.LogInfo("Retry worker stopping")
			return
		case <-ticker.C:
			retryBatch := []dto.RewardDistributionEvent{}
			// Drain retry queue without blocking
			for len(retryBatch) < 100 {
				select {
				case event := <-rdc.retryQueue:
					retryBatch = append(retryBatch, event)
				default:
					break
				}
			}

			for _, event := range retryBatch {
				key := fmt.Sprintf("%d_%s", event.BlockNumber, event.MinerAddress)
				attempts := retryAttempts[key]

				if attempts >= maxRetries {
					logger.LogInfo(fmt.Sprintf("Max retries exceeded for block %d, miner %s, dropping", event.BlockNumber, event.MinerAddress))
					delete(retryAttempts, key)
					continue
				}

				ctx, cancel := context.WithTimeout(rdc.stopCtx, rdc.cfg.ProcessingTimeout)
				rdc.process(ctx, event)
				cancel()

				retryAttempts[key]++
			}
		}
	}
}

// cleanupWorker periodically cleans up old processed reward entries to prevent memory leaks
func (rdc *RewardDistributionConsumer) cleanupWorker() {
	ticker := time.NewTicker(rdc.cfg.CleanupInterval)
	defer ticker.Stop()

	logger.LogInfo("Cleanup worker started")

	for {
		select {
		case <-rdc.stopCtx.Done():
			logger.LogInfo("Cleanup worker stopping")
			return
		case <-ticker.C:
			rdc.cleanupProcessedRewards()
		}
	}
}

// cleanupProcessedRewards removes expired entries from processed map
func (rdc *RewardDistributionConsumer) cleanupProcessedRewards() {
	rdc.processedMu.Lock()
	defer rdc.processedMu.Unlock()

	now := time.Now().Unix()
	removed := 0

	for key, entry := range rdc.processed {
		if now > entry.ExpiresAt {
			delete(rdc.processed, key)
			removed++
		}
	}

	if removed > 0 {
		logger.LogInfo(fmt.Sprintf("Cleaned up %d expired processed reward entries", removed))
	}
}

// isProcessed checks if a reward has already been distributed (idempotency check)
func (rdc *RewardDistributionConsumer) isProcessed(blockNumber int, minerAddress string) bool {
	rdc.processedMu.RLock()
	defer rdc.processedMu.RUnlock()

	key := fmt.Sprintf("%d_%s", blockNumber, minerAddress)
	processed, exists := rdc.processed[key]

	if !exists {
		return false
	}

	// Check if not expired
	return time.Now().Unix() <= processed.ExpiresAt
}

// markProcessed marks a reward as successfully distributed
func (rdc *RewardDistributionConsumer) markProcessed(blockNumber int, minerAddress string) {
	rdc.processedMu.Lock()
	defer rdc.processedMu.Unlock()

	key := fmt.Sprintf("%d_%s", blockNumber, minerAddress)
	rdc.processed[key] = ProcessedReward{
		BlockNumber: blockNumber,
		Timestamp:   time.Now().Unix(),
		ExpiresAt:   time.Now().Add(rdc.cfg.TTLDuration).Unix(),
	}
}

// updateStats updates metrics for reward distribution
func (rdc *RewardDistributionConsumer) updateStats(event dto.RewardDistributionEvent) {
	// Use RLock for reads, Lock for writes
	rdc.statsMu.Lock()
	defer rdc.statsMu.Unlock()

	s := rdc.stats[event.MinerAddress]
	s.MinerAddress = event.MinerAddress
	s.TotalRewards += event.MinerReward
	s.TotalUSDValue += event.MinerUSDValue
	s.RewardCount++
	s.LastRewardAt = event.Timestamp
	s.LastBlockNumber = event.BlockNumber

	rdc.stats[event.MinerAddress] = s
}

// logResult logs the successful reward distribution
func (rdc *RewardDistributionConsumer) logResult(event dto.RewardDistributionEvent, prev, now float64) {
	bd := event.RewardBreakdown

	logger.LogInfo(
		fmt.Sprintf(
			"DISTRIBUTED - Miner: %s, Block: #%d, Prev: %.8f YTE, Reward: %.8f YTE, New: %.8f YTE, USD: $%.2f, Breakdown: Block=%.8f | TxFee=%.8f | Bonus=%.8f",
			event.MinerAddress,
			event.BlockNumber,
			prev,
			event.MinerReward,
			now,
			event.MinerUSDValue,
			bd.BlockReward,
			bd.TransactionFees,
			bd.BonusReward,
		))
}

// GetStats retrieves stats for a specific miner address
func (rdc *RewardDistributionConsumer) GetStats(address string) RewardStats {
	rdc.statsMu.RLock()
	defer rdc.statsMu.RUnlock()
	return rdc.stats[address]
}

// GetAllStats retrieves all reward distribution stats
func (rdc *RewardDistributionConsumer) GetAllStats() map[string]RewardStats {
	rdc.statsMu.RLock()
	defer rdc.statsMu.RUnlock()

	statsCopy := make(map[string]RewardStats)
	for addr, stats := range rdc.stats {
		statsCopy[addr] = stats
	}

	return statsCopy
}

// Stop gracefully stops the reward distribution consumer
func (rdc *RewardDistributionConsumer) Stop() {
	rdc.mu.Lock()
	defer rdc.mu.Unlock()

	if !rdc.isRunning {
		return
	}

	logger.LogInfo("Stopping consumer")
	rdc.stopCancel()
	close(rdc.distQueue)
	close(rdc.retryQueue)
	rdc.isRunning = false
	logger.LogInfo("Consumer stopped")
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
