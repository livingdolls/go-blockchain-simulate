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
	mu             sync.RWMutex
	processingTime int64 // nanoseconds
}

func newAsyncQueue(size int) *asyncQueue {
	return &asyncQueue{
		ch: make(chan logEvent, size),
	}
}

func (q *asyncQueue) push(ev logEvent, dropOnFull bool) {
	atomic.AddUint64(&q.total, 1)

	if dropOnFull {
		select {
		case q.ch <- ev:
		default:
			atomic.AddUint64(&q.dropped, 1)
		}
	} else {
		q.ch <- ev
	}
}

// GetStats returns queue statistics
func (q *asyncQueue) GetStats() (dropped, total uint64, queueLen int) {
	return atomic.LoadUint64(&q.dropped), atomic.LoadUint64(&q.total), len(q.ch)
}
