package services

import (
	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
)

type CandleService interface {
	GetCandles(intervalType string, limit int) ([]models.Candle, error)
	GetCandlesFrom(intervalType string, startTime int64, limit int) ([]models.Candle, error)
	GetCandle(intervalType string, startTime int64) (models.Candle, error)
	UpsertCandleWithTx(tx *sqlx.Tx, candle models.Candle) error
	DeleteOldCandlesWithTx(tx *sqlx.Tx, intervalType string, beforeTime int64) error
}

type candleService struct {
	repo repository.CandlesRepository
}

func NewCandleService(repo repository.CandlesRepository) CandleService {
	return &candleService{
		repo: repo,
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
