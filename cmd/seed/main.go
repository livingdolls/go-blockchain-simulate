package main

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/config"
)

// Seed data untuk demo/development.
// Jalankan: go run cmd/seed/main.go
//
// Script ini membuat sample users, wallets, balances, dan transactions
// agar developer bisa langsung explore aplikasi tanpa manual input.
//
// Requirements:
// - Database sudah di-migrate (tables exist)
// - config.yaml atau config.local.yaml sudah di-set dengan DSN valid

var sampleUsers = []struct {
	Username string
}{
	{"alice"},
	{"bob"},
	{"charlie"},
	{"dave"},
	{"eve"},
	{"frank"},
	{"grace"},
	{"heidi"},
	{"ivan"},
	{"judy"},
}

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Gagal load config: %v", err)
	}

	db, err := sqlx.Connect(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Gagal connect DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	_ = ctx

	fmt.Println("🌱 Mulai seeding data...")

	// Generate users
	users := generateUsers(db)
	fmt.Printf("✅ %d users dibuat\n", len(users))

	// Create wallets
	createWallets(db, users)
	fmt.Printf("✅ %d wallets dibuat\n", len(users))

	// Create balances
	createBalances(db, users)
	fmt.Printf("✅ %d balances dibuat\n", len(users))

	// Generate transactions
	txCount := generateTransactions(db, users)
	fmt.Printf("✅ %d transactions dibuat\n", txCount)

	fmt.Println("🎉 Seeding selesai!")
}

type seedUser struct {
	Address   string
	PublicKey string
	Username  string
}

func generateUsers(db *sqlx.DB) []seedUser {
	var users []seedUser

	for _, u := range sampleUsers {
		privKey, err := crypto.GenerateKey()
		if err != nil {
			log.Printf("Gagal generate key untuk %s: %v", u.Username, err)
			continue
		}

		pubKey := privKey.Public().(*ecdsa.PublicKey)
		address := crypto.PubkeyToAddress(*pubKey).Hex()
		pubKeyBytes := crypto.FromECDSAPub(pubKey)
		pubKeyHex := hex.EncodeToString(pubKeyBytes)

		_, err = db.Exec(
			`INSERT INTO users (name, address, public_key, created_at)
			 VALUES (?, ?, ?, NOW())
			 ON DUPLICATE KEY UPDATE name = VALUES(name)`,
			u.Username, address, pubKeyHex,
		)
		if err != nil {
			log.Printf("Gagal insert user %s: %v", u.Username, err)
			continue
		}

		users = append(users, seedUser{
			Address:   address,
			PublicKey: pubKeyHex,
			Username:  u.Username,
		})
	}

	return users
}

func createWallets(db *sqlx.DB, users []seedUser) {
	for _, u := range users {
		_, err := db.Exec(
			`INSERT INTO user_wallets (user_address, yte_balance, last_transaction_at)
			 VALUES (?, 0, NOW())
			 ON DUPLICATE KEY UPDATE user_address = VALUES(user_address)`,
			u.Address,
		)
		if err != nil {
			log.Printf("Gagal create wallet untuk %s: %v", u.Username, err)
		}
	}
}

func createBalances(db *sqlx.DB, users []seedUser) {
	for _, u := range users {
		// Random USD balance antara 1000 - 10000
		usdBalance := 1000.0 + rand.Float64()*9000.0

		_, err := db.Exec(
			`INSERT INTO user_balances (user_address, usd_balance, locked_balance, total_deposited, total_withdrawn, total_traded)
			 VALUES (?, ?, 0, ?, 0, 0)
			 ON DUPLICATE KEY UPDATE usd_balance = VALUES(usd_balance)`,
			u.Address, usdBalance, usdBalance,
		)
		if err != nil {
			log.Printf("Gagal create balance untuk %s: %v", u.Username, err)
		}
	}
}

func generateTransactions(db *sqlx.DB, users []seedUser) int {
	if len(users) < 2 {
		return 0
	}

	count := 0
	txTypes := []string{"TRANSFER", "TRANSFER", "TRANSFER", "BUY", "SELL"}

	for i := 0; i < 50; i++ {
		// Random sender dan receiver (berbeda)
		senderIdx := rand.Intn(len(users))
		receiverIdx := rand.Intn(len(users))
		for receiverIdx == senderIdx {
			receiverIdx = rand.Intn(len(users))
		}

		sender := users[senderIdx]
		receiver := users[receiverIdx]
		txType := txTypes[rand.Intn(len(txTypes))]

		amount := 1.0 + rand.Float64()*99.0 // 1 - 100 YTE
		fee := calculateFee(amount)

		_, err := db.Exec(
			`INSERT INTO transactions (from_address, to_address, amount, fee, type, signature, status, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, 'PENDING', NOW())`,
			sender.Address, receiver.Address, amount, fee, txType, "seed-"+fmt.Sprintf("%d", time.Now().UnixNano()),
		)
		if err != nil {
			log.Printf("Gagal insert tx: %v", err)
			continue
		}

		count++
	}

	return count
}

func calculateFee(amount float64) float64 {
	if amount < 10 {
		return 0.001
	}
	if amount < 100 {
		return 0.01
	}
	fee := amount * 0.001
	if fee < 0.001 {
		fee = 0.001
	}
	return fee
}

// Helper: create empty SQL null string
func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
