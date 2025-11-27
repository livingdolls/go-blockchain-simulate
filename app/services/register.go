package services

import (
	"context"
	"encoding/hex"
	"fmt"
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
	Verify(ctx context.Context, address, signature string) (string, error)
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

	return nonce, nil
}

func (r *registerService) Verify(ctx context.Context, address, signature string) (string, error) {
	addr := strings.ToLower(address)
	nonce, ok := r.redis.Get(ctx, addr)

	if !ok {
		return "", fmt.Errorf("missing or expired challenge")
	}

	// cannonical message (must match exactly what was signed on client side)
	msg := fmt.Sprintf("Login to YuteBlockchain\nnonce:%s", nonce)
	sig, err := r.parseSignature(signature)
	if err != nil {
		return "", fmt.Errorf("signature not valid %w", err)
	}

	hash := prefixedHash([]byte(msg))
	pub, err := crypto.SigToPub(hash, sig)

	if err != nil {
		return "", fmt.Errorf("failed to sig to pub %w", err)
	}

	recovered := crypto.PubkeyToAddress(*pub).Hex()

	if !strings.EqualFold(recovered, address) {
		return "", fmt.Errorf("address missmatch")
	}

	// delete nonce after verification attempt
	r.redis.Del(ctx, addr)

	// generate jwt token
	token, err := r.jwt.GenerateToken(addr)

	if err != nil {
		return "", fmt.Errorf("failed to generate jwt web token %w", err)
	}

	return token, nil
}

func (r *registerService) parseSignature(sigHex string) ([]byte, error) {
	s := strings.TrimPrefix(sigHex, "0x")
	b, err := hex.DecodeString(s)

	if err != nil {
		return nil, fmt.Errorf("failed to hex decode to string %w", err)
	}

	if len(b) != 65 {
		return nil, fmt.Errorf("invalid signature length: %d", len(b))
	}

	v := int(b[64])
	if v == 0 || v == 1 {
		v += 27
	} else if v >= 35 {
		v = (v-35)%2 + 27
	}

	if v != 27 && v != 28 {
		return nil, fmt.Errorf("invalid signature recovery id: %d", v)
	}

	b[64] = byte(v)

	return b, nil
}

func prefixedHash(data []byte) []byte {
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(data))
	return crypto.Keccak256([]byte(prefix), data)
}
