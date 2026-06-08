package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/logger"

	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type CandleJob struct {
	Interval  string
	Timestamp int64
}

type GenerateCandleWorker struct {
	candleService services.CandleService
	workerCount   int
	jobTimeout    time.Duration

	jobs     chan CandleJob
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewGenerateCandlesWorker(candleService services.CandleService, workerCount int) *GenerateCandleWorker {
	return &GenerateCandleWorker{
		candleService: candleService,
		workerCount:   workerCount,
		// jobTimeout HARUS < shutdown timeout (30s di app/shutdown.go)
		// agar worker dapat menyelesaikan job dalam window shutdown.
		// Margin 5 detik: stop timeout (30s) - jobTimeout (25s) = 5s untuk cleanup.
		jobTimeout: 25 * time.Second,
		jobs:       make(chan CandleJob, len([]string{"1m", "5m", "15m", "30m", "1h", "4h", "1d"})*2),
		stopChan:   make(chan struct{}),
	}
}

func (w *GenerateCandleWorker) Start(interval time.Duration) {
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.worker(i)
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.dispatchJobs()
			case <-w.stopChan:
				close(w.jobs)
				return
			}
		}
	}()
}

func (w *GenerateCandleWorker) dispatchJobs() {
	now := time.Now().Unix()

	intervals := []string{"1m", "5m", "15m", "30m", "1h", "4h", "1d"}

	for _, interval := range intervals {
		select {
		case w.jobs <- CandleJob{
			Interval:  interval,
			Timestamp: now,
		}:
		case <-w.stopChan:
			return
		}
	}
}

func (w *GenerateCandleWorker) worker(id int) {
	defer w.wg.Done()

	for job := range w.jobs {
		// Pakai w.jobTimeout (bukan hard-coded 30s) agar konsisten dengan
		// default NewGenerateCandlesWorker dan mudah di-override via
		// SetJobTimeout. Shutdown timeout 30s harus > jobTimeout + margin
		// agar Stop() bisa selesai dalam window shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), w.jobTimeout)
		if err := w.candleService.AggregateCandle(ctx, job.Interval, job.Timestamp); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				logger.LogInfo(fmt.Sprintf("[worker-%d] timeout: interval %s took > %v", id, job.Interval, w.jobTimeout))

			} else {
				logger.LogError(
					fmt.Sprintf("[worker-%d] error interval %s", id, job.Interval),
					err,
				)
			}
		}
		// clean context
		cancel()
	}
}

func (w *GenerateCandleWorker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
	logger.LogInfo("Candles worker gracefully stopped")
}

func (w *GenerateCandleWorker) SetJobTimeout(timeout time.Duration) {
	w.jobTimeout = timeout
}
