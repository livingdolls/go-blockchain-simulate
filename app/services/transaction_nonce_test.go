package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newNonceTestService() (TransactionService, *mockMemoryAdapter) {
	mem := newMockMemoryAdapter()
	svc := NewTransactionService(nil, nil, nil, nil, nil, mem, nil, nil)
	return svc, mem
}

func TestGenerateTransactionNonce_Format(t *testing.T) {
	// Nonce harus UUID v4 format dan disimpan dengan key lowercased address.
	svc, mem := newNonceTestService()

	addr := "0xAbCdEf1234567890"
	nonce := svc.GenerateTransactionNonce(context.Background(), addr)

	assert.NotEmpty(t, nonce, "nonce harus non-empty")
	assert.Len(t, nonce, 36, "UUID v4 string harus 36 char (32 hex + 4 dash)")

	// Verify disimpan di memory dengan key lowercase
	val, ok := mem.Get(context.Background(), "tx_nonce:0xabcdef1234567890")
	assert.True(t, ok, "nonce harus tersimpan di memory dengan key lowercase")
	assert.Equal(t, nonce, string(val))
}

func TestGenerateTransactionNonce_Unique(t *testing.T) {
	// Generate 100 nonce untuk address sama: harus unik (overwrite, tapi
	// 100 UUID v4 akan unik secara astronomis).
	svc, _ := newNonceTestService()
	seen := make(map[string]bool)

	for i := 0; i < 100; i++ {
		n := svc.GenerateTransactionNonce(context.Background(), "0xabc")
		require.False(t, seen[n], "nonce duplikat pada iterasi %d: %s", i, n)
		seen[n] = true
	}
}

func TestGenerateTransactionNonce_OverwritesPrevious(t *testing.T) {
	// Generate nonce baru: nonce lama harus di-overwrite (TTL 5 menit).
	svc, mem := newNonceTestService()

	first := svc.GenerateTransactionNonce(context.Background(), "0xabc")
	second := svc.GenerateTransactionNonce(context.Background(), "0xabc")

	assert.NotEqual(t, first, second, "nonce baru harus beda")
	val, _ := mem.Get(context.Background(), "tx_nonce:0xabc")
	assert.Equal(t, second, string(val), "memory harus berisi nonce terbaru")
}

func TestGenerateTransactionNonce_AddressNormalization(t *testing.T) {
	// Address uppercase dan lowercase harus map ke key yang sama.
	svc, mem := newNonceTestService()

	svc.GenerateTransactionNonce(context.Background(), "0xABCDEF")
	_, okUpper := mem.Get(context.Background(), "tx_nonce:0xabcdef")
	assert.True(t, okUpper, "uppercase address harus dinormalisasi ke lowercase")
	// Key uppercase tidak boleh ada
	_, okRaw := mem.Get(context.Background(), "tx_nonce:0xABCDEF")
	assert.False(t, okRaw, "tidak boleh ada key dengan case original")
}
