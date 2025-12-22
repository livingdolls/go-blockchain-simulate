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
}

type marketRepository struct {
	db *sqlx.DB
}

func NewMarketRepository(db *sqlx.DB) MarketRepository {
	return &marketRepository{
		db: db,
	}
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
