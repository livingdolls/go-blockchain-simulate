package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type CandleService interface {
	GetCandles(intervalType string, limit int) ([]models.Candle, error)
	GetCandlesFrom(intervalType string, startTime int64, limit int) ([]models.Candle, error)
	GetCandle(intervalType string, startTime int64) (models.Candle, error)
	GetLatestCandleByInterval(intervalType string) (models.Candle, error)
	UpsertCandleWithTx(tx *sqlx.Tx, candle models.Candle) error
	DeleteOldCandlesWithTx(tx *sqlx.Tx, intervalType string, beforeTime int64) error
	AggregateCandle(ctx context.Context, interval string, timestamp int64) error
}

type candleService struct {
	repo   repository.CandlesRepository
	stream CandleStreamService
}

func NewCandleService(repo repository.CandlesRepository, stream CandleStreamService) CandleService {
	return &candleService{
		repo:   repo,
		stream: stream,
	}
}

// DeleteOldCandlesWithTx implements [CandleService].
func (c *candleService) DeleteOldCandlesWithTx(tx *sqlx.Tx, intervalType string, beforeTime int64) error {
	return c.repo.DeleteOldCandlesWithTx(tx, intervalType, beforeTime)
}

// GetCandle implements [CandleService].
func (c *candleService) GetCandle(intervalType string, startTime int64) (models.Candle, error) {
	return c.repo.GetCandleByIntervalAndStartTime(intervalType, startTime)
}

// GetCandles implements [CandleService].
func (c *candleService) GetCandles(intervalType string, limit int) ([]models.Candle, error) {
	if limit <= 0 {
		limit = 100
	}

	return c.repo.GetCandleByInterval(intervalType, limit)
}

// GetCandlesFrom implements [CandleService].
func (c *candleService) GetCandlesFrom(intervalType string, startTime int64, limit int) ([]models.Candle, error) {
	if limit <= 0 {
		limit = 100
	}

	return c.repo.GetCandleByIntervalAndTime(intervalType, startTime, limit)
}

// UpsertCandleWithTx implements [CandleService].
func (c *candleService) UpsertCandleWithTx(tx *sqlx.Tx, candle models.Candle) error {
	return c.repo.UpsertCandleWithTx(tx, candle)
}

func (c *candleService) AggregateCandle(ctx context.Context, interval string, timestamp int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	start := utils.FloorTime(timestamp, interval)
	duration := utils.IntervalDuration(interval)
	end := start + duration

	// fetch ticks in the interval
	ticks, err := c.repo.GetTicksRange(start, end)
	if err != nil {
		return err
	}

	if len(ticks) == 0 {
		// no ticks in this interval, skip
		return nil
	}

	// hitung ohlcv
	open := ticks[0].Price
	closep := ticks[len(ticks)-1].Price
	high, low := open, open

	var volume float64

	for _, tick := range ticks {
		if tick.Price > high {
			high = tick.Price
		}

		if tick.Price < low {
			low = tick.Price
		}

		volume += tick.BuyVolume + tick.SellVolume
	}

	candle := models.Candle{
		IntervalType: interval,
		StartTime:    start,
		OpenPrice:    open,
		HighPrice:    high,
		LowPrice:     low,
		ClosePrice:   closep,
		Volume:       volume,
	}

	// upsert candle with transaction
	tx, err := c.repo.CandleBeginTx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	rowsAffected, err := c.repo.UpsertCandleOnDuplicateWithTx(tx, candle)

	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if rowsAffected > 0 {
		logger.LogInfo(fmt.Sprintf("Aggregated candle for interval=%s, start=%d", interval, start))
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := c.stream.PublishCandle(ctx, candle); err != nil {
			logger.LogError("PublishCandle error", err)
		} else {
			logger.LogInfo(fmt.Sprintf("Published candle for interval=%s, start=%d", interval, start))
		}
	} else {
		// log.Printf("Candle for interval=%s, start=%d already up-to-date, skipped publish\n", interval, start)
	}

	return nil
}

func (c *candleService) GetLatestCandleByInterval(intervalType string) (models.Candle, error) {
	return c.repo.GetLatestCandleByInterval(intervalType)
}
