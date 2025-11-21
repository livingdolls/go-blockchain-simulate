package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
)

func CalculateMerkleRoot(transactions []models.Transaction) string {
	if len(transactions) == 0 {
		return ""
	}

	// 1. Hash each transaction (lead nodes)
	var hashes []string

	for _, tx := range transactions {
		txData := fmt.Sprintf("%d%s%s%.8f%s", tx.ID, tx.FromAddress, tx.ToAddress, tx.Amount, tx.Signature)

		hash := sha256.Sum256([]byte(txData))
		hashes = append(hashes, hex.EncodeToString(hash[:]))
	}

	// 2. Build merkle tree bottom-up
	for len(hashes) > 1 {
		var newLevel []string

		for i := 0; i < len(hashes); i += 2 {
			var combined string

			if i+1 < len(hashes) {
				combined = hashes[i] + hashes[i+1]
			} else {
				combined = hashes[i] + hashes[i] // duplicate last hash if odd
			}

			hash := sha256.Sum256([]byte(combined))
			newLevel = append(newLevel, hex.EncodeToString(hash[:]))
		}
		hashes = newLevel
	}
	return hashes[0]
}

func GetMerkleProof(transactions []models.Transaction, txIndex int) []string {
	if txIndex < 0 || txIndex >= len(transactions) {
		return nil
	}

	var proof []string
	var hashes []string

	// build initial leaf hashes
	for _, tx := range transactions {
		txData := fmt.Sprintf("%d%s%s%.8f%s", tx.ID, tx.FromAddress, tx.ToAddress, tx.Amount, tx.Signature)

		hash := sha256.Sum256([]byte(txData))
		hashes = append(hashes, hex.EncodeToString(hash[:]))
	}

	index := txIndex

	for len(hashes) > 1 {
		var newLevel []string

		for i := 0; i < len(hashes); i += 2 {
			var combined string

			if i == index || i+1 == index {
				if i == index && i+1 < len(hashes) {
					proof = append(proof, hashes[i+1])
				} else if i+1 == index {
					proof = append(proof, hashes[i])
				}
			}

			if i+1 < len(hashes) {
				combined = hashes[i] + hashes[i+1]
			} else {
				combined = hashes[i] + hashes[i] // duplicate last hash if odd
			}

			hash := sha256.Sum256([]byte(combined))
			newLevel = append(newLevel, hex.EncodeToString(hash[:]))
		}

		hashes = newLevel
		index = index / 2
	}
	return proof
}

func VerifyMerkleProof(txHash string, proof []string, merkleRoot string) bool {
	currentHash := txHash

	for _, siblingHash := range proof {
		combined := currentHash + siblingHash
		hash := sha256.Sum256([]byte(combined))
		currentHash = hex.EncodeToString(hash[:])
	}

	return currentHash == merkleRoot
}
