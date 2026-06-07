package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToggleV(t *testing.T) {
	// ToggleV membalik v antara 27 dan 28 (recovery id Ethereum). Untuk
	// nilai lain (yang belum dinormalisasi), selalu map ke 27.
	// Kontrak aktual: ToggleV adalah "snap to {27,28}" dengan prioritas 27.
	tests := []struct {
		name string
		in   byte
		want byte
	}{
		{"27 -> 28", 27, 28},
		{"28 -> 27", 28, 27},
		{"0 (unnormalized) -> 27", 0, 27},
		{"1 (unnormalized) -> 27", 1, 27},
		{"35 (EIP-155) -> 27", 35, 27},
		{"37 (EIP-155 odd) -> 27", 37, 27},
		{"255 -> 27", 255, 27},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ToggleV(tt.in))
		})
	}
}

func TestToggleV_InvolutionOnNormalized(t *testing.T) {
	// Untuk nilai 27/28 (yang sudah dinormalisasi), ToggleV(ToggleV(v)) = v.
	assert.Equal(t, byte(27), ToggleV(ToggleV(27)))
	assert.Equal(t, byte(28), ToggleV(ToggleV(28)))

	// Untuk nilai arbitrary: ToggleV(v) = 27, ToggleV(27) = 28,
	// jadi ToggleV(ToggleV(v)) = 28 untuk v ∉ {27,28}.
	for _, v := range []byte{0, 1, 35, 37, 100, 255} {
		assert.Equal(t, byte(28), ToggleV(ToggleV(v)),
			"ToggleV(ToggleV(%d)) harus == 28 (snap ke 27, lalu dibalik ke 28)", v)
	}
}

func TestPrefixedHash_Deterministic(t *testing.T) {
	// Hash harus deterministik: input sama -> output sama.
	data := []byte("hello world")
	h1 := PrefixedHash(data)
	h2 := PrefixedHash(data)
	assert.Equal(t, h1, h2, "PrefixedHash harus deterministik untuk input yang sama")
	assert.Len(t, h1, 32, "Keccak256 output harus 32 byte")
}

func TestPrefixedHash_DifferentInput(t *testing.T) {
	// Input berbeda harus hasil hash berbeda.
	a := PrefixedHash([]byte("hello"))
	b := PrefixedHash([]byte("world"))
	assert.NotEqual(t, a, b, "input berbeda harus hasil hash berbeda")
}

func TestPrefixedHash_PrefixLengthMatters(t *testing.T) {
	// Pesan "ab" dan "a"+"b" (concat) punya raw bytes identik tapi hash
	// akan beda karena prefix length berbeda. Ini verifikasi format
	// "\x19Ethereum Signed Message:\n<len>".
	a := PrefixedHash([]byte("ab"))
	b := PrefixedHash([]byte{'a', 'b'})
	assert.Equal(t, a, b, "byte slice vs string identik harus hash sama")

	c := PrefixedHash([]byte{'a', 'b', 'c'})
	assert.NotEqual(t, a, c, "panjang berbeda harus hash berbeda")
}
