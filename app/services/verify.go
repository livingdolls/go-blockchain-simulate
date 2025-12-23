package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type TransactionType string

const (
	BuyTransaction  TransactionType = "BUY"
	SellTransaction TransactionType = "SELL"
)

type VerifyTxService interface {
	VerifyTransactionSignature(ctx context.Context, fromAddt, toAddr string, amount float64, nonce, signature string) error
	VerifyBuySellSignature(ctx context.Context, address string, amount float64, nonce, signature string, txType TransactionType) error
}

type verifyTxService struct {
	memory redis.MemoryAdapter
}

func NewVerifyTxService(memory redis.MemoryAdapter) VerifyTxService {
	return &verifyTxService{
		memory: memory,
	}
}

func (s *verifyTxService) VerifyTransactionSignature(ctx context.Context, fromAddr, toAddr string, amount float64, nonce, signature string) error {
	from := strings.ToLower(strings.TrimSpace(fromAddr))
	to := strings.ToLower(strings.TrimSpace(toAddr))

	log.Printf("=== Transaction Verification Start ===")
	log.Printf("From: %s", from)
	log.Printf("To: %s", to)
	log.Printf("Amount: %.2f", amount)
	log.Printf("Nonce: %s", nonce)

	// verify nonce
	key := "tx_nonce:" + from
	storedNonce, ok := s.memory.Get(ctx, key)

	if !ok || string(storedNonce) != nonce {
		return fmt.Errorf("invalid nonce")
	}

	// build cannonical message same on frontend
	msg := fmt.Sprintf("Send %.2f to %s nonce:%s", amount, to, nonce)

	log.Printf("=== Message Details ===")
	log.Printf("Message: %s", msg)
	log.Printf("Message length: %d", len(msg))
	log.Printf("Message bytes: %x", []byte(msg))

	// parse signature
	sigHex := strings.TrimPrefix(strings.TrimSpace(signature), "0x")
	raw, err := hex.DecodeString(sigHex)

	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}

	if len(raw) != 65 {
		return fmt.Errorf("invalid signature length: %d", len(raw))
	}

	// exrtract r,s,v
	sPart := new(big.Int).SetBytes(raw[32:64])
	v := raw[64]

	// normalize v to 27/28
	origV := v
	if v == 0 || v == 1 {
		v += 27
	} else if v >= 35 {
		v = byte(((int(v) - 35) % 2) + 27)
	}

	if v != 27 && v != 28 {
		return fmt.Errorf("invalid v value in signature: %d (original %d)", v, origV)
	}

	// low check eip-2 cannonical signature
	curveN := crypto.S256().Params().N
	halfN := new(big.Int).Rsh(curveN, 1)

	if sPart.Cmp(halfN) == 1 {
		return fmt.Errorf("signature s too high (non-canonical)")
	}

	// hash the prefixed message
	hash := utils.PrefixedHash([]byte(msg))

	// recover public key
	var pubBytes []byte
	for _, vTry := range []byte{v, utils.ToggleV(v)} {
		testSig := make([]byte, 65)
		copy(testSig[:64], raw[:64])
		testSig[64] = vTry - 27 // ecrecover expects 0/1

		recovered, err := crypto.Ecrecover(hash, testSig)

		if err == nil {
			pubBytes = recovered
			break
		}
	}

	if pubBytes == nil {
		return fmt.Errorf("failed to recover public key from signature")
	}

	// convert to public key
	pubKey, err := crypto.UnmarshalPubkey(pubBytes)

	if err != nil {
		return fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	recovered := strings.ToLower(string(crypto.PubkeyToAddress(*pubKey).Hex()))

	log.Printf("=== Recovery Result ===")
	log.Printf("Recovered address: %s", recovered)
	log.Printf("Expected address: %s", from)
	log.Printf("Match: %v", recovered == from)

	// verify recovered address matches from address
	if recovered != from {
		return fmt.Errorf("address mismatch: recovered=%s expected=%s", recovered, from)
	}

	//delete nonce on success
	s.memory.Del(ctx, key)

	return nil
}

func (s *verifyTxService) VerifyBuySellSignature(ctx context.Context, address string, amount float64, nonce, signature string, txType TransactionType) error {
	addr := strings.ToLower(strings.TrimSpace(address))

	log.Printf("=== Transaction Verification Start ===")
	log.Printf("From: %s", addr)
	log.Printf("Amount: %.2f", amount)
	log.Printf("Nonce: %s", nonce)

	// verify nonce
	key := "tx_nonce:" + addr
	storedNonce, ok := s.memory.Get(ctx, key)

	if !ok || string(storedNonce) != nonce {
		return fmt.Errorf("invalid nonce")
	}

	// build cannonical message same on frontend
	msg := fmt.Sprintf(" %s %.2f nonce:%s", txType, amount, nonce)

	log.Printf("=== Message Details ===")
	log.Printf("Message: %s", msg)
	log.Printf("Message length: %d", len(msg))
	log.Printf("Message bytes: %x", []byte(msg))

	// parse signature
	sigHex := strings.TrimPrefix(strings.TrimSpace(signature), "0x")
	raw, err := hex.DecodeString(sigHex)

	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}

	if len(raw) != 65 {
		return fmt.Errorf("invalid signature length: %d", len(raw))
	}

	// exrtract r,s,v
	sPart := new(big.Int).SetBytes(raw[32:64])
	v := raw[64]

	// normalize v to 27/28
	origV := v
	if v == 0 || v == 1 {
		v += 27
	} else if v >= 35 {
		v = byte(((int(v) - 35) % 2) + 27)
	}

	if v != 27 && v != 28 {
		return fmt.Errorf("invalid v value in signature: %d (original %d)", v, origV)
	}

	// low check eip-2 cannonical signature
	curveN := crypto.S256().Params().N
	halfN := new(big.Int).Rsh(curveN, 1)

	if sPart.Cmp(halfN) == 1 {
		return fmt.Errorf("signature s too high (non-canonical)")
	}

	// hash the prefixed message
	hash := utils.PrefixedHash([]byte(msg))

	// recover public key
	var pubBytes []byte
	for _, vTry := range []byte{v, utils.ToggleV(v)} {
		testSig := make([]byte, 65)
		copy(testSig[:64], raw[:64])
		testSig[64] = vTry - 27 // ecrecover expects 0/1

		recovered, err := crypto.Ecrecover(hash, testSig)

		if err == nil {
			pubBytes = recovered
			break
		}
	}

	if pubBytes == nil {
		return fmt.Errorf("failed to recover public key from signature")
	}

	// convert to public key
	pubKey, err := crypto.UnmarshalPubkey(pubBytes)

	if err != nil {
		return fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	recovered := strings.ToLower(string(crypto.PubkeyToAddress(*pubKey).Hex()))

	log.Printf("=== Recovery Result ===")
	log.Printf("Recovered address: %s", recovered)

	// verify recovered address matches from address
	if recovered != addr {
		return fmt.Errorf("address mismatch: recovered=%s expected=%s", recovered, addr)
	}

	//delete nonce on success
	s.memory.Del(ctx, key)

	return nil
}
