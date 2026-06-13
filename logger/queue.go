package logger

import (
	"sync"
	"sync/atomic"
)

type logEvent struct {
	fn func()
}

// asyncQueue adalah MPSC (multi-producer single-consumer) queue
// dengan done-channel shutdown pattern.
//
// Kenapa TIDAK close(ch) saat shutdown:
//
//   Close channel sementara goroutine lain sedang send ke channel
//   itu adalah data race (runtime closechan vs chansend). Recover
//   di push() bisa handle panic, tapi TETAP race di level memory
//   model Go (race detector flag ini). Untuk production code yang
//   di-build dengan -race, ini unacceptable.
//
// Solusi: pakai done channel sebagai signal shutdown, JANGAN
// close ch. Senders select pada {ch <-, done} - kalau done closed,
// sender drop event. Receiver (worker) detect done, drain sisa
// event dari ch, lalu exit.
//
// Pattern ini umum di Go stdlib (context.Context, http.Server.Shutdown)
// dan zero-cost kalau tidak shutdown.
type asyncQueue struct {
	ch             chan logEvent
	done           chan struct{} // closed saat shutdown
	doneOnce       sync.Once     // guarantee close(done) sekali
	dropped        uint64        // atomic
	total          uint64        // atomic
	mu             sync.RWMutex  // protects state flag
	closed         uint32        // atomic flag; 1 = closed, push fast-path returns
	processingTime int64         // nanoseconds
}

func newAsyncQueue(size int) *asyncQueue {
	return &asyncQueue{
		ch:   make(chan logEvent, size),
		done: make(chan struct{}),
	}
}

// isClosed returns true kalau queue sudah di-shutdown. Fast-path
// check di push() untuk skip seluruh select machinery kalau queue
// jelas-jelas sudah closed (hot path optimization).
func (q *asyncQueue) isClosed() bool {
	return atomic.LoadUint32(&q.closed) == 1
}

// markClosed set flag close. Idempotent (aman dipanggil multiple kali).
// Channel close aktual dilakukan di closeCh() - lihat comment di sana
// untuk kenapa pakai done channel bukan close(ch).
func (q *asyncQueue) markClosed() {
	atomic.StoreUint32(&q.closed, 1)
}

// closeCh men-signal shutdown dengan close(done) sekali via sync.Once.
// JANGAN close(q.ch) di sini - lihat type comment di atas untuk
// alasannya (data race dengan concurrent senders).
//
// Setelah closeCh return, semua push() berikutnya akan return early
// (via select case <-q.done) sehingga ch aman di-drain oleh worker.
func (q *asyncQueue) closeCh() {
	q.doneOnce.Do(func() {
		close(q.done)
	})
}

// push memasukkan event ke queue. Tiga path:
//
//  1. Queue sudah closed (isClosed fast-path): drop silent, no counter
//     increment. Hindari misleading post-shutdown stats.
//  2. dropOnFull=true: non-blocking send via select. Kalau queue full
//     ATAU shutdown signal received, drop + increment counter.
//  3. dropOnFull=false: blocking send. Pilih ch <- atau done, mana
//     yang duluan dapat. Return tanpa increment kalau shutdown duluan.
//
// Race-free: done channel jadi synchronization point antara closeCh
// dan push. Tidak ada data race di level memory model.
func (q *asyncQueue) push(ev logEvent, dropOnFull bool) {
	if q.isClosed() {
		return
	}

	atomic.AddUint64(&q.total, 1)

	if dropOnFull {
		select {
		case q.ch <- ev:
			// sent
		case <-q.done:
			// shutdown during send attempt - count as dropped
			atomic.AddUint64(&q.dropped, 1)
		default:
			// queue full + not shutdown - count as dropped
			atomic.AddUint64(&q.dropped, 1)
		}
		return
	}

	// Blocking send: tunggu ch punya slot ATAU shutdown signal.
	// Tidak ada data race karena done channel jadi coordination point.
	select {
	case q.ch <- ev:
	case <-q.done:
		atomic.AddUint64(&q.dropped, 1)
	}
}

// GetStats returns queue statistics
func (q *asyncQueue) GetStats() (dropped, total uint64, queueLen int) {
	return atomic.LoadUint64(&q.dropped), atomic.LoadUint64(&q.total), len(q.ch)
}
