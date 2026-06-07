package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/livingdolls/go-blockchain-simulate/app"
	"github.com/livingdolls/go-blockchain-simulate/config"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Memuat konfigurasi terpusat. Prioritas: env > config.local.yaml > config.yaml.
	cfg, err := config.Load("")
	if err != nil {
		panic("Gagal memuat konfigurasi: " + err.Error())
	}

	// Inisialisasi logger berdasarkan environment. Field service/env/version
	// sudah ditambahkan global oleh logger (lihat async_logger.go), sehingga
	// pemanggil tidak perlu menambahkannya lagi per-call.
	env := cfg.App.Environment
	var logCfg logger.Config
	if env == "production" {
		logCfg = logger.ProductionConfig(cfg.App.Name, cfg.App.Version)
	} else {
		logCfg = logger.DevelopmentConfig(cfg.App.Name, cfg.App.Version)
	}
	// Pakai env dari config (bukan hard-coded "production"/"development")
	// agar staging/testing tidak salah label.
	logCfg.Env = env

	// Override log level jika LOG_LEVEL env di-set.
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
		panic("Gagal inisialisasi logger: " + err.Error())
	}
	defer logger.Shutdown(5 * time.Second)

	logger.L.Info("Aplikasi mulai")

	// Bangun seluruh dependensi infrastruktur dari konfigurasi.
	deps, err := app.NewAppDependencies(cfg)
	if err != nil {
		logger.LogError("[INIT] Gagal inisialisasi dependency", err)
		os.Exit(1)
	}

	// Inisialisasi app config yang lain (rabbitmq topology, websocket, repo, service, handler, worker, consumer).
	appConfig := &app.AppConfig{}
	appConfig.SetDeps(deps)

	if err := appConfig.SetupRabbitMQTopology(); err != nil {
		logger.LogError("[INIT] Gagal setup RabbitMQ topology", err)
	}

	appConfig.InitializeWebSocket()
	appConfig.InitializeRepositories()
	appConfig.InitializePublishers()
	appConfig.InitializeServices()
	appConfig.InitializeHandlers()
	appConfig.InitializeWorkers()
	appConfig.InitializeConsumers()
	appConfig.StartConsumers()

	// HTTP server dengan graceful shutdown.
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.RequestIDMiddleware())
	r.Use(logger.RequestLogMiddleware())
	r.Use(app.CORSMiddleware(&cfg.Server))
	appConfig.SetupRoutes(r)

	httpSrv := &http.Server{
		Addr:    cfg.Server.Addr(),
		Handler: r,
	}

	// Jalankan server di goroutine terpisah agar main thread bisa menunggu sinyal.
	go func() {
		logger.LogInfo("HTTP server mulai di " + cfg.Server.Addr())
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.LogError("HTTP server error", err)
		}
	}()

	// Tunggu sinyal shutdown (SIGINT atau SIGTERM).
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logger.LogInfo("Sinyal shutdown diterima: " + sig.String())

	// Graceful shutdown HTTP server dengan timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logger.LogError("gagal shutdown HTTP server", err)
	} else {
		logger.LogInfo("HTTP server berhenti dengan baik")
	}

	// Matikan worker, consumer, dan resource lain.
	appConfig.Shutdown()

	logger.LogInfo("Aplikasi berhenti dengan baik")
}
