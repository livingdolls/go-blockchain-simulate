package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// MarketPricingConsumer - Konsumer untuk market pricing events
// Mendengarkan queue market.pricing dan memproses pricing updates
// Multiple consumer instances bisa berjalan untuk scalability
type MarketPricingConsumer struct {
	client            *rabbitmq.Client
	marketRepo        repository.MarketRepository
	publisherWS       *publisher.PublisherWS
	mu                sync.Mutex
	isRunning         bool
	stopChan          chan struct{}
	workerCount       int
	processingTimeout time.Duration
	priceCache        map[int64]float64 // Cache harga terakhir per block ID
	priceCacheMu      sync.RWMutex
}

func NewMarketPricingConsumer(
	client *rabbitmq.Client,
	marketRepo repository.MarketRepository,
	publisherWS *publisher.PublisherWS,
	workerCount int,
) *MarketPricingConsumer {
	return &MarketPricingConsumer{
		client:            client,
		marketRepo:        marketRepo,
		publisherWS:       publisherWS,
		stopChan:          make(chan struct{}),
		workerCount:       workerCount,
		processingTimeout: 30 * time.Second,
		priceCache:        make(map[int64]float64),
	}
}

// Start - Memulai konsumer market pricing
// - Non-blocking: market updates tidak delay block generation
// - Multi-consumer: bisa scale dengan menambah worker count
// - Idempotent: safe untuk retry (menggunakan block_id sebagai key)
func (m *MarketPricingConsumer) Start() error {
	m.mu.Lock()

	if m.isRunning {
		m.mu.Unlock()
		return nil
	}

	m.isRunning = true
	m.mu.Unlock()

	logger.LogInfo("Starting market pricing consumer")

	if err := m.LoadInitialPriceCache(1000); err != nil {
		logger.LogWarn("Failed to load initial price cache", zap.Error(err))
	}

	return m.client.Consume(
		rabbitmq.MarketPricingQueue,
		m.workerCount,
		m.handleMessage,
	)
}

// handleMessage - Handler untuk setiap pesan market pricing
// strategi:
// - Parse pesan
// - Update price cache untuk tracking perubahab
// - Kirim update ke WebSocket clients
// - store ke database untuk historical data
// - ack message jika sukses
func (m *MarketPricingConsumer) handleMessage(msg amqp091.Delivery) {
	defer func() {
		if err := msg.Ack(false); err != nil {
			logger.LogError("Failed to ack message", err)
		}
	}()

	var event dto.MarketPricingEvent

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logger.LogError("Failed to unmarshal message", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.processingTimeout)
	defer cancel()

	// update price cache dan hitung perubahan harga
	m.priceCacheMu.Lock()
	previousPrice := m.priceCache[int64(event.BlockNumber)-1]
	m.priceCache[int64(event.BlockNumber)] = event.Price
	priceChage := event.Price - previousPrice
	priceChangePercent := 0.0

	if previousPrice > 0 {
		priceChangePercent = (priceChage / previousPrice) * 100
	}

	m.priceCacheMu.Unlock()

	// kirim update ke WebSocket clients
	if m.publisherWS != nil {
		priceUpdate := dto.PriceUpdate{
			BlockID:            event.BlockID,
			BlockNumber:        event.BlockNumber,
			Price:              event.Price,
			PriceChange:        priceChage,
			PriceChangePercent: priceChangePercent,
			Liquidity:          event.Liquidity,
			BuyVolume:          event.BuyVolume,
			SellVolume:         event.SellVolume,
			TxCount:            event.TxCount,
			Timestamp:          event.Timestamp,
			MinerAddress:       event.MinerAddress,
		}

		m.publisherWS.Publish(entity.EventMarketUpdate, priceUpdate)
	}

	// simpan historical data ke database
	go func() {
		if err := m.monitorMarketMetrics(ctx, event); err != nil {
			logger.LogError("Failed to store market snapshot", err)
		}
	}()
}

// monitorMarketMetrics - Monitor dan simpan metrik market ke database
func (m *MarketPricingConsumer) monitorMarketMetrics(ctx context.Context, event dto.MarketPricingEvent) error {
	tick, err := m.marketRepo.GetTickByBlockID(event.BlockID)
	if err != nil {
		return err
	}

	const priceEpsilon = 1e-8

	// validate consistency data
	if !floatEqual(tick.Price, event.Price, priceEpsilon) {
		logger.LogWarn("Data inconsistency detected",
			zap.Int64("block_id", event.BlockID),
			zap.Float64("tick_price", tick.Price),
			zap.Float64("event_price", event.Price),
		)
	}

	// calculate price metrics untuk monitoring
	m.priceCacheMu.RLock()
	previousPirce := m.priceCache[int64(event.BlockNumber)-1]
	m.priceCacheMu.RUnlock()

	priceChange := event.Price - previousPirce
	priceChangePercent := 0.0

	if previousPirce > 0 {
		priceChangePercent = (priceChange / previousPirce) * 100
	}

	// detect significant price changes
	if priceChangePercent > 5.0 || priceChangePercent < -5.0 {
		logger.LogWarn("Significant price movement detected",
			zap.Int("block_number", event.BlockNumber),
			zap.Float64("price_change_percent", priceChangePercent),
			zap.Float64("previous_price", previousPirce),
			zap.Float64("current_price", event.Price),
		)
	}

	// log successful verification with metrics
	logger.LogInfo("Market snapshot stored",
		zap.Int("block_number", event.BlockNumber),
		zap.Float64("price", event.Price),
		zap.Float64("price_change", priceChange),
		zap.Float64("price_change_percent", priceChangePercent),
		zap.Float64("liquidity", event.Liquidity),
		zap.Int("tx_count", event.TxCount),
	)

	m.triggerPriceAlerts(ctx, event, priceChangePercent)

	return nil
}

func (m *MarketPricingConsumer) triggerPriceAlerts(ctx context.Context, event dto.MarketPricingEvent, priceChangePercent float64) {
	const (
		MinorAlert    = 3.0
		MajorAlert    = 5.0
		CriticalAlert = 10.0
	)

	absChange := priceChangePercent
	if absChange < 0 {
		absChange = -absChange
	}

	if absChange >= CriticalAlert {
		logger.LogWarn("CRITICAL ALERT: Extreme price change",
			zap.Float64("price_change_percent", priceChangePercent),
			zap.Int("block_number", event.BlockNumber),
			zap.Float64("previous_price", event.Price-(event.Price*float64(priceChangePercent)/100)),
			zap.Float64("current_price", event.Price),
		)
	} else if absChange >= MajorAlert {
		logger.LogWarn("MAJOR ALERT: Significant price change",
			zap.Float64("price_change_percent", priceChangePercent),
			zap.Int("block_number", event.BlockNumber),
			zap.Float64("previous_price", event.Price-(event.Price*float64(priceChangePercent)/100)),
			zap.Float64("current_price", event.Price),
		)
	} else if absChange >= MinorAlert {
		logger.LogInfo("Minor price change alert",
			zap.Float64("price_change_percent", priceChangePercent),
			zap.Int("block_number", event.BlockNumber),
			zap.Float64("previous_price", event.Price-(event.Price*float64(priceChangePercent)/100)),
			zap.Float64("current_price", event.Price),
		)
	}
}

func (m *MarketPricingConsumer) GetPriceHistory(limit, offset int) ([]dto.PriceUpdate, error) {
	ticks, err := m.marketRepo.GetVolumeHistory(limit, offset)
	if err != nil {
		return nil, err
	}

	var updates []dto.PriceUpdate

	for i, tick := range ticks {
		priceChange := 0.0
		priceChangePercent := 0.0

		if i < len(ticks)-1 {
			previousPrice := ticks[i+1].Price
			if previousPrice > 0 {
				priceChange = tick.Price - previousPrice
				priceChangePercent = (priceChange / previousPrice) * 100
			}
		}

		updates = append(updates, dto.PriceUpdate{
			BlockID:            tick.BlockID,
			BlockNumber:        int(tick.BlockID),
			Price:              tick.Price,
			PriceChange:        priceChange,
			PriceChangePercent: priceChangePercent,
			Liquidity:          0,
			BuyVolume:          tick.BuyVolume,
			SellVolume:         tick.SellVolume,
			TxCount:            tick.TxCount,
			Timestamp:          tick.CreatedAt,
		})
	}

	return updates, nil
}

func (m *MarketPricingConsumer) LoadInitialPriceCache(limit int) error {
	ticks, err := m.marketRepo.GetVolumeHistory(limit, 0)
	if err != nil {
		return err
	}

	m.priceCacheMu.Lock()
	defer m.priceCacheMu.Unlock()

	for _, tick := range ticks {
		m.priceCache[int64(tick.BlockID)] = tick.Price
	}

	logger.LogInfo("Loaded initial price cache",
		zap.Int("blocks_loaded", len(ticks)),
	)
	return nil
}

func (m *MarketPricingConsumer) Stop() {
	m.mu.Lock()

	if !m.isRunning {
		m.mu.Unlock()
		return
	}

	m.isRunning = false
	m.mu.Unlock()

	logger.LogInfo("Stopping market pricing consumer")
	close(m.stopChan)
	logger.LogInfo("Market pricing consumer stopped")
}

func (m *MarketPricingConsumer) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isRunning
}

func floatEqual(a, b, epsilon float64) bool {
	if a == b {
		return true
	}

	absA := a
	if absA < 0 {
		absA = -absA
	}

	absB := b
	if absB < 0 {
		absB = -absB
	}

	maxAbs := absA
	if absB > maxAbs {
		maxAbs = absB
	}

	// relative tolerance
	if maxAbs > 1.0 {
		return (a-b)/(maxAbs) < epsilon
	}

	return a-b < epsilon
}
