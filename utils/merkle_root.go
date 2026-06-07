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
		txData := fmt.Sprintf("%d%s%s%.8f%.8f%s", tx.ID, tx.FromAddress, tx.ToAddress, tx.Amount, tx.Fee, tx.Signature)

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

	// build initial leaf hashes (sama format dengan CalculateMerkleRoot:20)
	for _, tx := range transactions {
		txData := fmt.Sprintf("%d%s%s%.8f%.8f%s", tx.ID, tx.FromAddress, tx.ToAddress, tx.Amount, tx.Fee, tx.Signature)

		hash := sha256.Sum256([]byte(txData))
		hashes = append(hashes, hex.EncodeToString(hash[:]))
	}

	index := txIndex

	for len(hashes) > 1 {
		var newLevel []string

		for i := 0; i < len(hashes); i += 2 {
			var combined string

			if i == index || i+1 == index {
				if i == index {
					// tx ada di kiri. Sibling di i+1, atau diri sendiri jika i+1 OOB (odd count).
					if i+1 < len(hashes) {
						proof = append(proof, hashes[i+1])
					} else {
						proof = append(proof, hashes[i])
					}
				} else if i+1 == index {
					// tx ada di kanan, sibling di i.
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

func VerifyMerkleProof(txHash string, proof []string, merkleRoot string, txIndex int) bool {
	// Tentukan posisi sibling di setiap level dari txIndex:
	// - bit terakhir index: 0 = left, 1 = right
	// - sibling index = index ^ 1 (XOR flip last bit)
	// - sibling di kiri (siblingIdx < index): combined = sibling + current
	// - sibling di kanan (siblingIdx > index): combined = current + sibling
	currentHash := txHash
	index := txIndex

	for _, siblingHash := range proof {
		var combined string
		if (index & 1) == 0 {
			// tx ada di kiri, sibling di kanan
			combined = currentHash + siblingHash
		} else {
			// tx ada di kanan, sibling di kiri
			combined = siblingHash + currentHash
		}
		hash := sha256.Sum256([]byte(combined))
		currentHash = hex.EncodeToString(hash[:])
		index = index / 2
	}

	return currentHash == merkleRoot
}
