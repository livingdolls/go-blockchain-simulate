package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

// MarketVolumeConsumer - Konsumer untuk market volume updates
// Digunakan untuk:
// - Historical volume tracking
// - Volume analysis dan pattern detection
// - Backup/archive purposes
// - Multiple independent instances bisa berjalan
type MarketVolumeConsumer struct {
	client            *rabbitmq.Client
	marketRepo        repository.MarketRepository
	mu                sync.Mutex
	isRunning         bool
	stopChan          chan struct{}
	workerCount       int
	processingTimeout time.Duration
	volumeStats       VolumeStats
	volumeStatsMu     sync.RWMutex
	processedBlocks   map[int64]bool
	processedBlocksMu sync.RWMutex
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
	return &MarketVolumeConsumer{
		client:            client,
		marketRepo:        marketRepo,
		stopChan:          make(chan struct{}),
		workerCount:       workerCount,
		processingTimeout: 30 * time.Second,
		volumeStats:       VolumeStats{},
		processedBlocks:   make(map[int64]bool),
	}
}

func (m *MarketVolumeConsumer) Start() error {
	m.mu.Lock()

	if m.isRunning {
		m.mu.Unlock()
		return nil
	}

	m.isRunning = true
	m.mu.Unlock()

	log.Println("[MARKET_VOLUME_CONSUMER] starting consumer...")

	return m.client.Consume(
		rabbitmq.MarketVolumeQueue,
		m.workerCount,
		m.handleMessage,
	)
}

func (m *MarketVolumeConsumer) handleMessage(msg amqp091.Delivery) {
	defer func() {
		if err := msg.Ack(false); err != nil {
			log.Printf("[MARKET_VOLUME_CONSUMER] failed to ack message: %v", err)
		}
	}()

	var volumeUpdate dto.MarketVolumeUpdate

	if err := json.Unmarshal(msg.Body, &volumeUpdate); err != nil {
		log.Printf("[MARKET_VOLUME_CONSUMER] failed to unmarshal message: %v", err)
		return
	}

	// check idempotency
	m.processedBlocksMu.RLock()
	if m.processedBlocks[volumeUpdate.BlockID] {
		m.processedBlocksMu.RUnlock()
		log.Printf("[MARKET_VOLUME_CONSUMER] duplicate block_id %d, skipping...", volumeUpdate.BlockID)
		return
	}

	m.processedBlocksMu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), m.processingTimeout)
	defer cancel()

	m.updateVolumeStats(volumeUpdate)

	go func() {
		if err := m.storeVolumeData(ctx, volumeUpdate); err != nil {
			log.Printf("[MARKET_VOLUME_CONSUMER] failed to store volume data: %v", err)
		}

		// mark as processed setelah sukses store
		m.processedBlocksMu.Lock()
		m.processedBlocks[volumeUpdate.BlockID] = true
		m.processedBlocksMu.Unlock()
	}()

	log.Printf(
		"[MARKET_VOLUME_CONSUMER] Processed volume update - Block #%d, Buy: %.2f, Sell: %.2f, Ratio: %.2f%%",
		volumeUpdate.BlockID,
		volumeUpdate.BuyVolume,
		volumeUpdate.SellVolume,
		volumeUpdate.VolumeRatio*100,
	)
}

func (m *MarketVolumeConsumer) updateVolumeStats(update dto.MarketVolumeUpdate) {
	m.volumeStatsMu.Lock()
	defer m.volumeStatsMu.Unlock()

	m.volumeStats.TotalBlocks++
	m.volumeStats.TotalBuyVol += update.BuyVolume
	m.volumeStats.TotalSellVol += update.SellVolume
	m.volumeStats.AvgBuyVol = m.volumeStats.TotalBuyVol / float64(m.volumeStats.TotalBlocks)
	m.volumeStats.AvgSellVol = m.volumeStats.TotalSellVol / float64(m.volumeStats.TotalBlocks)

	// update highest/lowest buy volume
	if update.BuyVolume > m.volumeStats.HighestBuyVol {
		m.volumeStats.HighestBuyVol = update.BuyVolume
	}

	if m.volumeStats.TotalBlocks == 1 || update.BuyVolume < m.volumeStats.LowestBuyVol {
		m.volumeStats.LowestBuyVol = update.BuyVolume
	}

	m.volumeStats.LastUpdated = time.Now()
}

func (m *MarketVolumeConsumer) storeVolumeData(ctx context.Context, update dto.MarketVolumeUpdate) error {
	tick, err := m.marketRepo.GetTickByBlockID(update.BlockID)
	if err != nil {
		return err
	}

	// validate consitency data
	if tick.BuyVolume != update.BuyVolume || tick.SellVolume != update.SellVolume {
		log.Printf(
			"[MARKET_VOLUME_CONSUMER] Warning: Data mismatch for block #%d - DB: (%.2f, %.2f), Event: (%.2f, %.2f)",
			update.BlockID,
			tick.BuyVolume, tick.SellVolume,
			update.BuyVolume, update.SellVolume,
		)
	}

	// Log successful storage
	log.Printf(
		"[MARKET_VOLUME_CONSUMER] Volume data verified - Block #%d, Buy: %.2f, Sell: %.2f, TxCount: %d, Timestamp: %d",
		tick.BlockID,
		tick.BuyVolume,
		tick.SellVolume,
		tick.TxCount,
		tick.CreatedAt,
	)

	//trigger aggregation
	if update.BlockID%100 == 0 {
		m.triggerAggregation(ctx, update.BlockID)
	}
	return nil
}

func (m *MarketVolumeConsumer) triggerAggregation(ctx context.Context, blockID int64) {
	blockRange := int64(100)
	startBlock := blockID - blockRange

	ticks, err := m.marketRepo.GetVolumeBlockRange(startBlock, blockID)
	if err != nil {
		log.Printf("[MARKET_VOLUME_CONSUMER] failed to get volume block range: %v", err)
		return
	}

	if len(ticks) == 0 {
		log.Printf("[MARKET_VOLUME_CONSUMER] no ticks found for aggregation in range %d - %d", startBlock, blockID)
		return
	}

	// calculate aggregated values
	var totalBuyVol, totalSellVol float64
	var avgPrice float64
	var minPrice, maxPrice float64 = ticks[0].Price, ticks[0].Price

	for i, tick := range ticks {
		totalBuyVol += tick.BuyVolume
		totalSellVol += tick.SellVolume
		avgPrice += tick.Price

		if tick.Price < minPrice {
			minPrice = tick.Price
		}

		if tick.Price > maxPrice {
			maxPrice = tick.Price
		}

		if i == len(ticks)-1 {
			log.Printf(
				"[MARKET_VOLUME_CONSUMER] Aggregated Volume Data - Blocks %d to %d: Total Buy: %.2f, Total Sell: %.2f, Avg Price: %.2f, Min Price: %.2f, Max Price: %.2f",
				startBlock,
				blockID,
				totalBuyVol,
				totalSellVol,
				avgPrice/float64(len(ticks)),
				minPrice,
				maxPrice,
			)
		}
	}
}

func (m *MarketVolumeConsumer) GetVolumeStats() VolumeStats {
	m.volumeStatsMu.RLock()
	defer m.volumeStatsMu.RUnlock()

	return m.volumeStats
}

func (m *MarketVolumeConsumer) GetHistoricalData(limit, offset int) ([]dto.MarketVolumeUpdate, error) {
	ticks, err := m.marketRepo.GetVolumeHistory(limit, offset)
	if err != nil {
		return nil, err
	}

	var updates []dto.MarketVolumeUpdate

	for _, tick := range ticks {
		netVolume := tick.BuyVolume - tick.SellVolume
		totalVolume := tick.BuyVolume + tick.SellVolume
		volumeRatio := 0.0

		if totalVolume > 0 {
			volumeRatio = tick.BuyVolume / totalVolume
		}

		updates = append(updates, dto.MarketVolumeUpdate{
			BlockID:     tick.BlockID,
			BuyVolume:   tick.BuyVolume,
			SellVolume:  tick.SellVolume,
			NetVolume:   netVolume,
			VolumeRatio: volumeRatio,
			TxCount:     tick.TxCount,
			Timestamp:   tick.CreatedAt,
		})
	}

	return updates, nil
}

func (m *MarketVolumeConsumer) Stop() {
	m.mu.Lock()

	if !m.isRunning {
		m.mu.Unlock()
		return
	}
	m.isRunning = false
	m.mu.Unlock()

	log.Println("[MARKET_VOLUME_CONSUMER] stopping consumer...")
	close(m.stopChan)
	log.Println("[MARKET_VOLUME_CONSUMER] consumer stopped.")
}

func (m *MarketVolumeConsumer) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isRunning
}
