package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
)

type MarketPricingPublisher interface {
	PublishPricingEvent(ctx context.Context, blockID int64, blockNumber int, priceData models.MarketEngine, volumeData models.MarketTick, minerAddress string) error
	PublishVolumeUpdate(ctx context.Context, volumeData models.MarketTick, blockNumber int) error
}

type marketPricingPublisher struct {
	rmqClient *rabbitmq.Client
}

func NewMarketPricingPublisher(rmqClient *rabbitmq.Client) MarketPricingPublisher {
	return &marketPricingPublisher{
		rmqClient: rmqClient,
	}
}

// PublishVolumeUpdate - Publikasikan update volume transaksi
func (m *marketPricingPublisher) PublishVolumeUpdate(ctx context.Context, volumeData models.MarketTick, blockNumber int) error {
	netVolume := volumeData.BuyVolume - volumeData.SellVolume
	totalVolume := volumeData.BuyVolume + volumeData.SellVolume

	var volumeRatio float64

	if totalVolume > 0 {
		volumeRatio = volumeData.BuyVolume / totalVolume
	}

	update := dto.MarketVolumeUpdate{
		BlockID:     volumeData.BlockID,
		BuyVolume:   volumeData.BuyVolume,
		SellVolume:  volumeData.SellVolume,
		NetVolume:   netVolume,
		VolumeRatio: volumeRatio,
		TxCount:     volumeData.TxCount,
		Timestamp:   volumeData.CreatedAt,
	}

	body, err := json.Marshal(update)

	if err != nil {
		return fmt.Errorf("failed to marshal market volume update: %w", err)
	}

	if err := m.rmqClient.Publish(
		ctx,
		rabbitmq.MarketExchange,
		rabbitmq.MarketVolumeUpdateKey,
		body,
	); err != nil {
		return fmt.Errorf("failed to publish market volume update: %w", err)
	}

	logger.LogInfo(fmt.Sprintf("[MARKET_VOLUME] Publisher volume update - Buy: %.2f, Sell: %.2f, Ratio: %.2f%%",
		volumeData.BuyVolume, volumeData.SellVolume, volumeRatio*100))

	return nil
}

// PublishPricingEvent - Publikasikan event pricing dari block yang baru dimined
// Routing Key: market.price.update
// Exchange: market (topic)
// Queue: market.pricing
func (m *marketPricingPublisher) PublishPricingEvent(ctx context.Context, blockID int64, blockNumber int, priceData models.MarketEngine, volumeData models.MarketTick, minerAddress string) error {
	event := dto.MarketPricingEvent{
		BlockID:      blockID,
		BlockNumber:  blockNumber,
		Price:        priceData.Price,
		Liquidity:    priceData.Liquidity,
		BuyVolume:    volumeData.BuyVolume,
		SellVolume:   volumeData.SellVolume,
		TxCount:      volumeData.TxCount,
		Timestamp:    volumeData.CreatedAt,
		MinerAddress: minerAddress,
	}

	body, err := json.Marshal(event)

	if err != nil {
		return fmt.Errorf("failed to marshal market pricing event: %w", err)
	}

	if err := m.rmqClient.Publish(
		ctx,
		rabbitmq.MarketExchange,
		rabbitmq.MarketPricingKey,
		body,
	); err != nil {
		return fmt.Errorf("failed to publish market pricing event: %w", err)
	}

	logger.LogInfo(fmt.Sprintf("[MARKET_PRICING] Published pricing event for block #%d at price %.2f", blockNumber, priceData.Price))

	return nil
}
