package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
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

	users := generateUsers(db)
	fmt.Printf("✅ %d users dibuat\n", len(users))

	createWallets(db, users)
	fmt.Printf("✅ %d wallets dibuat\n", len(users))

	createBalances(db, users)
	fmt.Printf("✅ %d balances dibuat\n", len(users))

	txCount := generateTransactions(db, users)
	fmt.Printf("✅ %d transactions dibuat\n", txCount)

	blocks := generateBlocks(db, txCount)
	fmt.Printf("✅ %d blocks dibuat\n", len(blocks))

	ticks := generateMarketTicks(db, blocks)
	fmt.Printf("✅ %d market ticks dibuat\n", len(ticks))

	candleCount := generateCandles(db, ticks)
	fmt.Printf("✅ %d candles dibuat\n", candleCount)

	fmt.Println("🎉 Seeding selesai! Chart dashboard sekarang terisi data.")
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
			 VALUES (?, 1000, NOW())
			 ON DUPLICATE KEY UPDATE yte_balance = GREATEST(yte_balance, 1000)`,
			u.Address,
		)
		if err != nil {
			log.Printf("Gagal create wallet untuk %s: %v", u.Username, err)
		}
	}
}

func createBalances(db *sqlx.DB, users []seedUser) {
	for _, u := range users {
		usdBalance := 1000.0 + rand.Float64()*9000.0

		_, err := db.Exec(
			`INSERT INTO user_balances (user_address, usd_balance, locked_balance, total_deposited, total_withdrawn, total_traded)
			 VALUES (?, ?, 0, ?, 0, 0)
			 ON DUPLICATE KEY UPDATE usd_balance = VALUES(usd_balance), total_deposited = VALUES(total_deposited)`,
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

	for i := 0; i < 200; i++ {
		senderIdx := rand.Intn(len(users))
		receiverIdx := rand.Intn(len(users))
		for receiverIdx == senderIdx {
			receiverIdx = rand.Intn(len(users))
		}

		sender := users[senderIdx]
		receiver := users[receiverIdx]
		txType := txTypes[rand.Intn(len(txTypes))]

		amount := 1.0 + rand.Float64()*99.0
		fee := calculateFee(amount)

		statuses := []string{"SUCCESS", "SUCCESS", "SUCCESS", "CONFIRMED", "PENDING"}
		status := statuses[rand.Intn(len(statuses))]

		_, err := db.Exec(
			`INSERT INTO transactions (from_address, to_address, amount, fee, type, signature, status, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, NOW())`,
			sender.Address, receiver.Address, amount, fee, txType,
			"seed-"+fmt.Sprintf("%d", time.Now().UnixNano()),
			status,
		)
		if err != nil {
			log.Printf("Gagal insert tx: %v", err)
			continue
		}

		count++
	}

	return count
}

type seedBlock struct {
	ID          int64
	BlockNumber int
}

func generateBlocks(db *sqlx.DB, txCount int) []seedBlock {
	numBlocks := 50
	blocks := make([]seedBlock, 0, numBlocks)

	prevHash := "0"
	for i := 0; i < numBlocks; i++ {
		blockNumber := i + 2
		nonce := rand.Int63n(999999)
		difficulty := 4

		ts := time.Now().Add(-time.Duration(numBlocks-i) * 10 * time.Second).Unix()

		hashInput := fmt.Sprintf("%d%s%d%d", blockNumber, prevHash, nonce, ts)
		hashSum := sha256.Sum256([]byte(hashInput))
		currentHash := hex.EncodeToString(hashSum[:])

		minerIdx := rand.Intn(len(sampleUsers))
		miner := "0x" + fmt.Sprintf("%040x", rand.Int63())

		if len(miner) > 10 {
			miner = miner[:42]
		}

		blockReward := 1.0 + rand.Float64()*0.5
		totalFees := rand.Float64() * 0.1

		merkleRoot := hex.EncodeToString(hashSum[16:])

		result, err := db.Exec(
			`INSERT INTO blocks (block_number, previous_hash, current_hash, nonce, difficulty, timestamp, merkle_root, miner_address, block_reward, total_fees, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, FROM_UNIXTIME(?))`,
			blockNumber, prevHash, currentHash, nonce, difficulty, ts, merkleRoot, miner, blockReward, totalFees, ts,
		)
		if err != nil {
			log.Printf("Gagal insert block %d: %v", blockNumber, err)
			continue
		}

		id, _ := result.LastInsertId()
		blocks = append(blocks, seedBlock{ID: id, BlockNumber: blockNumber})
		prevHash = currentHash

		_ = minerIdx
	}

	return blocks
}

func generateMarketTicks(db *sqlx.DB, blocks []seedBlock) []int64 {
	var tickIDs []int64
	price := 1.0

	for _, block := range blocks {
		delta := (rand.Float64() - 0.5) * 0.002
		price += delta
		if price < 0.998 {
			price = 0.998
		}
		if price > 1.002 {
			price = 1.002
		}

		buyVol := rand.Float64() * 50
		sellVol := rand.Float64() * 50
		txCount := rand.Intn(5) + 1

		result, err := db.Exec(
			`INSERT INTO market_ticks (block_id, price, buy_volume, sell_volume, tx_count, created_at)
			 VALUES (?, ?, ?, ?, ?, NOW())`,
			block.ID, price, buyVol, sellVol, txCount,
		)
		if err != nil {
			log.Printf("Gagal insert market tick untuk block %d: %v", block.BlockNumber, err)
			continue
		}

		id, _ := result.LastInsertId()
		tickIDs = append(tickIDs, id)
	}

	return tickIDs
}

func generateCandles(db *sqlx.DB, tickIDs []int64) int {
	intervals := []string{"1m", "5m", "15m", "1h", "4h", "1d"}
	count := 0
	baseTime := time.Now().Unix()

	for _, interval := range intervals {
		var stepSec int64
		switch interval {
		case "1m":
			stepSec = 60
		case "5m":
			stepSec = 300
		case "15m":
			stepSec = 900
		case "1h":
			stepSec = 3600
		case "4h":
			stepSec = 14400
		case "1d":
			stepSec = 86400
		}

		numCandles := 50
		if interval == "1d" {
			numCandles = 30
		}
		if interval == "4h" {
			numCandles = 42
		}

		price := 1.0
		for i := 0; i < numCandles; i++ {
			startTime := baseTime - int64(numCandles-i)*stepSec

			delta := (rand.Float64() - 0.45) * 0.0004
			openPrice := price
			closePrice := price + delta
			if closePrice < 0.995 {
				closePrice = 0.995
			}
			if closePrice > 1.005 {
				closePrice = 1.005
			}

			highPrice := max(openPrice, closePrice) + rand.Float64()*0.0002
			lowPrice := min(openPrice, closePrice) - rand.Float64()*0.0002
			volume := rand.Float64() * 100

			_, err := db.Exec(
				`INSERT INTO candles (interval_type, start_time, open_price, high_price, low_price, close_price, volume)
				 VALUES (?, ?, ?, ?, ?, ?, ?)
				 ON DUPLICATE KEY UPDATE close_price = VALUES(close_price)`,
				interval, startTime, openPrice, highPrice, lowPrice, closePrice, volume,
			)
			if err != nil {
				log.Printf("Gagal insert candle %s: %v", interval, err)
				continue
			}

			price = closePrice
			count++
		}
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

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
