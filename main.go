package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app"
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
