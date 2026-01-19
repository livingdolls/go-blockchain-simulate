package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type TransactionRepository interface {
	CreateWithTx(dbTx *sqlx.Tx, transaction models.Transaction) (int64, error)
	Create(transaction models.Transaction) (int64, error)
	UpdateStatusWithTx(dbTx *sqlx.Tx, id int64, status string) error
	GetPendingTransactionsWithTx(dbTx *sqlx.Tx) ([]models.Transaction, error)
	GetPendingTransactions(limit int) ([]models.Transaction, error)
	MarkConfirmedWithTx(dbTx *sqlx.Tx, txID int64) error
	BulkMarkConfirmedWithTx(dbTx *sqlx.Tx, txIDs []int64) error
	GetPendingTransactionsByAddress(address string) (float64, error)
	GetPendingBuyCostByBuyer(address string) (float64, error)
	GetTransactionsByBlockID(blockID int64) ([]models.Transaction, error)
	GetTransactionByID(id int64) (models.Transaction, error)
	GetTransactionByAddress(filter models.TransactionFilter) (models.TransactionWithTypeResponse, error)
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
		INSERT INTO transactions (from_address, to_address, amount, fee, type, signature, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := dbTx.Exec(query, transaction.FromAddress, transaction.ToAddress, transaction.Amount, transaction.Fee, transaction.Type, transaction.Signature, transaction.Status)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *transactionRepository) Create(transaction models.Transaction) (int64, error) {
	query := `
		INSERT INTO transactions (from_address, to_address, amount, fee, type, signature, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query, transaction.FromAddress, transaction.ToAddress, transaction.Amount, transaction.Fee, transaction.Type, transaction.Signature, transaction.Status)
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

func (r *transactionRepository) GetPendingTransactionsWithTx(dbTx *sqlx.Tx) ([]models.Transaction, error) {
	var list []models.Transaction

	query := `
        SELECT id, from_address, to_address, amount, fee, type, signature, status 
        FROM transactions 
        WHERE TRIM(status) = 'PENDING'
        ORDER BY id ASC
		FOR UPDATE
    `

	err := dbTx.Select(&list, query)

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

func (r *transactionRepository) MarkConfirmedWithTx(dbTx *sqlx.Tx, txID int64) error {
	query := `UPDATE transactions SET status = 'CONFIRMED' WHERE id = ?`

	_, err := dbTx.Exec(query, txID)

	return err
}

// GetPendingTransactions retrieves pending transactions without locking (for read-only validation)
func (r *transactionRepository) GetPendingTransactions(limit int) ([]models.Transaction, error) {
	var list []models.Transaction

	query := `
        SELECT id, from_address, to_address, amount, fee, type, signature, status 
        FROM transactions 
        WHERE TRIM(status) = 'PENDING'
        ORDER BY id ASC
        LIMIT ?
    `

	err := r.db.Select(&list, query, limit)

	if err != nil {
		fmt.Printf("GetPendingTransactions error: %v\n", err)
		return nil, err
	}

	if len(list) == 0 {
		return []models.Transaction{}, nil
	}

	return list, nil
}

// BulkMarkConfirmedWithTx marks multiple transactions as confirmed in a single query
func (r *transactionRepository) BulkMarkConfirmedWithTx(dbTx *sqlx.Tx, txIDs []int64) error {
	if len(txIDs) == 0 {
		return nil
	}

	query, args, err := sqlx.In(`UPDATE transactions SET status = 'CONFIRMED' WHERE id IN (?)`, txIDs)
	if err != nil {
		return err
	}

	_, err = dbTx.Exec(dbTx.Rebind(query), args...)
	return err
}

func (r *transactionRepository) GetPendingTransactionsByAddress(address string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount + fee), 0) as pending_amount
		FROM transactions
		WHERE from_address = ? AND status = 'PENDING'
	`

	var pendingAmount float64

	err := r.db.Get(&pendingAmount, query, address)
	return pendingAmount, err
}

func (r *transactionRepository) GetPendingBuyCostByBuyer(address string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount + fee), 0) as pending_amount
		FROM transactions
		WHERE LOWER(to_address) = LOWER(?) AND status = 'PENDING' AND LOWER(type) = 'buy'
		`

	var pendingAmount float64

	err := r.db.Get(&pendingAmount, query, address)
	return pendingAmount, err
}

func (r *transactionRepository) GetTransactionsByBlockID(blockID int64) ([]models.Transaction, error) {
	var transaction []models.Transaction

	query := `
		SELECT tx.id, tx.from_address, tx.to_address, tx.amount, tx.fee, tx.type, tx.signature, tx.status
		FROM transactions as tx
		JOIN block_transactions as bt ON tx.id = bt.transaction_id
		WHERE bt.block_id = ?
	`

	err := r.db.Select(&transaction, query, blockID)
	return transaction, err
}

func (r *transactionRepository) GetTransactionByID(id int64) (models.Transaction, error) {
	var transaction models.Transaction

	query := `
		SELECT id, from_address, to_address, amount, fee, type, signature, status
		FROM transactions
		WHERE id = ?
	`

	err := r.db.Get(&transaction, query, id)

	if err != nil {
		return models.Transaction{}, fmt.Errorf("error get transaction %w", err)
	}

	return transaction, err
}

func (r *transactionRepository) GetTransactionByAddress(filter models.TransactionFilter) (models.TransactionWithTypeResponse, error) {
	filter.Validate()

	var result models.TransactionWithTypeResponse

	result.Page = filter.Page
	result.Limit = filter.Limit

	whereCondition := []string{}
	args := []interface{}{}

	// Address condition
	whereCondition = append(whereCondition, "(LOWER(from_address) = LOWER(?) OR LOWER(to_address) = LOWER(?))")
	args = append(args, filter.Address, filter.Address)

	switch filter.Type {
	case "send":
		whereCondition = append(whereCondition, "LOWER(from_address) = LOWER(?)")
		args = append(args, filter.Address)
	case "received":
		whereCondition = append(whereCondition, "LOWER(to_address) = LOWER(?)")
		args = append(args, filter.Address)
	case "buy":
		whereCondition = append(whereCondition, "LOWER(from_address) = 'MINER_ACCOUNT'")
	case "sell":
		whereCondition = append(whereCondition, "LOWER(to_address) = 'MINER_ACCOUNT'")
	}

	// status filter
	if filter.Status != "ALL" {
		whereCondition = append(whereCondition, "status = ?")
		args = append(args, filter.Status)
	}

	whereClause := strings.Join(whereCondition, " AND ")

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM transactions WHERE %s`, whereClause)
	err := r.db.Get(&result.Total, countQuery, args...)
	if err != nil {
		return result, fmt.Errorf("failed to count transaction: %w", err)
	}

	// calculate total pages
	if result.Total > 0 {
		result.TotalPages = int((result.Total + int64(result.Limit) - 1) / int64(result.Limit))
	}

	//build main query with pagination and sorting
	offset := (result.Page - 1) * result.Limit

	query := fmt.Sprintf(`
		SELECT id, 
		CASE
			WHEN from_address = 'MINER_ACCOUNT' THEN 'BUYER SYSTEM'
			ELSE from_address
		END AS from_address,
		CASE
			WHEN to_address = 'MINER_ACCOUNT' THEN 'SELLER SYSTEM'
			ELSE to_address
		END AS to_address,
		amount, fee, signature, status, 
		CASE 
			WHEN LOWER(type) = 'transfer' THEN
				CASE
					WHEN LOWER(from_address) = LOWER(?) 
					THEN 'send'
					ELSE 'received'
				END
			ELSE LOWER(type)
		END AS type, created_at
		FROM transactions
		WHERE %s
		ORDER BY %s %s
		LIMIT ? OFFSET ?`, whereClause, filter.SortBy, filter.Order)

	// Append pagination args
	args = append([]interface{}{filter.Address}, args...)
	args = append(args, result.Limit, offset)

	err = r.db.Select(&result.Transactions, query, args...)
	if err != nil {
		return result, fmt.Errorf("failed to get transactions: %w", err)
	}

	return result, nil
}
