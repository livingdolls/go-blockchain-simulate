package repository

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
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

// Create menyimpan user baru ke DB. Return:
//
//   - nil: sukses
//   - entity.ErrAddressAlreadyRegistered: address sudah ada (UNIQUE constraint)
//   - entity.ErrUsernameAlreadyExists: name sudah ada (kalau ada UNIQUE di name)
//   - error lain: DB infrastructure error (caller map ke 500)
//
// Implementasi: deteksi MySQL error 1062 (duplicate entry) dari driver,
// lalu classify field mana yang duplicate dengan parse message string.
// trade-off: parse message string lebih fragile dari info schema, tapi
// tidak perlu query INFORMATION_SCHEMA setiap insert.
func (r *userRepository) Create(user models.User) error {
	query := `
		INSERT INTO users (name, address, public_key)
		VALUES (?, ?, ?)
	`

	_, err := r.db.Exec(query, user.Name, user.Address, user.PublicKey)
	if err != nil {
		return classifyMySQLInsertError(err, user)
	}
	return nil
}

// classifyMySQLInsertError menerjemahkan error MySQL ke entity error
// yang lebih semantik. Dipakai oleh semua repository yang insert ke
// tabel dengan UNIQUE constraint.
//
// MySQL error 1062 format: "Error 1062 (23000): Duplicate entry 'xxx' for key 'users.YY'"
// YY = nama index (biasanya: address, name, atau nama constraint).
func classifyMySQLInsertError(err error, _ models.User) error {
	if err == nil {
		return nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		// 1062 = ER_DUP_ENTRY (duplicate key)
		if mysqlErr.Number == 1062 {
			// Cek index name untuk tahu field mana yang duplicate.
			// Index name format: "table.column" atau "table.custom_name".
			// Pakai case-insensitive contains untuk handle kedua format.
			msg := strings.ToLower(mysqlErr.Message)
			switch {
			case strings.Contains(msg, "address"):
				return entity.ErrAddressAlreadyRegistered
			case strings.Contains(msg, "name"), strings.Contains(msg, "username"):
				return entity.ErrUsernameAlreadyExists
			default:
				// Duplicate di field lain - generic conflict.
				return entity.ErrConflict
			}
		}
	}

	// Bukan duplicate (mungkin connection lost, deadlock, syntax error, dll).
	// Wrap dengan entity.ErrDatabase untuk caller bisa distinguish dari
	// expected business errors.
	return errors.Join(entity.ErrDatabase, err)
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
		COALESCE(uw.yte_balance, 0)  AS yte_balance,
		COALESCE(ub.usd_balance, 0)  AS usd_balance
	FROM users us
	LEFT JOIN user_wallets uw ON uw.user_address = us.address
	LEFT JOIN user_balances ub ON ub.user_address = us.address
	WHERE us.address = ?
	LIMIT 1
	`

	err := r.db.Get(&user, query, address)
	return user, err
}

func (r *userRepository) GetMultipleByAddress(addresses []string) ([]models.User, error) {
	var users []models.User
	if len(addresses) == 0 {
		return users, nil
	}

	q, args, err := sqlx.In("SELECT id, name, address, public_key FROM users WHERE address IN (?)", addresses)
	if err != nil {
		return nil, err
	}
	q = r.db.Rebind(q)
	err = r.db.Select(&users, q, args...)
	return users, err
}

func (r *userRepository) GetMultipleByAddressWithTx(tx *sqlx.Tx, addresses []string) ([]models.User, error) {
	var users []models.User
	if len(addresses) == 0 {
		return users, nil
	}

	q, args, err := sqlx.In("SELECT id, name, address, public_key FROM users WHERE address IN (?) FOR UPDATE", addresses)
	if err != nil {
		return nil, err
	}
	q = tx.Rebind(q)
	err = tx.Select(&users, q, args...)
	return users, err
}

func (r *userRepository) GetUserWithWallet(address string) (models.User, models.UserWallet, error) {
	var (
		user   models.User
		wallet models.UserWallet
	)

	// Catatan: schema user_wallets saat ini tidak punya kolom
	// user_id/address/nonce - PK adalah user_address. Query ini join
	// lewat user_address. Field wallet_address & nonce di-comment
	// sementara sampai schema berkembang.
	const query = `
		SELECT
			us.id, us.name, us.address, us.public_key, us.role,
			uw.user_address, uw.yte_balance, uw.locked_balance,
			uw.available_balance, uw.total_received, uw.total_sent,
			uw.last_transaction_at
		FROM users us
		LEFT JOIN user_wallets uw ON uw.user_address = us.address
		WHERE us.address = ?
		LIMIT 1
	`

	type result struct {
		models.User
		UserAddress      sql.NullString `db:"user_address"`
		YTEBalance       sql.NullFloat64 `db:"yte_balance"`
		LockedBalance    sql.NullFloat64 `db:"locked_balance"`
		AvailableBalance sql.NullFloat64 `db:"available_balance"`
		TotalReceived    sql.NullFloat64 `db:"total_received"`
		TotalSent        sql.NullFloat64 `db:"total_sent"`
		LastTransaction  sql.NullString  `db:"last_transaction_at"`
	}

	var r2 result
	if err := r.db.Get(&r2, query, address); err != nil {
		return models.User{}, models.UserWallet{}, err
	}

	user = r2.User
	wallet = models.UserWallet{
		UserAddress:      r2.UserAddress.String,
		YTEBalance:       r2.YTEBalance.Float64,
		LockedBalance:    r2.LockedBalance.Float64,
		AvailableBalance: r2.AvailableBalance.Float64,
		TotalReceived:    r2.TotalReceived.Float64,
		TotalSent:        r2.TotalSent.Float64,
		LastTransaction:  r2.LastTransaction,
	}

	return user, wallet, nil
}
