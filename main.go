package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize logger based on environment
	var logCfg logger.Config
	env := os.Getenv("ENV")
	if env == "production" {
		logCfg = logger.ProductionConfig("blockchain", "1.0.0")
	} else {
		logCfg = logger.DevelopmentConfig("blockchain", "1.0.0")
	}

	// Override log level if specified
	if logLevelStr := os.Getenv("LOG_LEVEL"); logLevelStr != "" {
		switch logLevelStr {
		case "debug":
			logCfg.Level = zapcore.DebugLevel
		case "info":
			logCfg.Level = zapcore.InfoLevel
		case "warn":
			logCfg.Level = zapcore.WarnLevel
		case "error":
			logCfg.Level = zapcore.ErrorLevel
		}
	}

	if err := logger.Init(logCfg); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Shutdown(5 * time.Second)

	logger.L.Info("Application starting",
		zap.String("service", "blockchain"),
		zap.String("env", env),
		zap.String("version", "1.0.0"),
	)

	// Initialize application configuration and dependencies
	appConfig := &app.AppConfig{}

	// Initialize infrastructure (database, redis, rabbitmq, auth)
	if err := appConfig.InitializeInfrastructure(); err != nil {
		logger.LogError("[INIT] Failed to initialize infrastructure: %v\n", err)
	}

	// Setup RabbitMQ topology (queues, exchanges, bindings)
	if err := appConfig.SetupRabbitMQTopology(); err != nil {
		logger.LogError("[INIT] Failed to setup RabbitMQ topology: %v\n", err)
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
		logger.LogInfo("Server starting on port :3010")
		if err := r.Run(":3010"); err != nil && err.Error() != "http: Server closed" {
			logger.LogError("Server error", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.LogInfo("Received signal: " + sig.String() + ". Shutting down...")

	// Graceful shutdown with timeout
	appConfig.Shutdown()

	logger.LogInfo("Server gracefully stopped")
	os.Exit(0)
}
