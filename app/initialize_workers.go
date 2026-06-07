package app

import (
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/worker"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

// InitializeWorkers menginisialisasi background worker yang berjalan
// periodik (bukan dari RabbitMQ). Worker ini di-stop di Shutdown().
func (a *AppConfig) InitializeWorkers() {
	// Block generation: generate block baru setiap 10 detik (jika ada pending tx)
	a.BlockWorker = worker.NewGenerateBlockWorker(a.BlockService)
	a.BlockWorker.Start(10 * time.Second)

	// Candle generation: aggregate trade menjadi candle setiap 1 detik, 4 worker paralel
	a.CandleWorker = worker.NewGenerateCandlesWorker(a.CandleService, 4)
	a.CandleWorker.SetJobTimeout(45 * time.Second)
	a.CandleWorker.Start(1 * time.Second)

	logger.LogInfo("All background workers initialized successfully")
}
