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
	for i := 1; i < len(blocks); i++ {
		block := blocks[i]
		prevBlock := blocks[i-1]

		// 1. check previous hash link
		if block.PreviousHash != prevBlock.CurrentHash {
			return fmt.Errorf("block %d: previous hash mismatch", block.BlockNumber)
		}

		// 2. Proof ofWork Validation
		if !ValidateProofOfWork(block) {
			return fmt.Errorf("block %d: invalid proof of work", block.BlockNumber)
		}

		// 3. Hash recalculation
		calculatedHash := RecalculateBlockHash(block)
		if block.CurrentHash != calculatedHash {
			return fmt.Errorf("block %d: hash mismatch", block.BlockNumber)
		}

		// 4. Check timestamp squence
		if block.Timestamp <= prevBlock.Timestamp {
			return fmt.Errorf("block %d: timestamp not greater than previous block", block.BlockNumber)
		}

		// Skip genesis block hash validation
		if i > 1 {
			// 5. Block hash calculation
			calculatedHash := CalculateBlockHash(block)

			if block.CurrentHash != calculatedHash {
				return fmt.Errorf("block %d: calculated hash mismatch", block.BlockNumber)
			}
		}
	}

	fmt.Println("✅ Blockchain integrity check passed.")

	return nil
}
