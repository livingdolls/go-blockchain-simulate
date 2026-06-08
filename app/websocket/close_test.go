package websocket

import (
	"testing"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHub_Close_Idempotent(t *testing.T) {
	// Regression: Close() dipanggil dua kali tidak boleh panic.
	h := NewHub()
	go h.Run()

	// Tunggu Run() consume stop register
	time.Sleep(10 * time.Millisecond)

	require.NotPanics(t, func() {
		h.Close()
	}, "Close pertama harus sukses")

	require.NotPanics(t, func() {
		h.Close()
	}, "Close kedua harus idempotent (tidak panic)")

	// Close ketiga juga harus aman
	require.NotPanics(t, func() {
		h.Close()
	})
}

func TestHub_Close_StopsRun(t *testing.T) {
	// Run() harus exit setelah Close().
	h := NewHub()
	done := make(chan struct{})
	go func() {
		h.Run()
		close(done)
	}()

	// Tunggu Run consume
	time.Sleep(10 * time.Millisecond)

	h.Close()

	select {
	case <-done:
		// Run sudah exit
	case <-time.After(2 * time.Second):
		t.Fatal("Run() tidak exit setelah Close() dalam 2 detik")
	}
}

func TestHub_BroadcastAfterClose_Dropped(t *testing.T) {
	// Broadcast setelah Close() harus drop (select ke stopChan), bukan
	// panic.
	h := NewHub()
	go h.Run()
	time.Sleep(10 * time.Millisecond)
	h.Close()

	assert.NotPanics(t, func() {
		h.BroadCast(entity.MessageType("test-type"), "test")
	}, "Broadcast setelah Close() harus drop silent")
}
