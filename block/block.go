package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/livingdolls/go-blockchain-simulate/transaction"
)

type Block struct {
	Index        int
	Timestamp    int64
	Transactions []transaction.Transaction
	PrevHash     string
	Hash         string
}

func HashBlock(b Block) string {
	data, _ := json.Marshal(b)

	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}
