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
