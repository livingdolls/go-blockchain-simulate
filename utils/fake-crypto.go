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
