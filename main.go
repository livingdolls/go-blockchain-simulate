package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app"
	"github.com/livingdolls/go-blockchain-simulate/app/websocket"
	"github.com/livingdolls/go-blockchain-simulate/app/worker"
)

func main() {
	// Initialize application configuration and dependencies
	appConfig := &app.AppConfig{}

	// Initialize infrastructure (database, redis, rabbitmq, auth)
	if err := appConfig.InitializeInfrastructure(); err != nil {
		log.Fatalf("[INIT] Failed to initialize infrastructure: %v\n", err)
	}

	// Setup RabbitMQ topology (queues, exchanges, bindings)
	if err := appConfig.SetupRabbitMQTopology(); err != nil {
		log.Fatalf("[INIT] Failed to setup RabbitMQ topology: %v\n", err)
	}

	// Initialize WebSocket hub
	appConfig.InitializeWebSocket()

	// Initialize all repositories
	appConfig.InitializeRepositories()

	// Initialize publishers (market pricing, ledger)
	appConfig.InitializePublishers()

	// Initialize services
	appConfig.InitializeServices()

	// Initialize HTTP handlers
	appConfig.InitializeHandlers()

	// Initialize background workers (block generation, candle generation)
	appConfig.InitializeWorkers()

	// Initialize message consumers
	appConfig.InitializeConsumers()

	// Start all message consumers
	appConfig.StartConsumers()

	// Setup HTTP router with all routes
	r := gin.Default()
	r.Use(app.CORSMiddleware())
	appConfig.SetupRoutes(r)

	// Start HTTP server
	go func() {
		log.Println("Server starting on port :3010")
		if err := r.Run(":3010"); err != nil && err.Error() != "http: Server closed" {
			log.Printf("Server error: %v\n", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down...\n", sig)

	// Graceful shutdown with timeout
	appConfig.Shutdown()

	log.Println("Server gracefully stopped")
	os.Exit(0)
}

// stopWorkers stops all workers gracefully with timeout
func stopWorkers(ctx context.Context, workers ...interface{}) {
	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	wg.Add(len(workers))

	go func() {
		for _, w := range workers {
			go func(workerInstance interface{}) {
				defer wg.Done()
				switch v := workerInstance.(type) {
				case *worker.GenerateBlockWorker:
					v.Stop()
					log.Println("[WORKER] block worker stopped")
				case *worker.GenerateCandleWorker:
					v.Stop()
					log.Println("[WORKER] candle worker stopped")
				case *worker.TransactionConsumer:
					v.Stop()
					log.Println("[WORKER] transaction consumer stopped")
				case *worker.MarketPricingConsumer:
					v.Stop()
					log.Println("[WORKER] market pricing consumer stopped")
				case *worker.MarketVolumeConsumer:
					v.Stop()
					log.Println("[WORKER] market volume consumer stopped")
				case *worker.LedgerAuditConsumer:
					v.Stop()
					log.Println("[WORKER] ledger audit consumer stopped")
				case *worker.LedgerReconcileConsumer:
					v.Stop()
					log.Println("[WORKER] ledger reconcile consumer stopped")
				}
			}(w)
		}
		wg.Wait()
		close(stopChan)
	}()

	select {
	case <-stopChan:
		log.Println("[WORKER] All workers stopped")
	case <-ctx.Done():
		log.Println("[WORKER] Timeout while stopping workers")
	}
}

// closeHub closes WebSocket hub connections with timeout
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
		log.Println("[WEBSOCKET] WebSocket hub closed all connections")
	case <-ctx.Done():
		log.Println("[WEBSOCKET] Timeout while closing WebSocket hub connections")
	}
}
