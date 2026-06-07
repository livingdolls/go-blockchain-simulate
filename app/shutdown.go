package app

import (
	"context"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/websocket"
	"github.com/livingdolls/go-blockchain-simulate/app/worker"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

// Shutdown menghentikan seluruh komponen aplikasi secara graceful.
// Urutan: worker/consumer → websocket hub → RabbitMQ connection.
func (a *AppConfig) Shutdown() {
	logger.LogInfo("Starting graceful shutdown...")

	stopWorkers(
		a.BlockWorker,
		a.CandleWorker,
		a.TransactionConsumer,
		a.PricingConsumer,
		a.VolumeConsumer,
		a.AuditConsumer,
		a.ReconcileConsumer,
		a.RewardCalculationConsumer,
		a.RewardDistributionConsumer,
	)

	if a.Hub != nil {
		closeHub(a.Hub, 15*time.Second)
	}

	if a.RMQClient != nil {
		a.RMQClient.Close()
	}

	logger.LogInfo("Shutdown complete")
}

// stopWorkers menghentikan semua worker secara paralel dengan total timeout 30 detik.
// Setiap worker di-stop di goroutine-nya sendiri; kita tunggu semua selesai
// atau timeout. Tidak ada blocking antar worker.
func stopWorkers(workers ...interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	wg.Add(len(workers))

	go func() {
		for _, w := range workers {
			go func(workerInstance interface{}) {
				defer wg.Done()
				stopOne(workerInstance)
			}(w)
		}
		wg.Wait()
		close(stopChan)
	}()

	select {
	case <-stopChan:
		logger.LogInfo("All workers stopped")
	case <-ctx.Done():
		logger.LogInfo("Timeout while stopping workers")
	}
}

// stopOne memanggil Stop() sesuai tipe konkret worker. Type switch memastikan
// setiap worker dipanggil method yang sesuai. Worker yang tidak dikenal di-skip.
func stopOne(w interface{}) {
	switch v := w.(type) {
	case *worker.GenerateBlockWorker:
		v.Stop()
		logger.LogInfo("Block worker stopped")
	case *worker.GenerateCandleWorker:
		v.Stop()
		logger.LogInfo("Candle worker stopped")
	case *worker.TransactionConsumer:
		v.Stop()
		logger.LogInfo("Transaction consumer stopped")
	case *worker.MarketPricingConsumer:
		v.Stop()
		logger.LogInfo("Market pricing consumer stopped")
	case *worker.MarketVolumeConsumer:
		v.Stop()
		logger.LogInfo("Market volume consumer stopped")
	case *worker.LedgerAuditConsumer:
		v.Stop()
		logger.LogInfo("Ledger audit consumer stopped")
	case *worker.LedgerReconcileConsumer:
		v.Stop()
		logger.LogInfo("Ledger reconcile consumer stopped")
	case *worker.RewardCalculationConsumer:
		v.Stop()
		logger.LogInfo("Reward calculation consumer stopped")
	case *worker.RewardDistributionConsumer:
		v.Stop()
		logger.LogInfo("Reward distribution consumer stopped")
	}
}

// closeHub menutup WebSocket hub dengan timeout. Jika Close() tidak selesai
// dalam waktu yang ditentukan, log warning tapi lanjut shutdown.
func closeHub(hub *websocket.Hub, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		hub.Close()
		close(done)
	}()

	select {
	case <-done:
		logger.LogInfo("WebSocket hub closed all connections")
	case <-ctx.Done():
		logger.LogInfo("Timeout while closing WebSocket hub connections")
	}
}
