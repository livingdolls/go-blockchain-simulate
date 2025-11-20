package repository

import "github.com/jmoiron/sqlx"

type LedgerRepository interface {
	CreateWithTx(dbTx *sqlx.Tx, txID int64, address string, change float64, balanceAfter float64) error
}

type ledgerRepository struct {
	db *sqlx.DB
}

func NewLedgerRepository(db *sqlx.DB) LedgerRepository {
	return &ledgerRepository{
		db: db,
	}
}

// Create implements LedgerRepository.
func (l *ledgerRepository) CreateWithTx(dbTx *sqlx.Tx, txID int64, address string, change float64, balanceAfter float64) error {
	_, err := dbTx.Exec(`
		INSERT INTO ledger (tx_id, address, change_amount, balance_after) VALUES (?, ?, ?, ?)
	`, txID, address, change, balanceAfter)
	return err
}
