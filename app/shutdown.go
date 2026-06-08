package app

import (
	"context"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/websocket"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

// stoppable adalah interface internal untuk semua worker/consumer yang
// bisa di-Stop(). Implementasi interface ini WAJIB ada di setiap
// worker/consumer; tidak menggunakan type-switch (yang silent skip
// kalau struct baru ditambahkan tanpa update switch).
type stoppable interface {
	Stop()
}

// Shutdown menghentikan seluruh komponen aplikasi secara graceful.
// Urutan: worker/consumer → websocket hub → RabbitMQ → Redis → DB.
// Worker/consumer di-stop duluan agar tidak ada publish baru ke broker
// setelah broker ditutup. Redis & DB di-close paling akhir karena worker
// bisa saja masih menggunakan cache atau menulis audit log.
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
		a.NotificationWSConsumer,
	)

	if a.Hub != nil {
		closeHub(a.Hub, 15*time.Second)
	}

	if a.RMQClient != nil {
		a.RMQClient.Close()
		logger.LogInfo("RabbitMQ connection ditutup")
	}

	// Tutup Redis setelah RMQ. Worker bisa saja sempat publish di detik
	// terakhir; setelah itu baru pool Redis di-release.
	if a.deps != nil && a.deps.RedisClient != nil {
		if err := a.deps.RedisClient.Close(); err != nil {
			logger.LogError("gagal menutup Redis client", err)
		} else {
			logger.LogInfo("Redis client ditutup")
		}
	}

	// DB di-close paling akhir. Consumer bisa saja commit transaksi
	// final di detik-detik shutdown (mis. balance update).
	if a.deps != nil && a.deps.DBConn != nil {
		if err := a.deps.DBConn.Close(); err != nil {
			logger.LogError("gagal menutup DB connection", err)
		} else {
			logger.LogInfo("DB connection ditutup")
		}
	}

	logger.LogInfo("Shutdown complete")
}

// stopWorkers menghentikan semua worker secara paralel dengan total timeout 30 detik.
// Setiap worker di-stop di goroutine-nya sendiri; kita tunggu semua selesai
// atau timeout. Tidak ada blocking antar worker.
func stopWorkers(workers ...stoppable) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	wg.Add(len(workers))

	go func() {
		for _, w := range workers {
			if w == nil {
				wg.Done()
				continue
			}
			go func(workerInstance stoppable) {
				defer wg.Done()
				workerInstance.Stop()
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
