package database

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

// DBConn membungkus sqlx.DB untuk koneksi database utama.
type DBConn struct {
	db *sqlx.DB
}

// Database adalah interface abstraksi koneksi database.
// Dipakai agar unit test bisa mock dengan mudah.
type Database interface {
	GetDB() *sqlx.DB
	Close() error
}

// Config berisi parameter koneksi database.
// Dipakai agar package database tidak bergantung langsung ke package config.
type Config struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewDBConn membuat koneksi database baru dari konfigurasi.
func NewDBConn(cfg Config) (Database, error) {
	db, err := openDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("gagal konek ke database utama: %w", err)
	}

	return &DBConn{db: db}, nil
}

// Close menutup koneksi database.
func (d *DBConn) Close() error {
	return d.db.Close()
}

// GetDB mengembalikan instance sqlx.DB untuk diakses oleh repository.
func (d *DBConn) GetDB() *sqlx.DB {
	return d.db
}

func openDatabase(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping error: %w", err)
	}

	logger.LogInfo("koneksi database berhasil dibuka")

	return db, nil
}
