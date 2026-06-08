package logger

import (
	"sync"
	"sync/atomic"
)

type logEvent struct {
	fn func()
}

type asyncQueue struct {
	ch             chan logEvent
	dropped        uint64
	total          uint64
	closed         uint32 // atomic flag; 1 = closed, push returns early
	mu             sync.RWMutex
	processingTime int64 // nanoseconds
}

func newAsyncQueue(size int) *asyncQueue {
	return &asyncQueue{
		ch: make(chan logEvent, size),
	}
}

// isClosed returns true kalau queue sudah di-close. Setelah close,
// push() akan drop event secara silent untuk mencegah panic
// "send on closed channel" dari goroutine yang log di detik-detik
// terakhir shutdown.
func (q *asyncQueue) isClosed() bool {
	return atomic.LoadUint32(&q.closed) == 1
}

// markClosed set flag close. Idempotent (aman dipanggil multiple kali).
// Channel close aktual dilakukan di closeCh() untuk menjamin hanya
// satu goroutine yang close (mencegah double-close panic).
func (q *asyncQueue) markClosed() {
	atomic.StoreUint32(&q.closed, 1)
}

// closeCh menutup channel. Aman dari double-close: hanya markClosed
// caller pertama yang benar-benar close; caller kedua akan dapat
// panic "close of closed channel" di sini, yang kita recover.
func (q *asyncQueue) closeCh() {
	defer func() {
		// Recover dari double-close panic. Idempotent.
		_ = recover()
	}()
	close(q.ch)
}

func (q *asyncQueue) push(ev logEvent, dropOnFull bool) {
	// Jika queue sudah di-shutdown, drop event secara silent.
	// Jangan panic dan jangan increment counter (supaya stats
	// tidak misleading post-shutdown).
	if q.isClosed() {
		return
	}

	atomic.AddUint64(&q.total, 1)

	if dropOnFull {
		select {
		case q.ch <- ev:
		default:
			atomic.AddUint64(&q.dropped, 1)
		}
	} else {
		// Double-check isClosed sebelum blocking send: antara
		// check di atas dan push, Shutdown() bisa saja close channel.
		if q.isClosed() {
			atomic.AddUint64(&q.dropped, 1)
			return
		}
		// Wrap blocking send dengan recover untuk handle race
		// terakhir di mana channel di-close saat kita push.
		defer func() {
			if r := recover(); r != nil {
				atomic.AddUint64(&q.dropped, 1)
			}
		}()
		q.ch <- ev
	}
}

// GetStats returns queue statistics
func (q *asyncQueue) GetStats() (dropped, total uint64, queueLen int) {
	return atomic.LoadUint64(&q.dropped), atomic.LoadUint64(&q.total), len(q.ch)
}
