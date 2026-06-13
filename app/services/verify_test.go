package services

import (
	"context"
	"encoding/hex"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/livingdolls/go-blockchain-simulate/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMemoryAdapter adalah implementasi in-memory dari MemoryAdapter.
// TTL diabaikan; setiap nilai akan expire setelah 100ms (test pakai TTL besar).
type mockMemoryAdapter struct {
	mu   sync.Mutex
	data map[string][]byte
}

func newMockMemoryAdapter() *mockMemoryAdapter {
	return &mockMemoryAdapter{data: make(map[string][]byte)}
}

func (m *mockMemoryAdapter) Get(_ context.Context, key string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	return v, ok
}

func (m *mockMemoryAdapter) Set(_ context.Context, key string, value []byte, _ time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *mockMemoryAdapter) Del(_ context.Context, key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

func (m *mockMemoryAdapter) InvalidatePattern(_ context.Context, _ string) error {
	return nil
}

func (m *mockMemoryAdapter) Publish(_ context.Context, _ string, _ []byte) error {
	return nil
}

func (m *mockMemoryAdapter) Subscribe(_ context.Context, _ string, _ func([]byte) error) error {
	return nil
}

var _ redis.MemoryAdapter = (*mockMemoryAdapter)(nil)

// signMessage menandatangani message dengan private key dan return
// signature 65-byte (r || s || v) dalam format hex dengan prefix 0x.
func signMessage(t *testing.T, privKeyHex string, message []byte) string {
	t.Helper()
	privKey, err := crypto.HexToECDSA(privKeyHex)
	require.NoError(t, err)

	hash := utils.PrefixedHash(message)
	sig, err := crypto.Sign(hash, privKey)
	require.NoError(t, err)
	// crypto.Sign menghasilkan v=0/1 (recovery id). Untuk EIP-191 format
	// yang dipakai service: v = recId + 27 = 27 atau 28.
	sig[64] += 27
	return "0x" + hex.EncodeToString(sig)
}

func TestVerifyTransactionSignature_HappyPath(t *testing.T) {
	// Setup: generate keypair, store nonce di mock, sign canonical message.
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	fromAddr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())
	privKeyHex := hex.EncodeToString(crypto.FromECDSA(privKey))

	mem := newMockMemoryAdapter()
	nonce := "nonce-123"
	mem.Set(context.Background(), "tx_nonce:"+fromAddr, []byte(nonce), 0)

	toAddr := "0x" + strings.Repeat("b", 40)
	amount := 100.50
	msg := []byte("Send 100.50 to " + toAddr + " nonce:" + nonce)
	signature := signMessage(t, privKeyHex, msg)

	svc := NewVerifyTxService(mem)
	err = svc.VerifyTransactionSignature(context.Background(), fromAddr, toAddr, amount, nonce, signature)
	assert.NoError(t, err, "signature valid harus sukses")

	// Nonce harus dihapus setelah berhasil
	_, ok := mem.Get(context.Background(), "tx_nonce:"+fromAddr)
	assert.False(t, ok, "nonce harus dihapus setelah verifikasi sukses")
}

func TestVerifyTransactionSignature_InvalidNonce(t *testing.T) {
	privKey, _ := crypto.GenerateKey()
	fromAddr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())

	mem := newMockMemoryAdapter()
	mem.Set(context.Background(), "tx_nonce:"+fromAddr, []byte("real-nonce"), 0)

	svc := NewVerifyTxService(mem)
	err := svc.VerifyTransactionSignature(context.Background(), fromAddr, "0xabc", 10, "WRONG-NONCE", "0x00")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid nonce")
}

func TestVerifyTransactionSignature_NoNonceStored(t *testing.T) {
	privKey, _ := crypto.GenerateKey()
	fromAddr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())

	svc := NewVerifyTxService(mem_empty())
	err := svc.VerifyTransactionSignature(context.Background(), fromAddr, "0xabc", 10, "any", "0x00")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid nonce")
}

func TestVerifyTransactionSignature_InvalidHex(t *testing.T) {
	privKey, _ := crypto.GenerateKey()
	fromAddr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())

	mem := newMockMemoryAdapter()
	mem.Set(context.Background(), "tx_nonce:"+fromAddr, []byte("n"), 0)

	svc := NewVerifyTxService(mem)
	err := svc.VerifyTransactionSignature(context.Background(), fromAddr, "0xabc", 10, "n", "0xNOTHEX")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature hex")
}

func TestVerifyTransactionSignature_WrongLength(t *testing.T) {
	privKey, _ := crypto.GenerateKey()
	fromAddr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())

	mem := newMockMemoryAdapter()
	mem.Set(context.Background(), "tx_nonce:"+fromAddr, []byte("n"), 0)

	svc := NewVerifyTxService(mem)
	// 64 byte signature (bukan 65)
	short := "0x" + strings.Repeat("ab", 32)
	err := svc.VerifyTransactionSignature(context.Background(), fromAddr, "0xabc", 10, "n", short)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature length")
}

func TestVerifyTransactionSignature_AddressMismatch(t *testing.T) {
	// Sign dengan key A, klaim dari address B -> mismatch.
	privKey, _ := crypto.GenerateKey()
	privKeyHex := hex.EncodeToString(crypto.FromECDSA(privKey))
	recoveredAddr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())

	// Bikin address palsu (acak)
	bAddr := strings.ToLower(common.HexToAddress("0x" + strings.Repeat("c", 40)).Hex())

	mem := newMockMemoryAdapter()
	nonce := "n"
	mem.Set(context.Background(), "tx_nonce:"+bAddr, []byte(nonce), 0)

	toAddr := "0x" + strings.Repeat("d", 40)
	amount := 50.0
	// Sign dengan key A, tapi claim from=B
	msg := []byte("Send 50.00 to " + toAddr + " nonce:" + nonce)
	signature := signMessage(t, privKeyHex, msg)

	svc := NewVerifyTxService(mem)
	err := svc.VerifyTransactionSignature(context.Background(), bAddr, toAddr, amount, nonce, signature)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "address mismatch")
	assert.Contains(t, recoveredAddr, strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex()),
		"sanity: recovered address harus match privKey yang dipakai sign")
}

func TestVerifyTransactionSignature_NonceNormalizedToLower(t *testing.T) {
	// Nonce disimpan lowercase; query dengan uppercase harus tetap match
	// karena fromAddr di-normalize ke lowercase.
	privKey, _ := crypto.GenerateKey()
	fromAddrLower := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())
	privKeyHex := hex.EncodeToString(crypto.FromECDSA(privKey))

	mem := newMockMemoryAdapter()
	nonce := "nonce-mixed"
	mem.Set(context.Background(), "tx_nonce:"+fromAddrLower, []byte(nonce), 0)

	toAddr := "0x" + strings.Repeat("e", 40)
	amount := 25.0
	msg := []byte("Send 25.00 to " + toAddr + " nonce:" + nonce)
	signature := signMessage(t, privKeyHex, msg)

	svc := NewVerifyTxService(mem)
	// Panggil dengan fromAddr uppercase
	fromAddrUpper := strings.ToUpper(fromAddrLower)
	err := svc.VerifyTransactionSignature(context.Background(), fromAddrUpper, toAddr, amount, nonce, signature)
	assert.NoError(t, err, "fromAddr harus dinormalisasi ke lowercase sebelum lookup nonce")
}

func TestVerifyBuySellSignature_HappyPath(t *testing.T) {
	privKey, _ := crypto.GenerateKey()
	addr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())
	privKeyHex := hex.EncodeToString(crypto.FromECDSA(privKey))

	mem := newMockMemoryAdapter()
	nonce := "bs-nonce"
	mem.Set(context.Background(), "tx_nonce:"+addr, []byte(nonce), 0)

	amount := 200.0
	msg := []byte(" BUY 200.00 nonce:" + nonce) // leading space sesuai format
	signature := signMessage(t, privKeyHex, msg)

	svc := NewVerifyTxService(mem)
	err := svc.VerifyBuySellSignature(context.Background(), addr, amount, nonce, signature, BuyTransaction)
	assert.NoError(t, err)
}

func TestVerifyBuySellSignature_InvalidNonce(t *testing.T) {
	privKey, _ := crypto.GenerateKey()
	addr := strings.ToLower(crypto.PubkeyToAddress(privKey.PublicKey).Hex())

	mem := newMockMemoryAdapter()
	mem.Set(context.Background(), "tx_nonce:"+addr, []byte("real"), 0)

	svc := NewVerifyTxService(mem)
	err := svc.VerifyBuySellSignature(context.Background(), addr, 10, "WRONG", "0x00", SellTransaction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid nonce")
}

// helper: bikin mock memory kosong
func mem_empty() *mockMemoryAdapter {
	return newMockMemoryAdapter()
}
