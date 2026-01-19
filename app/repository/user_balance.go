package repository

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type UserBalanceRepository interface {
	BeginTx() (*sqlx.Tx, error)
	UpsertEmptyIfNotExistsWithTx(tx *sqlx.Tx, address string) error
	UpsertEmptyIfNotExists(address string) error
	GetForUpdateWithTx(tx *sqlx.Tx, address string) (models.UserBalance, error)
	UpdateBalanceWithTx(tx *sqlx.Tx, address string, newBalance, totalDeposited float64) error
	InsertHistoryWithTx(tx *sqlx.Tx, history models.BalanceHistory) error
	GetByAddress(address string) (models.UserBalance, error)
	GetMultipleByAddressWithTxForUpdate(tx *sqlx.Tx, addresses []string) ([]models.UserBalance, error)
	BulkUpdateBalancesWithTx(tx *sqlx.Tx, balances map[string]models.UserBalance) error
}

type userBalanceRepository struct {
	db *sqlx.DB
}

func NewUserBalanceRepository(db *sqlx.DB) UserBalanceRepository {
	return &userBalanceRepository{db: db}
}

// BeginTx implements [UserBalanceRepository].
func (u *userBalanceRepository) BeginTx() (*sqlx.Tx, error) {
	return u.db.Beginx()
}

func (u *userBalanceRepository) GetByAddress(address string) (models.UserBalance, error) {
	var ub models.UserBalance
	query := `
		SELECT user_address, usd_balance, locked_balance, total_deposited, total_withdrawn, total_traded, last_transaction_at
		FROM user_balances
		WHERE user_address = ?
	`
	err := u.db.Get(&ub, query, address)

	if err == sql.ErrNoRows {
		return models.UserBalance{}, entity.ErrUserBalanceNotFound
	}

	return ub, err
}

// GetForUpdateWithTx implements [UserBalanceRepository].
func (u *userBalanceRepository) GetForUpdateWithTx(tx *sqlx.Tx, address string) (models.UserBalance, error) {
	var ub models.UserBalance
	query := `
		SELECT user_address, usd_balance, locked_balance, total_deposited, total_withdrawn, total_traded, last_transaction_at
		FROM user_balances
		WHERE user_address = ?
		FOR UPDATE
	`
	err := tx.Get(&ub, query, address)

	if err == sql.ErrNoRows {
		return models.UserBalance{}, entity.ErrUserBalanceNotFound
	}

	return ub, err
}

// InsertHistoryWithTx implements [UserBalanceRepository].
func (u *userBalanceRepository) InsertHistoryWithTx(tx *sqlx.Tx, history models.BalanceHistory) error {
	query := `
		INSERT INTO balance_history (
			user_address, order_id, change_type, amount, balance_before, balance_after, locked_before, locked_after, reference_id, description
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := tx.Exec(query, history.UserAddress, history.OrderID, history.ChangeType, history.Amount, history.BalanceBefore, history.BalanceAfter, history.LockedBefore, history.LockedAfter, history.ReferenceID, history.Description)

	return err
}

// UpdateBalanceWithTx implements [UserBalanceRepository].
func (u *userBalanceRepository) UpdateBalanceWithTx(tx *sqlx.Tx, address string, newBalance float64, totalDeposited float64) error {
	query := `
		UPDATE user_balances
		SET usd_balance = ?, total_deposited = ?, last_transaction_at = NOW()
		WHERE user_address = ?`
	_, err := tx.Exec(query, newBalance, totalDeposited, address)

	return err
}

// UpsertEmptyIfNotExistsWithTx implements [UserBalanceRepository].
func (u *userBalanceRepository) UpsertEmptyIfNotExistsWithTx(tx *sqlx.Tx, address string) error {
	query := `
		INSERT INTO user_balances (user_address, usd_balance, locked_balance, total_deposited, total_withdrawn, total_traded)
		VALUES (?, 0, 0, 0, 0, 0)
		ON DUPLICATE KEY UPDATE user_address = user_address
	`
	_, err := tx.Exec(query, address)
	return err
}

// UpsertEmptyIfNotExists implements [UserBalanceRepository].
func (u *userBalanceRepository) UpsertEmptyIfNotExists(address string) error {
	query := `
		INSERT INTO user_balances (user_address, usd_balance, locked_balance, total_deposited, total_withdrawn, total_traded)
		VALUES (?, 0, 0, 0, 0, 0)
		ON DUPLICATE KEY UPDATE user_address = user_address
	`
	_, err := u.db.Exec(query, address)
	return err
}

func (u *userBalanceRepository) GetMultipleByAddressWithTxForUpdate(tx *sqlx.Tx, addresses []string) ([]models.UserBalance, error) {
	var balances []models.UserBalance

	query := `
		SELECT user_address, usd_balance, locked_balance, total_deposited, total_withdrawn, total_traded, last_transaction_at
		FROM user_balances
		WHERE user_address IN (?)
		FOR UPDATE
	`

	query, args, err := sqlx.In(query, addresses)

	if err != nil {
		return nil, err
	}

	err = tx.Select(&balances, tx.Rebind(query), args...)
	return balances, err
}

func (u *userBalanceRepository) BulkUpdateBalancesWithTx(tx *sqlx.Tx, balances map[string]models.UserBalance) error {
	for addr, balance := range balances {
		query := `UPDATE user_balances SET usd_balance = ?, total_withdrawn = ?, total_traded = ?, last_transaction_at = NOW() WHERE user_address = ?`

		if _, err := tx.Exec(query, balance.USDBalance, balance.TotalWithdrawn, balance.TotalTraded, addr); err != nil {
			return err
		}
	}

	return nil
}
