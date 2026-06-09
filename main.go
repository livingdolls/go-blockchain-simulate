package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/livingdolls/go-blockchain-simulate/app"
	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/app/middleware"
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

	// Wire up dari config (logger.level, logger.path, rotation settings).
	// Override via cfg.Logger di sini; env var LOG_LEVEL tetap di-support
	// untuk emergency override (mis. LOG_LEVEL=debug tanpa restart).
	if cfg.Logger.Path != "" {
		logCfg.LogPath = cfg.Logger.Path
		logCfg.MaxSize = cfg.Logger.MaxSize
		logCfg.MaxBackups = cfg.Logger.MaxBackups
		logCfg.MaxAge = cfg.Logger.MaxAge
		logCfg.Compress = cfg.Logger.Compress
	}
	if cfg.Logger.Level != "" {
		if lvl := parseLogLevel(cfg.Logger.Level); lvl != -1 {
			logCfg.Level = lvl
		}
	}

	// Override log level jika LOG_LEVEL env di-set (priority di atas cfg).
	if logLevelStr := os.Getenv("LOG_LEVEL"); logLevelStr != "" {
		if lvl := parseLogLevel(logLevelStr); lvl != -1 {
			logCfg.Level = lvl
		}
	}

	if err := logger.Init(logCfg); err != nil {
		panic("Gagal inisialisasi logger: " + err.Error())
	}

	// Register custom validator tags (eth_addr) ke validator/v10.
	// Harus dipanggil SEBELUM handler dipakai (sebelum ShouldBindJSON
	// pertama). Idempotent untuk init yang sama.
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		dto.RegisterCustomValidators(v)
	}
	// Defer Shutdown dan log error ke stderr kalau gagal (logger sudah
	// rusak sehingga kita tidak bisa pakai logger.LogError di sini).
	defer func() {
		if err := logger.Shutdown(5 * time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "logger shutdown error: %v\n", err)
		}
	}()

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
		os.Exit(1)
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
	// Body size limit: 1 MiB cukup untuk semua endpoint JSON.
	// Default rendah untuk mencegah DoS via JSON bomb; bisa di-override
	// per-route dengan middleware spesifik untuk upload besar.
	r.Use(middleware.MaxBodySizeMiddleware(1 << 20))
	r.Use(logger.RequestIDMiddleware())
	r.Use(logger.RequestLogMiddleware())
	r.Use(app.CORSMiddleware(&cfg.Server))
	appConfig.SetupRoutes(r)

	httpSrv := &http.Server{
		Addr:    cfg.Server.Addr(),
		Handler: r,
	}

	// Jalankan server di goroutine terpisah agar main thread bisa menunggu sinyal.
	// ListenAndServe() error (mis. port already-in-use) dikirim ke errorChan
	// agar main thread bisa trigger graceful shutdown alih-alih diam selamanya.
	errorChan := make(chan error, 1)
	go func() {
		logger.LogInfo("HTTP server mulai di " + cfg.Server.Addr())
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.LogError("HTTP server error", err)
			errorChan <- err
		}
		close(errorChan)
	}()

	// Tunggu sinyal shutdown (SIGINT atau SIGTERM) atau HTTP server error.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigChan:
		logger.LogInfo("Sinyal shutdown diterima: " + sig.String())
	case err := <-errorChan:
		logger.LogError("HTTP server fatal, trigger shutdown", err)
	}

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

// parseLogLevel mengkonversi string level ("debug", "info", "warn", "error")
// ke zapcore.Level. Return -1 kalau invalid (logger pakai nilai default).
func parseLogLevel(s string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	}
	return -1
}
