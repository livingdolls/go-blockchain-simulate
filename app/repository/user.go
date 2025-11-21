package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type UserRepository interface {
	Create(user models.User) error
	GetByAddress(address string) (models.User, error)
	UpdateBalanceWithTx(dbTx *sqlx.Tx, address string, balance float64) error
	BeginTx() (*sqlx.Tx, error)
	GetMultipleByAddress(addresses []string) ([]models.User, error)
	LockMultipleUsersWithTx(tx *sqlx.Tx, addresses []string) error
	BulkUpdateBalancesWithTx(tx *sqlx.Tx, balances map[string]float64) error
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) BeginTx() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r *userRepository) Create(user models.User) error {
	query := `
		INSERT INTO users (name, address, public_key, private_key, balance)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query, user.Name, user.Address, user.PublicKey, user.PrivateKey, user.Balance)
	return err
}

func (r *userRepository) GetByAddress(address string) (models.User, error) {
	var user models.User

	err := r.db.Get(&user, "SELECT id, name, address, public_key, private_key, balance FROM users WHERE address = ?", address)
	return user, err
}

func (r *userRepository) UpdateBalanceWithTx(dbTx *sqlx.Tx, address string, balance float64) error {
	_, err := dbTx.Exec("UPDATE users SET balance = ? WHERE address = ?", balance, address)
	return err
}

// GetMultipleByAddress retrieves multiple users by addresses in a single query
func (r *userRepository) GetMultipleByAddress(addresses []string) ([]models.User, error) {
	if len(addresses) == 0 {
		return []models.User{}, nil
	}

	var users []models.User
	query, args, err := sqlx.In(`SELECT id, name, address, public_key, private_key, balance FROM users WHERE address IN (?)`, addresses)
	if err != nil {
		return nil, err
	}

	err = r.db.Select(&users, r.db.Rebind(query), args...)
	return users, err
}

// LockMultipleUsersWithTx locks multiple user rows in a single query
func (r *userRepository) LockMultipleUsersWithTx(tx *sqlx.Tx, addresses []string) error {
	if len(addresses) == 0 {
		return nil
	}

	query, args, err := sqlx.In(`SELECT id FROM users WHERE address IN (?) FOR UPDATE`, addresses)
	if err != nil {
		return err
	}

	_, err = tx.Exec(tx.Rebind(query), args...)
	return err
}

// BulkUpdateBalancesWithTx updates multiple user balances in a single query using CASE statement
func (r *userRepository) BulkUpdateBalancesWithTx(tx *sqlx.Tx, balances map[string]float64) error {
	if len(balances) == 0 {
		return nil
	}

	// Build CASE statement for bulk update
	query := `UPDATE users SET balance = CASE address `
	var args []interface{}
	var addresses []interface{}

	for addr, bal := range balances {
		query += `WHEN ? THEN ? `
		args = append(args, addr, bal)
		addresses = append(addresses, addr)
	}

	query += `END WHERE address IN (?)`

	// Combine args
	finalArgs := append(args, addresses)

	finalQuery, finalQueryArgs, err := sqlx.In(query, finalArgs...)
	if err != nil {
		return err
	}

	_, err = tx.Exec(tx.Rebind(finalQuery), finalQueryArgs...)
	return err
}
