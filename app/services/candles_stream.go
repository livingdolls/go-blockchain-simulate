package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/redis"
)

type CandleStreamService interface {
	PublishCandle(ctx context.Context, candle models.Candle) error
	SubscribeCandle(ctx context.Context, interval string, callback func(models.Candle) error) error
	PublishAllIntervals(ctx context.Context, candle models.Candle) error
}

type candleStreamService struct {
	redis redis.MemoryAdapter
}

func NewCandleStreamService(redisAdapter redis.MemoryAdapter) CandleStreamService {
	return &candleStreamService{
		redis: redisAdapter,
	}
}

// PublishAllIntervals implements [CandleStreamService].
func (c *candleStreamService) PublishAllIntervals(ctx context.Context, candle models.Candle) error {
	intervals := []string{"1m", "5m", "15m", "1h", "4h", "1d"}

	for _, interval := range intervals {
		candleCopy := candle
		candleCopy.IntervalType = interval

		if err := c.PublishCandle(ctx, candleCopy); err != nil {
			log.Printf("PublishAllIntervals error: %v\n", err)
			return err
		}
	}

	return nil
}

// PublishCandle implements [CandleStreamService].
func (c *candleStreamService) PublishCandle(ctx context.Context, candle models.Candle) error {
	channel := fmt.Sprintf("candles:%s", candle.IntervalType)

	payload, err := json.Marshal(candle)
	if err != nil {
		log.Printf("Marshal Candle error: %v\n", err)
		return err
	}

	// publish to redis
	if err := c.redis.Publish(ctx, channel, payload); err != nil {
		log.Printf("Publish Candle error: %v\n", err)
		return err
	}

	log.Printf("Published candle to channel=%s\n", channel)
	return nil
}

// SubscribeCandle implements [CandleStreamService].
func (c *candleStreamService) SubscribeCandle(ctx context.Context, interval string, callback func(models.Candle) error) error {
	channel := fmt.Sprintf("candles:%s", interval)

	return c.redis.Subscribe(ctx, channel, func(message []byte) error {
		var candle models.Candle

		if err := json.Unmarshal(message, &candle); err != nil {
			log.Printf("Unmarshal Candle error: %v\n", err)
			return err
		}

		log.Printf("Received candle update: %s\n", interval)
		return callback(candle)
	})
}
