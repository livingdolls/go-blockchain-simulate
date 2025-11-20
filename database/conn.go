package database

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type DBConn struct {
	db *sqlx.DB
}

type Database interface {
	GetDB() *sqlx.DB
	Close() error
}

func NewDBConn() (Database, error) {
	db, err := openDatabase("mysql", "yurina:hirate@tcp(172.17.0.1:3306)/blockchain?parseTime=true&loc=Local")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to main DB: %w", err)
	}

	return &DBConn{
		db: db,
	}, nil
}

// Close implements Database.
func (d *DBConn) Close() error {
	return d.db.Close()
}

// GetDB implements Database.
func (d *DBConn) GetDB() *sqlx.DB {
	return d.db
}

func openDatabase(driver, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping error: %w", err)
	}

	log.Printf("Connected to %s database successfully!", dsn)

	return db, nil
}
