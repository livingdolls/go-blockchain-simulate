package signature

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func CreateMessage(from, to string, amount int64) string {
	raw := fmt.Sprintf("%s|%s|%d", from, to, amount)

	h := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(h[:])
}

func Sign(privateKey, message string) string {
	raw := privateKey + message

	h := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(h[:])
}

func VerifySignature(privateKey, message, signature string) bool {
	raw := privateKey + message

	h := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(h[:]) == signature
}
