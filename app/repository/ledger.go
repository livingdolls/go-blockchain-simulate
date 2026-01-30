package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type LedgerEntry struct {
	BlockID      int64   `db:"block_id"`
	TxID         *int64  `db:"tx_id"`
	Address      string  `db:"address"`
	Amount       float64 `db:"change_amount"`
	BalanceAfter float64 `db:"balance_after"`
}

type LedgerEntryWithID struct {
	ID           int64   `db:"id"`
	BlockID      int64   `db:"block_id"`
	TxID         *int64  `db:"tx_id"`
	Address      string  `db:"address"`
	Amount       float64 `db:"change_amount"`
	BalanceAfter float64 `db:"balance_after"`
}

type LedgerRepository interface {
	BulkCreateWithTx(dbTx *sqlx.Tx, entries []LedgerEntry) error
	BulkCreate(entries []LedgerEntry) error
	GetEntriesByBlockID(blockID int64) ([]LedgerEntryWithID, error)
	GetEntriesByAddress(address string, limit int) ([]LedgerEntryWithID, error)
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

	query := `INSERT INTO ledger (block_id, tx_id, address, change_amount, balance_after) VALUES `
	var values []interface{}

	for i, entry := range entries {
		if i > 0 {
			query += ","
		}
		query += "(?, ?, ?, ?, ?)"
		values = append(values, entry.BlockID, entry.TxID, entry.Address, entry.Amount, entry.BalanceAfter)
	}
	_, err := dbTx.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("bulk create with tx: %w", err)
	}
	return nil
}

func (l *ledgerRepository) BulkCreate(entries []LedgerEntry) error {
	if len(entries) == 0 {
		return nil
	}

	query := `INSERT INTO ledger (block_id, tx_id, address, change_amount, balance_after) VALUES `
	var values []interface{}

	for i, entry := range entries {
		if i > 0 {
			query += ","
		}
		query += "(?, ?, ?, ?, ?)"
		values = append(values, entry.BlockID, entry.TxID, entry.Address, entry.Amount, entry.BalanceAfter)
	}
	_, err := l.db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("bulk create: %w", err)
	}
	return nil
}

func (l *ledgerRepository) GetEntriesByBlockID(blockID int64) ([]LedgerEntryWithID, error) {
	var entries []LedgerEntryWithID
	query := `SELECT id, block_id, tx_id, address, change_amount, balance_after FROM ledger WHERE block_id = ? ORDER BY id ASC`
	err := l.db.Select(&entries, query, blockID)
	if err != nil {
		return nil, fmt.Errorf("get entries by block id: %w", err)
	}
	return entries, nil
}

func (l *ledgerRepository) GetEntriesByAddress(address string, limit int) ([]LedgerEntryWithID, error) {
	var entries []LedgerEntryWithID
	query := `SELECT id, block_id, tx_id, address, change_amount, balance_after FROM ledger WHERE address = ? ORDER BY id DESC LIMIT ?`
	err := l.db.Select(&entries, query, address, limit)
	if err != nil {
		return nil, fmt.Errorf("get entries by address: %w", err)
	}
	return entries, nil
}
