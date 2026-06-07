package services

import (
	"context"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jmoiron/sqlx"
	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/livingdolls/go-blockchain-simulate/app/repository"
	"github.com/livingdolls/go-blockchain-simulate/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeUserRepo minimal untuk Verify: hanya GetByAddress.
type fakeUserRepo struct {
	users map[string]models.User
}

func (f *fakeUserRepo) Create(models.User) error { return nil }
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
	mem := newMockMemoryAdapter()
	repo := &fakeUserRepo{users: users}
	svc := NewRegisterService(repo, &stubWalletRepo{}, &stubBalanceRepo{}, &fakeJWT{}, mem)
	return svc, mem
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
	assert.Contains(t, err.Error(), "address mismatch")
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
	assert.Contains(t, err.Error(), "user not found")
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


