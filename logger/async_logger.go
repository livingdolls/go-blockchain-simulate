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

	// Wrap with async queue
	L = baseLogger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.RegisterHooks(core, func(e zapcore.Entry) error {
			queue.push(logEvent{
				fn: func() {
					_ = core.Write(e, nil)
				},
			}, cfg.DropOnFull)
			return nil
		})
	}))

	return nil
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
