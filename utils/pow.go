package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

const (
	// target difficulty for proof-of-work
	// 3 = easy (1-2 second)
	// 4 = medium (5-10 seconds)
	// 5 = hard (30+ seconds)
	// 6 = very hard (5-10 minutes)
	DefaultDifficulty = 4

	// Target block time in seconds
	TargetBlockTime = 10

	// Difficulty adjustment interval in blocks
	DifficultyAdjustmentInterval = 10
)

type MiningResult struct {
	Hash       string
	Nonce      int64
	Duration   time.Duration
	HashRate   float64 // hashes per second
	Difficulty int
}

func MineBlock(blockNumber int, prevHash string, transactions []models.Transaction, difficulty int) MiningResult {
	startTime := time.Now()

	// calculate targtet (number with difficulty leading zeros)
	target := strings.Repeat("0", difficulty)

	var nonce int64 = 0
	var hash string
	var attempts int64 = 0

	fmt.Printf("Mining block #%d with difficulty %d (target: %s...)\n", blockNumber, difficulty, target)

	for {
		// Create block data string
		data := fmt.Sprintf("%d%s%v%d%d", blockNumber, prevHash, transactions, nonce, time.Now().UnixNano())

		// calculate SHA-256 hash
		hashBytes := sha256.Sum256([]byte(data))
		hash = hex.EncodeToString(hashBytes[:])

		attempts++

		// check if hash meets difficulty requirement
		if strings.HasPrefix(hash, target) {
			duration := time.Since(startTime)
			hashRate := float64(attempts) / duration.Seconds()

			fmt.Printf("Block mined! Hash: %s\n", hash)
			fmt.Printf("Nonce: %d, Attempts: %d, Time taken: %s, Hash Rate: %.2f hashes/sec\n", nonce, attempts, duration, hashRate)

			return MiningResult{
				Hash:       hash,
				Nonce:      nonce,
				Duration:   duration,
				HashRate:   hashRate,
				Difficulty: difficulty,
			}
		}

		nonce++

		// Proress indicator every 100k attempts

		if attempts%100000 == 0 {
			elapsed := time.Since(startTime)
			hashRate := float64(attempts) / elapsed.Seconds()
			fmt.Printf("  Tried %d hashes in %s (%.2f hashes/sec)\n", attempts, elapsed, hashRate)
		}

		// safety break to avoid infinite loop in testing

		if time.Since(startTime) > 10*time.Minute {
			return MiningResult{
				Hash:       "",
				Nonce:      nonce,
				Duration:   time.Since(startTime),
				HashRate:   0,
				Difficulty: difficulty,
			}
		}
	}
}

// ValidateProofOfWork, verifies that a block hash meets the required difficulty
func ValidateProofOfWork(block models.Block) bool {
	target := strings.Repeat("0", block.Difficulty)
	return strings.HasPrefix(block.CurrentHash, target)
}

// CalculateNextDifficulty adjust difficulty based on recent block times
func CalculateNextDifficulty(blocks []models.Block) int {
	if len(blocks) < DifficultyAdjustmentInterval {
		return DefaultDifficulty
	}

	// Get last N blocks for analysis
	recentBlocks := blocks[len(blocks)-DifficultyAdjustmentInterval:]

	// Calculate actual time taken for last N blocks
	firstBlockTime := recentBlocks[0].Timestamp
	lastBlockTime := recentBlocks[len(recentBlocks)-1].Timestamp
	actualTime := lastBlockTime - firstBlockTime

	// expected time for N blocks
	expectedTime := int64(TargetBlockTime * (DifficultyAdjustmentInterval - 1))

	// Current difficulty
	currentDifficulty := recentBlocks[len(recentBlocks)-1].Difficulty

	fmt.Printf("Difficulty adjustment analysis:\n")
	fmt.Printf("Expected time: %d seconds\n", expectedTime)
	fmt.Printf("Actual time: %d seconds\n", actualTime)
	fmt.Printf("Current difficulty: %d\n", currentDifficulty)

	// Adjust difficulty
	// if blocks are mined to fast, increase difficulty
	if actualTime < expectedTime/2 {
		newDifficulty := currentDifficulty + 1
		fmt.Printf("Increasing difficulty to %d\n", newDifficulty)
		return newDifficulty
	}

	// If blocks are mined too slow, decrease difficulty
	if actualTime > expectedTime*2 {
		if currentDifficulty > 1 {
			newDifficulty := currentDifficulty - 1
			fmt.Printf("Decreasing difficulty to %d\n", newDifficulty)
			return newDifficulty
		}
	}

	fmt.Printf("Keeping difficulty at %d\n", currentDifficulty)
	return currentDifficulty
}

// Recalculate block hash with nonce for validation
func RecalculateBlockHash(block models.Block) string {
	data := fmt.Sprintf("%d%s%v%d%d",
		block.BlockNumber,
		block.PreviousHash,
		block.Transactions,
		block.Nonce,
		block.Timestamp,
	)

	hashBytes := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hashBytes[:])
}

// Get difficulty target converts difficulty to target number
func GetDifficultyTarget(difficulty int) *big.Int {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty*4))
	return target
}
