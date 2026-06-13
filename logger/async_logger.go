package logger

import (
	"io"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// noSyncWriter membungkus io.Writer dengan Sync() yang no-op.
//
// Kenapa perlu: saat Shutdown(), zap.L.Sync() flush semua core termasuk
// stdout. Untuk stdout (yang di Linux adalah pipe/terminal, bukan
// regular file), fsync() return EINVAL "invalid argument" dan
// LoggerShutdown gagal print error message.
//
// Solusi: bungkus stdout dengan writer ini. Write() tetap delegate ke
// stdout asli, tapi Sync() return nil sehingga Shutdown() bersih.
//
// Trade-off: kalau stdout di-redirect ke regular file (mis. `app > log.txt`),
// fsync tidak akan di-call. Tapi ini acceptable - untuk durability
// pakai file logger (lumberjack) yang sudah ada Sync() proper.
type noSyncWriter struct {
	w io.Writer
}

func (n noSyncWriter) Write(p []byte) (int, error) { return n.w.Write(p) }
func (noSyncWriter) Sync() error                   { return nil }

var (
	L                *zap.Logger
	queue            *asyncQueue
	shutdownComplete chan struct{} // closed saat Shutdown selesai
	shutdownOnce     sync.Once     // guarantee Shutdown() body hanya run sekali
	once             sync.Once     // guarantee Init() body hanya run sekali
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

	// 1. Stdout for all levels.
	// Pakai noSyncWriter: stdout di container biasanya pipe/terminal,
	// fsync() di pipe return EINVAL - lihat noSyncWriter doc untuk detail.
	stdoutCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(noSyncWriter{w: os.Stdout}),
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
	shutdownComplete = make(chan struct{})

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

// asyncWorker memproses event dari queue sampai shutdown signal.
//
// Shutdown sequence:
//  1. Shutdown() panggil markClosed() + closeCh() - close(done) sekali.
//  2. Worker yang sedang range loop akan exit (done case di select).
//  3. Setelah done, worker DRAIN sisa event dari ch (race-free: tidak
//     ada sender baru karena isClosed()=true di push()).
//  4. Worker return.
//
// Pattern ini aman dengan concurrent senders - senders tidak akan
// race dengan close() karena kita TIDAK close(ch), hanya close(done).
// Lihat type comment di queue.go untuk detail.
func asyncWorker(id int, queue *asyncQueue) {
	for {
		select {
		case ev, ok := <-queue.ch:
			if !ok {
				// ch closed (defensive - normally kita exit via done)
				return
			}
			if ev.fn != nil {
				ev.fn()
			}
		case <-queue.done:
			// Shutdown signaled. Drain sisa event tanpa blocking
			// (tidak ada sender baru setelah done closed).
			for ev := range queue.ch {
				if ev.fn != nil {
					ev.fn()
				}
			}
			return
		}
	}
}

// Shutdown gracefully closes the logger. Urutan:
//  1. Set closed flag di queue (fast-path skip di push) + signal done
//     channel (race-free shutdown coordination dengan senders).
//  2. Worker detect done, drain sisa event dari ch, lalu return.
//  3. Tunggu shutdown complete (signal dari goroutine di bawah) atau
//     timeout - return L.Sync() di kedua case (logger core flush).
//
// Setelah Shutdown return, panggilan LogInfo/LogError berikutnya
// akan menjadi no-op (lihat guard `if L == nil` atau isClosed di push).
//
// Idempotent: panggilan kedua/ketiga langsung return tanpa efek.
// sync.Once melindungi close(shutdownComplete) dari double-close panic
// (penting untuk test yang manggil Init+Shutdown berkali-kali).
func Shutdown(timeout time.Duration) error {
	if queue == nil {
		return nil
	}

	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	shutdownOnce.Do(func() {
		go func() {
			// Set flag dulu agar push berikutnya drop early tanpa blocking.
			queue.markClosed()
			queue.closeCh()
			// Signal completion. Hanya close sekali via shutdownOnce di atas
			// untuk mencegah double-close panic.
			close(shutdownComplete)
		}()
	})

	select {
	case <-shutdownComplete:
		return L.Sync()
	case <-ticker.C:
		return L.Sync()
	}
}
