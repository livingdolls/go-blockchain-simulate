package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/publisher"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
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

	log.Println("[MARKET_PRICING_CONSUMER] starting consumer...")

	if err := m.LoadInitialPriceCache(1000); err != nil {
		log.Printf("[MARKET_PRICING_CONSUMER] Warning: failed to initial load cache")
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
			log.Printf("[MARKET_PRICING_CONSUMER] failed to ack message: %v", err)
		}
	}()

	var event dto.MarketPricingEvent

	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("[MARKET_PRICING_CONSUMER] failed to unmarshal message: %v", err)
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
			log.Printf("[MARKET_PRICING_CONSUMER] failed to store market snapshot: %v", err)
		}
	}()
}

// monitorMarketMetrics - Monitor dan simpan metrik market ke database
func (m *MarketPricingConsumer) monitorMarketMetrics(ctx context.Context, event dto.MarketPricingEvent) error {
	tick, err := m.marketRepo.GetTickByBlockID(event.BlockID)
	if err != nil {
		return err
	}

	// validate consistency data
	if tick.Price != event.Price {
		log.Printf("[MARKET_PRICING_CONSUMER] data inconsistency for block %d: tick price %.2f != event price %.2f", event.BlockID, tick.Price, event.Price)
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
		log.Printf(
			"[MARKET_PRICING_CONSUMER] 🚨 Significant price movement detected! Block #%d: %.2f%% (%.2f -> %.2f)",
			event.BlockNumber,
			priceChangePercent,
			previousPirce,
			event.Price,
		)
	}

	// log successful verification with metrics
	log.Printf(
		"[MARKET_PRICING_CONSUMER] Market snapshot stored - Block #%d, Price: %.2f, Change: %.2f (%.2f%%), Liquidity: %.2f, TxCount: %d",
		event.BlockNumber,
		event.Price,
		priceChange,
		priceChangePercent,
		event.Liquidity,
		event.TxCount,
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
		log.Printf(
			"[MARKET_PRICING_CONSUMER] 🚨 CRITICAL ALERT: Price changed by %.2f%% at Block #%d (%.2f -> %.2f)",
			priceChangePercent,
			event.BlockNumber,
			event.Price-(event.Price*float64(priceChangePercent)/100),
			event.Price,
		)
	} else if absChange >= MajorAlert {
		log.Printf(
			"[MARKET_PRICING_CONSUMER] ⚠️ MAJOR ALERT: Price changed by %.2f%% at Block #%d (%.2f -> %.2f)",
			priceChangePercent,
			event.BlockNumber,
			event.Price-(event.Price*float64(priceChangePercent)/100),
			event.Price,
		)
	} else if absChange >= MinorAlert {
		log.Printf(
			"[MARKET_PRICING_CONSUMER] ℹ️ Minor Alert: Price changed by %.2f%% at Block #%d (%.2f -> %.2f)",
			priceChangePercent,
			event.BlockNumber,
			event.Price-(event.Price*float64(priceChangePercent)/100),
			event.Price,
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

	log.Printf("[MARKET_PRICING_CONSUMER] Loaded initial price cache for %d blocks", len(ticks))
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

	log.Println("[MARKET_PRICING_CONSUMER] stopping consumer...")
	close(m.stopChan)
	log.Println("[MARKET_PRICING_CONSUMER] consumer stopped")
}

func (m *MarketPricingConsumer) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isRunning
}
