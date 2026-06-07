package utils

import (
	"strings"
	"testing"

	"github.com/livingdolls/go-blockchain-simulate/app/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateProofOfWork(t *testing.T) {
	tests := []struct {
		name string
		in   models.Block
		want bool
	}{
		{
			name: "hash with 4 leading zeros passes difficulty 4",
			in:   models.Block{CurrentHash: "0000abcdef", Difficulty: 4},
			want: true,
		},
		{
			name: "hash with 3 leading zeros fails difficulty 4",
			in:   models.Block{CurrentHash: "000abcdef", Difficulty: 4},
			want: false,
		},
		{
			name: "difficulty 0 always passes (empty target)",
			in:   models.Block{CurrentHash: "deadbeef", Difficulty: 0},
			want: true,
		},
		{
			name: "empty hash fails any positive difficulty",
			in:   models.Block{CurrentHash: "", Difficulty: 1},
			want: false,
		},
		{
			name: "hash exceeds required leading zeros (5 of 4 required)",
			in:   models.Block{CurrentHash: "00000abc", Difficulty: 4},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ValidateProofOfWork(tt.in))
		})
	}
}

func TestGetDifficultyTarget(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int
		// 2^(256 - difficulty*4). Validate by checking bit-length.
		wantBitLen int
	}{
		{"difficulty 1 -> 2^252", 1, 253},
		{"difficulty 4 -> 2^240", 4, 241},
		{"difficulty 64 -> 2^0", 64, 1},
		{"difficulty 0 -> 2^256", 0, 257},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := GetDifficultyTarget(tt.difficulty)
			assert.Equal(t, tt.wantBitLen, target.BitLen(),
				"GetDifficultyTarget(%d) harus punya BitLen=%d", tt.difficulty, tt.wantBitLen)
		})
	}
}

func TestRecalculateBlockHash_Deterministic(t *testing.T) {
	// Hash harus deterministik untuk input identik.
	b := models.Block{
		BlockNumber:  42,
		PreviousHash: "prevhash",
		Nonce:        12345,
		Timestamp:    1700000000,
		Transactions: []models.Transaction{
			{ID: 1, FromAddress: "0xA", ToAddress: "0xB", Amount: 10, Fee: 0.01, Signature: "sig"},
		},
	}

	h1 := RecalculateBlockHash(b)
	h2 := RecalculateBlockHash(b)
	assert.Equal(t, h1, h2, "RecalculateBlockHash harus deterministik")
	assert.Len(t, h1, 64, "SHA-256 hex output harus 64 char")
}

func TestRecalculateBlockHash_DifferentNonce(t *testing.T) {
	// Nonce berbeda harus hasil hash berbeda.
	base := models.Block{
		BlockNumber:  1,
		PreviousHash: "x",
		Timestamp:    100,
	}
	h0 := RecalculateBlockHash(base)
	base.Nonce = 1
	h1 := RecalculateBlockHash(base)
	assert.NotEqual(t, h0, h1)
}

func TestCalculateNextDifficulty_BelowInterval(t *testing.T) {
	// < 10 blocks: harus return DefaultDifficulty (4).
	blocks := make([]models.Block, 5)
	for i := range blocks {
		blocks[i] = models.Block{BlockNumber: i, Difficulty: 7} // kasih 7 utk verifikasi return default
	}
	assert.Equal(t, DefaultDifficulty, CalculateNextDifficulty(blocks),
		"kurang dari 10 blocks harus return DefaultDifficulty")
}

func TestCalculateNextDifficulty_AtInterval(t *testing.T) {
	// 10 blocks dengan timestamp rapat (< expected/2 = 45s) -> naik +1
	baseTime := int64(1000000)
	blocks := make([]models.Block, 10)
	for i := range blocks {
		// span total hanya 20s (well under expected 90s / 2 = 45s)
		blocks[i] = models.Block{
			BlockNumber: i,
			Timestamp:   baseTime + int64(i*2),
			Difficulty:  4,
		}
	}
	got := CalculateNextDifficulty(blocks)
	assert.Equal(t, 5, got, "10 blocks span 18s (< 45s) harus naik difficulty 4->5")
}

func TestCalculateNextDifficulty_TooSlow(t *testing.T) {
	// 10 blocks dengan span > expected*2 = 180s -> turun (min 1)
	baseTime := int64(1000000)
	blocks := make([]models.Block, 10)
	for i := range blocks {
		// span total 300s (> 180s)
		blocks[i] = models.Block{
			BlockNumber: i,
			Timestamp:   baseTime + int64(i*30),
			Difficulty:  4,
		}
	}
	got := CalculateNextDifficulty(blocks)
	assert.Equal(t, 3, got, "10 blocks span 270s (> 180s) harus turun difficulty 4->3")
}

func TestCalculateNextDifficulty_FloorAtOne(t *testing.T) {
	// Difficulty 1 + span > 2*expected -> tidak boleh turun di bawah 1
	baseTime := int64(1000000)
	blocks := make([]models.Block, 10)
	for i := range blocks {
		blocks[i] = models.Block{
			BlockNumber: i,
			Timestamp:   baseTime + int64(i*30),
			Difficulty:  1,
		}
	}
	got := CalculateNextDifficulty(blocks)
	assert.Equal(t, 1, got, "difficulty 1 dengan span lambat harus floor di 1 (bukan 0)")
}

func TestCalculateNextDifficulty_OnTarget(t *testing.T) {
	// Span persis expected (90s) -> tidak ada perubahan
	baseTime := int64(1000000)
	blocks := make([]models.Block, 10)
	for i := range blocks {
		// span total 90s (expected = 10 * 9 = 90s)
		blocks[i] = models.Block{
			BlockNumber: i,
			Timestamp:   baseTime + int64(i*10),
			Difficulty:  4,
		}
	}
	got := CalculateNextDifficulty(blocks)
	assert.Equal(t, 4, got, "span = expected (90s) harus tetap di difficulty saat ini")
}

func TestMineBlock_DifficultyOne(t *testing.T) {
	// Difficulty 1: probabilitas 1/16, harus cepat selesai.
	ts := int64(1700000000)
	result := MineBlock(1, "prevhash", []models.Transaction{
		{ID: 1, FromAddress: "0xA", ToAddress: "0xB", Amount: 10, Fee: 0.01, Signature: "sig"},
	}, 1, ts)
	assert.NotEmpty(t, result.Hash, "harus menemukan hash")
	assert.True(t, strings.HasPrefix(result.Hash, "0"),
		"hash harus dimulai dengan minimal 1 leading zero (difficulty 1), got=%s", result.Hash)
	assert.Equal(t, 1, result.Difficulty)
	assert.GreaterOrEqual(t, result.Nonce, int64(0))
	assert.Greater(t, result.HashRate, 0.0)

	// Hash harus sama dengan RecalculateBlockHash untuk block dengan nonce yang ditemukan
	rebuilt := models.Block{
		BlockNumber:  1,
		PreviousHash: "prevhash",
		Nonce:        result.Nonce,
		Difficulty:   1,
		Timestamp:    ts,
		Transactions: []models.Transaction{
			{ID: 1, FromAddress: "0xA", ToAddress: "0xB", Amount: 10, Fee: 0.01, Signature: "sig"},
		},
	}
	assert.Equal(t, result.Hash, RecalculateBlockHash(rebuilt),
		"hash yang ditambang harus bisa di-recover via RecalculateBlockHash dengan timestamp yang sama")
}
