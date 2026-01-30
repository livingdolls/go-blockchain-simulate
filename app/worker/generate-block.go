package worker

import (
	"errors"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/logger"

	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/livingdolls/go-blockchain-simulate/app/services"
)

type GenerateBlockWorker struct {
	blockService services.BlockService
	stopChan     chan struct{}
	ticker       *time.Ticker
}

func NewGenerateBlockWorker(blockService services.BlockService) *GenerateBlockWorker {
	return &GenerateBlockWorker{
		blockService: blockService,
		stopChan:     make(chan struct{}),
	}
}

func (w *GenerateBlockWorker) Start(interval time.Duration) {
	w.ticker = time.NewTicker(interval)

	go func() {
		defer w.ticker.Stop()

		for {
			select {
			case <-w.ticker.C:
				_, err := w.blockService.GenerateBlock()
				if err != nil {
					if errors.Is(err, entity.ErrNoPendingTransactions) {
						continue
					}
					logger.LogError("Generate block error", err)
				}
			case <-w.stopChan:
				logger.LogInfo("GenerateBlockWorker: Stopping ticker")
				return
			}
		}
	}()
}

func (w *GenerateBlockWorker) Stop() {
	logger.LogInfo("GenerateBlockWorker: Stopping worker")
	close(w.stopChan)
	logger.LogInfo("GenerateBlockWorker: Worker stopped")
}
