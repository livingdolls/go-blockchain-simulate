package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type TransactionRepository interface {
	CreateWithTx(dbTx *sqlx.Tx, transaction models.Transaction) (int64, error)
	UpdateStatusWithTx(dbTx *sqlx.Tx, id int64, status string) error
	GetPendingTransactions() ([]models.Transaction, error)
	MarkConfirmedWithTx(dbTx *sqlx.Tx, txIDs []int64) error
}

type transactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) TransactionRepository {
	return &transactionRepository{
		db: db,
	}
}

func (r *transactionRepository) CreateWithTx(dbTx *sqlx.Tx, transaction models.Transaction) (int64, error) {
	query := `
		INSERT INTO transactions (from_address, to_address, amount, signature, status)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := dbTx.Exec(query, transaction.FromAddress, transaction.ToAddress, transaction.Amount, transaction.Signature, transaction.Status)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *transactionRepository) UpdateStatusWithTx(dbTx *sqlx.Tx, id int64, status string) error {
	query := `
		UPDATE transactions
		SET status = ?
		WHERE id = ?
	`

	_, err := dbTx.Exec(query, status, id)
	return err
}

func (r *transactionRepository) GetPendingTransactions() ([]models.Transaction, error) {
	var list []models.Transaction

	query := `
        SELECT id, from_address, to_address, amount, signature, status 
        FROM transactions 
        WHERE TRIM(status) = 'PENDING'
        ORDER BY id ASC
    `

	err := r.db.Select(&list, query)

	if err != nil {
		// Log error detail
		fmt.Printf("GetPendingTransactions error: %v\n", err)
		return nil, err
	}

	// Empty result is not an error
	if len(list) == 0 {
		fmt.Println("No pending transactions found")
		return []models.Transaction{}, nil
	}

	return list, nil
}

func (r *transactionRepository) MarkConfirmedWithTx(dbTx *sqlx.Tx, txIDs []int64) error {
	query, args, _ := sqlx.In(`UPDATE transactions SET status = 'CONFIRMED' WHERE id IN (?)`, txIDs)

	_, err := dbTx.Exec(query, args...)

	return err
}
