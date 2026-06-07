package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// LogError logs error dengan stack trace
// Skips 1 level to show actual caller, not this helper
func LogError(msg string, err error, fields ...zap.Field) {
	if L == nil {
		return
	}
	L.WithOptions(zap.AddCallerSkip(1)).Error(msg,
		append(fields, zap.Error(err))...,
	)
}

// LogWarn logs warning
func LogWarn(msg string, fields ...zap.Field) {
	if L == nil {
		return
	}
	L.WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

// LogInfo logs info message
func LogInfo(msg string, fields ...zap.Field) {
	if L == nil {
		return
	}
	L.WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

// LogDebug logs debug message
func LogDebug(msg string, fields ...zap.Field) {
	if L == nil {
		return
	}
	L.WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

// LogErrorCtx logs error dengan field dari context (request_id, user_id, dll).
// Gunakan ini dari service/repository yang sudah menerima context.Context.
func LogErrorCtx(ctx context.Context, msg string, err error, fields ...zap.Field) {
	if L == nil {
		return
	}
	all := make([]zap.Field, 0, len(fields)+1)
	all = append(all, fields...)
	all = append(all, zap.Error(err))
	FromContext(ctx).WithOptions(zap.AddCallerSkip(1)).Error(msg, all...)
}

// LogWarnCtx logs warning dengan field dari context.
func LogWarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	if L == nil {
		return
	}
	FromContext(ctx).WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

// LogInfoCtx logs info dengan field dari context (request_id otomatis muncul).
func LogInfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	if L == nil {
		return
	}
	FromContext(ctx).WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

// LogDebugCtx logs debug dengan field dari context.
func LogDebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	if L == nil {
		return
	}
	FromContext(ctx).WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

// LogWithDuration logs message with execution duration
func LogWithDuration(msg string, start time.Time, fields ...zap.Field) {
	duration := time.Since(start)
	L.WithOptions(zap.AddCallerSkip(1)).Info(msg,
		append(fields, zap.Duration("duration_ms", duration))...,
	)
}

// LogBlockEvent logs blockchain event
func LogBlockEvent(blockNumber int64, action string, fields ...zap.Field) {
	L.WithOptions(zap.AddCallerSkip(1)).Info("block_event",
		append(fields,
			zap.Int64("block_number", blockNumber),
			zap.String("action", action),
		)...,
	)
}

// LogTransactionEvent logs transaction event
func LogTransactionEvent(txID int64, status string, fields ...zap.Field) {
	L.WithOptions(zap.AddCallerSkip(1)).Info("transaction_event",
		append(fields,
			zap.Int64("tx_id", txID),
			zap.String("status", status),
		)...,
	)
}

// LogWorkerEvent logs worker event
func LogWorkerEvent(workerID string, action string, fields ...zap.Field) {
	L.WithOptions(zap.AddCallerSkip(1)).Info("worker_event",
		append(fields,
			zap.String("worker_id", workerID),
			zap.String("action", action),
		)...,
	)
}

// GetQueueStats returns current queue statistics
func GetQueueStats() map[string]interface{} {
	return map[string]interface{}{
		"dropped": queue.dropped,
		"total":   queue.total,
	}
}
