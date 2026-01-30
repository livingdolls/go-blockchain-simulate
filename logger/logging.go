package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey string

const (
	RequestIDKey ctxKey = "request_id"
	TraceIDKey   ctxKey = "trace_id"
	WorkerIDKey  ctxKey = "worker_id"
	JobIDKey     ctxKey = "job_id"
	UserIDKey    ctxKey = "user_id"
	SpanKey      ctxKey = "span"
	BlockIDKey   ctxKey = "block_id"
	TxIDKey      ctxKey = "tx_id"
)

// FromContext extracts logger fields from context
func FromContext(ctx context.Context) *zap.Logger {
	fields := []zap.Field{}

	contextFields := map[ctxKey]string{
		RequestIDKey: "request_id",
		TraceIDKey:   "trace_id",
		WorkerIDKey:  "worker_id",
		JobIDKey:     "job_id",
		UserIDKey:    "user_id",
		SpanKey:      "span",
		BlockIDKey:   "block_id",
		TxIDKey:      "tx_id",
	}

	for key, fieldName := range contextFields {
		if val, ok := ctx.Value(key).(string); ok && val != "" {
			fields = append(fields, zap.String(fieldName, val))
		}
	}

	if len(fields) == 0 {
		return L
	}

	return L.With(fields...)
}

// ContextWithRequestID adds request ID to context
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// ContextWithTraceID adds trace ID to context
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// ContextWithWorkerID adds worker ID to context
func ContextWithWorkerID(ctx context.Context, workerID string) context.Context {
	return context.WithValue(ctx, WorkerIDKey, workerID)
}

// ContextWithJobID adds job ID to context
func ContextWithJobID(ctx context.Context, jobID string) context.Context {
	return context.WithValue(ctx, JobIDKey, jobID)
}

// ContextWithUserID adds user ID to context
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// ContextWithSpan adds span name to context
func ContextWithSpan(ctx context.Context, spanName string) context.Context {
	return context.WithValue(ctx, SpanKey, spanName)
}

// ContextWithBlockID adds block ID to context
func ContextWithBlockID(ctx context.Context, blockID string) context.Context {
	return context.WithValue(ctx, BlockIDKey, blockID)
}

// ContextWithTxID adds transaction ID to context
func ContextWithTxID(ctx context.Context, txID string) context.Context {
	return context.WithValue(ctx, TxIDKey, txID)
}
