package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type LedgerAuditConsumer struct {
	client            *rabbitmq.Client
	mu                sync.Mutex
	isRunning         bool
	stopChan          chan struct{}
	workerCount       int
	processingTimeout time.Duration
	auditTrail        []dto.AuditTrailEntry
	auditTrailMu      sync.RWMutex
}

func NewLedgerAuditConsumer(rmqClient *rabbitmq.Client, workerCount int) *LedgerAuditConsumer {
	return &LedgerAuditConsumer{
		client:            rmqClient,
		stopChan:          make(chan struct{}),
		workerCount:       workerCount,
		processingTimeout: 30 * time.Second,
		auditTrail:        make([]dto.AuditTrailEntry, 0),
	}
}

func (l *LedgerAuditConsumer) Start() error {
	l.mu.Lock()
	if l.isRunning {
		l.mu.Unlock()
		return nil
	}

	l.isRunning = true
	l.mu.Unlock()

	logger.LogInfo("Starting ledger audit consumer")

	return l.client.Consume(
		rabbitmq.LedgerAuditQueue,
		l.workerCount,
		l.handleMessage,
	)
}

func (l *LedgerAuditConsumer) handleMessage(msg amqp091.Delivery) {
	defer func() {
		if err := msg.Ack(false); err != nil {
			logger.LogError("Failed to ack message", err)
		}
	}()

	var batch dto.LedgerBatchEvent

	if err := json.Unmarshal(msg.Body, &batch); err != nil {
		logger.LogError("Failed to unmarshal ledger batch", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), l.processingTimeout)
	defer cancel()

	go l.processAuditTrail(ctx, batch)

	logger.LogInfo("Processed block",
		zap.Int("block_number", batch.BlockNumber),
		zap.Int("entries", batch.TotalEntries),
	)
}

func (l *LedgerAuditConsumer) processAuditTrail(ctx context.Context, batch dto.LedgerBatchEvent) {
	l.auditTrailMu.Lock()
	defer l.auditTrailMu.Unlock()

	for _, entry := range batch.Entries {
		action := l.determineAction(entry)

		auditEntry := dto.AuditTrailEntry{
			EntryID:     batch.BlockID,
			Action:      action,
			FromAddress: entry.Address,
			ToAddress:   entry.Address,
			Amount:      entry.Amount,
			BlockNumber: batch.BlockNumber,
			Timestamp:   entry.Timestamp,
			Reconciled:  false,
		}

		l.auditTrail = append(l.auditTrail, auditEntry)

		if entry.Amount < -1000 || entry.Amount > 1000 {
			logger.LogWarn("Large transaction detected",
				zap.Int("block_number", batch.BlockNumber),
				zap.String("address", entry.Address),
				zap.Float64("amount", entry.Amount),
			)
		}
	}

	// periodic audit log
	if batch.BlockNumber%100 == 0 {
		logger.LogInfo("Audit checkpoint",
			zap.Int("block_number", batch.BlockNumber),
			zap.Int("total_audit_entries", len(l.auditTrail)),
		)
	}
}

func (l *LedgerAuditConsumer) determineAction(entry dto.LedgerEntryEvent) string {
	if entry.TxID == nil {
		return "REWARD"
	}
	if entry.Amount > 0 {
		return "CREDIT"
	}

	return "DEBIT"
}

func (l *LedgerAuditConsumer) GetAuditTrail(limit int) []dto.AuditTrailEntry {
	l.auditTrailMu.RLock()
	defer l.auditTrailMu.RUnlock()

	if limit > len(l.auditTrail) {
		limit = len(l.auditTrail)
	}

	return l.auditTrail[len(l.auditTrail)-limit:]
}

func (l *LedgerAuditConsumer) Stop() {
	l.mu.Lock()

	if !l.isRunning {
		l.mu.Unlock()
		return
	}

	l.isRunning = false
	l.mu.Unlock()

	logger.LogInfo("Stopping ledger audit consumer")
	close(l.stopChan)
	logger.LogInfo("Ledger audit consumer stopped")
}
