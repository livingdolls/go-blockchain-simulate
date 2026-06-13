package services

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeUserRepo minimal untuk Verify: hanya GetByAddress.
type fakeUserRepo struct {
	users       map[string]models.User
	createErr   error // inject error untuk test duplicate path
	createdUser models.User
}

func (f *fakeUserRepo) Create(u models.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.createdUser = u
	return nil
}
func (f *fakeUserRepo) GetByAddress(address string) (models.User, error) {
	if u, ok := f.users[strings.ToLower(address)]; ok {
		return u, nil
	}
	return models.User{}, nil
}
func (f *fakeUserRepo) GetByAddressWithBalance(string) (models.UserWithBalance, error) {
	return models.UserWithBalance{}, nil
}
func (f *fakeUserRepo) BeginTx() (*sqlx.Tx, error) { return nil, nil }
func (f *fakeUserRepo) GetMultipleByAddress([]string) ([]models.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) GetMultipleByAddressWithTx(*sqlx.Tx, []string) ([]models.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) GetUserWithWallet(string) (models.User, models.UserWallet, error) {
	return models.User{}, models.UserWallet{}, nil
}

var _ repository.UserRepository = (*fakeUserRepo)(nil)// stubWalletRepo adalah stub UserWalletRepository.
// Method bulk/insert history tidak relevan untuk test Verify, return zero.
type stubWalletRepo struct{}

func (s *stubWalletRepo) UpsertEmptyIfNotExists(string) error { return nil }
func (s *stubWalletRepo) UpsertEmptyIfNotExistsWithTx(*sqlx.Tx, string) error {
	return nil
}
func (s *stubWalletRepo) GetForUpdateWithTx(*sqlx.Tx, string) (models.UserWallet, error) {
	return models.UserWallet{}, nil
}
func (s *stubWalletRepo) GetMultipleByAddress([]string) ([]models.UserWallet, error) {
	return nil, nil
}
func (s *stubWalletRepo) GetMultipleByAddressWithTx(*sqlx.Tx, []string) ([]models.UserWallet, error) {
	return nil, nil
}
func (s *stubWalletRepo) InsertHistoryWithTx(*sqlx.Tx, models.WalletHistory) error {
	return nil
}
func (s *stubWalletRepo) LockMultipleWalletsWithTx(*sqlx.Tx, []string) error { return nil }
func (s *stubWalletRepo) UpdateWalletWithTx(*sqlx.Tx, string, float64) error { return nil }
func (s *stubWalletRepo) GetByAddress(string) (models.UserWallet, error) {
	return models.UserWallet{}, nil
}
func (s *stubWalletRepo) GetTopBalances(int) ([]models.UserWallet, error) {
	return nil, nil
}
func (s *stubWalletRepo) BulkUpdateBalancesWithTx(*sqlx.Tx, map[string]float64) error {
	return nil
}
func (s *stubWalletRepo) BeginTx() (*sqlx.Tx, error) { return nil, nil }

var _ repository.UserWalletRepository = (*stubWalletRepo)(nil)

// stubBalanceRepo adalah stub UserBalanceRepository.
type stubBalanceRepo struct{}

func (s *stubBalanceRepo) UpsertEmptyIfNotExists(string) error { return nil }
func (s *stubBalanceRepo) UpsertEmptyIfNotExistsWithTx(*sqlx.Tx, string) error {
	return nil
}
func (s *stubBalanceRepo) GetForUpdateWithTx(*sqlx.Tx, string) (models.UserBalance, error) {
	return models.UserBalance{}, nil
}
func (s *stubBalanceRepo) UpdateBalanceWithTx(*sqlx.Tx, string, float64, float64) error {
	return nil
}
func (s *stubBalanceRepo) InsertHistoryWithTx(*sqlx.Tx, models.BalanceHistory) error {
	return nil
}
func (s *stubBalanceRepo) GetByAddress(string) (models.UserBalance, error) {
	return models.UserBalance{}, nil
}
func (s *stubBalanceRepo) GetMultipleByAddressWithTxForUpdate(*sqlx.Tx, []string) ([]models.UserBalance, error) {
	return nil, nil
}
func (s *stubBalanceRepo) BulkUpdateBalancesWithTx(*sqlx.Tx, map[string]models.UserBalance) error {
	return nil
}
func (s *stubBalanceRepo) BeginTx() (*sqlx.Tx, error) { return nil, nil }

var _ repository.UserBalanceRepository = (*stubBalanceRepo)(nil)

// fakeJWT menghasilkan token dummy (tidak ditandatangani secara kriptografis).
type fakeJWT struct{}

func (f *fakeJWT) GenerateToken(address string) (string, error) {
	return "token-" + address, nil
}
func (f *fakeJWT) ValidateToken(string) (*security.JWTClaims, error) { return nil, nil }

func newRegisterTestService(users map[string]models.User) (RegisterService, *mockMemoryAdapter) {
	svc, _, mem := newRegisterTestServiceWithRepo(users)
	return svc, mem
}

// newRegisterTestServiceWithRepo returns svc + repo + memory adapter.
// Dipakai oleh test Register() yang perlu inject createErr di repo.
func newRegisterTestServiceWithRepo(users map[string]models.User) (RegisterService, *fakeUserRepo, *mockMemoryAdapter) {
	mem := newMockMemoryAdapter()
	repo := &fakeUserRepo{users: users}
	svc := NewRegisterService(repo, &stubWalletRepo{}, &stubBalanceRepo{}, &fakeJWT{}, mem)
	return svc, repo, mem
}

func TestRegister_Challenge_StoresNonce(t *testing.T) {
	svc, mem := newRegisterTestService(nil)
	nonce, err := svc.Challenge(context.Background(), "0xABCDEF")
	require.NoError(t, err)
	assert.NotEmpty(t, nonce)
	assert.Len(t, nonce, 36, "UUID v4 string 36 char")

	// Disimpan dengan key lowercase
	val, ok := mem.Get(context.Background(), "0xabcdef")
	assert.True(t, ok)
	assert.Equal(t, nonce, string(val))
}

func TestRegister_Challenge_UniqueNonces(t *testing.T) {
	svc, _ := newRegisterTestService(nil)
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		n, _ := svc.Challenge(context.Background(), "0xabc")
		assert.False(t, seen[n])
		seen[n] = true
	}
}

func TestRegister_Verify_HappyPath(t *testing.T) {
	privKey, _ := crypto.GenerateKey()
	addr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())
	privKeyHex := hex.EncodeToString(crypto.FromECDSA(privKey))

	svc, mem := newRegisterTestService(map[string]models.User{
		addr: {ID: 1, Name: "alice", Address: addr},
	})
	nonce, _ := svc.Challenge(context.Background(), addr)

	msg := []byte("Login to YuteBlockchain nonce:" + nonce)
	signature := signMessage(t, privKeyHex, msg)

	token, err := svc.Verify(context.Background(), addr, nonce, signature, "alice")
	require.NoError(t, err)
	assert.Equal(t, "token-"+addr, token)

	// Nonce harus dihapus setelah berhasil
	_, ok := mem.Get(context.Background(), addr)
	assert.False(t, ok, "nonce harus dihapus setelah login sukses")
}

func TestRegister_Verify_MissingChallenge(t *testing.T) {
	svc, _ := newRegisterTestService(nil)
	_, err := svc.Verify(context.Background(), "0xabc", "any-nonce", "0x00", "alice")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing or expired challenge")
}

func TestRegister_Verify_StaleNonce(t *testing.T) {
	svc, _ := newRegisterTestService(nil)
	addr := "0xabc"
	svc.Challenge(context.Background(), addr)
	// Pakai nonce yang berbeda dari yang disimpan
	_, err := svc.Verify(context.Background(), addr, "WRONG-NONCE", "0x00", "alice")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stale challenge")
}

func TestRegister_Verify_InvalidHex(t *testing.T) {
	svc, _ := newRegisterTestService(nil)
	addr := "0xabc"
	nonce, _ := svc.Challenge(context.Background(), addr)

	_, err := svc.Verify(context.Background(), addr, nonce, "0xNOTHEX", "alice")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature hex")
}

func TestRegister_Verify_AddressMismatch(t *testing.T) {
	// Sign dengan key A, klaim address B
	privKey, _ := crypto.GenerateKey()
	privKeyHex := hex.EncodeToString(crypto.FromECDSA(privKey))
	bAddr := "0x" + strings.Repeat("c", 40)

	svc, _ := newRegisterTestService(map[string]models.User{
		strings.ToLower(bAddr): {ID: 1, Name: "bob", Address: bAddr},
	})
	nonce, _ := svc.Challenge(context.Background(), bAddr)

	msg := []byte("Login to YuteBlockchain nonce:" + nonce)
	signature := signMessage(t, privKeyHex, msg)

	_, err := svc.Verify(context.Background(), bAddr, nonce, signature, "bob")
	assert.Error(t, err)
	// Cek error code (typed) - lebih robust dari string match.
	appErr, ok := dto.AsAppError(err)
	require.True(t, ok, "expected *dto.AppError")
	assert.Equal(t, dto.CodeAddressMismatch, appErr.Code)
	assert.Equal(t, 400, appErr.Status)
	assert.Equal(t, "address", appErr.Field)
}

func TestNormalizeSignatureV(t *testing.T) {
	// Unit test untuk normalizeSignatureV. Dipisah dari full verify flow
	// agar bisa test semua format v (termasuk 29/30 yang go-ethereum's
	// crypto.Sign tidak bisa produce - test ini validasi compatibility
	// dengan BouncyCastle/library lain).
	tests := []struct {
		name    string
		input   byte
		want    byte
		wantErr bool
	}{
		// go-ethereum native
		{"v=0 → 27", 0, 27, false},
		{"v=1 → 28", 1, 28, false},

		// EIP-191 (recId 0/1, go-ethereum's Sign output)
		{"v=27 → 27", 27, 27, false},
		{"v=28 → 28", 28, 28, false},

		// EIP-191 (recId 2/3, BouncyCastle output) - 50% dari real signatures
		{"v=29 → 29", 29, 29, false},
		{"v=30 → 30", 30, 30, false},

		// EIP-155 (chain ID encoded)
		{"v=35 → 27 (chain ID 1, recId 0)", 35, 27, false},
		{"v=36 → 28 (chain ID 1, recId 1)", 36, 28, false},
		{"v=37 → 27 (chain ID 2, recId 0)", 37, 27, false},
		{"v=38 → 28 (chain ID 2, recId 1)", 38, 28, false},

		// Invalid values (gaps antara 0/1, 27-30, dan 35+)
		{"v=2", 2, 0, true},
		{"v=5", 5, 0, true},
		{"v=26 (just below range)", 26, 0, true},
		{"v=31 (just above range)", 31, 0, true},
		{"v=34 (just below EIP-155)", 34, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeSignatureV(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "input %d harus error", tt.input)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got, "input %d → want %d, got %d", tt.input, tt.want, got)
			}
		})
	}
}

func TestRegister_Verify_RejectsInvalidV(t *testing.T) {
	// v di luar range EIP-191 (27-30) dan EIP-155 (35+) harus di-reject.
	// Kita fabricate signature dengan v=5 - tidak valid untuk format apapun.
	// r/s disalin dari signature valid (tidak penting untuk test ini
	// karena validation v terjadi SEBELUM recovery).
	privKey, _ := crypto.GenerateKey()
	addr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())
	privKeyHex := hex.EncodeToString(crypto.FromECDSA(privKey))

	svc, _ := newRegisterTestService(map[string]models.User{
		addr: {ID: 1, Name: "alice", Address: addr},
	})
	nonce, _ := svc.Challenge(context.Background(), addr)
	msg := []byte("Login to YuteBlockchain nonce:" + nonce)

	// Ambil r,s dari signature valid, lalu set v=5 (invalid)
	validSig, _ := hex.DecodeString(strings.TrimPrefix(signMessage(t, privKeyHex, msg), "0x"))
	validSig[64] = 5 // v invalid: bukan 0/1, 27-30, atau 35+
	sigInvalid := "0x" + hex.EncodeToString(validSig)

	_, err := svc.Verify(context.Background(), addr, nonce, sigInvalid, "alice")
	require.Error(t, err)
	appErr, ok := dto.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, dto.CodeInvalidSignature, appErr.Code)
}

func TestRegister_Verify_UserNotFound(t *testing.T) {
	// Verifikasi signature sukses tapi user belum registered.
	privKey, _ := crypto.GenerateKey()
	addr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())
	privKeyHex := hex.EncodeToString(crypto.FromECDSA(privKey))

	// Tidak ada user di repo
	svc, _ := newRegisterTestService(map[string]models.User{})
	nonce, _ := svc.Challenge(context.Background(), addr)
	msg := []byte("Login to YuteBlockchain nonce:" + nonce)
	signature := signMessage(t, privKeyHex, msg)

	_, err := svc.Verify(context.Background(), addr, nonce, signature, "alice")
	assert.Error(t, err)
	appErr, ok := dto.AsAppError(err)
	require.True(t, ok, "expected *dto.AppError")
	assert.Equal(t, dto.CodeUserNotFound, appErr.Code)
	assert.Equal(t, 404, appErr.Status)
}

func TestRegister_Verify_UsernameMismatch(t *testing.T) {
	privKey, _ := crypto.GenerateKey()
	addr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())
	privKeyHex := hex.EncodeToString(crypto.FromECDSA(privKey))

	svc, _ := newRegisterTestService(map[string]models.User{
		addr: {ID: 1, Name: "alice", Address: addr},
	})
	nonce, _ := svc.Challenge(context.Background(), addr)
	msg := []byte("Login to YuteBlockchain nonce:" + nonce)
	signature := signMessage(t, privKeyHex, msg)

	// Verify dengan username yang tidak match
	_, err := svc.Verify(context.Background(), addr, nonce, signature, "wronguser")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username does not match")
}

func TestRegister_Verify_AddressNormalization(t *testing.T) {
	// Challenge dengan address lowercase; query dengan uppercase harus match
	privKey, _ := crypto.GenerateKey()
	addrLower := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())
	privKeyHex := hex.EncodeToString(crypto.FromECDSA(privKey))

	svc, _ := newRegisterTestService(map[string]models.User{
		addrLower: {ID: 1, Name: "alice", Address: addrLower},
	})
	nonce, _ := svc.Challenge(context.Background(), addrLower)
	msg := []byte("Login to YuteBlockchain nonce:" + nonce)
	signature := signMessage(t, privKeyHex, msg)

	// Panggil dengan uppercase
	addrUpper := strings.ToUpper(addrLower)
	_, err := svc.Verify(context.Background(), addrUpper, nonce, signature, "alice")
	assert.NoError(t, err, "address harus dinormalisasi ke lowercase")
}

// Hapus unused import warning untuk sql.NullInt64 (dipakai repository)
var _ = context.Background

// --- Test Register() error mapping ---

func TestRegister_Register_DuplicateAddress_Returns409(t *testing.T) {
	// Setup: fake repo mensimulasikan MySQL error 1062 duplicate address.
	// Service harus map ke AppError 409 dengan code ADDRESS_ALREADY_REGISTERED,
	// BUKAN return raw error 500.
	svc, repo, _ := newRegisterTestServiceWithRepo(map[string]models.User{})
	repo.createErr = entity.ErrAddressAlreadyRegistered

	_, err := svc.Register(models.UserRegister{
		Username:  "alice",
		Address:   "0x1234567890abcdef1234567890abcdef12345678",
		PublicKey: "abcdef",
	})

	require.Error(t, err)
	appErr, ok := dto.AsAppError(err)
	require.True(t, ok, "expected *dto.AppError, got %T", err)

	assert.Equal(t, 409, appErr.Status, "duplicate should be 409 Conflict, not 500")
	assert.Equal(t, dto.CodeAddressExists, appErr.Code)
	assert.Equal(t, "address", appErr.Field, "field name helps frontend highlight the offending field")
	assert.Equal(t, "address already registered", appErr.Message)
	assert.NotContains(t, appErr.Message, "SQL", "must not leak SQL details")
}

func TestRegister_Register_DuplicateUsername_Returns409(t *testing.T) {
	svc, repo, _ := newRegisterTestServiceWithRepo(map[string]models.User{})
	repo.createErr = entity.ErrUsernameAlreadyExists

	_, err := svc.Register(models.UserRegister{
		Username:  "alice",
		Address:   "0x1234567890abcdef1234567890abcdef12345678",
		PublicKey: "abcdef",
	})

	require.Error(t, err)
	appErr, ok := dto.AsAppError(err)
	require.True(t, ok)

	assert.Equal(t, 409, appErr.Status)
	assert.Equal(t, dto.CodeUsernameExists, appErr.Code)
	assert.Equal(t, "username", appErr.Field)
}

func TestRegister_Register_UnknownDBError_Returns500(t *testing.T) {
	// Unknown error (e.g., connection lost) harus jadi 500 generic.
	// Message TIDAK boleh bocor detail ke user, tapi cause harus di-log.
	svc, repo, _ := newRegisterTestServiceWithRepo(map[string]models.User{})
	repo.createErr = errors.New("connection refused to internal-db:3306")

	_, err := svc.Register(models.UserRegister{
		Username:  "alice",
		Address:   "0x1234567890abcdef1234567890abcdef12345678",
		PublicKey: "abcdef",
	})

	require.Error(t, err)
	appErr, ok := dto.AsAppError(err)
	require.True(t, ok)

	assert.Equal(t, 500, appErr.Status)
	assert.Equal(t, dto.CodeDatabaseError, appErr.Code)
	assert.Equal(t, "database error", appErr.Message, "must be generic, no leak")
	assert.NotContains(t, appErr.Message, "internal-db", "must not leak internal hostnames")
	assert.NotContains(t, appErr.Message, "3306", "must not leak internal ports")
	// Cause tetap ada untuk logging
	assert.NotNil(t, appErr.Cause)
}


