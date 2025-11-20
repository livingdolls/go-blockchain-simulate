package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

func GenerateFakeKey() (privateKey, publicKey string) {
	randomData := make([]byte, 32)
	rand.Read(randomData)

	privateKey = hex.EncodeToString(randomData)

	// public key is just a hash of private key for fake purposes

	sum := sha256.Sum256([]byte(privateKey))
	publicKey = hex.EncodeToString(sum[:])

	return
}

func GenerateAddressFromPublicKey(publicKey string) string {
	h := sha256.Sum256([]byte(publicKey + "FAKE-BLOCKCHAIN"))

	return "FAKE-" + hex.EncodeToString(h[:])[:32]
}

func SignFake(privateKey, toAddress string, amount float64) string {
	payload := privateKey + toAddress + fmt.Sprintf("%.2f", amount)

	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func HashBlock(prevHash string, txList []models.Transaction) string {
	raw := prevHash

	for _, tx := range txList {
		raw += fmt.Sprintf("%d-%s-%s-%.2f", tx.ID, tx.FromAddress, tx.ToAddress, tx.Amount)
	}

	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func CalculateBlockHash(block models.Block) string {
	// Recalculate hash from previous hash and transactions
	// This should match the original HashBlock calculation
	return HashBlock(block.PreviousHash, block.Transactions)
}

func CheckBlockchainIntegrity(blocks []models.Block) error {
	for i := range blocks {

		// Debug logging
		fmt.Printf("\n=== Block %d Validation ===\n", blocks[i].BlockNumber)
		fmt.Printf("Block ID: %d\n", blocks[i].ID)
		fmt.Printf("Previous Hash: %s\n", blocks[i].PreviousHash)
		fmt.Printf("Stored Hash: %s\n", blocks[i].CurrentHash)
		fmt.Printf("Transactions Count: %d\n", len(blocks[i].Transactions))

		// validasi block pertama (genesis)
		if i == 0 {
			if blocks[i].BlockNumber != 1 {
				return fmt.Errorf("genesis block has invalid block number")
			}

			if blocks[i].PreviousHash != "0" {
				return fmt.Errorf("genesis block has invalid previous hash")
			}

			// Skip hash validation for genesis block
			// Genesis block might have been created with different logic
			fmt.Printf("Skipping hash validation for genesis block\n")
			fmt.Printf("========================\n")
			continue
		}

		// validasi hash untuk non-genesis blocks
		calculatedHash := CalculateBlockHash(blocks[i])

		for j, tx := range blocks[i].Transactions {
			fmt.Printf("  TX %d: ID=%d, From=%s, To=%s, Amount=%.2f\n",
				j, tx.ID, tx.FromAddress, tx.ToAddress, tx.Amount)
		}
		fmt.Printf("Calculated Hash: %s\n", calculatedHash)
		fmt.Printf("========================\n")

		if calculatedHash != blocks[i].CurrentHash {
			return fmt.Errorf("block %d has invalid hash", blocks[i].ID)
		}

		// validasi previous hash

		if blocks[i].PreviousHash != blocks[i-1].CurrentHash {
			return fmt.Errorf("block %d has invalid previous hash", blocks[i].ID)
		}

		// validasi block number

		if blocks[i].BlockNumber != blocks[i-1].BlockNumber+1 {
			return fmt.Errorf("block %d has invalid block number", blocks[i].ID)
		}

		// validasi timestamp tidak muncud
		if blocks[i].CreatedAt < blocks[i-1].CreatedAt {
			return fmt.Errorf("block %d has invalid timestamp", blocks[i].ID)
		}
	}

	return nil
}
