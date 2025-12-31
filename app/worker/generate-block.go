package worker

import (
	"errors"
	"log"
	"time"

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
				log.Println("GenerateBlockWorker: Generating new block...")
				_, err := w.blockService.GenerateBlock()
				if err != nil {
					if errors.Is(err, entity.ErrNoPendingTransactions) {
						log.Println("No pending transactions to include in the new block.")
						continue
					}
					log.Printf("generate block error : %v", err)
				}
			case <-w.stopChan:
				log.Println("GenerateBlockWorker: Stopping ticker...")
				return
			}
		}
	}()
}

func (w *GenerateBlockWorker) Stop() {
	log.Println("GenerateBlockWorker: Stopping worker...")
	close(w.stopChan)
	log.Println("GenerateBlockWorker: Worker stopped.")
}
