package utils

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

func PrefixedHash(data []byte) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(data))
	return crypto.Keccak256([]byte(prefix), data)
}

func ToggleV(v byte) byte {
	if v == 27 {
		return 28
	}
	return 27
}
