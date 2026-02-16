package worker

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// MarketVolumeConsumer - Konsumer untuk market volume updates
// Digunakan untuk:
// - Historical volume tracking
// - Volume analysis dan pattern detection
// - Backup/archive purposes
// - Multiple independent instances bisa berjalan
type MarketVolumeConsumer struct {
	client     *rabbitmq.Client
	marketRepo repository.MarketRepository

	ctx    context.Context
	cancel context.CancelFunc

	workerCount int

	running atomic.Bool
	wg      sync.WaitGroup

	processedBlocks sync.Map

	maxRetention    time.Duration
	cleanupInterval time.Duration

	stats   VolumeStats
	statsMu sync.RWMutex
}

type VolumeStats struct {
	TotalBlocks   int64
	TotalBuyVol   float64
	TotalSellVol  float64
	AvgBuyVol     float64
	AvgSellVol    float64
	HighestBuyVol float64
	LowestBuyVol  float64
	LastUpdated   time.Time
}

func NewMarketVolumeConsumer(
	client *rabbitmq.Client,
	marketRepo repository.MarketRepository,
	workerCount int,
) *MarketVolumeConsumer {
	ctx, cancel := context.WithCancel(context.Background())

	return &MarketVolumeConsumer{
		client:          client,
		marketRepo:      marketRepo,
		ctx:             ctx,
		cancel:          cancel,
		workerCount:     workerCount,
		maxRetention:    1 * time.Hour,
		cleanupInterval: 15 * time.Minute,
	}
}

func (m *MarketVolumeConsumer) Start() error {
	if !m.running.CompareAndSwap(false, true) {
		logger.LogWarn("Market volume consumer is already running")
		return nil
	}

	logger.LogInfo("Starting market volume consumer")

	// start cleanup worker
	m.wg.Add(1)
	go m.cleanupLoop()

	// start consuming messages
	return m.client.ConsumeWithContext(
		m.ctx,
		rabbitmq.MarketVolumeQueue,
		m.workerCount,
		m.handleMessage,
	)
}

// stop gracefully
func (m *MarketVolumeConsumer) Stop() {
	if !m.running.CompareAndSwap(true, false) {
		return
	}

	logger.LogInfo("Stopping market volume consumer")

	m.cancel()
	m.wg.Wait()

	logger.LogInfo("Market volume consumer stopped")
}

// handle incoming messages
func (m *MarketVolumeConsumer) handleMessage(msg amqp091.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			logger.LogError("Panic in market volume consumer", errors.New("panic occurred"), zap.Any("recovered", r))
			msg.Nack(false, true)
		}
	}()

	var update dto.MarketVolumeUpdate

	if err := json.Unmarshal(msg.Body, &update); err != nil {
		logger.LogError("Failed to unmarshal market volume update", err)
		msg.Nack(false, false)
		return
	}

	// idempotency check
	if m.isProcessed(update.BlockID) {
		logger.LogInfo("Market volume update already processed for block", zap.Int64("blockID", update.BlockID))
		msg.Ack(false)
		return
	}

	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	// process
	err := m.processVolume(ctx, update)

	if err != nil {
		logger.LogError("Failed to process market volume update", err, zap.Int64("blockID", update.BlockID))
		msg.Nack(false, true)
		return
	}

	// mark as processed after successful processing
	m.markProcessed(update.BlockID)

	// update stats
	m.updateStats(update)

	msg.Ack(false)

	logger.LogInfo("Successfully processed market volume update", zap.Int64("blockID", update.BlockID), zap.Float64("buyVolume", update.BuyVolume), zap.Float64("sellVolume", update.SellVolume))
}

// process volume
func (m *MarketVolumeConsumer) processVolume(ctx context.Context, update dto.MarketVolumeUpdate) error {
	tick, err := m.marketRepo.GetTickByBlockID(update.BlockID)

	if err != nil {
		return err
	}

	if tick.BuyVolume != update.BuyVolume || tick.SellVolume != update.SellVolume {
		logger.LogWarn("Volume mismatch for block", zap.Int64("blockID", update.BlockID), zap.Float64("repoBuyVol", tick.BuyVolume), zap.Float64("repoSellVol", tick.SellVolume), zap.Float64("updateBuyVol", update.BuyVolume), zap.Float64("updateSellVol", update.SellVolume))
	}

	if update.BlockID%100 == 0 {
		return m.aggregate(ctx, update.BlockID)
	}

	return nil
}

// agregate
func (m *MarketVolumeConsumer) aggregate(ctx context.Context, blockID int64) error {
	start := blockID - 100

	ticks, err := m.marketRepo.GetVolumeBlockRange(start, blockID)
	if err != nil {
		return err
	}

	if len(ticks) == 0 {
		logger.LogWarn("No ticks found for aggregation", zap.Int64("startBlockID", start), zap.Int64("endBlockID", blockID))
		return nil
	}

	var buy, sell, price float64

	min := ticks[0].Price
	max := ticks[0].Price

	for _, t := range ticks {
		buy += t.BuyVolume
		sell += t.SellVolume
		price += t.Price

		if t.Price < min {
			min = t.Price
		}

		if t.Price > max {
			max = t.Price
		}
	}

	logger.LogInfo("Aggregated volume for block range", zap.Int64("startBlockID", start), zap.Int64("endBlockID", blockID), zap.Float64("totalBuyVol", buy), zap.Float64("totalSellVol", sell), zap.Float64("avgPrice", price/float64(len(ticks))), zap.Float64("minPrice", min), zap.Float64("maxPrice", max))

	return nil
}

// indemoptency check
func (m *MarketVolumeConsumer) isProcessed(blockID int64) bool {
	_, exists := m.processedBlocks.Load(blockID)
	return exists
}

func (m *MarketVolumeConsumer) markProcessed(blockID int64) {
	m.processedBlocks.Store(blockID, time.Now())
}

// stats
func (m *MarketVolumeConsumer) updateStats(update dto.MarketVolumeUpdate) {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()

	m.stats.TotalBlocks++

	m.stats.TotalBuyVol += update.BuyVolume
	m.stats.TotalSellVol += update.SellVolume

	m.stats.AvgBuyVol = m.stats.AvgBuyVol / float64(m.stats.TotalBlocks)
	m.stats.AvgSellVol = m.stats.AvgSellVol / float64(m.stats.TotalBlocks)

	if update.BuyVolume > m.stats.HighestBuyVol {
		m.stats.HighestBuyVol = update.BuyVolume
	}

	if m.stats.TotalBlocks == 1 || update.BuyVolume < m.stats.LowestBuyVol {
		m.stats.LowestBuyVol = update.BuyVolume
	}

	m.stats.LastUpdated = time.Now()
}

// cleanup Loop
func (m *MarketVolumeConsumer) cleanupLoop() {
	defer m.wg.Done()
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			m.processedBlocks.Range(func(key, value any) bool {
				t := value.(time.Time)
				if now.Sub(t) > m.maxRetention {
					m.processedBlocks.Delete(key)
				}

				return true
			})
		}
	}
}

// get stats
func (m *MarketVolumeConsumer) GetStats() VolumeStats {
	m.statsMu.RLock()
	defer m.statsMu.RUnlock()

	return m.stats
}
