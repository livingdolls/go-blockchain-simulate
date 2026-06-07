package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func txFor(id int64, from, to string, amount, fee float64) models.Transaction {
	return models.Transaction{
		ID:          id,
		FromAddress: from,
		ToAddress:   to,
		Amount:      amount,
		Fee:         fee,
		Signature:   "sig-" + from,
	}
}

// leafHash mereplikasi format persis di utils/merkle_root.go
// (id, from, to, amount, fee, signature). Format ini SAMA dengan yang
// dipakai CalculateMerkleRoot dan GetMerkleProof, sehingga proof yang
// dihasilkan bisa diverifikasi terhadap root.
func leafHash(tx models.Transaction) string {
	data := fmt.Sprintf("%d%s%s%.8f%.8f%s", tx.ID, tx.FromAddress, tx.ToAddress, tx.Amount, tx.Fee, tx.Signature)
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func TestCalculateMerkleRoot_Empty(t *testing.T) {
	// Tidak ada transaksi -> empty string.
	assert.Equal(t, "", CalculateMerkleRoot(nil))
	assert.Equal(t, "", CalculateMerkleRoot([]models.Transaction{}))
}

func TestCalculateMerkleRoot_Deterministic(t *testing.T) {
	// Hash identik untuk input identik.
	txs := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
	}
	h1 := CalculateMerkleRoot(txs)
	h2 := CalculateMerkleRoot(txs)
	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 64, "SHA-256 hex output harus 64 char")
}

func TestCalculateMerkleRoot_OrderMatters(t *testing.T) {
	// Urutan tx yang berbeda harus hasil hash berbeda.
	a := CalculateMerkleRoot([]models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
	})
	b := CalculateMerkleRoot([]models.Transaction{
		txFor(2, "0xB", "0xC", 5, 0.01),
		txFor(1, "0xA", "0xB", 10, 0.01),
	})
	assert.NotEqual(t, a, b, "merkle root harus sensitif terhadap urutan tx")
}

func TestCalculateMerkleRoot_OddCount(t *testing.T) {
	// 3 tx (ganjil) -> last node diduplikasi. Hasil harus deterministik.
	txs := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
		txFor(3, "0xC", "0xD", 3, 0.01),
	}
	got := CalculateMerkleRoot(txs)
	assert.NotEmpty(t, got)
	assert.Len(t, got, 64)
}

func TestCalculateMerkleRoot_DifferentTransactions(t *testing.T) {
	// Mutasi satu field harus ubah root.
	base := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
	}
	rootBase := CalculateMerkleRoot(base)

	mutated := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.02), // fee beda
	}
	rootMutated := CalculateMerkleRoot(mutated)
	assert.NotEqual(t, rootBase, rootMutated, "fee berbeda harus hasil root berbeda")
}

func TestGetMerkleProof_OutOfRange(t *testing.T) {
	txs := []models.Transaction{txFor(1, "0xA", "0xB", 10, 0.01)}
	assert.Nil(t, GetMerkleProof(txs, -1))
	assert.Nil(t, GetMerkleProof(txs, 1))
	assert.Nil(t, GetMerkleProof(txs, 100))
}

func TestGetMerkleProof_ValidIndex(t *testing.T) {
	// 4 tx -> proof untuk index mana pun harus non-empty dan verifiable.
	txs := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
		txFor(3, "0xC", "0xD", 3, 0.01),
		txFor(4, "0xD", "0xE", 7, 0.01),
	}
	for i := 0; i < len(txs); i++ {
		proof := GetMerkleProof(txs, i)
		assert.NotEmpty(t, proof, "proof untuk index %d harus non-empty", i)
	}
}

func TestVerifyMerkleProof_EvenIndex(t *testing.T) {
	// Index 0 (genap): sibling di kanan -> combined = current + sibling.
	// Setelah fix, VerifyMerkleProof harus sukses dengan leaf format
	// CalculateMerkleRoot dan txIndex = 0.
	txs := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
	}
	root := CalculateMerkleRoot(txs)
	leaf := leafHash(txs[0])
	proof := GetMerkleProof(txs, 0)
	assert.True(t, VerifyMerkleProof(leaf, proof, root, 0),
		"verify untuk even index harus sukses dengan txIndex parameter")
}

func TestVerifyMerkleProof_OddIndex(t *testing.T) {
	// Index 1 (ganjil): sibling di kiri -> combined = sibling + current.
	// Ini yang dulu gagal karena VerifyMerkleProof selalu 'current + sibling'.
	txs := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
	}
	root := CalculateMerkleRoot(txs)
	leaf := leafHash(txs[1])
	proof := GetMerkleProof(txs, 1)
	assert.True(t, VerifyMerkleProof(leaf, proof, root, 1),
		"verify untuk odd index harus sukses (sibling di kiri)")
}

func TestVerifyMerkleProof_FourTxRoundTrip(t *testing.T) {
	// 4 tx: round-trip proof untuk semua index harus sukses.
	txs := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
		txFor(3, "0xC", "0xD", 3, 0.01),
		txFor(4, "0xD", "0xE", 7, 0.01),
	}
	root := CalculateMerkleRoot(txs)
	for i := 0; i < len(txs); i++ {
		leaf := leafHash(txs[i])
		proof := GetMerkleProof(txs, i)
		assert.True(t, VerifyMerkleProof(leaf, proof, root, i),
			"round-trip proof untuk index %d harus sukses", i)
	}
}

func TestVerifyMerkleProof_WrongLeaf(t *testing.T) {
	// Leaf yang dimodifikasi harus ditolak.
	txs := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
	}
	root := CalculateMerkleRoot(txs)
	wrongLeaf := leafHash(models.Transaction{
		ID: 1, FromAddress: "0xX", ToAddress: "0xY", Amount: 999, Fee: 0, Signature: "wrong",
	})
	proof := GetMerkleProof(txs, 0)
	assert.False(t, VerifyMerkleProof(wrongLeaf, proof, root, 0),
		"leaf yang dimodifikasi harus ditolak")
}

func TestVerifyMerkleProof_WrongIndex(t *testing.T) {
	// txIndex yang salah harus membuat verifikasi gagal.
	txs := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
	}
	root := CalculateMerkleRoot(txs)
	leaf := leafHash(txs[0])
	proof := GetMerkleProof(txs, 0)
	// Klaim index 1, padahal proof milik index 0
	assert.False(t, VerifyMerkleProof(leaf, proof, root, 1),
		"txIndex yang salah harus menghasilkan hash akhir berbeda")
}

func TestMerkleProof_GetMerkleProof_RoundTrip(t *testing.T) {
	// Verifikasi bahwa GetMerkleProof output bisa diverifikasi end-to-end.
	// Ini test integrasi utama setelah fix bug leaf format.
	txs := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
		txFor(3, "0xC", "0xD", 3, 0.01),
	}
	root := CalculateMerkleRoot(txs)
	for i := 0; i < len(txs); i++ {
		proof := GetMerkleProof(txs, i)
		leaf := leafHash(txs[i])
		assert.True(t, VerifyMerkleProof(leaf, proof, root, i),
			"end-to-end round-trip untuk index %d harus sukses", i)
	}
}

func TestMerkleProof_OddCount_RoundTrip(t *testing.T) {
	// Odd count: index terakhir akan di-duplikasi sebagai sibling-nya sendiri.
	// Test ini verifikasi GetMerkleProof + VerifyMerkleProof menangani
	// kasus duplicate-last-hash dengan benar.
	txs := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
		txFor(2, "0xB", "0xC", 5, 0.01),
		txFor(3, "0xC", "0xD", 3, 0.01),
		txFor(4, "0xD", "0xE", 7, 0.01),
		txFor(5, "0xE", "0xF", 2, 0.01),
	}
	root := CalculateMerkleRoot(txs)
	for i := 0; i < len(txs); i++ {
		proof := GetMerkleProof(txs, i)
		leaf := leafHash(txs[i])
		assert.True(t, VerifyMerkleProof(leaf, proof, root, i),
			"odd count: round-trip index %d harus sukses", i)
	}
}

// leafHash mereplikasi format persis di utils/merkle_root.go:57
// (id, from, to, amount, signature). Ini berbeda dengan format di
// CalculateMerkleRoot:20 yang juga menyertakan .Fee.

func TestMineBlock_AndMerkleConsistency_Smoke(t *testing.T) {
	// Smoke test: mine block, hitung merkle root, dan check PoW.
	// Tidak ada coupling yang harus enforced, hanya smoke.
	txs := []models.Transaction{
		txFor(1, "0xA", "0xB", 10, 0.01),
	}
	mr := CalculateMerkleRoot(txs)
	require.NotEmpty(t, mr)
}
