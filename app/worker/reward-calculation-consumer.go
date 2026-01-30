package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/logger"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type RewardEngineConfig struct {
	ConsumerWorkers   int
	CalcWorkers       int
	ProcessingTimeout time.Duration
	QueueSize         int
}

type RewardCalculationConsumer struct {
	client          *rabbitmq.Client
	rewardPublisher services.RewardPublisher
	cfg             RewardEngineConfig
	mu              sync.Mutex
	isRunning       bool
	stopCtx         context.Context
	stopCancel      context.CancelFunc

	calcQueue   chan dto.RewardCalculationEvent
	processedMu sync.RWMutex
	processed   map[int64]ProcessedBlock

	// retry mechanism for pending blocks
	pendingBlocksMu sync.RWMutex
	pendingBlocks   map[int64]dto.RewardCalculationEvent

	// metrics
	metricsMu           sync.RWMutex
	processedCount      int64
	failedCount         int64
	retriedCount        int64
	totalRewardsIssued  float64
	totalUSDValueIssued float64
}

type ProcessedBlock struct {
	ProcessedAt int64
}

func NewRewardCalculationConsumer(rmqClient *rabbitmq.Client, rewardPublisher services.RewardPublisher, cfg RewardEngineConfig) *RewardCalculationConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &RewardCalculationConsumer{
		client:          rmqClient,
		rewardPublisher: rewardPublisher,
		cfg:             cfg,
		stopCtx:         ctx,
		stopCancel:      cancel,
		calcQueue:       make(chan dto.RewardCalculationEvent, cfg.QueueSize),
		processed:       make(map[int64]ProcessedBlock),
		pendingBlocks:   make(map[int64]dto.RewardCalculationEvent),
	}
}

func (r *RewardCalculationConsumer) Start() error {
	r.mu.Lock()
	if r.isRunning {
		r.mu.Unlock()
		return nil
	}

	r.isRunning = true
	r.mu.Unlock()

	logger.LogInfo("Starting reward calculation consumer")

	for i := 0; i < r.cfg.CalcWorkers; i++ {
		go r.calcWorker(i)
	}

	// start retry worker for handle pending blocks
	go r.retryPendingBlocksWorker()

	// start cleanup worker for processed blocks
	go r.cleanUpWorker()

	return r.client.Consume(
		rabbitmq.RewardCalculationQueue,
		r.cfg.ConsumerWorkers,
		r.handleMessage,
	)
}

func (r *RewardCalculationConsumer) handleMessage(msg amqp091.Delivery) {
	defer func() {
		if err := msg.Ack(false); err != nil {
			logger.LogError("Failed to ack message", err)
		}
	}()

	var event dto.RewardCalculationEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logger.LogError("Failed to unmarshal message", err)
		return
	}

	if r.isProcessed(event.BlockID) {
		logger.LogInfo(fmt.Sprintf("Block %d already processed, skipping", event.BlockNumber))
		return
	}

	select {
	case r.calcQueue <- event:
		logger.LogInfo(fmt.Sprintf("Enqueued reward calculation for block %d", event.BlockNumber))
	default:
		// durbality retry
		r.pendingBlocksMu.Lock()
		r.pendingBlocks[event.BlockID] = event
		r.pendingBlocksMu.Unlock()

		logger.LogInfo(fmt.Sprintf("Calculation queue full, storing block %d for retry", event.BlockNumber))
	}
}

func (r *RewardCalculationConsumer) calcWorker(workerID int) {
	logger.LogInfo(fmt.Sprintf("Calculation worker %d started", workerID))

	for {
		select {
		case <-r.stopCtx.Done():
			logger.LogInfo(fmt.Sprintf("Calculation worker %d stopping", workerID))
			return

		case event, ok := <-r.calcQueue:
			if !ok {
				logger.LogInfo(fmt.Sprintf("Calculation worker %d stopping - queue closed", workerID))
				return
			}

			ctx, cancel := context.WithTimeout(r.stopCtx, r.cfg.ProcessingTimeout)
			r.process(ctx, event)
			cancel()
		}
	}
}

func (r *RewardCalculationConsumer) retryPendingBlocksWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	logger.LogInfo("Starting retry pending blocks worker")

	for {
		select {
		case <-r.stopCtx.Done():
			logger.LogInfo("Retry pending blocks worker stopping")
			return
		case <-ticker.C:
			r.retryPendingBlocks()
		}
	}
}

func (r *RewardCalculationConsumer) retryPendingBlocks() {
	r.pendingBlocksMu.Lock()
	pendingList := make([]dto.RewardCalculationEvent, 0)
	pendingIDs := make([]int64, 0)

	for blockID, event := range r.pendingBlocks {
		pendingList = append(pendingList, event)
		pendingIDs = append(pendingIDs, blockID)
	}

	for i, event := range pendingList {
		select {
		case <-r.stopCtx.Done():
			return
		case r.calcQueue <- event:
			// retry successful, remove from pending
			r.pendingBlocksMu.Lock()
			delete(r.pendingBlocks, pendingIDs[i])
			r.pendingBlocksMu.Unlock()
			r.recordRetry()
			logger.LogInfo(fmt.Sprintf("Retry success for block %d", event.BlockNumber))
		default:
			// queue full, will retry later
			logger.LogInfo(fmt.Sprintf("Calculation queue full, will retry block %d later", event.BlockNumber))
		}
	}
}

func (r *RewardCalculationConsumer) addPending(event dto.RewardCalculationEvent) {
	r.pendingBlocksMu.Lock()
	r.pendingBlocks[event.BlockID] = event
	r.pendingBlocksMu.Unlock()
}

func (r *RewardCalculationConsumer) cleanUpWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	logger.LogInfo("Cleanup worker started")

	for {
		select {
		case <-r.stopCtx.Done():
			logger.LogInfo("Cleanup worker stopping")
			return
		case <-ticker.C:
			r.cleanupOldEntries()
		}
	}
}

func (r *RewardCalculationConsumer) cleanupOldEntries() {
	r.processedMu.Lock()
	defer r.processedMu.Unlock()

	now := time.Now().Unix()
	ttlSeconds := int64(3600)

	for id, entry := range r.processed {
		if now-entry.ProcessedAt > ttlSeconds {
			delete(r.processed, id)
			logger.LogInfo(fmt.Sprintf("Cleaned up processed entry for block ID %d", id))
		}
	}
}

func (r *RewardCalculationConsumer) process(ctx context.Context, event dto.RewardCalculationEvent) {
	select {
	case <-ctx.Done():
		logger.LogInfo(fmt.Sprintf("Context cancelled for block %d", event.BlockNumber))
		r.recordFailure()
		r.addPending(event)
		return
	default:
	}

	// calculate bonus reward
	bonusReward := r.calculationBonusReward(event)

	breakdown := dto.RewardBreakDown{
		BlockReward:     event.BlockReward,
		TransactionFees: event.TotalTransactionFee,
		BonusReward:     bonusReward,
	}

	breakdown.TotalReward = breakdown.BlockReward + breakdown.TransactionFees + breakdown.BonusReward
	breakdown.EstimatedUSDValue = breakdown.TotalReward * event.MarketPrice

	//validate calculations
	if breakdown.TotalReward < 0 {
		logger.LogInfo(fmt.Sprintf(
			"ERROR: Negative total reward for block #%d: %.8f",
			event.BlockNumber,
			breakdown.TotalReward,
		))
		r.recordFailure()
		r.addPending(event)
		return
	}

	if breakdown.EstimatedUSDValue < 0 {
		logger.LogInfo(fmt.Sprintf(
			"ERROR: Negative USD value for block #%d: %.2f",
			event.BlockNumber,
			breakdown.EstimatedUSDValue,
		))
		r.recordFailure()
		r.addPending(event)
		return
	}

	distributionEvent := dto.RewardDistributionEvent{
		BlockID:         event.BlockID,
		BlockNumber:     event.BlockNumber,
		MinerAddress:    event.MinerAddress,
		MinerReward:     breakdown.TotalReward,
		MinerUSDValue:   breakdown.EstimatedUSDValue,
		RewardBreakdown: breakdown,
		Timestamp:       time.Now().Unix(),
	}

	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Publish reward distribution event
	if err := r.rewardPublisher.PublishRewardDistribution(publishCtx, distributionEvent); err != nil {
		logger.LogError("Failed to publish reward distribution", err)
		r.recordFailure()
		r.addPending(event)
		return
	}

	r.markProcessed(event.BlockID)
	r.recordSuccess(breakdown.TotalReward, breakdown.EstimatedUSDValue)

	logger.LogInfo(fmt.Sprintf("Completed reward calculation for block %d", event.BlockNumber))
}

// bonus reward calculation rules
// - 0.1% bonus per transaction, max 5% (10 tx = 1%, 50 tx = 5%)
// - 0.5% bonus if miner address earned fees (active miner)
// - max total bonus: 10% of block reward
func (r *RewardCalculationConsumer) calculationBonusReward(event dto.RewardCalculationEvent) float64 {
	bonusPrecentage := 0.0

	// tx count
	if event.TransactionCount >= 10 {
		txBonus := float64(event.TransactionCount) / 10.0
		if txBonus > 5.0 {
			txBonus = 5.0
		}
		bonusPrecentage += txBonus
	}

	// active miner binus
	if event.MinerAddress != "" && event.TotalTransactionFee > 0 {
		bonusPrecentage += 0.5
	}

	// cap total bonus at 10% of block reward
	if bonusPrecentage > 10.0 {
		bonusPrecentage = 10.0
	}

	return event.BlockReward * (bonusPrecentage / 100.0)
}

// indempotentcy
func (r *RewardCalculationConsumer) isProcessed(blockID int64) bool {
	r.processedMu.RLock()
	defer r.processedMu.RUnlock()
	_, exits := r.processed[blockID]
	return exits
}

func (r *RewardCalculationConsumer) markProcessed(blockID int64) {
	r.processedMu.Lock()
	r.processed[blockID] = ProcessedBlock{ProcessedAt: time.Now().Unix()}
	r.processedMu.Unlock()
}

func (r *RewardCalculationConsumer) recordSuccess(totalReward, usdValue float64) {
	r.metricsMu.Lock()
	r.processedCount++
	r.totalRewardsIssued += totalReward
	r.totalUSDValueIssued += usdValue
	r.metricsMu.Unlock()
}

func (r *RewardCalculationConsumer) recordFailure() {
	r.metricsMu.Lock()
	r.failedCount++
	r.metricsMu.Unlock()
}

func (r *RewardCalculationConsumer) recordRetry() {
	r.metricsMu.Lock()
	r.retriedCount++
	r.metricsMu.Unlock()
}

func (r *RewardCalculationConsumer) GetMetrics() map[string]interface{} {
	r.metricsMu.RLock()
	processedCount := r.processedCount
	failedCount := r.failedCount
	retriedCount := r.retriedCount
	totalRewardsIssued := r.totalRewardsIssued
	totalUSDValueIssued := r.totalUSDValueIssued
	r.metricsMu.RUnlock()

	r.pendingBlocksMu.RLock()
	pendingCount := len(r.pendingBlocks)
	r.pendingBlocksMu.RUnlock()

	avgRewardPerBlock := 0.0
	if processedCount > 0 {
		avgRewardPerBlock = totalRewardsIssued / float64(processedCount)
	}

	return map[string]interface{}{
		"processed_count":        processedCount,
		"failed_count":           failedCount,
		"retried_count":          retriedCount,
		"pending_blocks_count":   pendingCount,
		"total_rewards_issued":   totalRewardsIssued,
		"total_usd_value_issued": totalUSDValueIssued,
		"avg_reward_per_block":   avgRewardPerBlock,
	}
}

func (r *RewardCalculationConsumer) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isRunning {
		return
	}

	logger.LogInfo("Stopping reward calculation consumer")

	// Cancel context terlebih dahulu (triggers all workers to stop)
	r.stopCancel()

	// Close channel setelah context cancelled
	close(r.calcQueue)

	r.isRunning = false

	logger.LogInfo("Reward calculation consumer stopped")
}
