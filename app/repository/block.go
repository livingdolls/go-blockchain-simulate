package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type BlockRepository interface {
	BeginTx() (*sqlx.Tx, error)
	CreateWithTx(tx *sqlx.Tx, block models.Block) (int64, error)
	GetLastBlock(tx *sqlx.Tx) (models.Block, error)
	InsertBlockTransactionWithTx(tx *sqlx.Tx, blockID, txID int64) error
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

	result, err := b.db.Exec(query, block.BlockNumber, block.PreviousHash, block.CurrentHash)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetLastBlock implements BlockRepository.
func (b *blockRepository) GetLastBlock(tx *sqlx.Tx) (models.Block, error) {
	var block models.Block

	err := tx.Get(&block, `SELECT * FROM blocks ORDER BY id DESC LIMIT 1`)
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
