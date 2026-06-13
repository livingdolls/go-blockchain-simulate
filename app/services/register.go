package services

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/livingdolls/go-blockchain-simulate/security"
	"go.uber.org/zap"
	"github.com/livingdolls/go-blockchain-simulate/utils"
)

type RegisterService interface {
	Register(req models.UserRegister) (models.UserRegisterResponse, error)
	Challenge(ctx context.Context, address string) (string, error)
	Verify(ctx context.Context, address, nonce, signature, username string) (string, error)
}

type registerService struct {
	repo        repository.UserRepository
	walletRepo  repository.UserWalletRepository
	balanceRepo repository.UserBalanceRepository
	jwt         security.JWTService
	redis       redis.MemoryAdapter
}

func NewRegisterService(repo repository.UserRepository, walletRepo repository.UserWalletRepository, balanceRepo repository.UserBalanceRepository, jwt security.JWTService, redis redis.MemoryAdapter) RegisterService {
	return &registerService{repo: repo, walletRepo: walletRepo, balanceRepo: balanceRepo, jwt: jwt, redis: redis}
}

// Register implements RegisterService. Return:
//
//   - Success: UserRegisterResponse + nil error
//   - dto.AppError 409: address/username sudah terdaftar (duplicate)
//   - dto.AppError 500: DB/JWT infrastructure error
//   - dto.AppError 500 dengan entity.ErrDatabase: unexpected DB error
//
// Service TIDAK return 4xx generic string - semua error kategori 4xx
// dikirim sebagai *dto.AppError agar handler bisa map ke status code
// + error_code yang tepat (lihat dto/apperror.go).
func (r *registerService) Register(req models.UserRegister) (models.UserRegisterResponse, error) {
	// save to db
	user := models.User{
		Name:      req.Username,
		Address:   req.Address,
		PublicKey: req.PublicKey,
	}

	err := r.repo.Create(user)
	if err != nil {
		return models.UserRegisterResponse{}, mapRegisterRepoError(err, req)
	}

	// create empty wallet
	err = r.walletRepo.UpsertEmptyIfNotExists(user.Address)
	if err != nil {
		// Wallet/balance error = infra error (500), bukan user error.
		// Log cause untuk debug tapi jangan expose ke user.
		logger.LogError(fmt.Sprintf("failed to create wallet for %s", user.Address), err)
		return models.UserRegisterResponse{}, dto.NewInternalError(err)
	}

	// create empty balance
	err = r.balanceRepo.UpsertEmptyIfNotExists(user.Address)
	if err != nil {
		logger.LogError(fmt.Sprintf("failed to create balance for %s", user.Address), err)
		return models.UserRegisterResponse{}, dto.NewInternalError(err)
	}

	token, err := r.jwt.GenerateToken(user.Address)
	if err != nil {
		// JWT generation error = infra (signing key missing, etc).
		// Bukan user error, jadi 500.
		logger.LogError(fmt.Sprintf("failed to generate JWT for %s", user.Address), err)
		return models.UserRegisterResponse{}, dto.NewInternalError(err)
	}

	userResponse := models.UserRegisterResponse{
		Username:   req.Username,
		Address:    req.Address,
		YTEBalance: 0,
		USDBalance: 0,
		Token:      token,
	}

	return userResponse, nil
}

// mapRegisterRepoError menerjemahkan entity error dari repository ke
// AppError dengan HTTP status yang sesuai. Dipakai agar handler tidak
// perlu switch/case by error string (yang fragile).
func mapRegisterRepoError(err error, req models.UserRegister) *dto.AppError {
	switch {
	case errors.Is(err, entity.ErrAddressAlreadyRegistered):
		return dto.NewConflict(dto.CodeAddressExists,
			"address already registered", "address")
	case errors.Is(err, entity.ErrUsernameAlreadyExists):
		return dto.NewConflict(dto.CodeUsernameExists,
			"username already taken", "username")
	case errors.Is(err, entity.ErrConflict):
		// Generic conflict (duplicate field lain).
		return dto.NewConflict(dto.CodeConflict,
			"resource already exists", "")
	default:
		// Unexpected DB error. Log + return 500 generic (jangan bocor
		// SQL error ke user).
		logger.LogError(fmt.Sprintf("register: unexpected error for address=%s name=%s",
			req.Address, req.Username), err)
		return dto.NewDatabaseError(err)
	}
}

func (r *registerService) Challenge(contex context.Context, address string) (string, error) {
	addr := strings.ToLower(address)

	nonce := uuid.NewString()
	r.redis.Set(contex, addr, []byte(nonce), 10*time.Minute)
	logger.LogInfo(fmt.Sprintf("Challenge created : address=%s, nonce=%s", addr, nonce))

	return nonce, nil
}

func (r *registerService) Verify(ctx context.Context, address, nonce, signature, username string) (string, error) {
	addr := strings.ToLower(strings.TrimSpace(address))

	stored, ok := r.redis.Get(ctx, addr)
	storedStr := string(stored)
	logger.LogWarn("Verify: pre-check",
		zap.String("nonce_sent", nonce),
		zap.String("nonce_stored", storedStr),
		zap.Bool("match", nonce == storedStr),
		zap.Bool("has_stored", ok && len(stored) > 0),
	)
	if !ok || len(stored) == 0 {
		// 400: user belum request challenge atau sudah expire (10 min TTL).
		return "", dto.NewBadRequest(dto.CodeMissingChallenge,
			"missing or expired challenge - request a new nonce first")
	}
	if nonce != storedStr {
		// 400: nonce yang di-sign tidak match dengan yang di Redis.
		// Berarti user pakai nonce lama atau salah paste.
		return "", dto.NewBadRequest(dto.CodeStaleChallenge,
			"stale challenge: request a new nonce")
	}

	msg := fmt.Sprintf("Login to YuteBlockchain nonce:%s", nonce)

	// DEBUG: log nonce and hash untuk diagnose ADDRESS_MISMATCH
	msgHash := utils.PrefixedHash([]byte(msg))
	logger.LogWarn("Verify: hash diagnostic",
		zap.String("nonce", nonce),
		zap.String("addr_input", addr),
		zap.String("msg", msg),
		zap.String("msgHash", fmt.Sprintf("%x", msgHash)),
	)

	// parse signature
	sigHex := strings.TrimPrefix(strings.TrimSpace(signature), "0x")
	raw, err := hex.DecodeString(sigHex)
	if err != nil {
		return "", dto.NewBadRequestField(dto.CodeInvalidSignature,
			"invalid signature hex format", "signature")
	}
	if len(raw) != 65 {
		return "", dto.NewBadRequestField(dto.CodeInvalidSignature,
			fmt.Sprintf("invalid signature length: %d (expected 65)", len(raw)), "signature")
	}

	// r,s,v
	sPart := new(big.Int).SetBytes(raw[32:64])
	v, err := normalizeSignatureV(raw[64])
	if err != nil {
		return "", dto.NewBadRequestField(dto.CodeInvalidSignature,
			err.Error(), "signature")
	}
	raw[64] = v

	// low-s check (EIP-2)
	halfN := new(big.Int).Rsh(crypto.S256().Params().N, 1)
	if sPart.Cmp(halfN) == 1 {
		return "", dto.NewBadRequestField(dto.CodeInvalidSignature,
			"signature s too high (non-canonical)", "signature")
	}

	// hash
	hash := utils.PrefixedHash([]byte(msg))

	// recover with Ecrecover using the given parity; toggle only if primary fails
	var pubBytes []byte
	for _, vTry := range []byte{v, utils.ToggleV(v)} {
		test := make([]byte, 65)
		copy(test, raw[:64])
		test[64] = vTry - 27 // Ecrecover expects 0/1
		if rec, er := crypto.Ecrecover(hash, test); er == nil {
			pubBytes = rec
			break
		}
	}
	if pubBytes == nil {
		return "", dto.NewBadRequest(dto.CodeSignatureRecoveryFail,
			"failed to recover public key from signature")
	}

	pubKey, err := crypto.UnmarshalPubkey(pubBytes)
	if err != nil {
		// Wrap internal - unmarshalPK failure biasanya = bug/corrupt sig.
		logger.LogError("Verify: failed to unmarshal public key", err)
		return "", dto.NewInternalError(err)
	}
	recovered := strings.ToLower(crypto.PubkeyToAddress(*pubKey).Hex())
	if recovered != addr {
		// 400: signature valid tapi address yg di-sign != address di body.
		// Bisa jadi user salah sign dengan key lain. Log di level Warn
		// (bukan Info) agar visible di dev dengan LOG_LEVEL=warn. Address
		// dan recovered address keduanya public info (sudah ada di chain),
		// jadi tidak ada privacy issue.
		logger.LogWarn("Verify: address mismatch",
			zap.String("expected", addr),
			zap.String("recovered", recovered),
		)
		return "", dto.NewBadRequestField(dto.CodeAddressMismatch,
			"signature does not match the provided address", "address")
	}

	// check username uniqueness
	existingUser, err := r.repo.GetByAddress(addr)
	if err != nil {
		// 500: DB error (bukan user not found - kita handle di bawah).
		logger.LogError(fmt.Sprintf("Verify: failed to get user by address %s", addr), err)
		return "", dto.NewDatabaseError(err)
	}
	if existingUser.ID == 0 {
		// 404: address ada di challenge Redis tapi belum register.
		return "", dto.NewNotFound(dto.CodeUserNotFound,
			"user not registered - call /auth/register first")
	}
	if existingUser.Name != username {
		// 400: username di body != username saat register. Attacker
		// mungkin coba ambil alih akun dengan address yg sudah register.
		return "", dto.NewBadRequestField(dto.CodeUsernameMismatch,
			"username does not match the registered username", "username")
	}

	if addr != strings.ToLower(existingUser.Address) {
		// Seharusnya tidak mungkin (address di DB selalu lowercased).
		// Tapi guard untuk paranoid case.
		return "", dto.NewBadRequestField(dto.CodeAddressMismatch,
			"address does not match the registered address", "address")
	}

	// success: delete nonce
	r.redis.Del(ctx, addr)

	token, err := r.jwt.GenerateToken(addr)
	if err != nil {
		logger.LogError(fmt.Sprintf("Verify: failed to generate JWT for %s", addr), err)
		return "", dto.NewInternalError(err)
	}
	return token, nil
}

// normalizeSignatureV menormalkan signature recovery id v ke bentuk
// EIP-191 (27-30) atau EIP-155 (35+). Return error kalau v di luar
// range yang dikenal.
//
// Format yang diterima:
//   - v ∈ {0, 1}        → go-ethereum native recovery id
//   - v ∈ {27, 28, 29, 30}  → EIP-191 personal_sign
//   - v >= 35           → EIP-155 (chain ID encoded)
//
// v=29/30 (recId 2/3) artinya r (signature's r value) >= N, sehingga
// recovery perlu pakai X = r + N. Normal untuk ~50% signature yang
// di-produce oleh BouncyCastle / library lain yang tidak menormalisasi r.
// go-ethereum's crypto.Sign selalu normalize r ke < N, jadi v dari
// go-ethereum Selalu 27/28. Test ini memastikan kompatibilitas dengan
// library lain.
func normalizeSignatureV(v byte) (byte, error) {
	switch {
	case v == 0 || v == 1:
		return v + 27, nil
	case v >= 27 && v <= 30:
		return v, nil
	case v >= 35:
		return byte(((int(v) - 35) % 2) + 27), nil
	default:
		return 0, fmt.Errorf("invalid signature recovery id: %d", v)
	}
}
