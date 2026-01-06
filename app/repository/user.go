package repository

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

type UserRepository interface {
	Create(user models.User) error
	GetByAddress(address string) (models.User, error)
	GetByAddressWithBalance(address string) (models.UserWithBalance, error)
	BeginTx() (*sqlx.Tx, error)
	GetMultipleByAddress(addresses []string) ([]models.User, error)
	GetMultipleByAddressWithTx(tx *sqlx.Tx, addresses []string) ([]models.User, error)
	GetUserWithWallet(address string) (models.User, models.UserWallet, error)
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
		INSERT INTO users (name, address, public_key)
		VALUES (?, ?, ?)
	`

	_, err := r.db.Exec(query, user.Name, user.Address, user.PublicKey)
	return err
}

func (r *userRepository) GetByAddress(address string) (models.User, error) {
	var user models.User

	err := r.db.Get(&user, "SELECT id, name, address, public_key FROM users WHERE address = ?", address)
	return user, err
}

func (r *userRepository) GetByAddressWithBalance(address string) (models.UserWithBalance, error) {
	var user models.UserWithBalance

	query := `
	SELECT 
		us.id, 
		us.name, 
		us.address, 
		us.public_key, 
		COALESCE(uw.yte_balance, 0) AS yte_balance,
    	COALESCE(ub.usd_balance, 0) AS usd_balance 
	FROM users as us 
	LEFT JOIN user_wallets as uw on us.address = uw.user_address 
	LEFT JOIN user_balances as ub on us.address = ub.user_address  
	WHERE us.address = ?
	`

	err := r.db.Get(&user, query, address)

	if err == sql.ErrNoRows {
		return user, entity.ErrUserNotFound
	}

	return user, err
}

// GetMultipleByAddress retrieves multiple users by addresses in a single query
func (r *userRepository) GetMultipleByAddress(addresses []string) ([]models.User, error) {
	if len(addresses) == 0 {
		return []models.User{}, nil
	}

	var users []models.User
	query, args, err := sqlx.In(`SELECT id, name, address, public_key FROM users WHERE address IN (?)`, addresses)
	if err != nil {
		return nil, err
	}

	err = r.db.Select(&users, r.db.Rebind(query), args...)
	return users, err
}

func (r *userRepository) GetMultipleByAddressWithTx(tx *sqlx.Tx, addresses []string) ([]models.User, error) {
	if len(addresses) == 0 {
		return []models.User{}, nil
	}

	var users []models.User
	query, args, err := sqlx.In(`SELECT id, name, address, public_key FROM users WHERE address in (?)`, addresses)

	if err != nil {
		return nil, err
	}

	err = tx.Select(&users, tx.Rebind(query), args...)
	return users, err
}

func (r *userRepository) GetUserWithWallet(address string) (models.User, models.UserWallet, error) {

	var user models.User
	var wallet models.UserWallet

	query := `
        SELECT 
            u.id, u.name, u.address, u.public_key,
            COALESCE(w.user_address, '') as user_address,
            COALESCE(w.yte_balance, 0) as yte_balance,
            COALESCE(w.locked_balance, 0) as locked_balance,
            COALESCE(w.total_received, 0) as total_received,
            COALESCE(w.total_sent, 0) as total_sent,
            w.last_transaction_at
        FROM users u
        LEFT JOIN user_wallets w ON u.address = w.user_address
        WHERE u.address = ?
    `

	type result struct {
		ID                int64   `db:"id"`
		Name              string  `db:"name"`
		Address           string  `db:"address"`
		PublicKey         string  `db:"public_key"`
		UserAddress       string  `db:"user_address"`
		YTEBalance        float64 `db:"yte_balance"`
		LockedBalance     float64 `db:"locked_balance"`
		TotalReceived     float64 `db:"total_received"`
		TotalSent         float64 `db:"total_sent"`
		LastTransactionAt string  `db:"last_transaction_at"`
	}

	var res result
	err := r.db.Get(&res, query, address)
	if err != nil {
		return user, wallet, err
	}

	user = models.User{
		ID:        int(res.ID),
		Name:      res.Name,
		Address:   res.Address,
		PublicKey: res.PublicKey,
	}

	wallet = models.UserWallet{
		UserAddress:     res.UserAddress,
		YTEBalance:      res.YTEBalance,
		LockedBalance:   res.LockedBalance,
		TotalReceived:   res.TotalReceived,
		TotalSent:       res.TotalSent,
		LastTransaction: res.LastTransactionAt,
	}

	return user, wallet, nil
}
