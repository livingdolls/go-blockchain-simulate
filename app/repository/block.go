package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type BlockRepository interface {
	BeginTx() (*sqlx.Tx, error)
	CreateWithTx(tx *sqlx.Tx, block models.Block) (int64, error)
	GetLastBlock() (models.Block, error)
	GetLastBlockForUpdateWithTx(tx *sqlx.Tx) (models.Block, error)
	InsertBlockTransactionWithTx(tx *sqlx.Tx, blockID, txID int64) error
	BulkInsertBlockTransactionsWithTx(tx *sqlx.Tx, blockID int64, txIDs []int64) error
	GetBlocks(limit, offset int) ([]models.Block, error)
	GetBlockByID(id int64) (models.Block, error)
	GetAllBlocks() ([]models.Block, error)
}

type blockRepository struct {
	db *sqlx.DB
}

func NewBlockRepository(db *sqlx.DB) BlockRepository {
	return &blockRepository{db: db}
}

func (b *blockRepository) BeginTx() (*sqlx.Tx, error) {
	return b.db.Beginx()
}

// CreateWithTx implements BlockRepository.
func (b *blockRepository) CreateWithTx(tx *sqlx.Tx, block models.Block) (int64, error) {
	query := `
		INSERT INTO blocks (block_number, previous_hash, current_hash)
		VALUES (?, ?, ?)
	`

	result, err := tx.Exec(query, block.BlockNumber, block.PreviousHash, block.CurrentHash)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetLastBlock implements BlockRepository (non-locking read-only).
func (b *blockRepository) GetLastBlock() (models.Block, error) {
	var block models.Block

	err := b.db.Get(&block, `SELECT * FROM blocks ORDER BY id DESC LIMIT 1`)
	return block, err
}

// GetLastBlockForUpdateWithTx implements BlockRepository (locking).
func (b *blockRepository) GetLastBlockForUpdateWithTx(tx *sqlx.Tx) (models.Block, error) {
	var block models.Block

	err := tx.Get(&block, `SELECT * FROM blocks ORDER BY id DESC LIMIT 1 FOR UPDATE`)
	return block, err
}

// InsertBlockTransaction implements BlockRepository.
func (b *blockRepository) InsertBlockTransactionWithTx(tx *sqlx.Tx, blockID, txID int64) error {
	_, err := tx.Exec(`
		INSERT INTO block_transactions (block_id, transaction_id)
		VALUES (?, ?)
	`, blockID, txID)

	return err
}

// BulkInsertBlockTransactionsWithTx inserts multiple block-transaction links in a single query
func (b *blockRepository) BulkInsertBlockTransactionsWithTx(tx *sqlx.Tx, blockID int64, txIDs []int64) error {
	if len(txIDs) == 0 {
		return nil
	}

	query := `INSERT INTO block_transactions (block_id, transaction_id) VALUES `
	var values []interface{}

	for i, txID := range txIDs {
		if i > 0 {
			query += ","
		}
		query += "(?, ?)"
		values = append(values, blockID, txID)
	}

	_, err := tx.Exec(query, values...)
	return err
}

func (b *blockRepository) GetBlocks(limit, offset int) ([]models.Block, error) {
	var blocks []models.Block

	err := b.db.Select(&blocks, `
		SELECT * FROM blocks
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, limit, offset)

	return blocks, err
}

func (b *blockRepository) GetBlockByID(id int64) (models.Block, error) {
	var block models.Block

	err := b.db.Get(&block, `
		SELECT * FROM blocks
		WHERE id = ?
	`, id)

	return block, err
}

func (b *blockRepository) GetAllBlocks() ([]models.Block, error) {
	var blocks []models.Block

	err := b.db.Select(&blocks, `
		SELECT * FROM blocks
		ORDER BY block_number ASC
	`)

	if err != nil {
		return nil, err
	}

	// Populate transactions for each block
	for i := range blocks {
		var txs []models.Transaction
		query := `
			SELECT t.id, t.from_address, t.to_address, t.amount, t.signature, t.status
			FROM transactions t
			INNER JOIN block_transactions bt ON t.id = bt.transaction_id
			WHERE bt.block_id = ?
			ORDER BY t.id ASC
		`
		err := b.db.Select(&txs, query, blocks[i].ID)
		if err != nil {
			return nil, err
		}
		blocks[i].Transactions = txs
	}

	return blocks, nil
}
