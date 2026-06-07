package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryLimiter_Allow(t *testing.T) {
	// rate=10/s, burst=3: 3 request pertama langsung OK, ke-4 ditolak.
	l := NewInMemoryLimiter(10, 3)

	for i := 0; i < 3; i++ {
		assert.True(t, l.Allow("client-a"), "request ke-%d harus diizinkan", i+1)
	}
	assert.False(t, l.Allow("client-a"), "request ke-4 harus ditolak (burst habis)")
}

func TestInMemoryLimiter_PerKey(t *testing.T) {
	// Setiap identifier punya bucket sendiri, tidak shared.
	l := NewInMemoryLimiter(10, 2)

	// Habiskan bucket untuk client-a
	assert.True(t, l.Allow("client-a"))
	assert.True(t, l.Allow("client-a"))
	assert.False(t, l.Allow("client-a"))

	// client-b harus masih punya token penuh
	assert.True(t, l.Allow("client-b"))
	assert.True(t, l.Allow("client-b"))
	assert.False(t, l.Allow("client-b"))
}

func TestInMemoryLimiter_Refill(t *testing.T) {
	// rate=10/s, burst=1: 1 request langsung OK, lalu tunggu 200ms untuk refill.
	l := NewInMemoryLimiter(10, 1)

	assert.True(t, l.Allow("client"))
	assert.False(t, l.Allow("client"))

	// Tunggu refill (10 token/detik = 1 token / 100ms).
	time.Sleep(150 * time.Millisecond)

	assert.True(t, l.Allow("client"), "setelah refill harus diizinkan")
}

func TestInMemoryLimiter_EmptyIdentifier(t *testing.T) {
	// Identifier kosong harus selalu diizinkan (fail-open untuk IP kosong).
	l := NewInMemoryLimiter(1, 1)
	for i := 0; i < 10; i++ {
		assert.True(t, l.Allow(""), "identifier kosong harus selalu diizinkan")
	}
}
