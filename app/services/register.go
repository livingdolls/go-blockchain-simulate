package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

type RegisterService interface {
	Register(req models.UserRegister) (models.UserRegisterResponse, error)
	Challenge(ctx context.Context, address string) (string, error)
	Verify(ctx context.Context, address, nonce, signature string) (string, error)
}

type registerService struct {
	repo  repository.UserRepository
	jwt   security.JWTService
	redis redis.MemoryAdapter
}

func NewRegisterService(repo repository.UserRepository, jwt security.JWTService, redis redis.MemoryAdapter) RegisterService {
	return &registerService{repo: repo, jwt: jwt, redis: redis}
}

// Registr implements RegisterService.
func (r *registerService) Register(req models.UserRegister) (models.UserRegisterResponse, error) {
	// save to db
	user := models.User{
		Name:      req.Username,
		Address:   req.Address,
		PublicKey: req.PublicKey,
		Balance:   1000,
	}

	err := r.repo.Create(user)
	if err != nil {
		return models.UserRegisterResponse{}, err
	}

	token, err := r.jwt.GenerateToken(user.Address)
	if err != nil {
		return models.UserRegisterResponse{}, err
	}

	userResponse := models.UserRegisterResponse{
		Username: req.Username,
		Address:  req.Address,
		Balance:  user.Balance,
		Token:    token,
	}

	return userResponse, nil
}

func (r *registerService) Challenge(contex context.Context, address string) (string, error) {
	addr := strings.ToLower(address)

	nonce := uuid.NewString()
	r.redis.Set(contex, addr, []byte(nonce), 10*time.Minute)
	log.Printf("Challenge created : address=%s, nonce=%s", addr, nonce)

	return nonce, nil
}

func (r *registerService) Verify(ctx context.Context, address, nonce, signature string) (string, error) {
	addr := strings.ToLower(strings.TrimSpace(address))

	stored, ok := r.redis.Get(ctx, addr)
	if !ok || len(stored) == 0 {
		return "", fmt.Errorf("missing or expired challenge")
	}
	if nonce != string(stored) {
		return "", fmt.Errorf("stale challenge: request a new nonce")
	}

	msg := fmt.Sprintf("Login to YuteBlockchain nonce:%s", nonce)

	// parse signature
	sigHex := strings.TrimPrefix(strings.TrimSpace(signature), "0x")
	raw, err := hex.DecodeString(sigHex)
	if err != nil {
		return "", fmt.Errorf("invalid signature hex: %w", err)
	}
	if len(raw) != 65 {
		return "", fmt.Errorf("invalid signature length: %d", len(raw))
	}

	// r,s,v
	sPart := new(big.Int).SetBytes(raw[32:64])
	v := raw[64]

	// normalize v to 27/28
	if v == 0 || v == 1 {
		v += 27
	} else if v >= 35 {
		v = byte(((int(v) - 35) % 2) + 27)
	}
	if v != 27 && v != 28 {
		return "", fmt.Errorf("invalid signature recovery id: %d", v)
	}
	raw[64] = v

	// low-s check (EIP-2)
	halfN := new(big.Int).Rsh(crypto.S256().Params().N, 1)
	if sPart.Cmp(halfN) == 1 {
		return "", fmt.Errorf("signature s too high (non-canonical)")
	}

	// hash
	hash := prefixedHash([]byte(msg))

	// recover with Ecrecover using the given parity; toggle only if primary fails
	var pubBytes []byte
	for _, vTry := range []byte{v, toggleV(v)} {
		test := make([]byte, 65)
		copy(test, raw[:64])
		test[64] = vTry - 27 // Ecrecover expects 0/1
		if rec, er := crypto.Ecrecover(hash, test); er == nil {
			pubBytes = rec
			break
		}
	}
	if pubBytes == nil {
		return "", fmt.Errorf("failed to recover public key")
	}

	pubKey, err := crypto.UnmarshalPubkey(pubBytes)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal public key: %w", err)
	}
	recovered := strings.ToLower(crypto.PubkeyToAddress(*pubKey).Hex())
	if recovered != addr {
		return "", fmt.Errorf("address mismatch recovered=%s expected=%s", recovered, addr)
	}

	// success: delete nonce
	r.redis.Del(ctx, addr)

	token, err := r.jwt.GenerateToken(addr)
	if err != nil {
		return "", fmt.Errorf("failed to generate jwt: %w", err)
	}
	return token, nil
}

func toggleV(v byte) byte {
	if v == 27 {
		return 28
	}
	return 27
}

func prefixedHash(data []byte) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(data))
	return crypto.Keccak256([]byte(prefix), data)
}
