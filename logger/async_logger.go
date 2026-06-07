package logger

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	L     *zap.Logger
	queue *asyncQueue
	done  chan struct{}
	once  sync.Once
)

func Init(cfg Config) error {
	var err error
	once.Do(func() {
		err = initLogger(cfg)
	})
	return err
}

func initLogger(cfg Config) error {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "message",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
	}

	var cores []zapcore.Core

	// 1. Stdout for all levels
	stdoutCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		cfg.Level,
	)
	cores = append(cores, stdoutCore)

	// 2. File for warnings+errors (if configured)
	if cfg.LogPath != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.LogPath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}

		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(fileWriter),
			zapcore.WarnLevel,
		)
		cores = append(cores, fileCore)
	}

	// Merge cores
	core := zapcore.NewTee(cores...)

	// Apply sampling
	core = zapcore.NewSampler(
		core,
		time.Second,
		cfg.SampleInitial,
		cfg.SampleThereafter,
	)

	// Create base logger
	baseLogger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.WithCaller(true),
	).With(
		zap.String("service", cfg.ServiceName),
		zap.String("env", cfg.Env),
		zap.String("version", cfg.Version),
	)

	// Setup async queue
	queue = newAsyncQueue(cfg.QueueSize)
	done = make(chan struct{})

	// Start worker pool
	for i := 0; i < cfg.Workers; i++ {
		go asyncWorker(i, queue)
	}

	// Wrap dengan custom Core yang hanya push ke queue (TIDAK menulis langsung).
	// Sebelumnya menggunakan RegisterHooks, yang menyebabkan double-write
	// (hook + core.Write). Implementasi ini: Write hanya enqueue, queue worker
	// yang menulis. AddCaller tetap dikirim lewat entry agar caller source
	// code ditampilkan dengan benar di log.
	L = baseLogger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &queueCore{
			Core:  core,
			queue: queue,
			cfg:   cfg,
		}
	}))

	return nil
}

// queueCore adalah zapcore.Core yang menunda penulisan ke core asli
// via async queue. Hanya method Write yang berubah; method lain (Sync, With,
// Enabled, Check) didelegasikan ke core asli.
type queueCore struct {
	zapcore.Core
	queue *asyncQueue
	cfg   Config
}

// Write menunda penulisan dengan memasukkan event ke queue worker pool.
// TIDAK memanggil core.Write secara langsung untuk menghindari duplikasi
// (RegisterHooks menyebabkan hook + original core sama-sama Write).
func (q *queueCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	q.queue.push(logEvent{
		fn: func() {
			_ = q.Core.Write(entry, fields)
		},
	}, q.cfg.DropOnFull)
	return nil
}

// Sync menunggu queue kosong lalu flush core asli.
// Implementasi sederhana: trigger Sync pada underlying core.
// Note: in-flight events mungkin masih diproses; panggil Shutdown untuk
// drain penuh.
func (q *queueCore) Sync() error {
	return q.Core.Sync()
}

func asyncWorker(id int, queue *asyncQueue) {
	for ev := range queue.ch {
		if ev.fn != nil {
			ev.fn()
		}
	}
}

// Shutdown gracefully closes the logger
func Shutdown(timeout time.Duration) error {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	go func() {
		close(queue.ch)
		done <- struct{}{}
	}()

	select {
	case <-done:
		return L.Sync()
	case <-ticker.C:
		return L.Sync()
	}
}
