package repository

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type UserWalletRepository interface {
	BeginTx() (*sqlx.Tx, error)
	UpsertEmptyIfNotExistsWithTx(tx *sqlx.Tx, address string) error
	UpsertEmptyIfNotExists(address string) error
	GetForUpdateWithTx(tx *sqlx.Tx, address string) (models.UserWallet, error)
	GetMultipleByAddressWithTx(tx *sqlx.Tx, addresses []string) ([]models.UserWallet, error)
	GetMultipleByAddress(addresses []string) ([]models.UserWallet, error)
	UpdateWalletWithTx(tx *sqlx.Tx, address string, newBalance float64) error
	BulkUpdateBalancesWithTx(tx *sqlx.Tx, balances map[string]float64) error
	InsertHistoryWithTx(tx *sqlx.Tx, history models.WalletHistory) error
	LockMultipleWalletsWithTx(tx *sqlx.Tx, addresses []string) error
	GetByAddress(address string) (models.UserWallet, error)
}

type userWalletRepository struct {
	db *sqlx.DB
}

func NewUserWalletRepository(db *sqlx.DB) UserWalletRepository {
	return &userWalletRepository{db: db}
}

// BeginTx implements [UserWalletRepository].
func (u *userWalletRepository) BeginTx() (*sqlx.Tx, error) {
	return u.db.Beginx()
}

// GetByAddress implements [UserWalletRepository].
func (u *userWalletRepository) GetByAddress(address string) (models.UserWallet, error) {
	var uw models.UserWallet
	query := `
		SELECT user_address, yte_balance, locked_balance, total_received, total_sent, last_transaction_at
		FROM user_wallets
		WHERE user_address = ?
	`

	err := u.db.Get(&uw, query, address)

	if err == sql.ErrNoRows {
		return models.UserWallet{}, entity.ErrUserWalletNotFound
	}

	return uw, err
}

// GetForUpdateWithTx implements [UserWalletRepository].
func (u *userWalletRepository) GetForUpdateWithTx(tx *sqlx.Tx, address string) (models.UserWallet, error) {
	var uw models.UserWallet
	query := `
		SELECT user_address, yte_balance, locked_balance, total_received, total_sent, last_transaction_at
		FROM user_wallets
		WHERE user_address = ?
		FOR UPDATE	
	`

	err := tx.Get(&uw, query, address)

	if err == sql.ErrNoRows {
		return models.UserWallet{}, entity.ErrUserWalletNotFound
	}

	return uw, err
}

// GetMultipleByAddressWithTx implements [UserWalletRepository].
func (u *userWalletRepository) GetMultipleByAddressWithTx(tx *sqlx.Tx, addresses []string) ([]models.UserWallet, error) {
	if len(addresses) == 0 {
		return []models.UserWallet{}, nil
	}

	var walets []models.UserWallet

	query, args, err := sqlx.In(`
		SELECT user_address, yte_balance, locked_balance, total_received, total_sent, last_transaction_at
		FROM user_wallets
		WHERE user_address IN (?)
	`, addresses)

	if err != nil {
		return nil, err
	}

	err = tx.Select(&walets, tx.Rebind(query), args...)
	return walets, err
}

// GetMultipleByAddress implements [UserWalletRepository].
func (u *userWalletRepository) GetMultipleByAddress(addresses []string) ([]models.UserWallet, error) {
	if len(addresses) == 0 {
		return []models.UserWallet{}, nil
	}

	var walets []models.UserWallet

	query, args, err := sqlx.In(`
		SELECT user_address, yte_balance, locked_balance, total_received, total_sent, last_transaction_at
		FROM user_wallets
		WHERE user_address IN (?)
	`, addresses)

	if err != nil {
		return nil, err
	}

	err = u.db.Select(&walets, u.db.Rebind(query), args...)
	return walets, err
}

// InsertHistoryWithTx implements [UserWalletRepository].
func (u *userWalletRepository) InsertHistoryWithTx(tx *sqlx.Tx, history models.WalletHistory) error {
	_, err := tx.Exec(`
		INSERT INTO wallet_history (
			user_address, tx_id, order_id, change_type, amount, balance_before, balance_after, locked_before, locked_after, reference_id, description
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		history.UserAddress, history.TxID, history.OrderID, history.ChangeType, history.Amount,
		history.BalanceBefore, history.BalanceAfter, history.LockedBefore, history.LockedAfter,
		history.ReferenceID, history.Description,
	)
	return err
}

// LockMultipleWalletsWithTx implements [UserWalletRepository].
func (u *userWalletRepository) LockMultipleWalletsWithTx(tx *sqlx.Tx, addresses []string) error {
	if len(addresses) == 0 {
		return nil
	}

	query, args, err := sqlx.In(`
		SELECT user_address
		FROM user_wallets
		WHERE user_address IN (?)
		FOR UPDATE
	`, addresses)

	if err != nil {
		return err
	}

	_, err = tx.Exec(tx.Rebind(query), args...)
	return err
}

// UpdateWalletWithTx implements [UserWalletRepository].
func (u *userWalletRepository) UpdateWalletWithTx(tx *sqlx.Tx, address string, newBalance float64) error {
	_, err := tx.Exec(`
		UPDATE user_wallets
		SET yte_balance = ?, last_transaction_at = NOW()
		WHERE user_address = ?`,
		newBalance, address,
	)
	return err
}

// UpsertEmptyIfNotExistsWithTx implements [UserWalletRepository].
func (u *userWalletRepository) UpsertEmptyIfNotExistsWithTx(tx *sqlx.Tx, address string) error {
	_, err := tx.Exec(`
		INSERT INTO user_wallets (user_address, yte_balance, locked_balance, total_received, total_sent)
		VALUES (?, 0, 0, 0, 0)
		ON DUPLICATE KEY UPDATE user_address = user_address
	`, address)

	return err
}

// UpsertEmptyIfNotExists implements [UserWalletRepository].
func (u *userWalletRepository) UpsertEmptyIfNotExists(address string) error {
	_, err := u.db.Exec(`
		INSERT INTO user_wallets (user_address, yte_balance, locked_balance, total_received, total_sent)
		VALUES (?, 0, 0, 0, 0)
		ON DUPLICATE KEY UPDATE user_address = user_address
	`, address)

	return err
}

// BulkUpdateBalancesWithTx implements [UserWalletRepository].
func (u *userWalletRepository) BulkUpdateBalancesWithTx(tx *sqlx.Tx, balances map[string]float64) error {
	if len(balances) == 0 {
		return nil
	}

	query := `UPDATE user_wallets SET yte_balance = CASE user_address `
	var args []interface{}
	var addresses []interface{}

	for addr, bal := range balances {
		query += ` WHEN ? THEN ? `
		args = append(args, addr, bal)
		addresses = append(addresses, addr)
	}

	query += `END, last_transaction_at = NOW() WHERE user_address IN (?)`
	finalArgs := append(args, addresses)

	finalQuery, finalQueryArgs, err := sqlx.In(query, finalArgs...)

	if err != nil {
		return err
	}

	_, err = tx.Exec(tx.Rebind(finalQuery), finalQueryArgs...)
	return err
}
