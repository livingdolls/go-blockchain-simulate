package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/dto"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
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

	log.Println("[LEDGER_AUDIT_CONSUMER] Starting ledger audit consumer...")

	return l.client.Consume(
		rabbitmq.LedgerEntriesQueue,
		l.workerCount,
		l.handleMessage,
	)
}

func (l *LedgerAuditConsumer) handleMessage(msg amqp091.Delivery) {
	defer func() {
		if err := msg.Ack(false); err != nil {
			log.Printf("[LEDGER_AUDIT_CONSUMER] Failed to ack message: %v", err)
		}
	}()

	var batch dto.LedgerBatchEvent

	if err := json.Unmarshal(msg.Body, &batch); err != nil {
		log.Printf("[LEDGER_AUDIT_CONSUMER] Failed to unmarshal ledger batch: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), l.processingTimeout)
	defer cancel()

	go l.processAuditTrail(ctx, batch)

	log.Printf(
		"[LEDGER_AUDIT_CONSUMER] Processed block #%d with %d entries",
		batch.BlockNumber,
		batch.TotalEntries,
	)
}

func (l *LedgerAuditConsumer) processAuditTrail(ctx context.Context, batch dto.LedgerBatchEvent) {
	panic("unimplement")
}
