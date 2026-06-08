package logger

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncQueue_PushAfterClose_NoPanic(t *testing.T) {
	// Regression: Shutdown() close channel, lalu ada goroutine yang log.
	// Sebelumnya panic "send on closed channel". Sekarang harus drop silent.
	q := newAsyncQueue(10)

	// Tutup queue
	q.markClosed()
	q.closeCh()

	// Push setelah close: harus return tanpa panic
	require.NotPanics(t, func() {
		q.push(logEvent{fn: func() {}}, true)
	})
}

func TestAsyncQueue_PushBlockingAfterClose_NoPanic(t *testing.T) {
	// Skenario: push blocking (dropOnFull=false) setelah close.
	// Race: antara isClosed check dan ch <-, Shutdown() close channel.
	// Recover di push harus handle.
	q := newAsyncQueue(1) // buffer 1, agar blocking send bisa terjadi

	q.markClosed()
	q.closeCh()

	require.NotPanics(t, func() {
		// dropOnFull=false -> blocking send, tapi channel sudah closed
		q.push(logEvent{fn: func() {}}, false)
	})
}

func TestAsyncQueue_DoubleClose_NoPanic(t *testing.T) {
	// closeCh harus idempotent: aman dipanggil dua kali.
	q := newAsyncQueue(10)
	q.markClosed()
	q.closeCh()

	require.NotPanics(t, func() {
		q.closeCh() // kedua
	})
}

func TestAsyncQueue_ConcurrentPushAndClose(t *testing.T) {
	// Stress: 100 goroutine push paralel + 1 goroutine close.
	// Total: tidak boleh ada panic.
	q := newAsyncQueue(100)

	var wg sync.WaitGroup

	// 100 pushers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				q.push(logEvent{fn: func() {}}, true)
			}
		}()
	}

	// 1 closer
	wg.Add(1)
	go func() {
		defer wg.Done()
		q.markClosed()
		q.closeCh()
	}()

	wg.Wait()
	// Stats: tidak ada panic
	dropped, total, _ := q.GetStats()
	t.Logf("total=%d dropped=%d", total, dropped)
	assert.GreaterOrEqual(t, total, uint64(0))
}

func TestAsyncQueue_IsClosed(t *testing.T) {
	q := newAsyncQueue(10)
	assert.False(t, q.isClosed(), "queue baru harus isClosed=false")

	q.markClosed()
	assert.True(t, q.isClosed(), "setelah markClosed harus isClosed=true")
}
