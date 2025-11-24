package repository

import (
	"github.com/jmoiron/sqlx"
)

type LedgerEntry struct {
	TxID         int64
	Address      string
	Amount       float64
	BalanceAfter float64
}

type LedgerRepository interface {
	BulkCreateWithTx(dbTx *sqlx.Tx, entries []LedgerEntry) error
}

type ledgerRepository struct {
	db *sqlx.DB
}

func NewLedgerRepository(db *sqlx.DB) LedgerRepository {
	return &ledgerRepository{
		db: db,
	}
}

// BulkCreateWithTx inserts multiple ledger entries in a single query
func (l *ledgerRepository) BulkCreateWithTx(dbTx *sqlx.Tx, entries []LedgerEntry) error {
	if len(entries) == 0 {
		return nil
	}

	query := `INSERT INTO ledger (tx_id, address, change_amount, balance_after) VALUES `
	var values []interface{}

	for i, entry := range entries {
		if i > 0 {
			query += ","
		}
		query += "(?, ?, ?, ?)"
		values = append(values, entry.TxID, entry.Address, entry.Amount, entry.BalanceAfter)
	}
	_, err := dbTx.Exec(query, values...)
	return err
}
