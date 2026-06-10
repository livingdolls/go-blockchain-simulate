package app

import (
	"context"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/worker"
	"github.com/livingdolls/go-blockchain-simulate/logger"
)

// InitializeConsumers menginisialisasi semua message consumer (RabbitMQ).
// Consumer belum start di sini; panggil StartConsumers() setelah semua
// dependency siap.
func (a *AppConfig) InitializeConsumers() {
	a.TransactionConsumer = worker.NewTransactionConsumer(a.RMQClient, a.TransactionService, 5)
	a.PricingConsumer = worker.NewMarketPricingConsumer(a.RMQClient, a.MarketRepo, a.PublisherWS, 3)
	a.VolumeConsumer = worker.NewMarketVolumeConsumer(a.RMQClient, a.MarketRepo, 2)
	a.AuditConsumer = worker.NewLedgerAuditConsumer(a.RMQClient, 3)
	a.LedgerPersistenceConsumer = worker.NewLedgerPersistenceConsumer(a.RMQClient, a.LedgerRepo, 5)

	reconcileConfig := worker.RecoilConfig{
		WorkerCount:       5,
		ReconcileWorkers:  3,
		ProcessingTimeout: 30 * time.Second,
		MaxDiscrepancies:  100000,
	}
	a.ReconcileConsumer = worker.NewLedgerReconcileConsumer(a.RMQClient, a.WalletRepo, a.LedgerRepo, a.DiscrepancyRepo, reconcileConfig)

	rewardCalcConfig := worker.RewardEngineConfig{
		ConsumerWorkers:   3,
		CalcWorkers:       5,
		ProcessingTimeout: 30 * time.Second,
		QueueSize:         1000,
	}
	a.RewardCalculationConsumer = worker.NewRewardCalculationConsumer(a.RMQClient, a.RewardPublisher, rewardCalcConfig)

	rewardDistConfig := worker.RewardDistConfig{
		ConsumerWorkers:   3,
		DistWorkers:       5,
		ProcessingTimeout: 30 * time.Second,
		QueueSize:         1000,
		CleanupInterval:   5 * time.Minute,
		TTLDuration:       24 * time.Hour,
	}
	a.RewardDistributionConsumer = worker.NewRewardDistributionConsumer(a.RMQClient, a.WalletRepo, rewardDistConfig)

	// NotificationWebSocketConsumer membaca dari queue 'notification.realtime'
	// (yang di-declare di topology) dan push ke WebSocket Hub. Sebelumnya
	// queue ini declared tapi TIDAK ada consumer, menyebabkan dead letter grows.
	a.NotificationWSConsumer = worker.NewNotificationWebSocketConsumer(a.RMQClient, a.PublisherWS, a.NotificationRepo, 3)

	logger.LogInfo("All message consumers initialized successfully")
}

// StartConsumers menjalankan semua consumer di goroutine terpisah.
// Error start di-log tapi tidak menghentikan consumer lain (best-effort).
func (a *AppConfig) StartConsumers() {
	starters := []struct {
		name string
		fn   func() error
	}{
		{"transaction", func() error { return a.TransactionConsumer.Start(context.Background()) }},
		{"pricing", func() error { return a.PricingConsumer.Start() }},
		{"volume", func() error { return a.VolumeConsumer.Start() }},
		{"ledger-persistence", func() error { return a.LedgerPersistenceConsumer.Start() }},
		{"audit", func() error { return a.AuditConsumer.Start() }},
		{"reconcile", func() error { return a.ReconcileConsumer.Start() }},
		{"reward-calculation", func() error { return a.RewardCalculationConsumer.Start() }},
		{"reward-distribution", func() error { return a.RewardDistributionConsumer.Start() }},
		{"notification-ws", func() error { return a.NotificationWSConsumer.Start() }},
	}

	for _, s := range starters {
		go func(name string, fn func() error) {
			if err := fn(); err != nil {
				logger.LogError("Error starting "+name+" consumer", err)
			}
		}(s.name, s.fn)
	}

	// Register reconnect callback: ketika koneksi RabbitMQ drop lalu
	// reconnect sukses, consumers harus di-restart. Tanpa ini, goroutine
	// consumer yang exit saat msgs channel close tidak akan pernah
	// di-restart → pipeline processing mati permanen sampai app restart.
	a.RMQClient.RegisterOnReconnect(func() {
		logger.LogInfo("Reconnect detected, re-registering all consumers")
		for _, s := range starters {
			go func(name string, fn func() error) {
				if err := fn(); err != nil {
					logger.LogError("Error re-starting "+name+" consumer after reconnect", err)
				}
			}(s.name, s.fn)
		}
	})

	logger.LogInfo("All message consumers started successfully")
}
