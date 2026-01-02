package repository

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type CandlesRepository interface {
	InsertCandleWithTx(tx *sqlx.Tx, candle models.Candle) (int64, error)
	GetCandleByInterval(intervalType string, limit int) ([]models.Candle, error)
	GetCandleByIntervalAndTime(intervalType string, startTime int64, limit int) ([]models.Candle, error)
	GetCandleByIntervalAndStartTime(intervalType string, startTime int64) (models.Candle, error)
	GetLatestCandleByInterval(intervalType string) (models.Candle, error)
	UpdateCandleWithTx(tx *sqlx.Tx, candle models.Candle) error
	UpsertCandleWithTx(tx *sqlx.Tx, candle models.Candle) error
	DeleteOldCandlesWithTx(tx *sqlx.Tx, intervalType string, beforeTime int64) error
	GetTicksRange(startTime, endTime int64) ([]models.MarketTick, error)
	UpsertCandleOnDuplicateWithTx(tx *sqlx.Tx, candle models.Candle) (int64, error)
	CandleBeginTx() (*sqlx.Tx, error)
}

type candleRepository struct {
	db *sqlx.DB
}

func NewCandleRepository(db *sqlx.DB) CandlesRepository {
	return &candleRepository{db: db}
}

// DeleteOldCandlesWithTx implements [CandlesRepository].
func (c *candleRepository) DeleteOldCandlesWithTx(tx *sqlx.Tx, intervalType string, beforeTime int64) error {
	_, err := tx.Exec(
		`DELETE FROM candles WHERE interval_type = ? AND start_time < ?`,
		intervalType,
		beforeTime,
	)

	return err
}

// GetCandleByInterval implements [CandlesRepository].
func (c *candleRepository) GetCandleByInterval(intervalType string, limit int) ([]models.Candle, error) {
	var candles []models.Candle

	err := c.db.Select(&candles,
		`SELECT id, interval_type, start_time, open_price, high_price, low_price, close_price, volume FROM candles 
		WHERE interval_type = ? 
		ORDER BY start_time DESC 
		LIMIT ?`,
		intervalType,
		limit,
	)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if candles == nil {
		candles = []models.Candle{}
	}

	return candles, err
}

// GetCandleByIntervalAndStartTime implements [CandlesRepository].
func (c *candleRepository) GetCandleByIntervalAndStartTime(intervalType string, startTime int64) (models.Candle, error) {
	var candle models.Candle

	err := c.db.Get(&candle,
		`SELECT id, interval_type, start_time, open_price, high_price, low_price, close_price, volume FROM candles 
		WHERE interval_type = ? AND start_time = ?`,
		intervalType,
		startTime,
	)

	return candle, err
}

// GetCandleByIntervalAndTime implements [CandlesRepository].
func (c *candleRepository) GetCandleByIntervalAndTime(intervalType string, startTime int64, limit int) ([]models.Candle, error) {
	var candles []models.Candle

	err := c.db.Select(&candles,
		`SELECT id, interval_type, start_time, open_price, high_price, low_price, close_price, volume FROM candles
		WHERE interval_type = ? AND start_time >= ?
		ORDER BY start_time DESC
		LIMIT ?`,
		intervalType,
		startTime,
		limit,
	)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if candles == nil {
		candles = []models.Candle{}
	}

	return candles, err
}

// InsertCandleWithTx implements [CandlesRepository].
func (c *candleRepository) InsertCandleWithTx(tx *sqlx.Tx, candle models.Candle) (int64, error) {
	res, err := tx.Exec(
		`INSERT INTO candles (interval_type, start_time, open_price, high_price, low_price, close_price, volume) VALUES (?,?,?,?,?,?,?)`,
		candle.IntervalType,
		candle.StartTime,
		candle.OpenPrice,
		candle.HighPrice,
		candle.LowPrice,
		candle.ClosePrice,
		candle.Volume,
	)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateCandleWithTx implements [CandlesRepository].
func (c *candleRepository) UpdateCandleWithTx(tx *sqlx.Tx, candle models.Candle) error {
	_, err := tx.Exec(
		`UPDATE candles
		SET open_price = ?, high_price = ?, low_price = ?, close_price = ?, volume = ?
		WHERE interval_type = ? AND start_time = ?`,
		candle.OpenPrice,
		candle.HighPrice,
		candle.LowPrice,
		candle.ClosePrice,
		candle.Volume,
		candle.IntervalType,
		candle.StartTime,
	)

	return err
}

// UpsertCandleWithTx implements [CandlesRepository].
func (c *candleRepository) UpsertCandleWithTx(tx *sqlx.Tx, candle models.Candle) error {
	var exists bool

	err := tx.Get(&exists,
		`SELECT COUNT(*) > 0 FROM candles WHERE interval_type = ? AND start_time = ?`,
		candle.IntervalType,
		candle.StartTime,
	)

	if err != nil {
		return err
	}

	if exists {
		return c.UpdateCandleWithTx(tx, candle)
	}

	_, err = c.InsertCandleWithTx(tx, candle)
	return err
}

func (c *candleRepository) UpsertCandleOnDuplicateWithTx(tx *sqlx.Tx, candle models.Candle) (int64, error) {
	res, err := tx.Exec(
		`INSERT INTO candles (interval_type, start_time, open_price, high_price, low_price, close_price, volume) VALUES (?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE 
			open_price = VALUES(open_price),
			high_price = VALUES(high_price),
			low_price = VALUES(low_price),
			close_price = VALUES(close_price),
			volume = VALUES(volume)`,
		candle.IntervalType,
		candle.StartTime,
		candle.OpenPrice,
		candle.HighPrice,
		candle.LowPrice,
		candle.ClosePrice,
		candle.Volume,
	)

	var rowsAffected int64 = 0
	if err == nil {
		rowsAffected, err = res.RowsAffected()
	}

	return rowsAffected, err
}

func (c *candleRepository) GetTicksRange(startTime, endTime int64) ([]models.MarketTick, error) {
	var ticks []models.MarketTick

	err := c.db.Select(&ticks,
		`SELECT id, block_id, price, buy_volume, sell_volume, tx_count, UNIX_TIMESTAMP(created_at) AS created_at FROM market_ticks
		WHERE created_at >= FROM_UNIXTIME(?) AND created_at < FROM_UNIXTIME(?)
		ORDER BY created_at ASC`,
		startTime,
		endTime,
	)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if ticks == nil {
		ticks = []models.MarketTick{}
	}

	return ticks, err
}

func (c *candleRepository) CandleBeginTx() (*sqlx.Tx, error) {
	return c.db.Beginx()
}

func (c *candleRepository) GetLatestCandleByInterval(intervalType string) (models.Candle, error) {
	var candle models.Candle

	err := c.db.Get(&candle,
		`SELECT id, interval_type, start_time, open_price, high_price, low_price, close_price, volume FROM candles 
		WHERE interval_type = ? 
		ORDER BY start_time DESC 
		LIMIT 1`,
		intervalType,
	)

	return candle, err
}
