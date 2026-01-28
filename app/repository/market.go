package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type MarketRepository interface {
	GetState() (models.MarketEngine, error)
	GetStateForUpdateWithTx(tx *sqlx.Tx) (models.MarketEngine, error)
	UpdateStateWithTx(tx *sqlx.Tx, market models.MarketEngine) error
	InsertTickWithTx(tx *sqlx.Tx, tick models.MarketTick) (int64, error)
	GetTickByBlockID(blockID int64) (models.MarketTick, error)
	GetVolumeHistory(limit, offset int) ([]models.MarketTick, error)
	GetVolumeBlockRange(startBlock, endBlock int64) ([]models.MarketTick, error)
	GetAverateVolume(blockRange int64) (float64, float64, error)
}

type marketRepository struct {
	db *sqlx.DB
}

func NewMarketRepository(db *sqlx.DB) MarketRepository {
	return &marketRepository{
		db: db,
	}
}

func (m *marketRepository) GetTickByBlockID(blockID int64) (models.MarketTick, error) {
	var tick models.MarketTick

	err := m.db.Get(&tick, `SELECT id, block_id, price, buy_volume, sell_volume, tx_count, UNIX_TIMESTAMP(created_at) as created_at FROM market_ticks WHERE block_id = ?`, blockID)
	return tick, err
}

func (m *marketRepository) GetVolumeHistory(limit, offset int) ([]models.MarketTick, error) {
	var ticks []models.MarketTick
	err := m.db.Select(&ticks, `SELECT id, block_id, price, buy_volume, sell_volume, tx_count, UNIX_TIMESTAMP(created_at) as created_at FROM market_ticks ORDER BY block_id DESC LIMIT ? OFFSET ?`, limit, offset)
	return ticks, err
}

func (m *marketRepository) GetVolumeBlockRange(startBlock, endBlock int64) ([]models.MarketTick, error) {
	var ticks []models.MarketTick
	err := m.db.Select(&ticks, `SELECT id, block_id, price, buy_volume, sell_volume, tx_count, UNIX_TIMESTAMP(created_at) as created_at FROM market_ticks WHERE block_id BETWEEN ? AND ? ORDER BY block_id ASC`, startBlock, endBlock)
	return ticks, err
}

func (m *marketRepository) GetAverateVolume(blockRange int64) (float64, float64, error) {
	var buyAvg, sellAvg float64
	result := struct {
		BuyAvg  float64 `db:"buy_avg"`
		SellAvg float64 `db:"sell_avg"`
	}{}
	err := m.db.Get(
		&result,
		`SELECT 
			AVG(buy_volume) as buy_avg,
			AVG(sell_volume) as sell_avg
		 FROM market_ticks 
		 WHERE block_id > (SELECT MAX(block_id) - ? FROM market_ticks)`,
		blockRange)

	if err != nil {
		return 0, 0, err
	}

	buyAvg = result.BuyAvg
	sellAvg = result.SellAvg

	return buyAvg, sellAvg, nil
}

// GetState implements MarketRepository.
func (m *marketRepository) GetState() (models.MarketEngine, error) {
	var state models.MarketEngine
	err := m.db.Get(&state, `SELECT id, price, liquidity, last_block, updated_at FROM market_engine WHERE id = 1`)
	return state, err
}

// GetStateForUpdateWithTx implements MarketRepository.
func (m *marketRepository) GetStateForUpdateWithTx(tx *sqlx.Tx) (models.MarketEngine, error) {
	var state models.MarketEngine
	err := tx.Get(&state, `SELECT id, price, liquidity, last_block, updated_at FROM market_engine WHERE id = 1 FOR UPDATE`)
	return state, err
}

// InsertTickWithTx implements MarketRepository.
func (m *marketRepository) InsertTickWithTx(tx *sqlx.Tx, tick models.MarketTick) (int64, error) {
	res, err := tx.Exec(`INSERT INTO market_ticks (block_id, price, buy_volume, sell_volume, tx_count) VALUES (?, ?, ?, ?, ?)`,
		tick.BlockID,
		tick.Price,
		tick.BuyVolume,
		tick.SellVolume,
		tick.TxCount,
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

// UpdateStateWithTx implements MarketRepository.
func (m *marketRepository) UpdateStateWithTx(tx *sqlx.Tx, market models.MarketEngine) error {
	res, err := tx.Exec(`UPDATE market_engine SET price = ?, liquidity = ?, last_block = ? WHERE id = ?`,
		market.Price,
		market.Liquidity,
		market.LastBlock,
		market.ID,
	)

	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		_, err = tx.Exec(`INSERT INTO market_engine (id, price, liquidity, last_block) VALUES (1, ?, ?, ?)`,
			market.Price,
			market.Liquidity,
			market.LastBlock,
		)

		return err
	}

	return nil
}
