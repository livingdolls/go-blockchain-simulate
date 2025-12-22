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
}

func NewGenerateBlockWorker(blockService services.BlockService) *GenerateBlockWorker {
	return &GenerateBlockWorker{
		blockService: blockService,
	}
}

func (w *GenerateBlockWorker) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			log.Println("GenerateBlockWorker: Generating new block...")
			_, err := w.blockService.GenerateBlock()
			if err != nil {
				if errors.Is(err, entity.ErrNoPendingTransactions) {
					log.Println("No pending transactions to include in the new block.")
					continue
				}
				log.Printf("generate block error : %v", err)
			}
		}
	}()
}
